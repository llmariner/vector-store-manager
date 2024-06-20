package ollama

import "fmt"

// Dimension returns the dimension of the model.
func Dimension(model string) (int, error) {
	dimsByModel := map[string]int{
		"all-minilm":       384,
		"nomic-embed-text": 768,
	}
	dim, ok := dimsByModel[model]
	if !ok {
		var models []string
		for m := range dimsByModel {
			models = append(models, m)
		}
		return -1, fmt.Errorf("model must be one of: %v", models)
	}
	return dim, nil
}
