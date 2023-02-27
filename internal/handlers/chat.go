package handlers

import (

	// "fmt"

	"github.com/yujen77300/goroom/pkg/chat"
	w "github.com/yujen77300/goroom/pkg/webrtc"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
)


func RoomChat(c *fiber.Ctx) error {
	return c.Render("chat", fiber.Map{}, "layouts/main")
}

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
	// fmt.Println("我在chat.go裡面")
	// fmt.Println(*room)
	// fmt.Println(room)
	// fmt.Println(&room)
	// fmt.Println(room.Hub)
	chat.PeerChatConn(c.Conn, room.Hub, uuid)
}

// func StreamChatWebsocket(c *websocket.Conn) {
// 	suuid := c.Params("suuid")
// 	if suuid == "" {
// 		return
// 	}

// 	w.RoomsLock.Lock()
// 	if stream, ok := w.Streams[suuid]; ok {
// 		w.RoomsLock.Unlock()
// 		if stream.Hub == nil {
// 			hub := chat.NewHub()
// 			stream.Hub = hub
// 			go hub.Run()
// 		}
// 		chat.PeerChatConn(c.Conn, stream.Hub)
// 		return
// 	}
// 	w.RoomsLock.Unlock()
// }
