package wsclient

import (
	"log"

	"github.com/otis-co-ltd/aihub-recorder/internal/recorder"
)

func (c *Client) handleMessage(msg WSMessage) {
	switch msg.Type {

	case MSG_START_RECORDING:
		c.handleStartRecording()

	case MSG_STOP_RECORDING:
		c.handleStopRecording()

	case MSG_STATUS:
		log.Println("?? Status from server:", string(msg.Data))

	case MSG_ERROR:
		log.Println("? Server error:", msg.Message)

	default:
		log.Println("?? Unknown WS type:", msg.Type)
		c.sendError("Unknown message type: " + msg.Type)
	}
}

func (c *Client) handleStartRecording() {
	log.Println("?? START RECORDING received")

	if err := recorder.Start(); err != nil {
		c.sendError("Failed to start recording: " + err.Error())
		return
	}

	c.sendSuccess("recording started")
}

func (c *Client) handleStopRecording() {
	log.Println("?? STOP RECORDING received")

	if err := recorder.Stop(); err != nil {
		c.sendError("Failed to stop recording: " + err.Error())
		return
	}

	c.sendSuccess("recording stopped")
}
