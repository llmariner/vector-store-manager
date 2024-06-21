package server

import (
	"context"
	"testing"

	v1 "github.com/llm-operator/vector-store-manager/api/v1"
	"github.com/llm-operator/vector-store-manager/server/internal/store"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	vectorStoreID  = "vector_store_1"
	collectionID   = int64(1)
	collectionName = "collection0"
)

func TestCreateVectorStoreFile(t *testing.T) {
	tcs := []struct {
		name    string
		req     *v1.CreateVectorStoreFileRequest
		wantErr bool
	}{
		{
			name: "success",
			req: &v1.CreateVectorStoreFileRequest{
				FileId:        fileID,
				VectorStoreId: vectorStoreID,
			},
			wantErr: false,
		},
		{
			name: "success with files",
			req: &v1.CreateVectorStoreFileRequest{
				FileId:        fileID,
				VectorStoreId: vectorStoreID,
				ChunkingStrategy: &v1.ChunkingStrategy{
					Type: string(store.ChunkingStrategyTypeAuto),
				},
			},
			wantErr: false,
		},
		{
			name: "invalid fileID",
			req: &v1.CreateVectorStoreFileRequest{
				FileId:        "unknown",
				VectorStoreId: vectorStoreID,
			},
			wantErr: true,
		},
		{
			name: "invalid vector store ID",
			req: &v1.CreateVectorStoreFileRequest{
				FileId:        fileID,
				VectorStoreId: "unknown",
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
					vs: map[string]int64{
						vectorStoreID: 1,
					},
				},
				&noopEmbedder{
					collectionName: vectorStoreID,
				},
				modelName,
				dimensions,
			)
			err := st.CreateCollection(&store.Collection{
				CollectionID:  collectionID,
				VectorStoreID: vectorStoreID,
				Name:          collectionName,
				Status:        store.CollectionStatusCompleted,
				ProjectID:     "default",
			})
			assert.NoError(t, err)
			resp, err := srv.CreateVectorStoreFile(context.Background(), tc.req)
			if tc.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, fileID, resp.Id)
			assert.Equal(t, vectorStoreID, resp.VectorStoreId)
			assert.Equal(t, vectorStoreFileObject, resp.Object)
			assert.Equal(t, string(store.ChunkingStrategyTypeAuto), resp.ChunkingStrategy.Type)
		})
	}
}

func TestCreateVectorStoreFile_AlreadyExists(t *testing.T) {
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
			vs: map[string]int64{
				vectorStoreID: 1,
			},
		},
		&noopEmbedder{
			collectionName: vectorStoreID,
		},
		modelName,
		dimensions,
	)
	err := st.CreateCollection(&store.Collection{
		CollectionID:  collectionID,
		VectorStoreID: vectorStoreID,
		Name:          collectionName,
		Status:        store.CollectionStatusCompleted,
		ProjectID:     "default",
	})
	assert.NoError(t, err)

	req := &v1.CreateVectorStoreFileRequest{
		FileId:        fileID,
		VectorStoreId: vectorStoreID,
	}
	_, err = srv.CreateVectorStoreFile(context.Background(), req)
	assert.NoError(t, err)

	_, err = srv.CreateVectorStoreFile(context.Background(), req)
	assert.Error(t, err)
	assert.Equal(t, codes.AlreadyExists, status.Code(err))
}

func TestListVectorStoreFiles(t *testing.T) {
	const (
		fileID         = "file0"
		vectorStoreID  = "vector_store_1"
		collectionID   = int64(1)
		collectionName = "collection0"
	)
	fs := []string{"file0", "file1", "file2"}

	tcs := []struct {
		name string
		req  *v1.ListVectorStoreFilesRequest
		resp *v1.ListVectorStoreFilesResponse
	}{
		{
			name: "no pagination",
			req: &v1.ListVectorStoreFilesRequest{
				VectorStoreId: vectorStoreID,
			},
			resp: &v1.ListVectorStoreFilesResponse{
				Object:  vectorStoreFileObject,
				FirstId: fs[2],
				LastId:  fs[0],
				HasMore: false,
				Data: []*v1.VectorStoreFile{
					{
						Id:            fs[2],
						VectorStoreId: vectorStoreID,
					},
					{
						Id:            fs[1],
						VectorStoreId: vectorStoreID,
					},
					{
						Id:            fs[0],
						VectorStoreId: vectorStoreID,
					},
				},
			},
		},
		{
			name: "success with after",
			req: &v1.ListVectorStoreFilesRequest{
				VectorStoreId: vectorStoreID,
				Limit:         1,
				After:         fs[2],
			},
			resp: &v1.ListVectorStoreFilesResponse{
				Object:  vectorStoreFileObject,
				FirstId: fs[1],
				LastId:  fs[1],
				HasMore: true,
				Data: []*v1.VectorStoreFile{
					{
						Id:            fs[1],
						VectorStoreId: vectorStoreID,
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
						fs[0]: true,
						fs[1]: true,
						fs[2]: true,
					},
				},
				&noopFileInternalClient{
					ids: map[string]string{
						fs[0]: "test0.txt",
						fs[1]: "test1.txt",
						fs[2]: "test2.txt",
					},
				},
				&noopVStoreClient{
					vs: map[string]int64{
						vectorStoreID: collectionID,
					},
				},
				&noopEmbedder{
					collectionName: vectorStoreID,
				},
				modelName,
				dimensions,
			)
			err := st.CreateCollection(&store.Collection{
				CollectionID:  collectionID,
				VectorStoreID: vectorStoreID,
				Name:          collectionName,
				Status:        store.CollectionStatusCompleted,
				ProjectID:     "default",
			})
			assert.NoError(t, err)
			ctx := context.Background()
			for _, f := range fs {
				resp, err := srv.CreateVectorStoreFile(ctx, &v1.CreateVectorStoreFileRequest{
					FileId:        f,
					VectorStoreId: vectorStoreID,
				})
				assert.NoError(t, err)
				assert.Equal(t, f, resp.Id)
				assert.Equal(t, vectorStoreID, resp.VectorStoreId)
				assert.Equal(t, vectorStoreFileObject, resp.Object)
				assert.Equal(t, string(store.ChunkingStrategyTypeAuto), resp.ChunkingStrategy.Type)
			}

			respList, err := srv.ListVectorStoreFiles(ctx, tc.req)
			assert.NoError(t, err)
			assert.Equal(t, len(tc.resp.Data), len(respList.Data))
			assert.Equal(t, tc.resp.FirstId, respList.FirstId)
			assert.Equal(t, tc.resp.LastId, respList.LastId)
			assert.Equal(t, tc.resp.HasMore, respList.HasMore)
		})
	}
}

