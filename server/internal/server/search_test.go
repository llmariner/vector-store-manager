package server

import (
	"context"
	"fmt"
	"testing"

	"github.com/go-logr/logr/testr"
	v1 "github.com/llmariner/vector-store-manager/api/v1"
	"github.com/stretchr/testify/assert"
)

func TestSearchVectorStore(t *testing.T) {
	tcs := []struct {
		name    string
		req     *v1.SearchVectorStoreRequest
		resp    *v1.SearchVectorStoreResponse
		wantErr bool
	}{
		{
			name: "found",
			req: &v1.SearchVectorStoreRequest{
				VectorStoreId: vectorStoreName,
				Query:         "hi",
			},
			resp: &v1.SearchVectorStoreResponse{
				Documents: []string{
					"hello",
					"hi",
				},
			},
			wantErr: false,
		},
		{
			name: "not found",
			req: &v1.SearchVectorStoreRequest{
				VectorStoreId: vectorStoreName,
				Query:         "unknown",
			},
			resp: &v1.SearchVectorStoreResponse{
				Documents: nil,
			},
			wantErr: false,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {

			srv := NewInternal(
				modelName,
				&noopRetriever{
					collectionName: vectorStoreName,
					docs: map[string][]string{
						"hi": []string{"hello", "hi"},
					},
				},
				testr.New(t),
			)
			ctx := context.Background()
			resp, err := srv.SearchVectorStore(ctx, tc.req)
			assert.NoError(t, err)
			assert.Equal(t, len(tc.resp.Documents), len(resp.Documents))
		})
	}
}

type noopRetriever struct {
	collectionName string
	docs           map[string][]string
}

func (c *noopRetriever) Search(ctx context.Context, collectionName, modelName, query string, numDocuments int) ([]string, error) {
	if collectionName != c.collectionName {
		return nil, fmt.Errorf("collection %s not found", collectionName)
	}
	return c.docs[query], nil
}
