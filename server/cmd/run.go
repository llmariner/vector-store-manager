package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/go-logr/stdr"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/llm-operator/inference-manager/pkg/llmkind"
	"github.com/llmariner/api-usage/pkg/sender"
	"github.com/llmariner/common/pkg/db"
	fv1 "github.com/llmariner/file-manager/api/v1"
	"github.com/llmariner/rbac-manager/pkg/auth"
	v1 "github.com/llmariner/vector-store-manager/api/v1"
	"github.com/llmariner/vector-store-manager/server/internal/config"
	"github.com/llmariner/vector-store-manager/server/internal/embedder"
	"github.com/llmariner/vector-store-manager/server/internal/milvus"
	"github.com/llmariner/vector-store-manager/server/internal/ollama"
	"github.com/llmariner/vector-store-manager/server/internal/s3"
	"github.com/llmariner/vector-store-manager/server/internal/server"
	"github.com/llmariner/vector-store-manager/server/internal/store"
	"github.com/llmariner/vector-store-manager/server/internal/vllm"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/protobuf/encoding/protojson"
)

func runCmd() *cobra.Command {
	var path string
	var logLevel int
	cmd := &cobra.Command{
		Use:   "run",
		Short: "run",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := config.Parse(path)
			if err != nil {
				return err
			}
			if err := c.Validate(); err != nil {
				return err
			}
			stdr.SetVerbosity(logLevel)
			if err := run(cmd.Context(), &c); err != nil {
				return err
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&path, "config", "", "Path to the config file")
	cmd.Flags().IntVar(&logLevel, "v", 0, "Log level")
	_ = cmd.MarkFlagRequired("config")
	return cmd
}

func run(ctx context.Context, c *config.Config) error {
	logger := stdr.New(log.Default())
	log := logger.WithName("boot")

	addr := fmt.Sprintf("localhost:%d", c.GRPCPort)
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	conn, err := grpc.NewClient(addr, opts...)
	if err != nil {
		return err
	}
	mux := runtime.NewServeMux(
		runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{
			// Do not use the camel case for JSON fields to follow OpenAI API.
			MarshalOptions: protojson.MarshalOptions{
				UseProtoNames:     true,
				EmitDefaultValues: true,
			},
			UnmarshalOptions: protojson.UnmarshalOptions{
				DiscardUnknown: true,
			},
		}),
		runtime.WithIncomingHeaderMatcher(auth.HeaderMatcher),
		runtime.WithHealthzEndpoint(grpc_health_v1.NewHealthClient(conn)),
	)
	if err := v1.RegisterVectorStoreServiceHandlerFromEndpoint(ctx, mux, addr, opts); err != nil {
		return err
	}

	conn, err = grpc.NewClient(c.FileManagerServerAddr, opts...)
	if err != nil {
		return err
	}
	fclient := fv1.NewFilesServiceClient(conn)

	conn, err = grpc.NewClient(c.FileManagerServerInternalAddr, opts...)
	if err != nil {
		return err
	}
	fwClient := fv1.NewFilesInternalServiceClient(conn)

	dbInst, err := db.OpenDB(c.Database)
	if err != nil {
		return err
	}

	st := store.New(dbInst)
	if err := st.AutoMigrate(); err != nil {
		return err
	}

	vstoreClient, err := milvus.New(ctx, c.VectorDatabase, logger)
	if err != nil {
		return err
	}

	var llm embedder.LLMClient
	var dim int
	switch c.LLMEngine {
	case llmkind.Ollama:
		llm = ollama.New(c.LLMEngineAddr)
		dim, err = ollama.Dimension(c.Model)
		if err != nil {
			return err
		}
	case llmkind.VLLM:
		llm = vllm.NewClient(c.LLMEngineAddr, logger)
		dim, err = vllm.Dimension(c.Model)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("unsupported llm engine: %s", c.LLMEngine)
	}
	s3Client, err := s3.NewClient(ctx, c.ObjectStore.S3)
	if err != nil {
		return err
	}
	e := embedder.New(llm, s3Client, vstoreClient, logger)

	s := server.New(st, fclient, fwClient, vstoreClient, e, c.Model, dim, logger)

	usage, err := sender.New(ctx, c.UsageSender, grpc.WithTransportCredentials(insecure.NewCredentials()), logger)
	if err != nil {
		return err
	}
	go func() { usage.Run(ctx) }()

	errCh := make(chan error)
	go func() {
		log.Info("Starting HTTP server...", "port", c.HTTPPort)
		errCh <- http.ListenAndServe(fmt.Sprintf(":%d", c.HTTPPort), mux)
	}()

	go func() {
		errCh <- s.Run(ctx, c.GRPCPort, c.AuthConfig, usage)
	}()

	go func() {
		s := server.NewInternal(c.Model, e, logger)
		errCh <- s.Run(c.InternalGRPCPort)
	}()

	return <-errCh
}
