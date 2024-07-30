package client

import (
	"math"
	"sync"
	"time"

	"github.com/er1cw00/comfy.go/base/logger"
	"github.com/gorilla/websocket"
)

type WebSocketCallback interface {
	OnMessage(message string)
	OnWebsocketConnected()
	OnWebsocketDisconnected()
}

type WebSocketClient struct {
	url         string
	conn        *websocket.Conn
	isRunning   bool
	isConnected bool
	retryCount  int
	mutex       sync.Mutex // For thread-safe access to the WebSocket connection
	callback    WebSocketCallback
	maxDelay    time.Duration // The maximum delay, e.g., 1 minute
}

func (c *WebSocketClient) Ping() error {
	// Attempt to send a ping message
	err := c.conn.WriteMessage(websocket.PingMessage, nil)
	if err != nil {
		return err
	}
	return nil
}

func (c *WebSocketClient) Start() {
	if c.isRunning {
		return
	}
	c.isRunning = true
	forceLog := true
	loop := func() {
		logger.Debug("websocket client loop enter >>")
		for c.isRunning {
			if err := c.connect(); err != nil {
				if c.isConnected || forceLog {
					logger.Errorf("websocket connecting failed,err: %v", err)
				}
				//forceLog = false
				c.isConnected = false
				delay := c.getReconnectDelay()
				time.Sleep(delay)
				continue
			}
			if !c.isConnected {
				logger.Info("comfy websocket connected >>")
			}
			c.isConnected = true
			c.handleMessages()
		}
		logger.Debug("websocket client loop exit <<")
	}
	go loop()
}

func (c *WebSocketClient) connect() error {
	conn, _, err := websocket.DefaultDialer.Dial(c.url, nil)
	if err != nil {
		return err
	}
	c.conn = conn
	return nil
}

func (c *WebSocketClient) handleMessages() {
	defer func() {
		logger.Debug("handleMessages defer")
		c.conn.Close()
	}()
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			logger.Warn("Read error: %v", err)
			break
		}
		if c.callback != nil {
			c.callback.OnMessage(string(message))
		}
	}
	logger.Debug("handleMessages exit .")
}

func (c *WebSocketClient) getReconnectDelay() time.Duration {
	// Calculate the delay as BaseDelay * 2^(RetryCount), capped at MaxDelay
	delay := time.Duration(5 + math.Pow(2, float64(c.retryCount)))
	if delay > c.maxDelay {
		delay = c.maxDelay
	} else {
		c.retryCount++ // Increment the retry counter for the next attempt
	}
	if c.retryCount >= 60 {
		c.retryCount = 60
	}
	logger.Debugf("RetryCount: %d, delaty: %v", c.retryCount, delay)
	return delay
}

func (c *WebSocketClient) LockRead() {
	c.mutex.Lock()
}

func (c *WebSocketClient) UnlockRead() {
	c.mutex.Unlock()
}
