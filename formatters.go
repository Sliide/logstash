package logstash

import (
	"encoding/json"
	"fmt"

	"github.com/sirupsen/logrus"
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

	data["message"] = entry.Message
	data["level"] = entry.Level.String()
	data["env"] = f.Env
	data["service"] = f.Service

	serialized, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("Failed to marshal fields to JSON, %v", err)
	}
	return append(serialized, '\n'), nil
}
