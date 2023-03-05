package handlers

import (
	// "fmt"

	"github.com/gofiber/fiber/v2"
)

func Welcome(c *fiber.Ctx) error {
	// Get the request header
	// header := c.Request().Header

	// Check if the X-Forwarded-For header is present
	// 回傳的網址季的要在%s後面加上8080
	// host := header.Peek("Host")
	// xff := header.Peek("X-Forwarded-For")
	// xrip := header.Peek("X-Real-IP")
	// xfp := header.Peek("X-Forwarded-Proto")
	// if len(xff) > 0 {
	// 	fmt.Printf("客戶請求的主機名稱 ,%s\n", string(host))
	// 	fmt.Printf("X-Forwarded-For是 ,%s\n", string(xff))
	// 	fmt.Printf("客戶的真實ip: ,%s\n", string(xrip))
	// 	fmt.Printf("客戶的請求的協議: ,%s\n", string(xfp))
	// 	fmt.Println("有收到來自Nginx")
	// } else {
	// 	fmt.Println("沒有收到來自Nginx")
	// 	fmt.Printf("X-Forwarded-For是,%s\n", string(xff))
	// 	fmt.Printf("客戶請求的主機名稱 ,%s\n", string(host))
	// 	fmt.Printf("客戶的請求的協議: ,%s\n", string(xfp))
	// }

	return c.Render("index", nil)
}

// a2163e26-a673-4481-be74-37ebe0d68ab2
