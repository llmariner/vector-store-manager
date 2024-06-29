package server

import (
	"context"

	v1 "github.com/llm-operator/vector-store-manager/api/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	defaultNumDocuments = 10
	maxNumDocuments     = 100
)

// SearchVectorStore searches documents for the given query from a vector store.
func (s *IS) SearchVectorStore(
	ctx context.Context,
	req *v1.SearchVectorStoreRequest,
) (*v1.SearchVectorStoreResponse, error) {
	if req.VectorStoreId == "" {
		return nil, status.Error(codes.InvalidArgument, "vector_store_id is required")
	}

	if req.Query == "" {
		return nil, status.Error(codes.InvalidArgument, "query is required")
	}

	if req.NumDocuments < 0 {
		return nil, status.Errorf(codes.InvalidArgument, "num_documents must be non-negative")
	}

	numDocs := int(req.NumDocuments)
	if numDocs == 0 {
		numDocs = defaultNumDocuments
	}
	if numDocs > maxNumDocuments {
		numDocs = maxNumDocuments
	}

	docs, err := s.retriever.Search(ctx, req.VectorStoreId, s.model, req.Query, numDocs)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "search vector store: %s", err)
	}
	return &v1.SearchVectorStoreResponse{
		Documents: docs,
	}, nil
}
