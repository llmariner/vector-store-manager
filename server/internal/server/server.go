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
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
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

type fileInternalClient interface {
	GetFilePath(ctx context.Context, in *fv1.GetFilePathRequest, opts ...grpc.CallOption) (*fv1.GetFilePathResponse, error)
}

type vstoreClient interface {
	CreateVectorStore(ctx context.Context, name string, dimensions int) (int64, error)
	DeleteVectorStore(ctx context.Context, name string) error
	ListVectorStores(ctx context.Context) ([]int64, error)
}

type embedder interface {
	AddFile(ctx context.Context, collectionName, modelName, fileID, fileName, filePath string, chunkSizeTokens, chunkOverlapTokens int64) error
	DeleteFile(ctx context.Context, collectionName, fileID string) error
}

// New creates a server.
func New(
	store *store.S,
	fileGetClient fileGetClient,
	fileInternalClient fileInternalClient,
	vstoreClient vstoreClient,
	e embedder,
	model string,
	dimensions int,
) *S {
	return &S{
		store:              store,
		fileGetClient:      fileGetClient,
		fileInternalClient: fileInternalClient,
		vstoreClient:       vstoreClient,
		embedder:           e,
		model:              model,
		dimensions:         dimensions,
	}
}

// S is a server.
type S struct {
	v1.UnimplementedVectorStoreServiceServer

	model      string
	dimensions int
	embedder   embedder

	fileInternalClient fileInternalClient
	fileGetClient      fileGetClient
	vstoreClient       vstoreClient
	store              *store.S
	srv                *grpc.Server

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
		authFn := ai.Unary()
		healthSkip := func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
			if info.FullMethod == "/grpc.health.v1.Health/Check" {
				// Skip authentication for health check
				return handler(ctx, req)
			}
			return authFn(ctx, req, info, handler)
		}
		opts = append(opts, grpc.ChainUnaryInterceptor(healthSkip))
		s.enableAuth = true
	}

	grpcServer := grpc.NewServer(opts...)
	v1.RegisterVectorStoreServiceServer(grpcServer, s)
	reflection.Register(grpcServer)

	healthCheck := health.NewServer()
	healthCheck.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)
	grpc_health_v1.RegisterHealthServer(grpcServer, healthCheck)

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
