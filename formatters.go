package logstash

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"os"
	"runtime"
	"strings"
)

type LogstashJsonFormatter struct {
	Env     string
	Service string
}

func (f *LogstashJsonFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	data := make(logrus.Fields, len(entry.Data)+3)
	for k, v := range entry.Data {
		switch v := v.(type) {
		case error:
			// Otherwise errors are ignored by `encoding/json`
			// https://github.com/Sirupsen/logrus/issues/137
			data[k] = v.Error()
		default:
			data[k] = v
		}
	}

	var pc uintptr
	var fileName string
	var lineNumber int
	for i := 2; i < 9; i++ {
		pc, fileName, lineNumber, _ = runtime.Caller(i)
		// If we need to debug the callstack, add this line
		//data[fmt.Sprintf("caller_%d", i)] = fmt.Sprintf("%s:%d", fileName, lineNumber)
		if !strings.Contains(fileName, "sirupsen") {
			break
		}
	}
	functionName := runtime.FuncForPC(pc).Name()

	// logstash will trim timestamp to milliseconds,
	// but this means we won't need to edit the code in the future, when nanosecond support comes out.
	data["@timestamp"] = time.Now().UnixNano()
	data["hostname"] = os.Getenv("HOSTNAME")
	data["message"] = entry.Message
	data["logger"] = strings.TrimPrefix(functionName, "github.com/sliide/")
	data["lineno"] = fmt.Sprintf("%s:%d", fileName, lineNumber)
	data["level"] = entry.Level.String()
	data["env"] = f.Env
	data["service"] = f.Service

	serialized, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("Failed to marshal fields to JSON, %v", err)
	}
	return append(serialized, '\n'), nil
}
