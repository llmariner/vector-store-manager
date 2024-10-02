package vllm

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/sashabaranov/go-openai"
)

// NewClient creates a new OpenAI client.
func NewClient(url string, log logr.Logger) *Client {
	authToken := ""
	cfg := openai.DefaultConfig(authToken)
	cfg.BaseURL = fmt.Sprintf("http://%s/v1", url)
	return &Client{
		client: openai.NewClientWithConfig(cfg),
		log:    log.WithName("vllm"),
	}
}

// Client wraps OpenAI client.
type Client struct {
	client *openai.Client
	log    logr.Logger
}

// Embed creates embeddings.
func (c *Client) Embed(ctx context.Context, modelName, prompt string) ([]float32, error) {
	req := openai.EmbeddingRequest{
		Input:          []string{prompt},
		Model:          openai.EmbeddingModel(modelName),
		EncodingFormat: openai.EmbeddingEncodingFormatFloat,
	}
	resp, err := c.client.CreateEmbeddings(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("create embeddings: %s", err)
	}
	return resp.Data[0].Embedding, nil
}

// PullModel pulls a model.
func (c *Client) PullModel(ctx context.Context, modelName string) error {
	resp, err := c.client.ListModels(ctx)
	if err != nil {
		return nil
	}
	for _, m := range resp.Models {
		if m.ID == modelName {
			c.log.Info("Model found", "name", modelName)
			return nil
		}
	}
	// TODO(guangrui): bring up a vLLM instance with the required model.
	return fmt.Errorf("pulling model is not implemented in vLLM")
}
