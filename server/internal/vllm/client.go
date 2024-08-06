package vllm

import (
	"context"
	"fmt"
	"log"

	"github.com/sashabaranov/go-openai"
)

// NewClient creates a new OpenAI client.
func NewClient(url string) *Client {
	authToken := ""
	cfg := openai.DefaultConfig(authToken)
	cfg.BaseURL = fmt.Sprintf("http://%s/v1", url)
	return &Client{
		client: openai.NewClientWithConfig(cfg),
	}
}

// Client wraps OpenAI client.
type Client struct {
	client *openai.Client
}

// Embed creates embeddings.
func (c *Client) Embed(ctx context.Context, modelName, prompt string) ([]float32, error) {
	req := openai.EmbeddingRequest{
		Input: []string{prompt},
		Model: openai.EmbeddingModel(modelName),
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
			log.Printf("Model %s found\n", modelName)
			return nil
		}
	}
	// TODO(guangrui): bring up a vLLM instance with the required model.
	return fmt.Errorf("Pulling model is not implemented in vLLM.\n")
}
