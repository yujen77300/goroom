package handlers

import (
	"fmt"
	"os"

	"time"

	"github.com/yujen77300/goroom/pkg/chat"
	w "github.com/yujen77300/goroom/pkg/webrtc"

	"crypto/sha256"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	guuid "github.com/google/uuid"
	"github.com/pion/webrtc/v3"
)

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
		uuid := c.Params("uuid")
		if uuid == "" {
			c.Status(400)
			return nil
		}

		ws := "ws"
		if os.Getenv("ENVIRONMENT") == "PRODUCTION" {
			ws = "wss"
		}

		uuid, suuid, _ := createOrGetRoom(uuid)
		return c.Render("peer", fiber.Map{
			"RoomWebsocketAddr":   fmt.Sprintf("%s://%s/room/%s/websocket", ws, c.Hostname(), uuid),
			"RoomLink":            fmt.Sprintf("%s://%s/room/%s", c.Protocol(), c.Hostname(), uuid),
			"ChatWebsocketAddr":   fmt.Sprintf("%s://%s/room/%s/chat/websocket", ws, c.Hostname(), uuid),
			"ViewerWebsocketAddr": fmt.Sprintf("%s://%s/room/%s/viewer/websocket", ws, c.Hostname(), uuid),
			"StreamLink":          fmt.Sprintf("%s://%s/stream/%s", c.Protocol(), c.Hostname(), suuid),
			"Type":                "room",
		})
	}

}

// WebSocket 處理器，接收連接請求並處理客戶端和伺服器間的通訊。
func RoomWebsocket(c *websocket.Conn) {
	uuid := c.Params("uuid")
	if uuid == "" {
		return
	}

	_, _, room := createOrGetRoom(uuid)
	w.RoomConn(c, room.Peers)
}

func createOrGetRoom(uuid string) (string, string, *w.Room) {
	// 鎖住w還是鎖住Rooms?
	w.RoomsLock.Lock()
	defer w.RoomsLock.Unlock()

	// 產生hash.Hash的一個interface
	h := sha256.New()
	h.Write([]byte(uuid))
	suuid := fmt.Sprintf("%x", h.Sum(nil))

	if room := w.Rooms[uuid]; room != nil {

		if _, ok := w.Streams[suuid]; !ok {
			w.Streams[suuid] = room
		}
		fmt.Println("測試房間號碼")
		fmt.Println(room)
		fmt.Println("測試uuid")
		fmt.Println(uuid)
		fmt.Println("測試suuid")
		fmt.Println(suuid)
		return uuid, suuid, room
	}

	hub := chat.NewHub()
	fmt.Println("進來測試一下hub")
	fmt.Println(hub)
	fmt.Println(*hub)
	// 創建一個指向 webrtc.Peers 結構體的指標，建立新的Peers結構體，並返為地址給指標變數p
	p := &w.Peers{}
	// var p *w.Peers = &w.Peers{}
	fmt.Printf("p的資料型態是%T\n", p)
	fmt.Println(p)

	p.TrackLocals = make(map[string]*webrtc.TrackLocalStaticRTP)
	room := &w.Room{
		Peers: p,
		Hub:   hub,
	}
	w.Rooms[uuid] = room
	w.Streams[suuid] = room

	go hub.Run()
	return uuid, suuid, room
}

func RoomViewerWebsocket(c *websocket.Conn) {
	uuid := c.Params("uuid")
	if uuid == "" {
		return
	}

	w.RoomsLock.Lock()
	fmt.Println("進來RoomViewerWebsocket")
	if peer, ok := w.Rooms[uuid]; ok {
		fmt.Println("有近來這一層")
		w.RoomsLock.Unlock()
		roomViewerConn(c, peer.Peers)
		return
	}
	fmt.Println("有離開了")
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
		w.Write([]byte(fmt.Sprintf("%d", len(p.Connections))))
	}
}
