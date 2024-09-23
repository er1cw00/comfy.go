package comfy

import (
	"encoding/json"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/er1cw00/comfy.go/base"
	"github.com/er1cw00/comfy.go/base/logger"
	"github.com/google/uuid"
)

type QueuedItemStoppedReason string

const (
	QueuedItemStoppedReasonFinished    QueuedItemStoppedReason = "finished"
	QueuedItemStoppedReasonInterrupted QueuedItemStoppedReason = "interrupted"
	QueuedItemStoppedReasonError       QueuedItemStoppedReason = "error"
)

type ComfyClientCallbacks struct {
	WebsocketConnected    func(*ComfyClient)
	WebsocketDisconnected func(*ComfyClient)
	// ClientQueueCountChanged func(*ComfyClient, int)
	// QueuedItemStarted       func(*ComfyClient, *QueueItem)
	// QueuedItemStopped       func(*ComfyClient, *QueueItem, QueuedItemStoppedReason)
	// QueuedItemDataAvailable func(*ComfyClient, *QueueItem, *PromptMessageData)
}

// ComfyClient is the top level object that allows for interaction with the ComfyUI backend
type ComfyClient struct {
	baseAddr              string
	clientId              string
	websocket             *WebSocketClient
	nodeObjects           *NodeObjects
	queuecount            int
	callbacks             *ComfyClientCallbacks
	lastProcessedPromptID string
	messages              chan PromptMessage
	//queueditems           map[string]*QueueItem
}

// NewComfyClientWithTimeout creates a new instance of a Comfy2go client with a connection timeout
func NewComfyClientWithTimeout(baseAddr string, callbacks *ComfyClientCallbacks, timeout int64) *ComfyClient {
	cid := uuid.New().String()
	retv := &ComfyClient{
		baseAddr: baseAddr,
		clientId: cid,
		websocket: &WebSocketClient{
			url:         "ws://" + baseAddr + "/ws?clientId=" + cid,
			isConnected: false,
			isRunning:   false,
			maxDelay:    time.Duration(timeout) * time.Second,
		},
		queuecount: 0,
		callbacks:  callbacks,
		messages:   nil,
		//queueditems: make(map[string]*QueueItem),
	}
	retv.messages = make(chan PromptMessage)
	// golang uses mark-sweep GC, so this circular reference should be fine
	retv.websocket.callback = retv
	return retv
}

// NewComfyClient creates a new instance of a Comfy2go client
func NewComfyClient(baseAddr string, callbacks *ComfyClientCallbacks) *ComfyClient {
	cid := uuid.New().String()
	retv := &ComfyClient{
		baseAddr: baseAddr,
		clientId: cid,
		websocket: &WebSocketClient{
			url:         "ws://" + baseAddr + "/ws?clientId=" + cid,
			isConnected: false,
			isRunning:   false,
			maxDelay:    60 * time.Second,
		},
		queuecount: 0,
		callbacks:  callbacks,
		messages:   nil,
		//queueditems: make(map[string]*QueueItem),
	}
	retv.messages = make(chan PromptMessage)
	// golang uses mark-sweep GC, so this circular reference should be fine
	retv.websocket.callback = retv
	return retv
}

func (cc *ComfyClient) GetMessages() chan PromptMessage {
	return cc.messages
}
func (cc *ComfyClient) OnMessage(message string) {
	cc.OnWindowSocketMessage(message)
}
func (cc *ComfyClient) OnWebsocketConnected() {
	if cc.callbacks != nil && cc.callbacks.WebsocketConnected != nil {
		cc.callbacks.WebsocketConnected(cc)
	}
}
func (cc *ComfyClient) OnWebsocketDisconnected() {
	if cc.callbacks != nil && cc.callbacks.WebsocketDisconnected != nil {
		cc.callbacks.WebsocketDisconnected(cc)
	}
}

// Init starts the websocket connection (if not already connected) and retrieves the collection of node objects
func (cc *ComfyClient) Start() {
	if !cc.websocket.isConnected {
		// as soon as the ws is connected, it will receive a "status" message of the current status
		// of the ComfyUI server
		cc.websocket.Start()
	}
	return
}

func (cc *ComfyClient) IsInitialized() bool {
	if cc.websocket.isConnected && cc.nodeObjects != nil {
		return true
	}
	return false
}
func (cc *ComfyClient) QueryNodeObjects() error {
	if !cc.websocket.isConnected {
		return ErrComfyDisconnected
	}
	// Get the object infos for the Comfy Server
	if cc.nodeObjects == nil {
		object_infos, err := cc.GetObjectInfos()
		if err != nil {
			return err
		}
		cc.nodeObjects = object_infos
	}
	return nil
}

// ClientID returns the unique client ID for the connection to the ComfyUI backend
func (c *ComfyClient) ClientID() string {
	return c.clientId
}

// NewGraphFromJsonReader creates a new graph from the data read from an io.Reader
func (cc *ComfyClient) NewGraphFromJsonReader(r io.Reader) (*Graph, *[]string, error) {
	if cc.nodeObjects == nil {
		return nil, nil, ErrNotNodeObjects
	}
	return NewGraphFromJsonReader(r, cc.nodeObjects)
}

// NewGraphFromJsonFile creates a new graph from a JSON file
func (cc *ComfyClient) NewGraphFromJsonFile(path string) (*Graph, *[]string, error) {
	if cc.nodeObjects == nil {
		return nil, nil, ErrNotNodeObjects
	}
	return NewGraphFromJsonFile(path, cc.nodeObjects)
}

