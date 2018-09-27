package logstash

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

var (
	LogLevels = map[string]log.Level{
		"DEBUG":   log.DebugLevel,
		"INFO":    log.InfoLevel,
		"WARNING": log.WarnLevel,
		"ERROR":   log.ErrorLevel,
		"FATAL":   log.FatalLevel,
		"PANIC":   log.PanicLevel,
	}

	rotationCheckInterval       = 10 * time.Second
	maxLogFileSize        int64 = 2 << 30 // 2GiB

	stopRotation chan struct{}
)

// Init sets up logging
// This function should only be called once when the service is started
func Init(logLevel string, logFileNameBase string, env string, service string, maxSize int64) bool {

	if maxSize != 0 {
		maxLogFileSize = maxSize
	}
	logFileName := addTimestampToFilename(logFileNameBase)
	logFile, err := os.Create(logFileName)

	if err != nil {
		panic(err)
	}

	log.SetFormatter(&LogstashJsonFormatter{
		Env:     env,
		Service: service,
	})

	log.SetLevel(LogLevels[logLevel])
	setLogFile(logFile)
	go rotate(logFileName, logFileNameBase)
	return true
}

func setLogFile(writer io.Writer) {
	log.SetOutput(io.MultiWriter(os.Stdout, writer))
}

// rotate checks periodically if the
// when the current logfile
func rotate(logFileName, logFileBaseName string) {
	defer func() {
		rotatedFilename := addTimestampToFilename(logFileName)
		os.Rename(logFileName, rotatedFilename)
	}()
	log.Debug("starting rotation for", logFileName)

	if stopRotation != nil {
		log.Debug("stopping rotation")
		close(stopRotation)
	}
	stopRotation = make(chan struct{})

	ticker := time.NewTicker(rotationCheckInterval)
	for {
		select {
		case <-stopRotation:
			log.Debug("stopping rotation for", logFileName)
			ticker.Stop()
			return
		case <-ticker.C:
			f, err := os.Stat(logFileName)
			if err != nil {
				log.Errorf("couldn't get file size for %s : %s", logFileName, err)
				continue
			}
			if f.Size() > maxLogFileSize {
				log.Debugf("log file too large, rotating the log file")

				// create new log file with current timestamp
				logFileName = addTimestampToFilename(logFileBaseName)
				NewLogFile, err := os.Create(logFileName)
				if err != nil {
					log.Error("cannot create new log file: %s", err)
					continue
				}
				// add end timestamp
				rotatedFilename := addTimestampToFilename(logFileName)
				err = os.Rename(logFileName, rotatedFilename)
				if err != nil {
					log.Error("couldn't rename log file", logFileName, err)
					continue
				}

				setLogFile(NewLogFile)
			}
		}

	}
}

func addTimestampToFilename(baseName string) string {
	return fmt.Sprintf(
		"%s-%s.log",
		strings.TrimSuffix(baseName, ".log"),
		time.Now().Format(time.RFC3339),
	)
}
