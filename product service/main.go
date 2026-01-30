package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
)

func main() {
	log.Println("Product service running")

	app := fiber.New()
	app.Get("/products", func(c *fiber.Ctx) error {
		return c.SendString("Product service")
	})
	log.Fatal(app.Listen(":3000"))

}
