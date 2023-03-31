package handlers

import (
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
	"github.com/spf13/viper"
)

// type PcpsWsPayload struct {
// 	Event string `json:"event"`
// 	Data  string `json:"data"`
// }

// type PcpsInfo struct {
// 	StreamID string `json:"streamId"`
// 	PCPEmail string `json:"pcpEmail"`
// 	PCPId    int    `json:"pcpId"`
// 	PCPName  string `json:"pcpName"`
// }

// type LeavePcpInfo struct {
// 	StreamID string `json:"streamId"`
// }

// type PcpsConnection struct {
// 	Conn     *websocket.Conn
// 	RoomUuid string
// }

// 存放全部上線的人
// var pcpsList []PcpsInfo

func RoomCreate(c *fiber.Ctx) error {
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		log.Fatalln(err)
	}

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
			"TurnName":   viper.GetString("TURNNAME"),
			"TurnPwd":   viper.GetString("TURNPWD"),
			"TurnName2":   viper.GetString("TURNNAME2"),
			"TurnPwd2":   viper.GetString("TURNPWD2"),
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
	pcphub := chat.NewPcpHub()
	// fmt.Println("進來測試一下hub")
	// fmt.Println(hub)
	// fmt.Println(pcphub)
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
	pcpRoom := &w.PcpRoom{
		Peers:  p,
		PcpHub: pcphub,
	}
	w.Rooms[uuid] = room
	w.PcpRooms[uuid] = pcpRoom

	go hub.Run()
	go pcphub.Run()
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

func RoomPcpsWebsocket(c *websocket.Conn) {
	uuid := c.Params("uuid")
	if uuid == "" {
		return
	}
	w.PcpRoomsLock.Lock()
	room := w.PcpRooms[uuid]
	// fmt.Println("我在這個裡面")
	// fmt.Println(room)
	w.PcpRoomsLock.Unlock()
	if room == nil {
		return
	}
	if room.PcpHub == nil {
		return
	}
	chat.PcpRoomConn(c.Conn, room.PcpHub, uuid)
}
