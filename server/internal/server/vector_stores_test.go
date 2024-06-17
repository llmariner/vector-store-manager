package server

import (
	"context"
	"fmt"
	"testing"

	fv1 "github.com/llm-operator/file-manager/api/v1"
	v1 "github.com/llm-operator/vector-store-manager/api/v1"
	"github.com/llm-operator/vector-store-manager/server/internal/store"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func TestCreateVectorStore(t *testing.T) {
	const (
		fileID          = "file0"
		vectorStoreName = "vector_store_1"
		vectorStoreID   = int64(1)
	)

	tcs := []struct {
		name    string
		req     *v1.CreateVectorStoreRequest
		wantErr bool
	}{
		{
			name: "success",
			req: &v1.CreateVectorStoreRequest{
				Name: vectorStoreName,
			},
			wantErr: false,
		},
		{
			name: "success with files",
			req: &v1.CreateVectorStoreRequest{
				FileIds: []string{fileID},
				Name:    vectorStoreName,
			},
			wantErr: false,
		},
		{
			name: "invalid file",
			req: &v1.CreateVectorStoreRequest{
				FileIds: []string{"unknown file"},
				Name:    vectorStoreName,
			},
			wantErr: true,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			st, tearDown := store.NewTest(t)
			defer tearDown()

			srv := New(
				st,
				&noopFileGetClient{
					ids: map[string]bool{
						fileID: true,
					},
				},
				&noopVStoreClient{
					vs: map[string]int64{
						vectorStoreName: vectorStoreID,
					},
				},
			)
			ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("Authorization", "dummy"))
			resp, err := srv.CreateVectorStore(ctx, tc.req)
			if tc.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, fmt.Sprintf("%d", vectorStoreID), resp.Id)
			assert.Equal(t, vectorStoreName, resp.Name)
		})
	}
}

func TestListVectorStores(t *testing.T) {
	const (
		fileID          = "file0"
		vectorStoreName = "vector_store_1"
		vectorStoreID   = int64(1)
	)

	vs := map[string]int64{
		vectorStoreName:  vectorStoreID,
		"vector_store_2": int64(2),
		"vector_store_3": int64(3),
	}

	tcs := []struct {
		name string
		req  *v1.ListVectorStoresRequest
		resp *v1.ListVectorStoresResponse
	}{
		{
			name: "empty body",
			req:  &v1.ListVectorStoresRequest{},
			resp: &v1.ListVectorStoresResponse{
				Object:  vectorStoreName,
				FirstId: "3",
				LastId:  "1",
				HasMore: false,
				Data: []*v1.VectorStore{
					{
						Id:   "3",
						Name: "vector_store_3",
					},
					{
						Id:   "2",
						Name: "vector_store_2",
					},
					{
						Id:   "1",
						Name: "vector_store_1",
					},
				},
			},
		},
		{
			name: "success with after",
			req: &v1.ListVectorStoresRequest{
				Limit: 1,
				After: "3",
			},
			resp: &v1.ListVectorStoresResponse{
				Object:  vectorStoreName,
				FirstId: "2",
				LastId:  "2",
				HasMore: true,
				Data: []*v1.VectorStore{
					{
						Id:   "2",
						Name: "vector_store_2",
					},
				},
			},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			st, tearDown := store.NewTest(t)
			defer tearDown()

			srv := New(
				st,
				&noopFileGetClient{
					ids: map[string]bool{
						fileID: true,
					},
				},
				&noopVStoreClient{
					vs: vs,
				},
			)
			ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("Authorization", "dummy"))
			for name, id := range vs {
				resp, err := srv.CreateVectorStore(ctx, &v1.CreateVectorStoreRequest{
					Name: name,
				})
				assert.NoError(t, err)
				assert.Equal(t, name, resp.Name)
				assert.Equal(t, fmt.Sprintf("%d", id), resp.Id)
			}

			respList, err := srv.ListVectorStores(ctx, tc.req)
			assert.NoError(t, err)
			assert.Equal(t, len(tc.resp.Data), len(respList.Data))
			assert.Equal(t, tc.resp.FirstId, respList.FirstId)
			assert.Equal(t, tc.resp.LastId, respList.LastId)
			assert.Equal(t, tc.resp.HasMore, respList.HasMore)
		})
	}
}

func TestGetVectorStores(t *testing.T) {
	vs := map[string]int64{
		"vector_store_1": int64(1),
		"vector_store_2": int64(2),
	}

	tcs := []struct {
		name    string
		req     *v1.GetVectorStoreRequest
		resp    *v1.VectorStore
		wantErr bool
	}{
		{
			name: "success",
			req: &v1.GetVectorStoreRequest{
				Id: "1",
			},
			resp: &v1.VectorStore{
				Id:   "1",
				Name: "vector_store_1",
			},
			wantErr: false,
		},
		{
			name: "not found",
			req: &v1.GetVectorStoreRequest{
				Id: "10",
			},
			wantErr: true,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			st, tearDown := store.NewTest(t)
			defer tearDown()

			srv := New(
				st,
				&noopFileGetClient{},
				&noopVStoreClient{
					vs: vs,
				},
			)
			ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("Authorization", "dummy"))
			for name, id := range vs {
				resp, err := srv.CreateVectorStore(ctx, &v1.CreateVectorStoreRequest{
					Name: name,
				})
				assert.NoError(t, err)
				assert.Equal(t, name, resp.Name)
				assert.Equal(t, fmt.Sprintf("%d", id), resp.Id)
			}

			got, err := srv.GetVectorStore(ctx, tc.req)
			if tc.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tc.resp.Name, got.Name)
		})
	}
}

