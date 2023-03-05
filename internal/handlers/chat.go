package handlers

import (
	"github.com/yujen77300/goroom/pkg/chat"
	w "github.com/yujen77300/goroom/pkg/webrtc"


	"github.com/gofiber/websocket/v2"
)


func RoomChatWebsocket(c *websocket.Conn) {
	uuid := c.Params("uuid")
	if uuid == "" {
		return
	}

	w.RoomsLock.Lock()
	room := w.Rooms[uuid]
	w.RoomsLock.Unlock()
	if room == nil {
		return
	}
	if room.Hub == nil {
		return
	}
	chat.PeerChatConn(c.Conn, room.Hub, uuid)
}