// NewGraphFromJsonString creates a new graph from a JSON string
func (cc *ComfyClient) NewGraphFromJsonString(path string) (*Graph, *[]string, error) {
	if cc.nodeObjects == nil {
		return nil, nil, ErrNotNodeObjects
	}
	return NewGraphFromJsonString(path, cc.nodeObjects)
}

// NewGraphFromPNGReader extracts the workflow from PNG data read from an io.Reader and creates a new graph
func (c *ComfyClient) NewGraphFromPNGReader(r io.Reader) (*Graph, *[]string, error) {
	metadata, err := base.GetPngMetadata(r)
	if err != nil {
		return nil, nil, err
	}

	workflow, ok := metadata["workflow"]
	if !ok {
		return nil, nil, ErrNotWorkflowInPNG
	}
	reader := strings.NewReader(workflow)
	graph, missing, err := c.NewGraphFromJsonReader(reader)
	if err != nil {
		return nil, missing, err
	}
	return graph, missing, nil
}

// NewGraphFromPNGReader extracts the workflow from PNG data read from a file and creates a new graph
func (c *ComfyClient) NewGraphFromPNGFile(path string) (*Graph, *[]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, nil, err
	}
	defer file.Close()
	return c.NewGraphFromPNGReader(file)
}

// GetQueuedItem returns a QueueItem that was queued with the ComfyClient, that has not been processed yet
// or is currently being processed.  Once a QueueItem has been processed, it will not be available with this method.
// func (c *ComfyClient) GetQueuedItem(promptId string) *QueueItem {
// 	val, ok := c.queueditems[promptId]
// 	if ok {
// 		return val
// 	}
// 	return nil
// }

// OnWindowSocketMessage processes each message received from the websocket connection to ComfyUI.
// The messages are parsed, and translated into PromptMessage structs and placed into the correct QueuedItem's message channel.
func (c *ComfyClient) OnWindowSocketMessage(msg string) {
	message := &StatusMessage{}
	err := json.Unmarshal([]byte(msg), &message)
	if err != nil {
		logger.Errorf("Deserializing Status Message: %v", err)
	}

	switch message.Type {
	case "status":
	case "execution_start":
		s := message.Data.(*MessageDataExecutionStart)

		// update lastProcessedPromptID to indicate we are processing a new prompt
		c.lastProcessedPromptID = s.PromptID
		m := PromptMessage{
			Type: "started",
			Message: &PromptMessageStarted{
				PromptID: s.PromptID,
			},
		}
		if c.messages != nil {
			c.messages <- m
		}
	case "execution_cached":
		// this is probably not usefull for us
	case "executing":
		s := message.Data.(*MessageDataExecuting)

		if s.Node == nil {
			// final node was processed
			stop := false
			if s.BatchClose != nil {
				stop = *s.BatchClose
			} else {
				stop = true
			}
			m := PromptMessage{
				Type: "stopped",
				Message: &PromptMessageStopped{
					PromptID:  s.PromptID, //qi,
					Exception: nil,
					Stop:      stop,
				},
			}
			// remove the Item from our Queue before sending the message
			// no other messages will be sent to the channel after this
			if c.messages != nil {
				c.messages <- m
			}
		} else {
			//node := qi.Workflow.GetNodeById(*s.Node)
			m := PromptMessage{
				Type: "executing",
				Message: &PromptMessageExecuting{
					NodeID: *s.Node,
					Title:  "unknown", //node.DisplayName,
				},
			}
			if c.messages != nil {
				c.messages <- m
			}
		}
	case "progress":
		s := message.Data.(*MessageDataProgress)
		m := PromptMessage{
			Type: "progress",
			Message: &PromptMessageProgress{
				Value: s.Value,
				Max:   s.Max,
			},
		}
		if c.messages != nil {
			c.messages <- m
		}
	case "executed":
		s := message.Data.(*MessageDataExecuted)
		// collect the data from the output
		mdata := &PromptMessageData{
			NodeID: s.Node,
			Data:   make(map[string][]DataOutput),
		}

		for k, v := range s.Output {
			mdata.Data[k] = *v
		}
		m := PromptMessage{
			Type:    "data",
			Message: mdata,
		}
		if c.messages != nil {
			c.messages <- m
		}
	case "execution_interrupted":
		s := message.Data.(*MessageExecutionInterrupted)
		m := PromptMessage{
			Type: "stopped",
			Message: &PromptMessageStopped{
				PromptID:  s.PromptID, //qi,
				Exception: nil,
				Stop:      true,
			},
		}
		// remove the Item from our Queue before sending the message
		// no other messages will be sent to the channel after this
		if c.messages != nil {
			c.messages <- m
		}
	case "execution_error":
		s := message.Data.(*MessageExecutionError)
		nindex, _ := strconv.Atoi(s.Node) // the node id is serialized as a string
		m := PromptMessage{
			Type: "stopped",
			Message: &PromptMessageStopped{
				PromptID: s.PromptID,
				Exception: &PromptMessageStoppedException{
					NodeID:           nindex,
					NodeType:         s.NodeType,
					NodeName:         "unknonw",
					ExceptionMessage: s.ExceptionMessage,
					ExceptionType:    s.ExceptionType,
					Traceback:        s.Traceback,
				},
				Stop: true,
			},
		}
		if c.messages != nil {
			c.messages <- m
		}
	default:
		// Handle unknown data types or return a dedicated error here
		logger.Warn("Unhandled message type: %s", message.Type)
	}
}
