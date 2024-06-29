package ollama

import (
	"context"
	"net/http"
	"net/url"

	"github.com/ollama/ollama/api"
)

// New returns a new Ollama.
func New(addr string) *Ollama {
	url := &url.URL{
		Scheme: "http",
		Host:   addr,
	}
	return &Ollama{
		client: api.NewClient(url, http.DefaultClient),
	}
}

// Ollama wraps the Ollama client.
type Ollama struct {
	client *api.Client
}

// Embed creates embeddings.
func (o *Ollama) Embed(ctx context.Context, modelName, prompt string) ([]float32, error) {
	req := api.EmbeddingRequest{
		Model:  modelName,
		Prompt: prompt,
	}
	resp, err := o.client.Embeddings(ctx, &req)
	if err != nil {
		return nil, err
	}

	// ollama generates embeddings as []float64, but milvus takes []float32 only, so convert []float64 to []float32.
	var es32 []float32
	for _, e := range resp.Embedding {
		es32 = append(es32, float32(e))
	}
	return es32, nil
}

// PullModel pulls a model.
func (o *Ollama) PullModel(ctx context.Context, modelName string) error {
	req := api.PullRequest{
		Model: modelName,
	}
	// Noop progress function.
	fn := func(api.ProgressResponse) error {
		return nil
	}
	return o.client.Pull(ctx, &req, fn)
}
