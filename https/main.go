package main

import (
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

func main() {
	fiberApp := fiber.New(fiber.Config{
		Prefork: true,
	})

	fiberApp.Use(logger.New())

	fiberApp.Get("/", func(c *fiber.Ctx) error {
		c.SendString("Hello World")
		return nil
	})
	log.Println(os.Getwd())
	fiberApp.ListenTLS(":8000", "./cert.pem", "./key.pem")
}
