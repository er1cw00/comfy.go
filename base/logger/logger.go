package logger

import (
	"github.com/op/go-logging"
)

var logger *logging.Logger = nil

func SetLogger(l *logging.Logger) {
	logger = l
}

func Debug(i ...interface{}) {
	logger.Debug(i...)
}

func Debugf(format string, args ...interface{}) {
	logger.Debugf(format, args...)
}

func Info(i ...interface{}) {
	logger.Info(i...)
}

func Infof(format string, args ...interface{}) {
	logger.Infof(format, args...)
}

func Warn(i ...interface{}) {
	logger.Warning(i...)
}

func Warnf(format string, args ...interface{}) {
	logger.Warningf(format, args...)
}

func Error(i ...interface{}) {
	logger.Error(i...)
}

func Errorf(format string, args ...interface{}) {
	logger.Errorf(format, args...)
}

func Fatal(i ...interface{}) {
	logger.Fatal(i...)
}

func Fatalf(format string, args ...interface{}) {
	logger.Fatalf(format, args...)
}

func Panic(i ...interface{}) {
	logger.Panic(i...)
}

func Panicf(format string, args ...interface{}) {
	logger.Panicf(format, args...)
}

func Assert(exp bool, i ...interface{}) {
	if !exp {
		logger.Panic(i...)
	}
}

func Assertf(exp bool, format string, args ...interface{}) {
	if !exp {
		logger.Panicf(format, args...)
	}
}
