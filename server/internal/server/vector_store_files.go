package server

import (
	"context"
	"errors"
	"log"
	"strings"
	"time"

	fv1 "github.com/llm-operator/file-manager/api/v1"
	"github.com/llm-operator/rbac-manager/pkg/auth"
	v1 "github.com/llm-operator/vector-store-manager/api/v1"
	"github.com/llm-operator/vector-store-manager/server/internal/store"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

const (
	vectorStoreFileObject = "vector_store_file"

	minMaxChunkSizeTokens     = int64(100)
	maxMaxChunkSizeTokens     = int64(4096)
	defaultMaxChunkSizeTokens = int64(800)
	defaultChunkOverlapTokens = int64(400)
)

// CreateVectorStoreFile adds a new file to the vector store.
func (s *S) CreateVectorStoreFile(
	ctx context.Context,
	req *v1.CreateVectorStoreFileRequest,
) (*v1.VectorStoreFile, error) {
	userInfo, err := s.extractUserInfoFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if req.VectorStoreId == "" {
		return nil, status.Error(codes.InvalidArgument, "vector store id is required")
	}
	if req.FileId == "" {
		return nil, status.Error(codes.InvalidArgument, "file id is required")
	}

	maxChunkSizeTokens := defaultMaxChunkSizeTokens
	chunkOverlapTokens := defaultChunkOverlapTokens
	if cs := req.ChunkingStrategy; cs != nil {
		if err := validateChunkingStrategy(cs); err != nil {
			return nil, err
		}
		if cs.Static != nil {
			maxChunkSizeTokens = cs.Static.MaxChunkSizeTokens
			chunkOverlapTokens = cs.Static.ChunkOverlapTokens
		}
	} else {
		// TODO(kenji): The OpenAI API reference says the strategy is always "static".
		// https://platform.openai.com/docs/api-reference/vector-stores-files/file-object
		//
		// We might need to do the conversion.
		req.ChunkingStrategy = &v1.ChunkingStrategy{
			Type: string(store.ChunkingStrategyTypeAuto),
		}
	}

	if err := s.validateFile(auth.CarryMetadata(ctx), req.FileId); err != nil {
		return nil, err
	}

	resp, err := s.fileInternalClient.GetFilePath(ctx, &fv1.GetFilePathRequest{
		Id: req.FileId,
	})
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, status.Errorf(codes.NotFound, "file %q not found", req.FileId)
		}
		return nil, status.Errorf(codes.Internal, "get file path: %s", err)
	}

	c, err := s.store.GetCollectionByVectorStoreID(userInfo.ProjectID, req.VectorStoreId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Errorf(codes.NotFound, "vector store %q not found", req.VectorStoreId)
		}
		return nil, status.Errorf(codes.Internal, "get collection: %s", err)
	}

	log.Printf("Added file %q to vector store %q.\n", req.FileId, req.VectorStoreId)
	if err := s.embedder.AddFile(ctx, c.Name, c.EmbeddingModel, resp.Path, maxChunkSizeTokens, chunkOverlapTokens); err != nil {
		return nil, status.Errorf(codes.Internal, "add file: %s", err)
	}
	f := &store.File{
		FileID:               req.FileId,
		VectorStoreID:        req.VectorStoreId,
		UsageBytes:           0,
		Status:               store.FileStatusInProgress,
		ChunkingStrategyType: store.ChunkingStrategyType(req.ChunkingStrategy.Type),
		MaxChunkSizeTokens:   maxChunkSizeTokens,
		ChunkOverlapTokens:   chunkOverlapTokens,
	}
	if err := s.store.CreateFile(f); err != nil {
		return nil, status.Errorf(codes.Internal, "create file: %s", err)
	}
	log.Printf("Added file %q to vector store %q", f.FileID, f.VectorStoreID)

	return toVectorStoreFileProto(f), nil
}

func validateChunkingStrategy(cs *v1.ChunkingStrategy) error {
	if cs.Type != string(store.ChunkingStrategyTypeAuto) && cs.Type != string(store.ChunkingStrategyTypeStatic) {
		return status.Errorf(codes.InvalidArgument, "chunking strategy type must be either auto or static")
	}
	if cs.Static == nil {
		return nil
	}
	if cs.Static.MaxChunkSizeTokens < minMaxChunkSizeTokens {
		return status.Errorf(codes.InvalidArgument, "chunk size tokens must be no less than %d", minMaxChunkSizeTokens)
	}
	if cs.Static.MaxChunkSizeTokens > maxMaxChunkSizeTokens {
		return status.Errorf(codes.InvalidArgument, "chunk size tokens must be no more than %d", maxMaxChunkSizeTokens)
	}
	if cs.Static.ChunkOverlapTokens <= 0 {
		return status.Errorf(codes.InvalidArgument, "chunk overlap tokens must be greater than 0")
	}
	if cs.Static.ChunkOverlapTokens > (cs.Static.MaxChunkSizeTokens / 2) {
		return status.Errorf(codes.InvalidArgument, "chunk overlap tokens must be no more than %d", cs.Static.MaxChunkSizeTokens/2)
	}
	return nil
}

