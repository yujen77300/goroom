package handlers

import (
	"github.com/gofiber/fiber/v2"
)

func Member(c *fiber.Ctx) error {
	livedToken := c.Cookies("MyJWT")
	if len(livedToken) == 0 {
		return c.Redirect("/")
	} else {
		return c.Render("member", nil)
	}

}
