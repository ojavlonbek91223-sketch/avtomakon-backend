package handler

import (
	"time"

	"github.com/gofiber/fiber/v2"
)

func Health(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"status": "ok",
		"time":   time.Now().UTC(),
	})
}

func Ready(c *fiber.Ctx) error {
	// TODO: DB va Redis ulanishini tekshirish
	return c.JSON(fiber.Map{
		"status": "ready",
	})
}
