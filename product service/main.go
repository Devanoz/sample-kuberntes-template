package main

import (
	"context"
	"encoding/json"
	"log"
	"log/slog"
	"os"
	"time"

	"product-service/orderpb"
	"product-service/telemetry"

	"github.com/gofiber/contrib/otelfiber/v2"
	"github.com/gofiber/fiber/v2"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var orderClient orderpb.OrderServiceClient
var logger *slog.Logger

func initOrderClient() {
	addr := os.Getenv("ORDER_SERVICE_ADDR")
	if addr == "" {
		addr = "localhost:50051"
	}

	// Add OTel interceptor for client-side tracing
	conn, err := grpc.NewClient(addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
	)
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
	// Initialize structured logger
	logger = slog.New(slog.NewJSONHandler(os.Stdout, nil))

	// Initialize telemetry
	shutdown, err := telemetry.InitTracer("product-service")
	if err != nil {
		log.Fatalf("Failed to init tracer: %v", err)
	}
	defer func() {
		if err := shutdown(context.Background()); err != nil {
			log.Printf("Error shutting down tracer: %v", err)
		}
	}()

	log.Println("Product service running")

	initOrderClient()

	app := fiber.New()

	// Add OTel middleware for HTTP tracing
	app.Use(otelfiber.Middleware())

	app.Get("/products", func(c *fiber.Ctx) error {
		return c.SendString("Product service")
	})

	app.Post("/products/:productId/orders", func(c *fiber.Ctx) error {
		span := trace.SpanFromContext(c.UserContext())
		traceID := span.SpanContext().TraceID().String()

		span.SetAttributes(attribute.String("http.request.body", string(c.Body())))
		logger.Info("request",
			"traceId", traceID,
			"method", "POST",
			"path", c.Path(),
			"body", string(c.Body()))

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

		ctx, cancel := context.WithTimeout(c.UserContext(), 5*time.Second)
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

		response := orderToResponse(resp.Order)
		responseJSON, _ := json.Marshal(response)
		span.SetAttributes(attribute.String("http.response.body", string(responseJSON)))
		logger.Info("response",
			"traceId", traceID,
			"status", 201,
			"body", string(responseJSON))

		return c.Status(fiber.StatusCreated).JSON(response)
	})

	app.Get("/products/:productId/orders", func(c *fiber.Ctx) error {
		span := trace.SpanFromContext(c.UserContext())
		traceID := span.SpanContext().TraceID().String()

		logger.Info("request",
			"traceId", traceID,
			"method", "GET",
			"path", c.Path())

		if orderClient == nil {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"error": "Order service unavailable",
			})
		}

		productId := c.Params("productId")

		ctx, cancel := context.WithTimeout(c.UserContext(), 5*time.Second)
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

		responseJSON, _ := json.Marshal(orders)
		span.SetAttributes(attribute.String("http.response.body", string(responseJSON)))
		logger.Info("response",
			"traceId", traceID,
			"status", 200,
			"body", string(responseJSON))

		return c.JSON(orders)
	})

	log.Fatal(app.Listen(":3000"))
}
