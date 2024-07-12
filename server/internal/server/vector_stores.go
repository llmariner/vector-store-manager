package server

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/llm-operator/common/pkg/id"
	fv1 "github.com/llm-operator/file-manager/api/v1"
	"github.com/llm-operator/rbac-manager/pkg/auth"
	v1 "github.com/llm-operator/vector-store-manager/api/v1"
	"github.com/llm-operator/vector-store-manager/server/internal/store"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

const (
	vectorStoreObject      = "vector_store"
	maxMetadataEntries     = 16
	maxMetadataKeyLength   = 64
	maxMetadataValueLength = 512
)

// CreateVectorStore creates a new vector store.
func (s *S) CreateVectorStore(
	ctx context.Context,
	req *v1.CreateVectorStoreRequest,
) (*v1.VectorStore, error) {
	userInfo, err := s.extractUserInfoFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if req.Name == "" {
		return nil, status.Errorf(codes.InvalidArgument, "name is required")
	}

	if err := validateMetadata(req.Metadata); err != nil {
		return nil, err
	}

	// Pass the Authorization to the context for downstream gRPC calls.
	ctx = auth.CarryMetadata(ctx)

	var fs []*fv1.File
	for _, fid := range req.FileIds {
		f, err := s.validateFile(ctx, fid)
		if err != nil {
			return nil, err
		}
		fs = append(fs, f)
	}

	if _, err := s.store.GetCollectionByName(userInfo.ProjectID, req.Name); err == nil {
		return nil, status.Errorf(codes.AlreadyExists, "vector store %q already exists", req.Name)
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, status.Errorf(codes.Internal, "get collection: %s", err)
	}

	// vector store ID is not a k8s resource, but the ID is used as a Milivus collection name,
	// which can only contain numbers, letters and underscores.
	vsID, err := id.GenerateIDForK8SResource("vs_")
	if err != nil {
		return nil, status.Errorf(codes.Internal, "generate id: %s", err)
	}

	cid, err := s.vstoreClient.CreateVectorStore(ctx, vsID, s.dimensions)
	if err != nil {
		return nil, err
	}

	// TODO(kenji): If the RPC fails after this point, a dangling Milvus collection will be left behind.
	// We need some background cleaning processing.

	c := &store.Collection{
		VectorStoreID:       vsID,
		CollectionID:        cid,
		Name:                req.Name,
		Status:              store.CollectionStatusCompleted,
		OrganizationID:      userInfo.OrganizationID,
		ProjectID:           userInfo.ProjectID,
		TenantID:            userInfo.TenantID,
		LastActiveAt:        time.Now().Unix(),
		EmbeddingModel:      s.model,
		EmbeddingDimensions: s.dimensions,
	}
	if ea := req.ExpiresAfter; ea != nil {
		if err := validateExpiresAfter(ea); err != nil {
			return nil, err
		}
		c.Anchor = store.ExpiresAfterAnchor(ea.Anchor)
		c.ExpiresAfterDays = ea.Days
	}

	var cms []*store.CollectionMetadata
	for k, v := range req.Metadata {
		cms = append(cms, &store.CollectionMetadata{
			VectorStoreID: c.VectorStoreID,
			Key:           k,
			Value:         v,
		})
	}

	if err := s.store.Transaction(func(tx *gorm.DB) error {
		if err := store.CreateCollectionInTransaction(tx, c); err != nil {
			return fmt.Errorf("create collection: %s", err)
		}

		for _, cm := range cms {
			if err := store.CreateCollectionMetadataInTransaction(tx, cm); err != nil {
				return fmt.Errorf("create collection metadata: %s", err)
			}
		}
		return nil
	}); err != nil {
		return nil, status.Errorf(codes.Internal, "transaction: %s", err)
	}

	cs, err := getChunkingStrategy(req.ChunkingStrategy)
	if err != nil {
		return nil, err
	}

	fileCompleted := int64(0)
	var errMsgs []string
	for _, f := range fs {
		if _, err := s.createVectorStoreFile(ctx, c, f, cs); err != nil {
			log.Printf("Failed to add file %q to vector store %q: %s", f.Id, c.VectorStoreID, err)
			errMsgs = append(errMsgs, fmt.Sprintf("file %q: %s", f.Id, err))
			continue
		}
		fileCompleted++
	}

	c, err = s.store.GetCollectionByVectorStoreID(userInfo.ProjectID, c.VectorStoreID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "get collection: %s", err)
	}
	c.FileCountsCompleted += fileCompleted
	c.FileCountsTotal += fileCompleted
	if err := s.store.UpdateCollection(c); err != nil {
		return nil, status.Errorf(codes.Internal, "update collection: %s", err)
	}

	vsProto := toVectorStoreProto(c, cms)
	if len(errMsgs) > 0 {
		return vsProto, status.Errorf(codes.Internal, "create vector store file: %s", strings.Join(errMsgs, ";"))
	}
	return vsProto, nil
}

func (s *S) validateFile(ctx context.Context, fileID string) (*fv1.File, error) {
	f, err := s.fileGetClient.GetFile(ctx, &fv1.GetFileRequest{Id: fileID})
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, status.Errorf(codes.InvalidArgument, "file %q not found", fileID)
		}
		return nil, status.Errorf(codes.Internal, "get file: %s", err)
	}
	return f, nil
}

