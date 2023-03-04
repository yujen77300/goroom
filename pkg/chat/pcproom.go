package chat

import (
	// "bytes"

	"fmt"
	"log"
	"time"

	"github.com/fasthttp/websocket"
)

type PcpClient struct {
	Hub  *PcpHub
	Conn *websocket.Conn
	Send chan []byte
}

func (c *PcpClient) writePump(){
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

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

			fmt.Println("玉山")
			fmt.Println(string(message))

			w.Write(message)

			// w.Write([]byte("{\"account\":\"dylan\",\"message\":\"你只能打這樣啦\"}"))

			fmt.Println("測試測試這邊")
			fmt.Println(c.Send)
			fmt.Println(len(c.Send))
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
	// 如果有收到pong重新設定
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

	// 送到Run()這個receiver function的資訊
	pcpclient.Hub.register <- pcpclient

	go pcpclient.writePump()
	pcpclient.readPump()
}
