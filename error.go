package comfy

import (
	"errors"
)

var ErrComfyDisconnected = errors.New("comfy disconnected")
var ErrNotNodeObjects = errors.New("not node objects")
var ErrNotWorkflowInPNG = errors.New("png does not contain workflow metadata")