func validateMetadata(metadata map[string]string) error {
	if len(metadata) > maxMetadataEntries {
		return status.Errorf(codes.InvalidArgument, "No more than %d metadata entries are allowed", maxMetadataEntries)
	}
	for k, v := range metadata {
		if len(k) > maxMetadataKeyLength {
			return status.Errorf(codes.InvalidArgument, "Metadata key %q is too long, max allowed is %d", k, maxMetadataKeyLength)
		}
		if len(v) > maxMetadataValueLength {
			return status.Errorf(codes.InvalidArgument, "Metadata value for key %q is too long, max allowed is %d", k, maxMetadataValueLength)
		}
	}
	return nil
}

func validateExpiresAfter(ea *v1.ExpiresAfter) error {
	if ea.Anchor != "" && ea.Anchor != string(store.ExpiresAfterAnchorLastActiveAt) {
		return status.Errorf(codes.InvalidArgument, "expires_after.anchor must be %q", store.ExpiresAfterAnchorLastActiveAt)
	}
	if ea.Days <= 0 {
		return status.Errorf(codes.InvalidArgument, "expires_after.days must be greater than 0")
	}
	return nil
}

// ListVectorStores lists all vector stores.
func (s *S) ListVectorStores(
	ctx context.Context,
	req *v1.ListVectorStoresRequest,
) (*v1.ListVectorStoresResponse, error) {
	userInfo, err := s.extractUserInfoFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if req.Limit < 0 {
		return nil, status.Errorf(codes.InvalidArgument, "limit must be non-negative")
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
	if vid := req.After; vid != "" {
		c, err := s.store.GetCollectionByVectorStoreID(userInfo.ProjectID, vid)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, status.Errorf(codes.NotFound, "invalid value of after: %q", vid)
			}
			return nil, status.Errorf(codes.Internal, "get collection: %s", err)
		}
		afterCreatedAt = c.CreatedAt
		afterID = c.ID
	}

	cs, hasMore, err := s.store.ListCollectionsWithPagination(userInfo.ProjectID, afterCreatedAt, afterID, order, int(limit))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "list collections with pagination: %s", err)
	}

	var protos []*v1.VectorStore
	for _, c := range cs {
		cm, err := s.store.ListCollectionMetadataByVectorStoreID(c.VectorStoreID)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "list collection metadata: %s", err)
		}
		protos = append(protos, toVectorStoreProto(c, cm))
	}
	first := ""
	last := ""
	if len(protos) > 0 {
		first = protos[0].Id
		last = protos[len(protos)-1].Id
	}
	return &v1.ListVectorStoresResponse{
		Object:  vectorStoreObject,
		Data:    protos,
		FirstId: first,
		LastId:  last,
		HasMore: hasMore,
	}, nil
}

// GetVectorStore gets a vector store.
func (s *S) GetVectorStore(
	ctx context.Context,
	req *v1.GetVectorStoreRequest,
) (*v1.VectorStore, error) {
	userInfo, err := s.extractUserInfoFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	c, err := s.store.GetCollectionByVectorStoreID(userInfo.ProjectID, req.Id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Errorf(codes.NotFound, "collection %q not found", req.Id)
		}
		return nil, status.Errorf(codes.Internal, "get collection: %s", err)
	}

	cm, err := s.store.ListCollectionMetadataByVectorStoreID(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "list collection metadata: %s", err)
	}

	return toVectorStoreProto(c, cm), nil
}

// GetVectorStoreByName gets a vector store by its name.
func (s *S) GetVectorStoreByName(
	ctx context.Context,
	req *v1.GetVectorStoreByNameRequest,
) (*v1.VectorStore, error) {
	userInfo, err := s.extractUserInfoFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if req.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "name is required")
	}

	c, err := s.store.GetCollectionByName(userInfo.ProjectID, req.Name)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Errorf(codes.NotFound, "collection %q not found", req.Name)
		}
		return nil, status.Errorf(codes.Internal, "get collection: %s", err)
	}

	cm, err := s.store.ListCollectionMetadataByVectorStoreID(c.VectorStoreID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "list collection metadata: %s", err)
	}

	return toVectorStoreProto(c, cm), nil
}

