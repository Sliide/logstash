package logstash

import (
	"os"

	log "github.com/sirupsen/logrus"
)

var LogLevels = map[string]log.Level{
	"DEBUG":   log.DebugLevel,
	"INFO":    log.InfoLevel,
	"WARNING": log.WarnLevel,
	"ERROR":   log.ErrorLevel,
	"FATAL":   log.FatalLevel,
	"PANIC":   log.PanicLevel,
}

func Init(logLevel string, logFile string, env string, service string) bool {

	f, err := os.Create(logFile)
	if err != nil {
		panic(err)
	}

	log.SetFormatter(&LogstashJsonFormatter{
		Env:     env,
		Service: service,
	})
	log.SetOutput(f)
	log.SetLevel(LogLevels[logLevel])
	return true
}
