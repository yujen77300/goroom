package handlers

import (
	"github.com/gofiber/fiber/v2"
)

func Welcome(c *fiber.Ctx) error {
	return c.Render("index", nil)
}

// a2163e26-a673-4481-be74-37ebe0d68ab2
