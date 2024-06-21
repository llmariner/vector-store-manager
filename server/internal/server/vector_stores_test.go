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
	"google.golang.org/grpc/status"
)

const (
	fileID          = "file0"
	vectorStoreName = "vector_store_1"
	modelName       = "model"
	dimensions      = 10
)

func TestCreateVectorStore(t *testing.T) {
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
				&noopFileInternalClient{
					ids: map[string]string{
						fileID: "test.txt",
					},
				},
				&noopVStoreClient{
					vs: map[string]int64{},
				},
				&noopEmbedder{},
				modelName,
				dimensions,
			)
			ctx := context.Background()
			resp, err := srv.CreateVectorStore(ctx, tc.req)
			if tc.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, vectorStoreName, resp.Name)
		})
	}
}

func TestListVectorStores(t *testing.T) {
	names := []string{
		vectorStoreName,
		"vector_store_2",
		"vector_store_3",
	}

	st, tearDown := store.NewTest(t)
	defer tearDown()

	srv := New(
		st,
		&noopFileGetClient{
			ids: map[string]bool{
				fileID: true,
			},
		},
		&noopFileInternalClient{
			ids: map[string]string{
				fileID: "test.txt",
			},
		},
		&noopVStoreClient{
			vs: map[string]int64{},
		},
		&noopEmbedder{},
		modelName,
		dimensions,
	)

	ctx := context.Background()
	var vss []*v1.VectorStore
	for _, name := range names {
		vs, err := srv.CreateVectorStore(ctx, &v1.CreateVectorStoreRequest{
			Name: name,
		})
		assert.NoError(t, err)
		assert.Equal(t, name, vs.Name)
		vss = append(vss, vs)
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
				FirstId: vss[2].Id,
				LastId:  vss[0].Id,
				HasMore: false,
				Data: []*v1.VectorStore{
					{
						Id:   vss[2].Id,
						Name: "vector_store_3",
					},
					{
						Id:   vss[1].Id,
						Name: "vector_store_2",
					},
					{
						Id:   vss[0].Id,
						Name: "vector_store_1",
					},
				},
			},
		},
		{
			name: "success with after",
			req: &v1.ListVectorStoresRequest{
				Limit: 1,
				After: vss[1].Id,
			},
			resp: &v1.ListVectorStoresResponse{
				Object:  vectorStoreName,
				FirstId: vss[2].Id,
				LastId:  vss[2].Id,
				HasMore: false,
				Data: []*v1.VectorStore{
					{
						Id:   vss[0].Id,
						Name: "vector_store_1",
					},
				},
			},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			respList, err := srv.ListVectorStores(ctx, tc.req)
			assert.NoError(t, err)
			assert.Len(t, respList.Data, len(tc.resp.Data))
			for i, want := range tc.resp.Data {
				assert.Equal(t, want.Id, respList.Data[i].Id)
				assert.Equal(t, want.Name, respList.Data[i].Name)
			}
			assert.Equal(t, tc.resp.HasMore, respList.HasMore)
		})
	}
}

func TestGetVectorStores(t *testing.T) {
	st, tearDown := store.NewTest(t)
	defer tearDown()

	srv := New(
		st,
		&noopFileGetClient{},
		&noopFileInternalClient{},
		&noopVStoreClient{
			vs: map[string]int64{},
		},
		&noopEmbedder{},
		modelName,
		dimensions,
	)

	names := []string{
		"vector_store_1",
		"vector_store_2",
	}
	ctx := context.Background()
	var vss []*v1.VectorStore
	for _, name := range names {
		vs, err := srv.CreateVectorStore(ctx, &v1.CreateVectorStoreRequest{
			Name: name,
		})
		assert.NoError(t, err)
		assert.Equal(t, name, vs.Name)
		vss = append(vss, vs)
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
				Id: vss[0].Id,
			},
			resp: &v1.VectorStore{
				Id:   vss[0].Id,
				Name: vss[0].Name,
			},
			wantErr: false,
		},
		{
			name: "not found",
			req: &v1.GetVectorStoreRequest{
				Id: "dummy",
			},
			wantErr: true,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			got, err := srv.GetVectorStore(ctx, tc.req)
			if tc.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tc.resp.Id, got.Id)
			assert.Equal(t, tc.resp.Name, got.Name)
		})
	}
}

