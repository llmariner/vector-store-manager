package server

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strconv"
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

	for _, fid := range req.FileIds {
		if err := s.validateFile(ctx, fid); err != nil {
			return nil, err
		}
	}

	if _, err := s.store.GetCollectionByName(userInfo.ProjectID, req.Name); err == nil {
		return nil, status.Errorf(codes.AlreadyExists, "vector store %q already exists", req.Name)
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, status.Errorf(codes.Internal, "get collection: %s", err)
	}

	cid, err := s.vstoreClient.CreateVectorStore(ctx, req.Name)
	if err != nil {
		return nil, err
	}

	c := &store.Collection{
		CollectionID:   cid,
		Name:           req.Name,
		Status:         store.CollectionStatusInProgress,
		OrganizationID: userInfo.OrganizationID,
		ProjectID:      userInfo.ProjectID,
		TenantID:       userInfo.TenantID,
		LastActiveAt:   time.Now().Unix(),
	}
	if ea := req.ExpiresAfter; ea != nil {
		if err := validateExpiresAfter(ea); err != nil {
			return nil, err
		}
		c.Anchor = store.ExpiresAfterAnchor(ea.Anchor)
		c.ExpiresAfterDays = ea.Days
	}

	if err := s.store.CreateCollection(c); err != nil {
		return nil, status.Errorf(codes.Internal, "create collection: %s", err)
	}

	// TODO(guangrui): Make CreateCollection and CreatCollectionMetadata in a transaction.
	var cms []*store.CollectionMetadata
	for k, v := range req.Metadata {
		cms = append(cms, &store.CollectionMetadata{
			CollectionID: cid,
			Key:          k,
			Value:        v,
		})
	}
	for _, cm := range cms {
		if err := s.store.CreateCollectionMetadata(cm); err != nil {
			return nil, status.Errorf(codes.Internal, "create collection metadata: %s", err)
		}
	}

	c, err = s.store.GetCollectionByCollectionID(userInfo.ProjectID, cid)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "get collection: %s", err)
	}

	return toVectorStoreProto(c, cms), nil
}

func (s *S) validateFile(ctx context.Context, fileID string) error {
	if _, err := s.fileGetClient.GetFile(ctx, &fv1.GetFileRequest{
		Id: fileID,
	}); err != nil {
		if status.Code(err) == codes.NotFound {
			return status.Errorf(codes.InvalidArgument, "file %q not found", fileID)
		}
		return status.Errorf(codes.Internal, "get file: %s", err)
	}
	return nil
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

	afterID := int64(-1)
	if req.After != "" {
		var err error
		afterID, err = getCollectionID(req.After)
		if err != nil {
			return nil, err
		}
	}

	cs, hasMore, err := s.store.ListCollectionsWithPagination(userInfo.ProjectID, afterID, order, int(limit))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "list collections with pagination: %s", err)
	}

	var protos []*v1.VectorStore
	for _, c := range cs {
		cm, err := s.store.ListCollectionMetadataByCollectionID(c.CollectionID)
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

	id, err := getCollectionID(req.Id)
	if err != nil {
		return nil, err
	}

	c, err := s.store.GetCollectionByCollectionID(userInfo.ProjectID, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Errorf(codes.NotFound, "collection %q not found", req.Id)
		}
		return nil, status.Errorf(codes.Internal, "get collection: %s", err)
	}

	cm, err := s.store.ListCollectionMetadataByCollectionID(c.CollectionID)
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

	if err := validateMetadata(req.Metadata); err != nil {
		return nil, err
	}

	id, err := getCollectionID(req.Id)
	if err != nil {
		return nil, err
	}

	c, err := s.store.GetCollectionByCollectionID(userInfo.ProjectID, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Errorf(codes.NotFound, "collection %q not found", req.Id)
		}
		return nil, status.Errorf(codes.Internal, "get collection: %s", err)
	}

	// update collection metadata
	cms, err := s.store.ListCollectionMetadataByCollectionID(c.CollectionID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "list collection metadata: %s", err)
	}
	cur := make(map[string]*store.CollectionMetadata)
	for _, cm := range cms {
		cur[cm.Key] = cm
	}
	// TODO(guangrui): Update collection and collection metadata in a transaction.
	for k, v := range req.Metadata {
		found, ok := cur[k]
		if !ok {
			cm := &store.CollectionMetadata{
				CollectionID: c.CollectionID,
				Key:          k,
				Value:        v,
			}
			if err := s.store.CreateCollectionMetadata(cm); err != nil {
				return nil, status.Errorf(codes.Internal, "create collection metadata: %s", err)
			}
		} else if found.Value != v {
			found.Value = v
			if err := s.store.UpdateCollectionMetadata(found); err != nil {
				return nil, status.Errorf(codes.Internal, "update collection metadata: %s", err)
			}
		}
	}
	for _, cm := range cms {
		if _, ok := req.Metadata[cm.Key]; !ok {
			if err := s.store.DeleteCollectionMetadata(cm.CollectionID, cm.Key); err != nil {
				return nil, status.Errorf(codes.Internal, "delete collection metadata: %s", err)
			}
		}
	}
	cms, err = s.store.ListCollectionMetadataByCollectionID(c.CollectionID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "list collection metadata: %s", err)
	}

	// update collection
	if req.Name != "" && req.Name != c.Name {
		if err := s.vstoreClient.UpdateVectorStoreName(ctx, c.Name, req.Name); err != nil {
			return nil, status.Errorf(codes.Internal, "update vector store: %s", err)
		}
		c.Name = req.Name
	}
	// TODO(guangrui): support clearing ExpiresAfter. Considering to use 'google.protobuf.Int32Value' to differentiate
	// between clearing and unsetting.
	if ea := req.ExpiresAfter; ea != nil {
		if err := validateExpiresAfter(ea); err != nil {
			return nil, err
		}
		c.Anchor = store.ExpiresAfterAnchor(ea.Anchor)
		c.ExpiresAfterDays = ea.Days
	}
	if err := s.store.UpdateCollection(c); err != nil {
		return nil, status.Errorf(codes.Internal, "update collection: %s", err)
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

	id, err := getCollectionID(req.Id)
	if err != nil {
		return nil, err
	}

	if err := s.vstoreClient.DeleteVectorStore(ctx, req.Id); err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Errorf(codes.Internal, "delete collection: %s", err)
		}
		log.Printf("Collection %q not found in vector store server.", req.Id)
	}

	// TODO(guangrui): Delete collection and collection metadata in a transaction.
	if err := s.store.DeleteCollection(userInfo.ProjectID, id); err != nil {
		return nil, status.Errorf(codes.Internal, "delete collection: %s", err)
	}
	if err := s.store.DeleteCollectionMetadataByCollectionID(id); err != nil {
		return nil, status.Errorf(codes.Internal, "delete collection metadata: %s", err)
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
		Id:         fmt.Sprintf("%d", c.CollectionID),
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

func getCollectionID(vid string) (int64, error) {
	id, err := strconv.ParseInt(vid, 10, 64)
	if err != nil {
		return 0, status.Errorf(codes.Internal, "get collection id %s", err)
	}
	return id, nil
}
