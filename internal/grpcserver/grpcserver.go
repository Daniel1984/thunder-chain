package grpcserver

import (
	"fmt"
	"net"

	"com.perkunas/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func Serve(port string, service proto.TransactionServiceServer) error {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		return fmt.Errorf("failed starting net listener %w", err)
	}

	server := grpc.NewServer()
	reflection.Register(server)

	proto.RegisterTransactionServiceServer(server, service)

	return server.Serve(listener)
}
