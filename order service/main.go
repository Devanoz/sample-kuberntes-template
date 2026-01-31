package main

import (
	"context"
	"log"
	"net"
	"sync"
	"time"

	"order-service/orderpb"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

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
	log.Printf("Created order: %s for product: %s, quantity: %d", order.Id, order.ProductId, order.Quantity)

	return &orderpb.CreateOrderResponse{Order: order}, nil
}

func (s *orderServer) GetOrder(ctx context.Context, req *orderpb.GetOrderRequest) (*orderpb.GetOrderResponse, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	order, exists := s.orders[req.Id]
	if !exists {
		return &orderpb.GetOrderResponse{Order: nil}, nil
	}

	return &orderpb.GetOrderResponse{Order: order}, nil
}

func (s *orderServer) ListOrders(ctx context.Context, req *orderpb.ListOrdersRequest) (*orderpb.ListOrdersResponse, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var orders []*orderpb.Order
	for _, order := range s.orders {
		if req.ProductId == "" || order.ProductId == req.ProductId {
			orders = append(orders, order)
		}
	}

	log.Printf("Listed %d orders for product: %s", len(orders), req.ProductId)
	return &orderpb.ListOrdersResponse{Orders: orders}, nil
}

func main() {
	log.Println("Order service starting...")

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	orderpb.RegisterOrderServiceServer(grpcServer, newOrderServer())

	log.Printf("Order service listening on :50051")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
