package main

import (
	"context"
	"log"
	"os"
	"time"

	"product-service/orderpb"

	"github.com/gofiber/fiber/v2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var orderClient orderpb.OrderServiceClient

func initOrderClient() {
	addr := os.Getenv("ORDER_SERVICE_ADDR")
	if addr == "" {
		addr = "localhost:50051"
	}

	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Printf("Warning: Failed to connect to order service: %v", err)
		return
	}

	orderClient = orderpb.NewOrderServiceClient(conn)
	log.Printf("Connected to order service at %s", addr)
}

type CreateOrderRequest struct {
	Quantity int32 `json:"quantity"`
}

type OrderResponse struct {
	ID        string `json:"id"`
	ProductID string `json:"product_id"`
	Quantity  int32  `json:"quantity"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
}

func orderToResponse(o *orderpb.Order) OrderResponse {
	createdAt := ""
	if o.CreatedAt != nil {
		createdAt = o.CreatedAt.AsTime().Format(time.RFC3339)
	}
	return OrderResponse{
		ID:        o.Id,
		ProductID: o.ProductId,
		Quantity:  o.Quantity,
		Status:    o.Status,
		CreatedAt: createdAt,
	}
}

func main() {
	log.Println("Product service running")

	initOrderClient()

	app := fiber.New()

	app.Get("/products", func(c *fiber.Ctx) error {
		return c.SendString("Product service")
	})

	app.Post("/products/:productId/orders", func(c *fiber.Ctx) error {
		if orderClient == nil {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"error": "Order service unavailable",
			})
		}

		productId := c.Params("productId")

		var req CreateOrderRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid request body",
			})
		}

		if req.Quantity <= 0 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Quantity must be positive",
			})
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		resp, err := orderClient.CreateOrder(ctx, &orderpb.CreateOrderRequest{
			ProductId: productId,
			Quantity:  req.Quantity,
		})
		if err != nil {
			log.Printf("Failed to create order: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to create order",
			})
		}

		return c.Status(fiber.StatusCreated).JSON(orderToResponse(resp.Order))
	})

	app.Get("/products/:productId/orders", func(c *fiber.Ctx) error {
		if orderClient == nil {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"error": "Order service unavailable",
			})
		}

		productId := c.Params("productId")

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		resp, err := orderClient.ListOrders(ctx, &orderpb.ListOrdersRequest{
			ProductId: productId,
		})
		if err != nil {
			log.Printf("Failed to list orders: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to list orders",
			})
		}

		orders := make([]OrderResponse, 0, len(resp.Orders))
		for _, o := range resp.Orders {
			orders = append(orders, orderToResponse(o))
		}

		return c.JSON(orders)
	})

	log.Fatal(app.Listen(":3000"))
}
