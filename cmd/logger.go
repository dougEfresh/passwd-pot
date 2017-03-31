package cmd

import (
	"github.com/Sirupsen/logrus/hooks/syslog"
)

var syslogHook *logrus_syslog.SyslogHook