func TestDeleteVectorStore(t *testing.T) {
	tcs := []struct {
		name    string
		req     func(id string) *v1.DeleteVectorStoreRequest
		resp    func(id string) *v1.DeleteVectorStoreResponse
		wantErr bool
	}{
		{
			name: "success",
			req: func(id string) *v1.DeleteVectorStoreRequest {
				return &v1.DeleteVectorStoreRequest{
					Id: id,
				}
			},
			resp: func(id string) *v1.DeleteVectorStoreResponse {
				return &v1.DeleteVectorStoreResponse{
					Id:      id,
					Object:  vectorStoreName,
					Deleted: true,
				}
			},
			wantErr: false,
		},
		{
			name: "not found",
			req: func(id string) *v1.DeleteVectorStoreRequest {
				return &v1.DeleteVectorStoreRequest{
					Id: "not-existing",
				}
			},
			resp: func(id string) *v1.DeleteVectorStoreResponse {
				return &v1.DeleteVectorStoreResponse{}
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
				&noopFileInternalClient{
					ids: map[string]string{
						fileID: "test.txt",
					},
				},
				&noopVStoreClient{
					vs: map[string]int64{},
				},
				&noopEmbedder{},
				modelName,
				dimensions,
			)
			ctx := context.Background()
			resp, err := srv.CreateVectorStore(ctx, &v1.CreateVectorStoreRequest{
				Name: vectorStoreName,
			})
			assert.NoError(t, err)
			assert.Equal(t, vectorStoreName, resp.Name)

			respDelete, err := srv.DeleteVectorStore(ctx, tc.req(resp.Id))
			if tc.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			want := tc.resp(resp.Id)
			assert.Equal(t, want.Id, respDelete.Id)
			assert.Equal(t, want.Deleted, respDelete.Deleted)
		})
	}
}

func TestUpdateVectorStore(t *testing.T) {
	tcs := []struct {
		name    string
		req     func(id string) *v1.UpdateVectorStoreRequest
		resp    *v1.VectorStore
		wantErr bool
	}{
		{
			name: "update name",
			req: func(id string) *v1.UpdateVectorStoreRequest {
				return &v1.UpdateVectorStoreRequest{
					Id:   id,
					Name: "new_vector_store_1",
				}
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
			req: func(id string) *v1.UpdateVectorStoreRequest {
				return &v1.UpdateVectorStoreRequest{
					Id: id,
					Metadata: map[string]string{
						"key0": "value0",
					},
				}
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
				&noopFileInternalClient{
					ids: map[string]string{
						fileID: "test.txt",
					},
				},
				&noopVStoreClient{
					vs: map[string]int64{},
				},
				&noopEmbedder{},
				modelName,
				dimensions,
			)
			ctx := context.Background()
			resp, err := srv.CreateVectorStore(ctx, &v1.CreateVectorStoreRequest{
				Name: vectorStoreName,
			})
			assert.NoError(t, err)
			assert.Equal(t, vectorStoreName, resp.Name)

			resp, err = srv.UpdateVectorStore(ctx, tc.req(resp.Id))
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

type noopFileInternalClient struct {
	ids map[string]string
}

func (c *noopFileInternalClient) GetFilePath(ctx context.Context, in *fv1.GetFilePathRequest, opts ...grpc.CallOption) (*fv1.GetFilePathResponse, error) {
	path, ok := c.ids[in.Id]
	if !ok {
		return nil, status.Error(codes.NotFound, "file not found")
	}

	return &fv1.GetFilePathResponse{
		Path: path,
	}, nil
}

type noopVStoreClient struct {
	vs map[string]int64
}

func (c *noopVStoreClient) CreateVectorStore(ctx context.Context, name string, dimensions int) (int64, error) {
	newID := int64(len(c.vs) + 1)
	c.vs[name] = newID
	return newID, nil
}

func (c *noopVStoreClient) DeleteVectorStore(ctx context.Context, name string) error {
	for k := range c.vs {
		if k == name {
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

type noopEmbedder struct {
	collectionName string
}

func (c *noopEmbedder) AddFile(ctx context.Context, collectionName, modelName, filePath string, chunkSizeTokens, chunkOverlapTokens int64) error {
	if collectionName == c.collectionName {
		return nil
	}
	return fmt.Errorf("collection %s not found", collectionName)
}
