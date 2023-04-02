package chat

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/fasthttp/websocket"
	"github.com/yujen77300/goroom/internal/models"
)

type PcpClient struct {
	Hub  *PcpHub
	Conn *websocket.Conn
	Send chan []byte
}

func (c *PcpClient) writePump(roomUuid string) {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	userIteration := 1

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				return
			}
			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}

			var dataMap map[string]interface{}
			error := json.Unmarshal(message, &dataMap)
			if error != nil {
				fmt.Println("JSON unmarshal failed:", err)
				return
			}

			eventType, ok := dataMap["event"].(string)
			if !ok {
				fmt.Println("There is no event column")
				return
			}

			switch eventType {
			case "join":
				dataStr, ok := dataMap["data"].(string)
				if !ok {
					fmt.Println("There is no data column")
					return
				}
				dataByteSlice := []byte(dataStr)
				if userIteration == 1 {
					models.UpdateParticipantInfo(dataByteSlice, roomUuid)
					userIteration += 1
				}

				w.Write(message)
			case "leave":
				dataStr, ok := dataMap["data"].(string)
				if !ok {
					fmt.Println("There is no data column")
					return
				}
				dataByteSlice := []byte(dataStr)
				models.DeleteParticipantInfo(dataByteSlice, roomUuid)
				w.Write(message)
			case "hand":
				w.Write(message)
			}

			n := len(c.Send)
			for i := 0; i < n; i++ {
				w.Write(newline)
				w.Write(<-c.Send)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (c *PcpClient) readPump() {
	defer func() {
		c.Hub.unregister <- c
		c.Conn.Close()
	}()
	c.Conn.SetReadLimit(maxMessageSize)
	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error { c.Conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}
		c.Hub.broadcast <- message
	}
}

func PcpRoomConn(c *websocket.Conn, hub *PcpHub, roomUuid string) {
	pcpclient := &PcpClient{Hub: hub, Conn: c, Send: make(chan []byte, 256)}

	pcpclient.Hub.register <- pcpclient

	go pcpclient.writePump(roomUuid)
	pcpclient.readPump()
}
