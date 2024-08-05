package logger

import (
	"github.com/op/go-logging"
	"io"
	"os"
)

var logger *logging.Logger = nil

var format = logging.MustStringFormatter(
	`%{color}%{time:15:04:05.000} %{level:.4s} %{shortfile} %{message} %{color:reset}`,
)

func SetLogger(l *logging.Logger) {
	logger = l
}

func createBackend(w io.Writer, level logging.Level) logging.Backend {
	backend := logging.NewLogBackend(w, "", 0)
	backendLeveled := logging.AddModuleLevel(logging.NewBackendFormatter(backend, format))
	backendLeveled.SetLevel(level, "")
	return backendLeveled
}

func New(logLevel string, module string) error {
	level, err := logging.LogLevel(logLevel)
	if err != nil {
		level = logging.WARNING
	}
	if logger == nil {
		logger = logging.MustGetLogger(module)
	}
	logger.ExtraCalldepth = 1
	consoleBackend := createBackend(os.Stdout, level)
	logging.SetBackend(consoleBackend)
	return nil

}
func check() {
	if logger == nil {
		New("Debug", "default")
	}
}
func Debug(i ...interface{}) {
	if logger == nil {
		return
	}
	logger.Debug(i...)
}

func Debugf(format string, args ...interface{}) {
	if logger == nil {
		return
	}
	logger.Debugf(format, args...)
}

func Info(i ...interface{}) {
	if logger == nil {
		return
	}
	logger.Info(i...)
}

func Infof(format string, args ...interface{}) {
	if logger == nil {
		return
	}
	logger.Infof(format, args...)
}

func Warn(i ...interface{}) {
	if logger == nil {
		return
	}
	logger.Warning(i...)
}

func Warnf(format string, args ...interface{}) {
	if logger == nil {
		return
	}
	logger.Warningf(format, args...)
}

func Error(i ...interface{}) {
	if logger == nil {
		return
	}
	logger.Error(i...)
}

func Errorf(format string, args ...interface{}) {
	if logger == nil {
		return
	}
	logger.Errorf(format, args...)
}

func Fatal(i ...interface{}) {
	if logger == nil {
		return
	}
	logger.Fatal(i...)
}

func Fatalf(format string, args ...interface{}) {
	if logger == nil {
		return
	}
	logger.Fatalf(format, args...)
}

func Panic(i ...interface{}) {
	if logger == nil {
		return
	}
	logger.Panic(i...)
}

func Panicf(format string, args ...interface{}) {
	if logger == nil {
		return
	}
	logger.Panicf(format, args...)
}

func Assert(exp bool, i ...interface{}) {
	if !exp && logger != nil {
		logger.Panic(i...)
	}
}

func Assertf(exp bool, format string, args ...interface{}) {
	if !exp && logger != nil {
		logger.Panicf(format, args...)
	}
}
