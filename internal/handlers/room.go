package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"time"

	"github.com/yujen77300/goroom/pkg/chat"
	w "github.com/yujen77300/goroom/pkg/webrtc"

	// "crypto/sha256"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	guuid "github.com/google/uuid"
	"github.com/pion/webrtc/v3"
)

type PcpsWsPayload struct {
	Event string `json:"event"`
	Data  string `json:"data"`
}

type PcpsInfo struct {
	StreamID string `json:"streamId"`
	PCPEmail string `json:"pcpEmail"`
	PCPId    int    `json:"pcpId"`
	PCPName  string `json:"pcpName"`
}

type LeavePcpInfo struct {
	StreamID string `json:"streamId"`
}

type PcpsConnection struct {
	Conn *websocket.Conn
}

// 存放全部上線的人
var pcpsList []PcpsInfo

func RoomCreate(c *fiber.Ctx) error {
	livedToken := c.Cookies("MyJWT")
	if len(livedToken) == 0 {
		return c.Redirect("/")
	} else {
		return c.Redirect(fmt.Sprintf("/room/%s", guuid.New().String()))
	}

}

func Room(c *fiber.Ctx) error {
	livedToken := c.Cookies("MyJWT")
	if len(livedToken) == 0 {
		return c.Redirect("/")
	} else {
		ruuid := c.Params("uuid")
		if ruuid == "" {
			c.Status(400)
			return nil
		}

		ws := "ws"
		if os.Getenv("ENVIRONMENT") == "PRODUCTION" {
			ws = "wss"
		}

		uuid, _ := createOrGetRoom(ruuid)
		return c.Render("peer", fiber.Map{
			"RoomWebsocketAddr":   fmt.Sprintf("%s://%s/room/%s/websocket", ws, c.Hostname(), uuid),
			"RoomLink":            fmt.Sprintf("%s://%s/room/%s", c.Protocol(), c.Hostname(), uuid),
			"ChatWebsocketAddr":   fmt.Sprintf("%s://%s/room/%s/chat/websocket", ws, c.Hostname(), uuid),
			"ViewerWebsocketAddr": fmt.Sprintf("%s://%s/room/%s/viewer/websocket", ws, c.Hostname(), uuid),
			"PcpsWebsocketAddr":   fmt.Sprintf("%s://%s/room/%s/pcps/websocket", ws, c.Hostname(), uuid),
		})
	}

}

// WebSocket 處理器，接收連接請求並處理客戶端和伺服器間的通訊。
func RoomWebsocket(c *websocket.Conn) {
	uuid := c.Params("uuid")
	if uuid == "" {
		return
	}

	_, room := createOrGetRoom(uuid)
	w.RoomConn(c, room.Peers)
}

func createOrGetRoom(uuid string) (string, *w.Room) {
	w.RoomsLock.Lock()
	defer w.RoomsLock.Unlock()

	if room := w.Rooms[uuid]; room != nil {
		return uuid, room
	}

	hub := chat.NewHub()
	// fmt.Println("進來測試一下hub")
	// fmt.Println(hub)
	// fmt.Println(*hub)
	// 創建一個指向 webrtc.Peers 結構體的指標，建立新的Peers結構體，並返為地址給指標變數p
	p := &w.Peers{}
	// var p *w.Peers = &w.Peers{}
	// fmt.Printf("p的資料型態是%T\n", p)
	// fmt.Println(p)

	p.TrackLocals = make(map[string]*webrtc.TrackLocalStaticRTP)
	room := &w.Room{
		Peers: p,
		Hub:   hub,
	}
	w.Rooms[uuid] = room

	go hub.Run()
	return uuid, room
}

func RoomViewerWebsocket(c *websocket.Conn) {
	uuid := c.Params("uuid")
	if uuid == "" {
		return
	}

	w.RoomsLock.Lock()
	if peer, ok := w.Rooms[uuid]; ok {
		w.RoomsLock.Unlock()
		roomViewerConn(c, peer.Peers)
		return
	}
	w.RoomsLock.Unlock()
}

func roomViewerConn(c *websocket.Conn, p *w.Peers) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	defer c.Close()

	// ticker.C=> 建立定時器事件的channel
	for range ticker.C {
		// NextWriter返回一個用於寫入下一個消息的寫入器
		// websocket.TextMessage代表數據類型為text
		w, err := c.Conn.NextWriter(websocket.TextMessage)
		if err != nil {
			return
		}
		// viewer := fmt.Sprintf("%d", len(p.Connections))
		// w.Write([]byte("{\"account\":\"dylan\",\"email\":\"dylan@gmail\",\"url\":\"https:google.com\",\"viewer\":\"" + viewer + "\"}"))
		w.Write([]byte(fmt.Sprintf("%d", len(p.Connections))))
	}
}

var connections []PcpsConnection

func RoomPcpsWebsocket(c *websocket.Conn) {

	connections = append(connections, PcpsConnection{Conn: c})

	pcpsInfo := &PcpsWsPayload{}
	for {
		_, raw, err := c.ReadMessage()
		fmt.Println("先來看一下row")
		fmt.Println(string(raw))
		if err != nil {
			log.Println(err)
			return
		} else if err := json.Unmarshal(raw, &pcpsInfo); err != nil {
			log.Println(err)
			return
		}
		fmt.Println("我只是進來這裡了了啦")
		fmt.Println(pcpsInfo)
		fmt.Println(pcpsInfo.Event)
		fmt.Println(pcpsInfo.Data)
		switch pcpsInfo.Event {
		case "join":
			var newPcpsInfo PcpsInfo
			err := json.Unmarshal([]byte(pcpsInfo.Data), &newPcpsInfo)
			if err != nil {
				fmt.Println(err)
				return
			}

			pcpsList = append(pcpsList, newPcpsInfo)

			fmt.Println("我有近來唷")
			for _, conn := range connections {
				err = conn.Conn.WriteJSON(fiber.Map{
					"event": "join",
					"data":  pcpsList,
				})
				if err != nil {
					log.Println(err)
				}
			}
		// c.WriteJSON(fiber.Map{
		// 	"event": "join",
		// 	"data": fiber.Map{
		// 		"streamId": newPcpsInfo.StreamID,
		// 		"pcpName":  newPcpsInfo.PCPName,
		// 		"pcpId":    newPcpsInfo.PCPId,
		// 		"pcpEmail": newPcpsInfo.PCPEmail,
		// 	},
		// })
		case "leave":
			var leavePcpInfo LeavePcpInfo
			err := json.Unmarshal([]byte(pcpsInfo.Data), &leavePcpInfo)
			if err != nil {
				fmt.Println(err)
				return
			}
			fmt.Println("我離開會議室")
			fmt.Println(leavePcpInfo)

		}
	}
}
