.PHONY: proto clean

proto:
	protoc --go_out=. --go_opt=paths=source_relative \
	       --go-grpc_out=. --go-grpc_opt=paths=source_relative \
	       proto/order.proto
	cp proto/order.pb.go "order service/orderpb/"
	cp proto/order_grpc.pb.go "order service/orderpb/"
	cp proto/order.pb.go "product service/orderpb/"
	cp proto/order_grpc.pb.go "product service/orderpb/"
	rm proto/order.pb.go proto/order_grpc.pb.go

clean:
	rm -f "order service/orderpb/"*.pb.go
	rm -f "product service/orderpb/"*.pb.go