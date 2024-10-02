package embedder

import (
	"context"
	"fmt"
	"io"
	"testing"

	"github.com/go-logr/logr/testr"
	"github.com/stretchr/testify/assert"
	"github.com/tmc/langchaingo/schema"
)

func TestAddSearchDeleteFile(t *testing.T) {
	const (
		fileID             = "file-001"
		fileName           = "test.txt"
		filePath           = "testdata/test.txt"
		collectionName0    = "collection0"
		vectorStoreName    = "vector_store_1"
		embeddingModel     = "model1"
		chunkSizeTokens    = 10
		chunkOverlapTokens = 2
		modelName          = "model1"
	)

	tcs := []struct {
		name     string
		fileName string
		path     string
		wantErr  bool
	}{
		{
			name:     "success",
			fileName: fileName,
			path:     filePath,
			wantErr:  false,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			e := New(
				&noopLLMClient{
					e: map[string][]float32{
						"line1": {1.111, 0.122},
						"line2": {2.211, 0.222},
					},
				},
				&noopS3Client{},
				&noopVStoreClient{
					collectionName: collectionName0,
					docs: map[int][]string{
						1: {"line1"},
						2: {"line2"},
					},
				},
				testr.New(t),
			)
			ctx := context.Background()
			err := e.AddFile(ctx, collectionName0, modelName, fileID, tc.fileName, tc.path, chunkSizeTokens, chunkOverlapTokens)
			if tc.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)

			docs, err := e.Search(ctx, collectionName0, modelName, "line1", 1)
			assert.NoError(t, err)
			assert.Equal(t, 1, len(docs))
			assert.Equal(t, "line1", docs[0])

			err = e.DeleteFile(ctx, collectionName0, fileID)
			assert.NoError(t, err)
		})
	}
}

func TestSplitFile(t *testing.T) {
	tcs := []struct {
		name               string
		path               string
		chunkSizeTokens    int64
		chunkOverlapTokens int64
		exp                []schema.Document
		wantErr            bool
	}{
		{
			name:               "text",
			path:               "testdata/test.txt",
			chunkSizeTokens:    5,
			chunkOverlapTokens: 2,
			exp: []schema.Document{
				{
					PageContent: "Tokens can be",
				},
				{
					PageContent: "can be thought of as",
				},
			},
			wantErr: false,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			got, err := splitFile(ctx, tc.path, ".txt", tc.chunkSizeTokens, tc.chunkOverlapTokens)
			if tc.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Less(t, 2, len(got))
			assert.Equal(t, tc.exp[0].PageContent, got[0].PageContent)
			assert.Equal(t, tc.exp[1].PageContent, got[1].PageContent)
		})
	}

}

type noopLLMClient struct {
	// e is keyed by prompt
	e map[string][]float32
}

func (c *noopLLMClient) Embed(ctx context.Context, modelName, prompt string) ([]float32, error) {
	e, ok := c.e[prompt]
	if !ok {
		return nil, fmt.Errorf("no embedding found")
	}
	return e, nil
}

func (c *noopLLMClient) PullModel(ctx context.Context, modelName string) error {
	return nil
}

// noopS3Client is a no-op S3 client.
type noopS3Client struct{}

// Download is a no-op implementation of Download
func (n *noopS3Client) Download(ctx context.Context, w io.WriterAt, key string) error {
	return nil
}

type noopVStoreClient struct {
	collectionName string
	docs           map[int][]string
}

func (c *noopVStoreClient) InsertDocuments(
	ctx context.Context,
	collectionName string,
	fileIDs, texts []string,
	vectors [][]float32,
) error {
	if collectionName != c.collectionName {
		return fmt.Errorf("collection %s not found", collectionName)
	}
	return nil
}

func (c *noopVStoreClient) DeleteDocuments(ctx context.Context, collectionName, fileID string) error {
	if collectionName != c.collectionName {
		return fmt.Errorf("collection %s not found", collectionName)
	}
	return nil
}

func (c *noopVStoreClient) Search(ctx context.Context, collectionName string, vectors []float32, numDocuments int) ([]string, error) {
	if collectionName != c.collectionName {
		return nil, fmt.Errorf("collection %s not found", collectionName)
	}
	return c.docs[int(vectors[0])], nil
}
