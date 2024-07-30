package client

import (
	"github.com/er1cw00/comfy.go/graph"
)

type QueueItem struct {
	PromptID   string                 `json:"prompt_id"`
	Number     int                    `json:"number"`
	NodeErrors map[string]interface{} `json:"node_errors"`
	Messages   chan PromptMessage     `json:"-"`
	Workflow   *graph.Graph           `json:"-"`
}
