package server

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/template/html"
	"github.com/gofiber/websocket/v2"
	"github.com/yujen77300/goroom/internal/handlers"
	"github.com/yujen77300/goroom/internal/models"
	w "github.com/yujen77300/goroom/pkg/webrtc"
)

var (
	addr = flag.String("addr", ":"+os.Getenv("PORT"), "")
	cert = flag.String("cert", os.Getenv("CERT"), "")
	key  = flag.String("key", os.Getenv("KEY"), "")
)

func Run() error {
	flag.Parse()
	fmt.Println("進來主程式")
	fmt.Println(*cert)
	fmt.Println(*key)
	fmt.Println(&cert)
	fmt.Println(&key)

	if *addr == ":" {
		*addr = ":8080"
	}
	engine := html.New("./views", ".html")
	app := fiber.New(fiber.Config{Views: engine})
	app.Use(logger.New())
	app.Use(cors.New())

	app.Get("/", handlers.Welcome)
	app.Get("/member", handlers.Member)
	app.Get("/room/create", handlers.RoomCreate)
	app.Get("/room/:uuid", handlers.Room)
	app.Get("/room/:uuid/websocket", websocket.New(handlers.RoomWebsocket, websocket.Config{
		HandshakeTimeout: 10 * time.Second,
	}))
	app.Get("/room/:uuid/chat", handlers.RoomChat)
	app.Get("/room/:uuid/chat/websocket", websocket.New(handlers.RoomChatWebsocket))
	// 影響觀看人數
	app.Get("/room/:uuid/viewer/websocket", websocket.New(handlers.RoomViewerWebsocket))
	app.Post("/api/user", models.NewUser)
	app.Get("/api/alluser", models.FindALLUsers)
	app.Get("/api/user/auth", models.GetUser)
	app.Put("/api/user/auth", models.PutUser)
	app.Delete("/api/user/auth", models.SignOutUser)
	app.Get("/api/user/avatar", models.GetAvatar)
	app.Post("/api/user/avatar", models.UpdateAvatar)
	app.Static("/", "./static")

	// 讓這兩個變量進行初始化
	w.Rooms = make(map[string]*w.Room)
	w.Streams = make(map[string]*w.Room)

	if *cert != "" {
		return app.ListenTLS(*addr, *cert, *key)
	}
	return app.Listen(*addr)

}
