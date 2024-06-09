package server

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"

	"github.com/llm-operator/rbac-manager/pkg/auth"
	v1 "github.com/llm-operator/vector-store-manager/api/v1"
	"github.com/llm-operator/vector-store-manager/server/internal/config"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type reqIntercepter interface {
	InterceptHTTPRequest(req *http.Request) (int, auth.UserInfo, error)
}

// New creates a server.
func New() *S {
	return &S{}
}

// S is a server.
type S struct {
	v1.UnimplementedVectorStoreServiceServer

	srv *grpc.Server

	reqIntercepter reqIntercepter
	enableAuth     bool
}

// Run starts the gRPC server.
func (s *S) Run(ctx context.Context, port int, authConfig config.AuthConfig) error {
	log.Printf("Starting server on port %d\n", port)

	var opts []grpc.ServerOption
	if authConfig.Enable {
		ai, err := auth.NewInterceptor(ctx, auth.Config{
			RBACServerAddr: authConfig.RBACInternalServerAddr,
			AccessResource: "api.files",
		})
		if err != nil {
			return err
		}
		opts = append(opts, grpc.ChainUnaryInterceptor(ai.Unary()))

		s.reqIntercepter = ai
		s.enableAuth = true
	}

	grpcServer := grpc.NewServer(opts...)
	v1.RegisterVectorStoreServiceServer(grpcServer, s)
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

// Stop stops the gRPC server.
func (s *S) Stop() {
	s.srv.Stop()
}
