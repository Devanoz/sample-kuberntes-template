package main

import (
	"context"
	"encoding/json"
	"log"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"sync"
	"time"

	"order-service/orderpb"
	"order-service/telemetry"

	"github.com/google/uuid"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var logger *slog.Logger

type orderServer struct {
	orderpb.UnimplementedOrderServiceServer
	mu     sync.RWMutex
	orders map[string]*orderpb.Order
}

func newOrderServer() *orderServer {
	return &orderServer{
		orders: make(map[string]*orderpb.Order),
	}
}

func (s *orderServer) CreateOrder(ctx context.Context, req *orderpb.CreateOrderRequest) (*orderpb.CreateOrderResponse, error) {
	span := trace.SpanFromContext(ctx)
	traceID := span.SpanContext().TraceID().String()

	// Log request payload to span and structured log
	reqJSON, _ := json.Marshal(map[string]interface{}{
		"product_id": req.ProductId,
		"quantity":   req.Quantity,
	})
	span.SetAttributes(attribute.String("rpc.request.body", string(reqJSON)))
	logger.Info("request",
		"traceId", traceID,
		"method", "CreateOrder",
		"body", string(reqJSON))

	s.mu.Lock()
	defer s.mu.Unlock()

	order := &orderpb.Order{
		Id:        uuid.New().String(),
		ProductId: req.ProductId,
		Quantity:  req.Quantity,
		Status:    "pending",
		CreatedAt: timestamppb.New(time.Now()),
	}

	s.orders[order.Id] = order

	// Log response payload to span and structured log
	respJSON, _ := json.Marshal(map[string]interface{}{
		"id":         order.Id,
		"product_id": order.ProductId,
		"quantity":   order.Quantity,
		"status":     order.Status,
		"created_at": order.CreatedAt.AsTime().Format(time.RFC3339),
	})
	span.SetAttributes(attribute.String("rpc.response.body", string(respJSON)))
	logger.Info("response",
		"traceId", traceID,
		"method", "CreateOrder",
		"body", string(respJSON))

	return &orderpb.CreateOrderResponse{Order: order}, nil
}

func (s *orderServer) GetOrder(ctx context.Context, req *orderpb.GetOrderRequest) (*orderpb.GetOrderResponse, error) {
	span := trace.SpanFromContext(ctx)
	traceID := span.SpanContext().TraceID().String()

	// Log request payload
	reqJSON, _ := json.Marshal(map[string]interface{}{"id": req.Id})
	span.SetAttributes(attribute.String("rpc.request.body", string(reqJSON)))
	logger.Info("request",
		"traceId", traceID,
		"method", "GetOrder",
		"body", string(reqJSON))

	s.mu.RLock()
	defer s.mu.RUnlock()

	order, exists := s.orders[req.Id]
	if !exists {
		logger.Info("response",
			"traceId", traceID,
			"method", "GetOrder",
			"body", "null")
		return &orderpb.GetOrderResponse{Order: nil}, nil
	}

	// Log response payload
	respJSON, _ := json.Marshal(map[string]interface{}{
		"id":         order.Id,
		"product_id": order.ProductId,
		"quantity":   order.Quantity,
		"status":     order.Status,
		"created_at": order.CreatedAt.AsTime().Format(time.RFC3339),
	})
	span.SetAttributes(attribute.String("rpc.response.body", string(respJSON)))
	logger.Info("response",
		"traceId", traceID,
		"method", "GetOrder",
		"body", string(respJSON))

	return &orderpb.GetOrderResponse{Order: order}, nil
}

func (s *orderServer) ListOrders(ctx context.Context, req *orderpb.ListOrdersRequest) (*orderpb.ListOrdersResponse, error) {
	span := trace.SpanFromContext(ctx)
	traceID := span.SpanContext().TraceID().String()

	// Log request payload
	reqJSON, _ := json.Marshal(map[string]interface{}{"product_id": req.ProductId})
	span.SetAttributes(attribute.String("rpc.request.body", string(reqJSON)))
	logger.Info("request",
		"traceId", traceID,
		"method", "ListOrders",
		"body", string(reqJSON))

	s.mu.RLock()
	defer s.mu.RUnlock()

	var orders []*orderpb.Order
	for _, order := range s.orders {
		if req.ProductId == "" || order.ProductId == req.ProductId {
			orders = append(orders, order)
		}
	}

	// Log response payload
	var ordersList []map[string]interface{}
	for _, o := range orders {
		ordersList = append(ordersList, map[string]interface{}{
			"id":         o.Id,
			"product_id": o.ProductId,
			"quantity":   o.Quantity,
			"status":     o.Status,
			"created_at": o.CreatedAt.AsTime().Format(time.RFC3339),
		})
	}
	respJSON, _ := json.Marshal(ordersList)
	span.SetAttributes(attribute.String("rpc.response.body", string(respJSON)))
	logger.Info("response",
		"traceId", traceID,
		"method", "ListOrders",
		"body", string(respJSON))

	return &orderpb.ListOrdersResponse{Orders: orders}, nil
}

func main() {
	// Initialize structured logger
	logger = slog.New(slog.NewJSONHandler(os.Stdout, nil))

	// Initialize telemetry
	shutdown, err := telemetry.InitTracer("order-service")
	if err != nil {
		log.Fatalf("Failed to init tracer: %v", err)
	}
	defer func() {
		if err := shutdown(context.Background()); err != nil {
			log.Printf("Error shutting down tracer: %v", err)
		}
	}()

	log.Println("Order service starting...")

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	// Add OTel interceptors for automatic tracing
	grpcServer := grpc.NewServer(
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
	)
	orderpb.RegisterOrderServiceServer(grpcServer, newOrderServer())

	// Graceful shutdown
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, os.Interrupt)
		<-sigCh
		log.Println("Shutting down...")
		grpcServer.GracefulStop()
	}()

	log.Printf("Order service listening on :50051")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
