package vllm

import (
	"fmt"
)

// Dimension returns the dimension of the model.
func Dimension(model string) (int, error) {
	dimsByModel := map[string]int{
		"intfloat/e5-mistral-7b-instruct": 4096,
	}
	dim, ok := dimsByModel[model]
	if !ok {
		var models []string
		for m := range dimsByModel {
			models = append(models, m)
		}
		return -1, fmt.Errorf("embedding model must be one of: %v", models)
	}
	return dim, nil
}
