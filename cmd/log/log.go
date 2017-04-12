// Copyright © 2017 Douglas Chimento <dchimento@gmail.com>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package log

import (
	"fmt"
	klog "github.com/go-kit/kit/log"
)

type FieldLogger interface {
	Debugf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
	Fatalf(format string, args ...interface{})
	Panicf(format string, args ...interface{})

	Debug(args ...interface{})
	Info(args ...interface{})
	Warn(args ...interface{})
	Error(args ...interface{})
	Fatal(args ...interface{})
	Panic(args ...interface{})
}

// Level type
type Level uint8

const (
	// ErrorLevel level. Logs. Used for errors that should definitely be noted.
	// Commonly used for hooks to send errors to an error tracking service.
	ErrorLevel Level = iota
	// WarnLevel level. Non-critical entries that deserve eyes.
	WarnLevel
	// InfoLevel level. General operational entries about what's going on inside the
	// application.
	InfoLevel
	// DebugLevel level. Usually only enabled when debugging. Very verbose logging.
	DebugLevel
)

type Logger struct {
	loggers []klog.Logger
	level   Level
}

func (logger Logger) Log(keyvals ...interface{}) error {
	logger.Info(fmt.Sprint(keyvals))
	return nil
}

func (logger *Logger) SetLevel(l Level) {
	logger.level = l
}

func (logger *Logger) GetLevel() Level {
	return logger.level
}

func (logger *Logger) With(key string, value interface{}) {
	for i, l := range logger.loggers {
		logger.loggers[i] = klog.With(l, key, value)
	}
}

func (logger *Logger) AddLogger(l klog.Logger) {
	if logger.loggers == nil {
		logger.loggers = make([]klog.Logger, 1)
		logger.loggers[0] = l
	} else {
		logger.loggers = append(logger.loggers, l)
	}

}

func (logger *Logger) Debugf(format string, args ...interface{}) {
	if logger.level >= DebugLevel {
		for _, l := range logger.loggers {
			l.Log("message", fmt.Sprintf(format, args...), "level", DebugLevel)
		}
	}
}

func (logger *Logger) Infof(format string, args ...interface{}) {
	if logger.level >= InfoLevel {
		for _, l := range logger.loggers {
			l.Log("message", fmt.Sprintf(format, args...), "level", InfoLevel)
		}
	}
}

func (logger *Logger) Warnf(format string, args ...interface{}) {
	if logger.level >= WarnLevel {
		for _, l := range logger.loggers {
			l.Log("message", fmt.Sprintf(format, args...), "level", WarnLevel)
		}
	}
}

func (logger *Logger) Errorf(format string, args ...interface{}) {
	for _, l := range logger.loggers {
		l.Log("message", fmt.Sprintf(format, args...), "level", ErrorLevel)
	}
}

func (logger *Logger) Debug(msg interface{}) {
	if logger.level >= DebugLevel {
		for _, l := range logger.loggers {
			l.Log("message", msg, "level", DebugLevel)
		}
	}
}

func (logger *Logger) Info(msg interface{}) {
	if logger.level >= InfoLevel {
		for _, l := range logger.loggers {
			l.Log("message", msg, "level", InfoLevel)
		}
	}
}
func (logger *Logger) Warn(msg interface{}) {
	if logger.level >= WarnLevel {
		for _, l := range logger.loggers {
			l.Log("message", msg, "level", WarnLevel)
		}
	}
}

func (logger *Logger) Error(msg interface{}) {
	for _, l := range logger.loggers {
		l.Log("message", msg, "level", ErrorLevel)
	}
}