// GetVectorStoreFile gets a file from the vector store.
func (s *S) GetVectorStoreFile(
	ctx context.Context,
	req *v1.GetVectorStoreFileRequest,
) (*v1.VectorStoreFile, error) {
	userInfo, err := s.extractUserInfoFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if req.VectorStoreId == "" {
		return nil, status.Error(codes.InvalidArgument, "vector store id is required")
	}
	if req.FileId == "" {
		return nil, status.Error(codes.InvalidArgument, "file id is required")
	}

	if err := s.validateVectorStore(req.VectorStoreId, userInfo.ProjectID); err != nil {
		return nil, err
	}

	f, err := s.store.GetFileByFileID(req.VectorStoreId, req.FileId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Errorf(codes.NotFound, "file %q not found in vector store %q", req.FileId, req.VectorStoreId)
		}
		return nil, status.Errorf(codes.Internal, "get file: %s", err)
	}
	return toVectorStoreFileProto(f), nil
}

// ListVectorStoreFiles lists files in the vector store.
func (s *S) ListVectorStoreFiles(
	ctx context.Context,
	req *v1.ListVectorStoreFilesRequest,
) (*v1.ListVectorStoreFilesResponse, error) {
	userInfo, err := s.extractUserInfoFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if req.VectorStoreId == "" {
		return nil, status.Error(codes.InvalidArgument, "vector store id is required")
	}
	if req.Limit < 0 {
		return nil, status.Errorf(codes.InvalidArgument, "limit must be non-negative")
	}

	if err := s.validateVectorStore(req.VectorStoreId, userInfo.ProjectID); err != nil {
		return nil, err
	}

	limit := req.Limit
	if limit == 0 {
		limit = defaultPageSize
	}
	if limit > maxPageSize {
		limit = maxPageSize
	}

	order := strings.ToLower(req.Order)
	if order != "" && order != "asc" && order != "desc" {
		return nil, status.Errorf(codes.InvalidArgument, "order must be one of 'asc' or 'desc'")
	}

	var afterCreatedAt time.Time
	var afterID uint
	if fid := req.After; fid != "" {
		f, err := s.store.GetFileByFileID(req.VectorStoreId, fid)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, status.Errorf(codes.NotFound, "invalid value of after: %q", fid)
			}
			return nil, status.Errorf(codes.Internal, "get file: %s", err)
		}
		afterCreatedAt = f.CreatedAt
		afterID = f.ID
	}

	fs, hasMore, err := s.store.ListFilesWithPagination(req.VectorStoreId, afterCreatedAt, afterID, order, int(limit))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "list files with pagination: %s", err)
	}

	var protos []*v1.VectorStoreFile
	for _, f := range fs {
		protos = append(protos, toVectorStoreFileProto(f))
	}
	first := ""
	last := ""
	if len(protos) > 0 {
		first = protos[0].Id
		last = protos[len(protos)-1].Id
	}
	return &v1.ListVectorStoreFilesResponse{
		Object:  vectorStoreFileObject,
		Data:    protos,
		FirstId: first,
		LastId:  last,
		HasMore: hasMore,
	}, nil
}

// DeleteVectorStoreFile deletes a file from the vector store.
func (s *S) DeleteVectorStoreFile(
	ctx context.Context,
	req *v1.DeleteVectorStoreFileRequest,
) (*v1.DeleteVectorStoreFileResponse, error) {
	userInfo, err := s.extractUserInfoFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if req.VectorStoreId == "" {
		return nil, status.Error(codes.InvalidArgument, "vector store id is required")
	}
	if req.FileId == "" {
		return nil, status.Error(codes.InvalidArgument, "file id is required")
	}

	if err := s.validateVectorStore(req.VectorStoreId, userInfo.ProjectID); err != nil {
		return nil, err
	}

	if err := s.store.DeleteFile(req.VectorStoreId, req.FileId); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Errorf(codes.NotFound, "file %q not found in vector store %q", req.FileId, req.VectorStoreId)
		}
		return nil, status.Errorf(codes.Internal, "delete file: %s", err)
	}
	return &v1.DeleteVectorStoreFileResponse{
		Id:      req.FileId,
		Object:  vectorStoreFileObject,
		Deleted: true,
	}, nil
}

func toVectorStoreFileProto(f *store.File) *v1.VectorStoreFile {
	proto := &v1.VectorStoreFile{
		Id:            f.FileID,
		Object:        vectorStoreFileObject,
		UsageBytes:    f.UsageBytes,
		CreatedAt:     f.CreatedAt.Unix(),
		VectorStoreId: f.VectorStoreID,
		Status:        string(f.Status),
		ChunkingStrategy: &v1.ChunkingStrategy{
			Type: string(f.ChunkingStrategyType),
		},
	}
	if f.LastErrorCode != store.LastErrorCodeNone {
		proto.LastError = &v1.VectorStoreFile_Error{
			Code:    string(f.LastErrorCode),
			Message: f.LastErrorMessage,
		}
	}
	if f.ChunkingStrategyType == store.ChunkingStrategyTypeStatic {
		proto.ChunkingStrategy.Static = &v1.ChunkingStrategy_Static{
			MaxChunkSizeTokens: f.MaxChunkSizeTokens,
			ChunkOverlapTokens: f.ChunkOverlapTokens,
		}
	}
	return proto
}
