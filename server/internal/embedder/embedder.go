package embedder

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/tmc/langchaingo/documentloaders"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/textsplitter"
)

const (
	charactersPerToken = 4
)

type llmClient interface {
	Embed(ctx context.Context, modelName, prompt string) ([]float64, error)
	PullModel(ctx context.Context, modelName string) error
}

// s3Client is an interface for an S3 client.
type s3Client interface {
	Download(w io.WriterAt, key string) error
}

type vstoreClient interface {
	InsertDocuments(ctx context.Context, collectionName string, vectors [][]float32) error
}

// E is an embedder.
type E struct {
	llmClient    llmClient
	s3Client     s3Client
	vstoreClient vstoreClient
}

// New creates a new Embedder.
func New(
	llmClient llmClient,
	s3Client s3Client,
	vstoreClient vstoreClient,
) *E {
	return &E{
		llmClient:    llmClient,
		s3Client:     s3Client,
		vstoreClient: vstoreClient,
	}
}

// AddFile adds a file to the embedder.
func (e *E) AddFile(ctx context.Context, collectionName, modelName, filePath string, chunkSizeTokens, chunkOverlapTokens int64) error {
	log.Printf("Downloading file from %q\n", filePath)
	f, err := os.CreateTemp("/tmp", "rag-file-")
	if err != nil {
		return err
	}
	defer func() {
		if err := os.Remove(f.Name()); err != nil {
			log.Printf("Failed to remove %q: %s", f.Name(), err)
		}
	}()

	if err := e.s3Client.Download(f, filePath); err != nil {
		return fmt.Errorf("download: %s", err)
	}
	log.Printf("Downloaded file to %q\n", f.Name())
	if err := f.Close(); err != nil {
		return err
	}

	ext := filepath.Ext(filePath)
	docs, err := splitFile(ctx, f.Name(), ext, chunkSizeTokens, chunkSizeTokens)
	if err != nil {
		return fmt.Errorf("split file: %s", err)
	}

	if err := e.llmClient.PullModel(ctx, modelName); err != nil {
		return fmt.Errorf("pull model: %s", err)
	}

	embeddings := make([][]float32, 0, len(docs))
	for _, doc := range docs {
		es, err := e.llmClient.Embed(ctx, modelName, doc.PageContent)
		if err != nil {
			return err
		}

		// ollama generates embeddings as []float64, and milvus takes []float32 only, so convert []float64 to []float32.
		var es32 []float32
		for _, e := range es {
			es32 = append(es32, float32(e))
		}
		embeddings = append(embeddings, es32)
	}
	return e.vstoreClient.InsertDocuments(ctx, collectionName, embeddings)
}

func splitFile(ctx context.Context, fileName, fileType string, chunkSizeTokens, chunkOverlapTokens int64) ([]schema.Document, error) {
	log.Printf("Spliting file %q into chunks\n", fileName)
	file, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = file.Close()
	}()

	splitter := textsplitter.NewRecursiveCharacter()
	splitter.ChunkSize = int(chunkSizeTokens) * charactersPerToken
	splitter.ChunkOverlap = int(chunkOverlapTokens) * charactersPerToken

	switch fileType {
	case ".pdf":
		finfo, err := file.Stat()
		if err != nil {
			return nil, err
		}
		return documentloaders.NewPDF(file, finfo.Size()).LoadAndSplit(ctx, splitter)
	case ".html":
		return documentloaders.NewHTML(file).LoadAndSplit(ctx, splitter)
	case ".txt":
		return documentloaders.NewText(file).LoadAndSplit(ctx, splitter)
	// TODO(guangrui): support more file types.
	default:
		return nil, fmt.Errorf("unexpected file type: fileName=%q, fileType=%q", fileName, fileType)
	}
}
