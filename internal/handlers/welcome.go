package handlers

import (
	// "fmt"

	"github.com/gofiber/fiber/v2"
)

func Welcome(c *fiber.Ctx) error {
	livedToken := c.Cookies("MyJWT")
	if len(livedToken) != 0 {
		return c.Redirect("/member")
	}
	
	return c.Render("index", nil)
}

