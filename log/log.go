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
	"io"

	"github.com/dougEfresh/zapz"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// FieldLogger for passwd-pot
type FieldLogger interface {
	Debugf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
	Fatalf(format string, args ...interface{})
	Panicf(format string, args ...interface{})

	Debug(args string)
	Info(args string)
	Warn(args string)
	Error(args interface{})
	Fatal(args string)
	Panic(args string)

	With(f zapcore.Field) FieldLogger

	AddLogger(l *zap.Logger)
	SetLevel(level Level)
	GetLevel() Level
	Sync() error
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

// Logger structure of Logger
type Logger struct {
	loggers []*zap.Logger
	level   Level
}

// DefaultLogger to STDOUT
func DefaultLogger(w io.Writer) FieldLogger {
	l := &Logger{}
	l.SetLevel(InfoLevel)
	en := zapcore.NewJSONEncoder(zapz.DefaultConfig)
	c := zapcore.NewCore(en, zapcore.AddSync(w), zap.DebugLevel)

	l.AddLogger(zap.New(c))
	return l
}

func (logger *Logger) Log(keyvals string) error {
	return nil
}

func (logger *Logger) SetLevel(l Level) {
	logger.level = l
}

func (logger *Logger) GetLevel() Level {
	return logger.level
}

func (logger *Logger) IsDebug() bool {
	return logger.level == DebugLevel
}

func (logger *Logger) AddLogger(l *zap.Logger) {
	logger.loggers = append(logger.loggers, l)
}

func (logger *Logger) Debugf(format string, args ...interface{}) {
	if logger.level >= DebugLevel {
		for _, l := range logger.loggers {
			l.Debug(fmt.Sprintf(format, args...))
		}
	}
}

func (logger *Logger) Infof(format string, args ...interface{}) {
	if logger.level >= InfoLevel {
		for _, l := range logger.loggers {
			l.Info(fmt.Sprintf(format, args...))
		}
	}
}

func (logger *Logger) Warnf(format string, args ...interface{}) {
	if logger.level >= WarnLevel {
		for _, l := range logger.loggers {
			l.Warn(fmt.Sprintf(format, args...))
		}
	}
}

func (logger *Logger) Errorf(format string, args ...interface{}) {
	for _, l := range logger.loggers {
		l.Error(fmt.Sprintf(format, args...))
	}
}

func (logger *Logger) Debug(msg string) {
	if logger.level >= DebugLevel {
		for _, l := range logger.loggers {
			l.Debug(msg)
		}
	}
}

func (logger *Logger) Info(msg string) {
	if logger.level >= InfoLevel {
		for _, l := range logger.loggers {
			l.Info(msg)
		}
	}
}
func (logger *Logger) Warn(msg string) {
	if logger.level >= WarnLevel {
		for _, l := range logger.loggers {
			l.Warn(msg)
		}
	}
}

func (logger *Logger) Error(msg interface{}) {
	for _, l := range logger.loggers {
		l.Error(fmt.Sprintf("%s", msg))
	}
}

func (logger *Logger) Fatal(msg string) {
	logger.Error(msg)
}

func (logger *Logger) Fatalf(format string, msg ...interface{}) {
	logger.Errorf(format, msg)
}

func (logger *Logger) Panic(msg string) {
	logger.Error(msg)
}

func (logger *Logger) Panicf(format string, msg ...interface{}) {
	logger.Errorf(format, msg...)
}

func (logger *Logger) Sync() error {
	for _, l := range logger.loggers {
		l.Sync()
	}
	return nil
}

func (logger *Logger) With(f zapcore.Field) FieldLogger {
	return logger
}