func TestDeleteVectorStore(t *testing.T) {
	const (
		fileID          = "file0"
		vectorStoreName = "vector_store_1"
		vectorStoreID   = int64(1)
	)

	tcs := []struct {
		name    string
		req     *v1.DeleteVectorStoreRequest
		resp    *v1.DeleteVectorStoreResponse
		wantErr bool
	}{
		{
			name: "success",
			req: &v1.DeleteVectorStoreRequest{
				Id: "1",
			},
			resp: &v1.DeleteVectorStoreResponse{
				Id:      "1",
				Object:  vectorStoreName,
				Deleted: true,
			},
			wantErr: false,
		},
		{
			name: "not found",
			req: &v1.DeleteVectorStoreRequest{
				Id: "2",
			},
			resp:    &v1.DeleteVectorStoreResponse{},
			wantErr: true,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			st, tearDown := store.NewTest(t)
			defer tearDown()

			srv := New(
				st,
				&noopFileGetClient{
					ids: map[string]bool{
						fileID: true,
					},
				},
				&noopVStoreClient{
					vs: map[string]int64{
						vectorStoreName: vectorStoreID,
					},
				},
			)
			ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("Authorization", "dummy"))
			resp, err := srv.CreateVectorStore(ctx, &v1.CreateVectorStoreRequest{
				Name: vectorStoreName,
			})
			assert.NoError(t, err)
			assert.Equal(t, vectorStoreName, resp.Name)

			respDelete, err := srv.DeleteVectorStore(ctx, tc.req)
			if tc.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tc.resp.Id, respDelete.Id)
			assert.Equal(t, tc.resp.Deleted, respDelete.Deleted)
		})
	}
}

func TestUpdateVectorStore(t *testing.T) {
	const (
		fileID          = "file0"
		vectorStoreName = "vector_store_1"
		vectorStoreID   = int64(1)
	)

	tcs := []struct {
		name    string
		req     *v1.UpdateVectorStoreRequest
		resp    *v1.VectorStore
		wantErr bool
	}{
		{
			name: "update name",
			req: &v1.UpdateVectorStoreRequest{
				Id:   "1",
				Name: "new_vector_store_1",
			},
			resp: &v1.VectorStore{
				Id:       "1",
				Object:   vectorStoreName,
				Name:     "new_vector_store_1",
				Metadata: map[string]string{},
			},
			wantErr: false,
		},
		{
			name: "update metadata",
			req: &v1.UpdateVectorStoreRequest{
				Id: "1",
				Metadata: map[string]string{
					"key0": "value0",
				},
			},
			resp: &v1.VectorStore{
				Id:     "1",
				Object: vectorStoreObject,
				Name:   vectorStoreName,
				Metadata: map[string]string{
					"key0": "value0",
				},
			},
			wantErr: false,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			st, tearDown := store.NewTest(t)
			defer tearDown()

			srv := New(
				st,
				&noopFileGetClient{
					ids: map[string]bool{
						fileID: true,
					},
				},
				&noopVStoreClient{
					vs: map[string]int64{
						vectorStoreName: vectorStoreID,
					},
				},
			)
			ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("Authorization", "dummy"))
			resp, err := srv.CreateVectorStore(ctx, &v1.CreateVectorStoreRequest{
				Name: vectorStoreName,
			})
			assert.NoError(t, err)
			assert.Equal(t, vectorStoreName, resp.Name)

			resp, err = srv.UpdateVectorStore(ctx, tc.req)
			assert.NoError(t, err)
			assert.Equal(t, tc.resp.Name, resp.Name)
			assert.Equal(t, tc.resp.Metadata, resp.Metadata)
		})
	}
}

type noopFileGetClient struct {
	ids map[string]bool
}

func (c *noopFileGetClient) GetFile(ctx context.Context, in *fv1.GetFileRequest, opts ...grpc.CallOption) (*fv1.File, error) {
	if _, ok := c.ids[in.Id]; !ok {
		return nil, status.Error(codes.NotFound, "file not found")
	}

	return &fv1.File{}, nil
}

type noopVStoreClient struct {
	vs map[string]int64
}

func (c *noopVStoreClient) CreateVectorStore(ctx context.Context, name string) (int64, error) {
	found, ok := c.vs[name]
	if !ok {
		return 0, status.Error(codes.Internal, "failed to create vector store")
	}
	return found, nil
}

func (c *noopVStoreClient) UpdateVectorStoreName(ctx context.Context, oldName, newName string) error {
	found, ok := c.vs[oldName]
	if !ok {
		return status.Error(codes.NotFound, "name not found")
	}
	c.vs[newName] = found

	return nil
}
func (c *noopVStoreClient) DeleteVectorStore(ctx context.Context, name string) error {
	for k, v := range c.vs {
		if fmt.Sprintf("%d", v) == name {
			delete(c.vs, k)
			return nil
		}
	}

	return status.Error(codes.NotFound, "name not found")
}

func (c *noopVStoreClient) ListVectorStores(ctx context.Context) ([]int64, error) {
	var ids []int64
	for _, id := range c.vs {
		ids = append(ids, id)
	}
	return ids, nil
}
