package logstash

import (
	"os"

	log "github.com/sirupsen/logrus"
	"io"
)

var LogLevels = map[string]log.Level{
	"DEBUG":   log.DebugLevel,
	"INFO":    log.InfoLevel,
	"WARNING": log.WarnLevel,
	"ERROR":   log.ErrorLevel,
	"FATAL":   log.FatalLevel,
	"PANIC":   log.PanicLevel,
}

func Init(logLevel string, logFileName string, env string, service string) bool {

	logFile, err := os.Create(logFileName)

	if err != nil {
		panic(err)
	}

	log.SetFormatter(&LogstashJsonFormatter{
		Env:     env,
		Service: service,
	})

	log.SetLevel(LogLevels[logLevel])
	mw := io.MultiWriter(os.Stdout, logFile)

	log.SetOutput(mw)
	return true
}
