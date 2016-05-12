package logstash_test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"testing"

	. "gopkg.in/check.v1"

	log "github.com/sirupsen/logrus"
	"github.com/sliide/logstash"
)

func TestLogstash(t *testing.T) { TestingT(t) }

type LogstashTestSuite struct{}

var _ = Suite(
	&LogstashTestSuite{},
)

func (s *LogstashTestSuite) TestFileLogger(c *C) {
	var filename = "./logs"

	logstash.Init(
		"INFO",
		filename,
		"local",
		"logstash",
	)

	log.WithFields(log.Fields{"my_field": fmt.Sprintf("%d", 1)}).Error("This is an error message")

	type LogMessage struct {
		Message string `json:"message"`
		Level   string `json:"level"`
		MyField string `json:"my_field"`
	}

	jsonString, err := ioutil.ReadFile(filename)

	fmt.Println(string(jsonString))
	if err != nil {
		c.Fail()
	}

	var logMessage LogMessage

	if err := json.Unmarshal(jsonString, &logMessage); err != nil {
		c.Fail()
	}

	c.Assert(logMessage.MyField, Equals, "1")
	c.Assert(logMessage.Level, Equals, "error")
	c.Assert(logMessage.Message, Equals, "This is an error message")
}
