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
		err := c.SendString("Hello World")
		if err != nil {
			return err
		}
		return nil
	})
	log.Println(os.Getwd())
	err := fiberApp.ListenTLS(":8000", "server.crt", "server.key")
	if err != nil {
		log.Println(os.Getwd())
		log.Fatal(err)
		return
	}
}
