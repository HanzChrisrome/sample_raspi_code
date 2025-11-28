package wsclient

import (
	"fmt"
	"log"
	"net/url"
	"time"

	//External Packages:
	"github.com/gorilla/websocket"

	//Internal Packages:
	"github.com/otis-co-ltd/aihub-recorder/internal/config"
	"github.com/otis-co-ltd/aihub-recorder/internal/recorder"
)

func Start(piID string) {
	for {
		err := connectAndListen(piID)
		if err != nil {
			log.Println("Disconnected, retrying in", config.ReconnectSeconds, "seconds...")
			time.Sleep(time.Duration(config.ReconnectSeconds) * time.Second)
		}
	}
}

func connectAndListen(piID string) error {
	wsURL := url.URL{
		Scheme: "ws",
		Host:   config.BackendHost,
		Path:   config.WebSocketPath,
	}

	q := wsURL.Query()
	q.Set("pi_id", piID)
	wsURL.RawQuery = q.Encode()

	log.Println("Connecting to:", wsURL.String())

	conn, _, err := websocket.DefaultDialer.Dial(wsURL.String(), nil)
	if err != nil {
		log.Println("WebSocket connect failed:", err)
		return err
	}

	defer conn.Close()
	log.Println("Connected to backend.")

	for {
		msgType, msg, err := conn.ReadMessage()
		if err != nil {
			log.Println("Read error:", err)
			return err
		}

		if msgType != websocket.TextMessage {
			continue
		}

		cmd := string(msg)
		fmt.Println("Received command:", cmd)

		switch cmd {
		case "START":
			recorder.Start()
		case "STOP":
			recorder.Stop()
		case "PING":
			fmt.Println("tangina mo aeron")
			conn.WriteMessage(websocket.TextMessage, []byte("PONG"))
		}
	}
}
