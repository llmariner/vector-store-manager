package server

import (
	"context"
	"fmt"
	"log"
	"net"

	v1 "github.com/llmariner/vector-store-manager/api/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type retriever interface {
	Search(ctx context.Context, collectionName, modelName, query string, numDocs int) ([]string, error)
}

// NewInternal creates an internal server.
func NewInternal(model string, r retriever) *IS {
	return &IS{
		model:     model,
		retriever: r,
	}
}

// IS is an internal server.
type IS struct {
	v1.UnimplementedVectorStoreInternalServiceServer

	model     string
	retriever retriever
	srv       *grpc.Server
}

// Run starts the internal gRPC server.
func (s *IS) Run(port int) error {
	log.Printf("Starting internal server on port %d\n", port)

	grpcServer := grpc.NewServer()
	v1.RegisterVectorStoreInternalServiceServer(grpcServer, s)
	reflection.Register(grpcServer)

	s.srv = grpcServer

	l, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return fmt.Errorf("listen: %s", err)
	}
	if err := grpcServer.Serve(l); err != nil {
		return fmt.Errorf("serve: %s", err)
	}
	return nil
}
