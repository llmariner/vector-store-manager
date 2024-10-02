package server

import (
	"context"
	"fmt"
	"net"

	"github.com/go-logr/logr"
	"github.com/llmariner/api-usage/pkg/sender"
	fv1 "github.com/llmariner/file-manager/api/v1"
	"github.com/llmariner/rbac-manager/pkg/auth"
	v1 "github.com/llmariner/vector-store-manager/api/v1"
	"github.com/llmariner/vector-store-manager/server/internal/config"
	"github.com/llmariner/vector-store-manager/server/internal/store"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
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
	log logr.Logger,
) *S {
	return &S{
		store:              store,
		fileGetClient:      fileGetClient,
		fileInternalClient: fileInternalClient,
		vstoreClient:       vstoreClient,
		embedder:           e,
		model:              model,
		dimensions:         dimensions,
		log:                log.WithName("grpc"),
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
	log                logr.Logger

	srv *grpc.Server
}

// Run starts the gRPC server.
func (s *S) Run(ctx context.Context, port int, authConfig config.AuthConfig, usage sender.UsageSetter) error {
	s.log.Info("Starting gRPC server...", "port", port)

	var opt grpc.ServerOption
	if authConfig.Enable {
		ai, err := auth.NewInterceptor(ctx, auth.Config{
			RBACServerAddr: authConfig.RBACInternalServerAddr,
			AccessResource: "api.vector-stores",
		})
		if err != nil {
			return err
		}
		opt = grpc.ChainUnaryInterceptor(ai.Unary("/grpc.health.v1.Health/Check"), sender.Unary(usage))
	} else {
		fakeAuth := func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
			return handler(fakeAuthInto(ctx), req)
		}
		opt = grpc.ChainUnaryInterceptor(fakeAuth, sender.Unary(usage))
	}

	grpcServer := grpc.NewServer(opt)
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

// fakeAuthInto sets dummy user info and token into the context.
func fakeAuthInto(ctx context.Context) context.Context {
	return auth.AppendUserInfoToContext(ctx, auth.UserInfo{
		OrganizationID: defaultOrganizationID,
		ProjectID:      defaultProjectID,
		TenantID:       defaultTenantID,
	})
}
