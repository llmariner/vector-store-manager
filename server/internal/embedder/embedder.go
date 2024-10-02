package embedder

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/go-logr/logr"
	"github.com/tmc/langchaingo/documentloaders"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/textsplitter"
)

const (
	charactersPerToken = 4
)

// LLMClient is an interface to handle embedding requests.
type LLMClient interface {
	Embed(ctx context.Context, modelName, prompt string) ([]float32, error)
	PullModel(ctx context.Context, modelName string) error
}

// s3Client is an interface for an S3 client.
type s3Client interface {
	Download(ctx context.Context, w io.WriterAt, key string) error
}

type vstoreClient interface {
	InsertDocuments(ctx context.Context, collectionName string, files, texts []string, vectors [][]float32) error
	DeleteDocuments(ctx context.Context, collectionName, fileID string) error
	Search(ctx context.Context, collectionName string, vectors []float32, numDocuments int) ([]string, error)
}

// E is an embedder.
type E struct {
	llmClient    LLMClient
	s3Client     s3Client
	vstoreClient vstoreClient
	log          logr.Logger
}

// New creates a new Embedder.
func New(
	llmClient LLMClient,
	s3Client s3Client,
	vstoreClient vstoreClient,
	log logr.Logger,
) *E {
	return &E{
		llmClient:    llmClient,
		s3Client:     s3Client,
		vstoreClient: vstoreClient,
		log:          log.WithName("embed"),
	}
}

// AddFile adds a file to the embedder.
func (e *E) AddFile(
	ctx context.Context,
	collectionName,
	modelName,
	fileID,
	fileName,
	filePath string,
	chunkSizeTokens,
	chunkOverlapTokens int64,
) error {
	e.log.Info("Downloading file", "from", filePath)
	f, err := os.CreateTemp("/tmp", "rag-file-")
	if err != nil {
		return err
	}

	log := e.log.WithValues("file", f.Name())
	defer func() {
		if err := os.Remove(f.Name()); err != nil {
			e.log.Error(err, "Failed to remove")
		}
	}()

	if err := e.s3Client.Download(ctx, f, filePath); err != nil {
		return fmt.Errorf("download: %s", err)
	}
	log.Info("Downloaded file")
	if err := f.Close(); err != nil {
		return err
	}

	docs, err := splitFile(logr.NewContext(ctx, log), f.Name(), filepath.Ext(fileName), chunkSizeTokens, chunkSizeTokens)
	if err != nil {
		return fmt.Errorf("split file: %s", err)
	}
	log.Info("Splitted file into chunks", "count", len(docs))

	if err := e.llmClient.PullModel(ctx, modelName); err != nil {
		return fmt.Errorf("pull model: %s", err)
	}

	var embeddings [][]float32
	var texts []string
	var files []string
	for _, doc := range docs {
		es, err := e.llmClient.Embed(ctx, modelName, doc.PageContent)
		if err != nil {
			return fmt.Errorf("llm embed: %s", err)
		}
		embeddings = append(embeddings, es)
		texts = append(texts, doc.PageContent)
		files = append(files, fileID)
	}
	return e.vstoreClient.InsertDocuments(ctx, collectionName, files, texts, embeddings)
}

func splitFile(ctx context.Context, fileName, fileType string, chunkSizeTokens, chunkOverlapTokens int64) ([]schema.Document, error) {
	logr.FromContextOrDiscard(ctx).Info("Splitting file into chunks")
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
		return nil, fmt.Errorf("unexpected file type: fileType=%q", fileType)
	}
}

// DeleteFile deletes a file from the embedder.
func (e *E) DeleteFile(ctx context.Context, collectionName, fileID string) error {
	return e.vstoreClient.DeleteDocuments(ctx, collectionName, fileID)
}

// Search searches for the matched documents in the embedder for the given query.
func (e *E) Search(ctx context.Context, collectionName, modelName, query string, numDocs int) ([]string, error) {
	if err := e.llmClient.PullModel(ctx, modelName); err != nil {
		return nil, fmt.Errorf("pull model: %s", err)
	}

	es, err := e.llmClient.Embed(ctx, modelName, query)
	if err != nil {
		return nil, fmt.Errorf("embed: %s", err)
	}

	results, err := e.vstoreClient.Search(ctx, collectionName, es, numDocs)
	if err != nil {
		return nil, fmt.Errorf("vector search: %s", err)
	}
	e.log.Info("search result", "query", query, "results", results)
	return results, nil
}