func TestGetVectorStoreFile(t *testing.T) {
	const (
		fileID         = "file0"
		vectorStoreID  = "vector_store_1"
		collectionID   = int64(1)
		collectionName = "collection0"
	)

	tcs := []struct {
		name    string
		req     *v1.GetVectorStoreFileRequest
		resp    *v1.VectorStoreFile
		wantErr bool
	}{
		{
			name: "success",
			req: &v1.GetVectorStoreFileRequest{
				FileId:        fileID,
				VectorStoreId: vectorStoreID,
			},
			resp: &v1.VectorStoreFile{
				Id:            fileID,
				VectorStoreId: vectorStoreID,
			},
			wantErr: false,
		},
		{
			name: "file not found",
			req: &v1.GetVectorStoreFileRequest{
				FileId:        "unknown",
				VectorStoreId: vectorStoreID,
			},
			wantErr: true,
		},
		{
			name: "vector store not found",
			req: &v1.GetVectorStoreFileRequest{
				FileId:        "unknown",
				VectorStoreId: vectorStoreID,
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
					vs: map[string]int64{
						vectorStoreID: collectionID,
					},
				},
				&noopEmbedder{
					collectionName: vectorStoreID,
				},
				modelName,
				dimensions,
			)
			err := st.CreateCollection(&store.Collection{
				CollectionID:  collectionID,
				VectorStoreID: vectorStoreID,
				Name:          collectionName,
				Status:        store.CollectionStatusCompleted,
				ProjectID:     "default",
			})
			assert.NoError(t, err)
			ctx := context.Background()
			resp, err := srv.CreateVectorStoreFile(ctx, &v1.CreateVectorStoreFileRequest{
				FileId:        fileID,
				VectorStoreId: vectorStoreID,
			})
			assert.NoError(t, err)
			assert.Equal(t, fileID, resp.Id)
			assert.Equal(t, vectorStoreID, resp.VectorStoreId)

			got, err := srv.GetVectorStoreFile(ctx, tc.req)
			if tc.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, fileID, got.Id)
			assert.Equal(t, vectorStoreID, got.VectorStoreId)
		})
	}
}

func TestDeleteVectorStoreFile(t *testing.T) {
	tcs := []struct {
		name    string
		req     *v1.DeleteVectorStoreFileRequest
		resp    *v1.DeleteVectorStoreFileResponse
		wantErr bool
	}{
		{
			name: "success",
			req: &v1.DeleteVectorStoreFileRequest{
				FileId:        fileID,
				VectorStoreId: vectorStoreID,
			},
			resp: &v1.DeleteVectorStoreFileResponse{
				Id:      fileID,
				Object:  vectorStoreFileObject,
				Deleted: true,
			},
			wantErr: false,
		},
		{
			name: "not found",
			req: &v1.DeleteVectorStoreFileRequest{
				FileId:        "file2",
				VectorStoreId: vectorStoreID,
			},
			resp:    &v1.DeleteVectorStoreFileResponse{},
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
					vs: map[string]int64{
						vectorStoreID: collectionID,
					},
				},
				&noopEmbedder{
					collectionName: vectorStoreID,
				},
				modelName,
				dimensions,
			)
			err := st.CreateCollection(&store.Collection{
				CollectionID:  collectionID,
				VectorStoreID: vectorStoreID,
				Name:          "collection0",
				Status:        store.CollectionStatusCompleted,
				ProjectID:     "default",
			})
			assert.NoError(t, err)
			ctx := context.Background()
			resp, err := srv.CreateVectorStoreFile(ctx, &v1.CreateVectorStoreFileRequest{
				FileId:        fileID,
				VectorStoreId: vectorStoreID,
			})
			assert.NoError(t, err)
			assert.Equal(t, fileID, resp.Id)
			assert.Equal(t, vectorStoreID, resp.VectorStoreId)

			respDelete, err := srv.DeleteVectorStoreFile(ctx, tc.req)
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
