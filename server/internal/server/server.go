package server

import (
	"context"
	"fmt"
	"log"
	"net"

	fv1 "github.com/llm-operator/file-manager/api/v1"
	"github.com/llm-operator/rbac-manager/pkg/auth"
	v1 "github.com/llm-operator/vector-store-manager/api/v1"
	"github.com/llm-operator/vector-store-manager/server/internal/config"
	"github.com/llm-operator/vector-store-manager/server/internal/store"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

const (
	defaultProjectID      = "default"
	defaultOrganizationID = "default"
	defaultTenantID       = "default-tenant-id"

	defaultPageSize = 20
	maxPageSize     = 100
)

type fileGetClient interface {
	GetFile(ctx context.Context, in *fv1.GetFileRequest, opts ...grpc.CallOption) (*fv1.File, error)
}

type vstoreClient interface {
	CreateVectorStore(ctx context.Context, name string) (int64, error)
	DeleteVectorStore(ctx context.Context, name string) error
	ListVectorStores(ctx context.Context) ([]int64, error)
}

// New creates a server.
func New(
	store *store.S,
	fileGetClient fileGetClient,
	vstoreClient vstoreClient,
) *S {
	return &S{
		store:         store,
		fileGetClient: fileGetClient,
		vstoreClient:  vstoreClient,
	}
}

// S is a server.
type S struct {
	v1.UnimplementedVectorStoreServiceServer

	fileGetClient fileGetClient
	vstoreClient  vstoreClient
	store         *store.S
	srv           *grpc.Server

	enableAuth bool
}

// Run starts the gRPC server.
func (s *S) Run(ctx context.Context, port int, authConfig config.AuthConfig) error {
	log.Printf("Starting server on port %d\n", port)

	var opts []grpc.ServerOption
	if authConfig.Enable {
		ai, err := auth.NewInterceptor(ctx, auth.Config{
			RBACServerAddr: authConfig.RBACInternalServerAddr,
			AccessResource: "api.vector-stores",
		})
		if err != nil {
			return err
		}
		opts = append(opts, grpc.ChainUnaryInterceptor(ai.Unary()))
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

func (s *S) extractUserInfoFromContext(ctx context.Context) (*auth.UserInfo, error) {
	if !s.enableAuth {
		return &auth.UserInfo{
			OrganizationID: defaultOrganizationID,
			ProjectID:      defaultProjectID,
			TenantID:       defaultTenantID,
		}, nil
	}
	var ok bool
	userInfo, ok := auth.ExtractUserInfoFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "user info not found")
	}
	return userInfo, nil
}