// UpdateVectorStore updates a vector store.
func (s *S) UpdateVectorStore(
	ctx context.Context,
	req *v1.UpdateVectorStoreRequest,
) (*v1.VectorStore, error) {
	userInfo, err := s.extractUserInfoFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	if ea := req.ExpiresAfter; ea != nil {
		if err := validateExpiresAfter(ea); err != nil {
			return nil, err
		}
	}

	if err := validateMetadata(req.Metadata); err != nil {
		return nil, err
	}

	c, err := s.store.GetCollectionByVectorStoreID(userInfo.ProjectID, req.Id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Errorf(codes.NotFound, "collection %q not found", req.Id)
		}
		return nil, status.Errorf(codes.Internal, "get collection: %s", err)
	}

	// update collection metadata
	cms, err := s.store.ListCollectionMetadataByVectorStoreID(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "list collection metadata: %s", err)
	}
	cur := make(map[string]*store.CollectionMetadata)
	for _, cm := range cms {
		cur[cm.Key] = cm
	}

	if err := s.store.Transaction(func(tx *gorm.DB) error {
		for k, v := range req.Metadata {
			found, ok := cur[k]
			if !ok {
				cm := &store.CollectionMetadata{
					VectorStoreID: req.Id,
					Key:           k,
					Value:         v,
				}
				if err := store.CreateCollectionMetadataInTransaction(tx, cm); err != nil {
					return fmt.Errorf("create collection metadata: %s", err)
				}
			} else if found.Value != v {
				found.Value = v
				if err := store.UpdateCollectionMetadataInTransaction(tx, found); err != nil {
					return fmt.Errorf("update collection metadata: %s", err)
				}
			}
		}
		for _, cm := range cms {
			if _, ok := req.Metadata[cm.Key]; !ok {
				if err := store.DeleteCollectionMetadataInTransaction(tx, cm.ID); err != nil {
					return fmt.Errorf("delete collection metadata: %s", err)
				}
			}
		}

		// Update collection.

		if req.Name != "" {
			c.Name = req.Name
		}
		// TODO(guangrui): support clearing ExpiresAfter. Considering to use 'google.protobuf.Int32Value' to differentiate
		// between clearing and unsetting.
		if ea := req.ExpiresAfter; ea != nil {
			c.Anchor = store.ExpiresAfterAnchor(ea.Anchor)
			c.ExpiresAfterDays = ea.Days
		}
		if err := store.UpdateCollectionInTransaction(tx, c); err != nil {
			return fmt.Errorf("update collection: %s", err)
		}
		return nil
	}); err != nil {
		return nil, status.Errorf(codes.Internal, "transaction: %s", err)
	}

	cms, err = s.store.ListCollectionMetadataByVectorStoreID(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "list collection metadata: %s", err)
	}

	return toVectorStoreProto(c, cms), nil
}

// DeleteVectorStore deletes a vector store.
func (s *S) DeleteVectorStore(
	ctx context.Context,
	req *v1.DeleteVectorStoreRequest,
) (*v1.DeleteVectorStoreResponse, error) {
	userInfo, err := s.extractUserInfoFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	if err := s.validateVectorStore(req.Id, userInfo.ProjectID); err != nil {
		return nil, err
	}

	if err := s.store.Transaction(func(tx *gorm.DB) error {
		if err := store.DeleteCollectionInTransaction(tx, userInfo.ProjectID, req.Id); err != nil {
			return fmt.Errorf("delete collection: %s", err)
		}
		if err := store.DeleteAllCollectionMetadatasByVectorStoreIDInTransaction(tx, req.Id); err != nil {
			return fmt.Errorf("delete collection metadatas: %s", err)
		}
		if err := store.DeleteAllFilesByVectorStoreIDInTransaction(tx, req.Id); err != nil {
			return fmt.Errorf("delete files: %s", err)
		}
		return nil
	}); err != nil {
		return nil, status.Errorf(codes.Internal, "transaction: %s", err)
	}

	// TODO(kenji): If the RPC fails after this point, a dangling Milvus collection will be left behind.
	// We need some background cleaning processing.

	if err := s.vstoreClient.DeleteVectorStore(ctx, req.Id); err != nil {
		return nil, status.Errorf(codes.Internal, "delete collection: %s", err)
	}

	return &v1.DeleteVectorStoreResponse{
		Id:      req.Id,
		Object:  vectorStoreObject,
		Deleted: true,
	}, nil
}

func toVectorStoreProto(c *store.Collection, cms []*store.CollectionMetadata) *v1.VectorStore {
	m := map[string]string{}
	for _, cm := range cms {
		m[cm.Key] = cm.Value
	}

	return &v1.VectorStore{
		Id:         c.VectorStoreID,
		Object:     vectorStoreObject,
		CreatedAt:  c.CreatedAt.Unix(),
		Name:       c.Name,
		UsageBytes: c.UsageBytes,
		FileCounts: &v1.VectorStore_FileCounts{
			InProgress: c.FileCountsInProgress,
			Completed:  c.FileCountsCompleted,
			Failed:     c.FileCountsFailed,
			Cancelled:  c.FileCountsCancelled,
			Total:      c.FileCountsTotal,
		},
		Status: string(c.Status),
		ExpiresAfter: &v1.ExpiresAfter{
			Anchor: string(c.Anchor),
			Days:   c.ExpiresAfterDays,
		},
		ExpiresAt:    c.ExpiresAt,
		LastActiveAt: c.LastActiveAt,
		Metadata:     m,
	}
}

// validateVectorStore checks if the specified vector is visible to the user.
func (s *S) validateVectorStore(vectorStoreID, projectID string) error {
	if _, err := s.store.GetCollectionByVectorStoreID(projectID, vectorStoreID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return status.Errorf(codes.NotFound, "collection %q not found", vectorStoreID)
		}
		return status.Errorf(codes.Internal, "get collection: %s", err)
	}
	return nil
}
