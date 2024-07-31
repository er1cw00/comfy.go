package comfy

import (
	"github.com/er1cw00/comfy.go/base/logger"
	"github.com/op/go-logging"
)

func SetLogger(l *logging.Logger) {
	logger.SetLogger(l)
}
