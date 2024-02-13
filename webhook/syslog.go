package webhook

import (
	"log/syslog"
	"os"

	"github.com/ferretcode/locomotive/graphql"
	"github.com/ferretcode/locomotive/railway"
)

type Syslog struct {
	Syslog syslog.Writer
}

func NewSyslog() (Syslog, error) {
	address := os.Getenv("SYSLOG_ADDRESS")

	sysLog, err := syslog.Dial("tcp", address, syslog.LOG_INFO, "locomotive")

	if err != nil {
		return Syslog{}, err
	}

	return Syslog{
		Syslog: *sysLog,
	}, nil
}

func (s *Syslog) SendSyslogLog(log graphql.Log) error {
	var err error

	switch log.Severity {
	case railway.SEVERITY_INFO:
		s.Syslog.Info(log.Message)
	case railway.SEVERITY_WARN:
		s.Syslog.Warning(log.Message)
	case railway.SEVERITY_ERROR:
		s.Syslog.Err(log.Message)
	}

	if err != nil {
		return err
	}

	return nil
}
