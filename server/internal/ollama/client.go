package ollama

import (
	"context"

	"github.com/ollama/ollama/api"
)

// New returns a new Ollama.
func New() (*Ollama, error) {
	client, err := api.ClientFromEnvironment()
	if err != nil {
		return nil, err
	}
	return &Ollama{
		client: client,
	}, nil
}

// Ollama wraps the Ollama client.
type Ollama struct {
	client *api.Client
}

// Embed creates embeddings.
func (o *Ollama) Embed(ctx context.Context, modelName, prompt string) ([]float64, error) {
	req := api.EmbeddingRequest{
		Model:  modelName,
		Prompt: prompt,
	}
	resp, err := o.client.Embeddings(ctx, &req)
	if err != nil {
		return nil, err
	}
	return resp.Embedding, nil
}

// PullModel pulls a model.
func (o *Ollama) PullModel(ctx context.Context, modelName string) error {
	req := api.PullRequest{
		Model: modelName,
	}
	return o.client.Pull(ctx, &req, nil)
}
