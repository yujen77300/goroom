package chat

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/fasthttp/websocket"
	"github.com/google/uuid"
	"github.com/yujen77300/goroom/internal/models"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

var upgrader = websocket.FastHTTPUpgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type Client struct {
	Hub  *Hub
	Conn *websocket.Conn
	Send chan []byte
}

func (c *Client) readPump() {
	defer func() {
		// 送到Run()這個receiver function的資訊
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
		message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))
		// message = ([]byte("{\"account\":\"dylan\",\"message\":\"你只能打這樣啦\"}"))
		// 送到Run()這個receiver function的資訊
		c.Hub.broadcast <- message
	}
}

func (c *Client) writeAndStorePump(roomUuid string) {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	uuidObj, err := uuid.NewRandom()
	participantuuid := uuidObj.String()[0:6]
	fmt.Println(err)
	fmt.Println("個人代號")
	fmt.Println(participantuuid)

	userIteration := 1

	for {
		select {
			// 接受 Run()這個receiver function的資訊
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				return
			}
			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}

			var m map[string]interface{}
			testerr := json.Unmarshal(message, &m)
			if testerr != nil {
				fmt.Println(err)
			}
			w.Write(message)

			if len(m) > 2 {
				if userIteration == 1 {
					fmt.Println("第一個迴圈")
					fmt.Println(string(message))
					// fmt.Println("存到資料庫")
					models.UpdateParticipantInfo(message, roomUuid, participantuuid)
					userIteration += 1
				}
	
			}

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

// 建立一個websocket連接的客戶端，與Hub進行通訊。
func PeerChatConn(c *websocket.Conn, hub *Hub, roomUuid string) {
	client := &Client{Hub: hub, Conn: c, Send: make(chan []byte, 256)}

	// 送到Run()這個receiver function的資訊
	client.Hub.register <- client

	go client.writeAndStorePump(roomUuid)
	client.readPump()
}
