//go:build integration
// +build integration

package milvus

import (
	"context"
	"testing"

	"github.com/llm-operator/common/pkg/db"
	"github.com/stretchr/testify/assert"
)

func TestMilvusCreateListUpdateDeleteVectorStores(t *testing.T) {
	collectionName := "test_collection_1"
	collectionNameNew := "test_collection_1_new"

	cfg := db.Config{
		Host: "localhost:19530",
	}
	ctx := context.Background()
	s, err := New(ctx, cfg)
	assert.NoError(t, err)

	preExist, err := s.ListVectorStores(ctx)
	assert.NoError(t, err)

	_, err = s.CreateVectorStore(ctx, collectionName)
	assert.NoError(t, err)

	vss, err := s.ListVectorStores(ctx)
	assert.NoError(t, err)
	assert.Equal(t, len(preExist)+1, len(vss))

	err = s.UpdateVectorStoreName(ctx, collectionName, collectionNameNew)

	c, err := s.client.DescribeCollection(ctx, collectionNameNew)
	assert.NoError(t, err)
	assert.Equal(t, collectionNameNew, c.Name)

	err = s.DeleteVectorStore(ctx, collectionNameNew)
	assert.NoError(t, err)

	vss, err = s.ListVectorStores(ctx)
	assert.NoError(t, err)
	assert.Equal(t, len(preExist), len(vss))
}
