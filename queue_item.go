package comfy

type QueueItem struct {
	PromptID   string                 `json:"prompt_id"`
	Number     int                    `json:"number"`
	NodeErrors map[string]interface{} `json:"node_errors"`
	Workflow   *Graph                 `json:"-"`
}
