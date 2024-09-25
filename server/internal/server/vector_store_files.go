package server

import (
	"context"
	"errors"
	"log"
	"strings"
	"time"

	fv1 "github.com/llm-operator/file-manager/api/v1"
	"github.com/llmariner/rbac-manager/pkg/auth"
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

type chunkingStrategy struct {
	maxChunkSizeTokens   int64
	chunkOverlapTokens   int64
	chunkingStrategyType store.ChunkingStrategyType
}

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

	cs, err := getChunkingStrategy(req.ChunkingStrategy)
	if err != nil {
		return nil, err
	}

	c, err := s.store.GetCollectionByVectorStoreID(userInfo.ProjectID, req.VectorStoreId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Errorf(codes.NotFound, "vector store %q not found", req.VectorStoreId)
		}
		return nil, status.Errorf(codes.Internal, "get collection: %s", err)
	}

	file, err := s.validateFile(auth.CarryMetadata(ctx), req.FileId)
	if err != nil {
		return nil, err
	}
	f, err := s.createVectorStoreFile(ctx, c, file, cs)
	if err != nil {
		return nil, err
	}

	c, err = s.store.GetCollectionByVectorStoreID(userInfo.ProjectID, c.VectorStoreID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "get collection: %s", err)
	}
	c.FileCountsCompleted++
	c.FileCountsTotal++
	if err := s.store.UpdateCollection(c); err != nil {
		return nil, status.Errorf(codes.Internal, "update collection: %s", err)
	}
	return toVectorStoreFileProto(f), nil
}

func (s *S) createVectorStoreFile(ctx context.Context, c *store.Collection, f *fv1.File, cs *chunkingStrategy) (*store.File, error) {
	if _, err := s.store.GetFileByFileID(c.VectorStoreID, f.Id); err == nil {
		return nil, status.Errorf(codes.AlreadyExists, "file %q already exists in vector store %q", f.Id, c.VectorStoreID)
	}

	resp, err := s.fileInternalClient.GetFilePath(ctx, &fv1.GetFilePathRequest{Id: f.Id})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "get file path: %s", err)
	}

	log.Printf("Adding file %q to vector store %q.\n", f.Id, c.VectorStoreID)
	if err := s.embedder.AddFile(
		ctx,
		c.VectorStoreID,
		c.EmbeddingModel,
		f.Id,
		f.Filename,
		resp.Path,
		cs.maxChunkSizeTokens,
		cs.chunkOverlapTokens,
	); err != nil {
		return nil, status.Errorf(codes.Internal, "add file: %s", err)
	}
	file := &store.File{
		FileID:               f.Id,
		VectorStoreID:        c.VectorStoreID,
		UsageBytes:           0,
		Status:               store.FileStatusCompleted,
		ChunkingStrategyType: cs.chunkingStrategyType,
		MaxChunkSizeTokens:   cs.maxChunkSizeTokens,
		ChunkOverlapTokens:   cs.chunkOverlapTokens,
	}
	if err := s.store.CreateFile(file); err != nil {
		return nil, status.Errorf(codes.Internal, "create file: %s", err)
	}
	log.Printf("Added file %q to vector store %q", file.FileID, file.VectorStoreID)
	return file, nil
}

func getChunkingStrategy(cs *v1.ChunkingStrategy) (*chunkingStrategy, error) {
	ret := &chunkingStrategy{
		maxChunkSizeTokens:   defaultMaxChunkSizeTokens,
		chunkOverlapTokens:   defaultChunkOverlapTokens,
		chunkingStrategyType: store.ChunkingStrategyTypeStatic,
	}
	if cs != nil {
		if err := validateChunkingStrategy(cs); err != nil {
			return nil, err
		}
		if cs.Static != nil {
			ret.maxChunkSizeTokens = cs.Static.MaxChunkSizeTokens
			ret.chunkOverlapTokens = cs.Static.ChunkOverlapTokens
		}
	}
	return ret, nil
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

	// TODO(guangrui): Gracefully handle the deletion error.
	if err := s.embedder.DeleteFile(ctx, req.VectorStoreId, req.FileId); err != nil {
		// milvus does not return error if the file does not exist.
		return nil, status.Errorf(codes.Internal, "embedder delete file: %s", err)
	}

	if err := s.store.DeleteFile(req.VectorStoreId, req.FileId); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Errorf(codes.NotFound, "file %q not found in vector store %q", req.FileId, req.VectorStoreId)
		}
		return nil, status.Errorf(codes.Internal, "delete file: %s", err)
	}

	c, err := s.store.GetCollectionByVectorStoreID(userInfo.ProjectID, req.VectorStoreId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "get collection: %s", err)
	}
	c.FileCountsCompleted--
	c.FileCountsTotal--
	if err := s.store.UpdateCollection(c); err != nil {
		return nil, status.Errorf(codes.Internal, "update collection: %s", err)
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
