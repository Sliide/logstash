package logstash

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"testing"
	"time"

	. "gopkg.in/check.v1"

	log "github.com/sirupsen/logrus"
)

func TestLogstash(t *testing.T) { TestingT(t) }

type LogstashTestSuite struct{}

type LogMessage struct {
	Message string `json:"message"`
	Level   string `json:"level"`
	MyField string `json:"my_field"`
	LineNo  string `json:"lineno"`
	Logger  string `json:"logger"`
}

var _ = Suite(
	&LogstashTestSuite{},
)

func (s *LogstashTestSuite) TestFileLogger(c *C) {
	var filename = "./logs"
	defer os.Remove(filename)

	Init(
		"INFO",
		filename,
		"local",
		"logstash",
		0,
	)

	log.WithFields(log.Fields{"my_field": fmt.Sprintf("%d", 1)}).Error("This is an error message")

	jsonString, err := ioutil.ReadFile(filename)

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
	c.Assert(logMessage.LineNo, Matches, ".*logstash_test.go.*")
	c.Assert(logMessage.Logger, Matches, ".*TestFileLogger")
}

func (s *LogstashTestSuite) TestOutputLogger(c *C) {

	b := bytes.NewBuffer(nil)

	InitWithOutput(
		"INFO",
		"local",
		"logstash",
		b,
	)

	log.WithFields(log.Fields{"my_field": fmt.Sprintf("%d", 1)}).Error("This is an error message")

	var logMessage LogMessage

	if err := json.Unmarshal(b.Bytes(), &logMessage); err != nil {
		c.Fail()
	}

	c.Assert(logMessage.MyField, Equals, "1")
	c.Assert(logMessage.Level, Equals, "error")
	c.Assert(logMessage.Message, Equals, "This is an error message")
	c.Assert(logMessage.LineNo, Matches, ".*logstash_test.go.*")
	c.Assert(logMessage.Logger, Matches, ".*TestOutputLogger")
}

func (s *LogstashTestSuite) TestFileRotation(c *C) {
	logsDir := "./test_logs"
	os.Mkdir(logsDir, os.ModePerm)
	logfile := logsDir + "/log"
	defer os.RemoveAll(logsDir)

	rotationCheckInterval = 100 * time.Millisecond

	Init(
		"DEBUG",
		logfile,
		"local",
		"logstash",
		1000,
	)

	stop := make(chan struct{})

	go func() {
		for {
			select {
			case <-stop:
				return
			default:
				log.Info("the cake is a lie")
				time.Sleep(10 * time.Millisecond)
			}
		}
	}()

	time.Sleep(210 * time.Millisecond)
	close(stop)

	dir, err := os.Open(logsDir)
	c.Assert(err, IsNil)

	files, err := dir.Readdir(100)
	c.Assert(err, IsNil)
	c.Assert(len(files), Equals, 3)

	for _, f := range files {
		switch f.Name() {
		case "log":
			c.Assert(f.Size() < 1000, Equals, true)
		case "log.1":
			c.Assert(f.Size() > 1000, Equals, true)
		default:
			c.Assert(strings.HasSuffix(f.Name(), ".gz"), Equals, true)
		}
	}
}
