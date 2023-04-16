package handlers

import (
	"fmt"
	"log"
	"os"

	"time"

	"github.com/yujen77300/goroom/pkg/chat"
	w "github.com/yujen77300/goroom/pkg/webrtc"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	guuid "github.com/google/uuid"
	"github.com/pion/webrtc/v3"
	"github.com/spf13/viper"
)



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
			"RoomWebsocketAddr":   fmt.Sprintf("%s://%s:8080/room/%s/websocket", ws, c.Hostname(), uuid),
			"RoomLink":            fmt.Sprintf("%s://%s:8080/room/%s", c.Protocol(), c.Hostname(), uuid),
			"ChatWebsocketAddr":   fmt.Sprintf("%s://%s:8080/room/%s/chat/websocket", ws, c.Hostname(), uuid),
			"ViewerWebsocketAddr": fmt.Sprintf("%s://%s:8080/room/%s/viewer/websocket", ws, c.Hostname(), uuid),
			"PcpsWebsocketAddr":   fmt.Sprintf("%s://%s:8080/room/%s/pcps/websocket", ws, c.Hostname(), uuid),
			"TurnName":   viper.GetString("TURNNAME"),
			"TurnPwd":   viper.GetString("TURNPWD"),
			"TurnName2":   viper.GetString("TURNNAME2"),
			"TurnPwd2":   viper.GetString("TURNPWD2"),
		})
	}

}


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

	p := &w.Peers{}

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

	for range ticker.C {

		w, err := c.Conn.NextWriter(websocket.TextMessage)
		if err != nil {
			return
		}

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

	w.PcpRoomsLock.Unlock()
	if room == nil {
		return
	}
	if room.PcpHub == nil {
		return
	}
	chat.PcpRoomConn(c.Conn, room.PcpHub, uuid)
}
