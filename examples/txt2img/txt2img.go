package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	comfy "github.com/er1cw00/comfy.go"
	"github.com/er1cw00/comfy.go/base/logger"
)

var serverAddr string
var workflow string
var help bool

func usage() {
	fmt.Printf("simple_api\r\n")
	fmt.Printf("Usage: simple_api [-sh]\r\n")
	fmt.Printf("           -h help\r\n")
	fmt.Printf("           -s comfyui host\r\n")
	fmt.Printf("           -w workflow json file\r\n")
}

func init() {
	flag.BoolVar(&help, "h", false, "show this help")
	flag.StringVar(&serverAddr, "s", "localhost:7680", "ComfyUI hostname")
	flag.StringVar(&workflow, "w", "./simple_api.json", "ComfyUI hostname")
	flag.Usage = usage
}

func main() {
	var (
		err error              = nil
		g   *comfy.Graph       = nil
		cc  *comfy.ComfyClient = nil
	)
	flag.Parse()
	if help {
		usage()
		os.Exit(0)
	}
	if len(serverAddr) == 0 && len(workflow) == 0 {
		usage()
		os.Exit(0)
	}
	logger.New("Debug", "txt2img")
	callbacks := &comfy.ComfyClientCallbacks{
		ClientQueueCountChanged: func(c *comfy.ComfyClient, queuecount int) {
			logger.Debugf("Client %s at %s Queue size: %d\n", c.ClientID(), serverAddr, queuecount)
		},
		QueuedItemStarted: func(c *comfy.ComfyClient, qi *comfy.QueueItem) {
			logger.Debugf("Queued item %s started\n", qi.PromptID)
		},
		QueuedItemStopped: func(cc *comfy.ComfyClient, qi *comfy.QueueItem, reason comfy.QueuedItemStoppedReason) {
			logger.Debugf("Queued item %s stopped\n", qi.PromptID)
		},
		QueuedItemDataAvailable: func(cc *comfy.ComfyClient, qi *comfy.QueueItem, pmd *comfy.PromptMessageData) {
			logger.Debugf("image data available:\n")
		},
		WebsocketConnected: func(cc *comfy.ComfyClient) {
			log.Printf("websocket connected\n")
			if err := cc.QueryNodeObjects(); err != nil {
				log.Printf("query node objects fail, err: %v", err)
				return
			}
		},
		WebsocketDisconnected: func(cc *comfy.ComfyClient) {
			log.Printf("websocket disconnected\n")
		},
	}
	cc = comfy.NewComfyClient(serverAddr, callbacks)
	cc.Start()
	for {
		if cc.IsInitialized() {
			break
		}
		time.Sleep(time.Second * 1)
	}
	if g, _, err = cc.NewGraphFromJsonFile(workflow); err != nil {
		logger.Errorf("create graph fail, err: %v\n", err)
		return
	}
	loader := g.GetNodeById(4)
	positive := g.GetNodeById(20)
	negative := g.GetNodeById(21)
	// for k, v := range loader.Properties {
	// 	logger.Debugf("%s, %v", k, v)
	// }
	logger.Debugf("ckpt_name: %v", loader.GetPropertyWithName("ckpt_name").GetValue())
	loader.GetPropertyWithName("ckpt_name").SetValue("v1-5-pruned-emaonly.safetensors")
	positive.GetPropertyWithName("text").SetValue("1girl, dancing, outdoor, large breasts")
	negative.GetPropertyWithName("text").SetValue("text, watermark")

	item, err := cc.QueuePrompt(g)
	if err != nil {
		log.Println("Failed to queue prompt: ", err)
		os.Exit(1)
	}

	//var currentNodeTitle string
	for continueLoop := true; continueLoop; {
		msg := <-item.Messages
		switch msg.Type {
		case "started":
			qm := msg.ToPromptMessageStarted()
			log.Printf("Start executing prompt ID %s\n", qm.PromptID)
		case "executing":

			qm := msg.ToPromptMessageExecuting()
			// store the node's title so we can use it in the progress bar
			//currentNodeTitle = qm.Title
			log.Printf("Executing Node: %d\n", qm.NodeID)
		case "progress":
			// update our progress bar
			qm := msg.ToPromptMessageProgress()
			fmt.Printf("\rprogress: %d", qm.Value)
		case "stopped":
			// if we were stopped for an exception, display the exception message
			qm := msg.ToPromptMessageStopped()
			if qm.Exception != nil {
				log.Println(qm.Exception)
				os.Exit(1)
			}
			continueLoop = false
		case "data":
			qm := msg.ToPromptMessageData()
			// data objects have the fields: Filename, Subfolder, Type
			// * Subfolder is the subfolder in the output directory
			// * Type is the type of the image temp/
			for k, v := range qm.Data {
				if k == "images" || k == "gifs" {
					for _, output := range v {
						img_data, err := cc.GetImage(output)
						if err != nil {
							log.Println("Failed to get image:", err)
							os.Exit(1)
						}
						f, err := os.Create(output.Filename)
						if err != nil {
							log.Println("Failed to write image:", err)
							os.Exit(1)
						}
						f.Write(*img_data)
						f.Close()
						log.Println("Got data: ", output.Filename)
					}
				}
			}
		}
	}
}
