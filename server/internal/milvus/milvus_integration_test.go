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
	const (
		collectionName    = "test_collection_1"
		collectionNameNew = "test_collection_1_new"
		dimensions        = 128
	)

	cfg := db.Config{
		Host: "localhost",
		Port: 19530,
	}
	ctx := context.Background()
	s, err := New(ctx, cfg)
	assert.NoError(t, err)

	preExist, err := s.ListVectorStores(ctx)
	assert.NoError(t, err)

	_, err = s.CreateVectorStore(ctx, collectionName, dimensions)
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

func TestInsertDocuments(t *testing.T) {
	const (
		collectionName = "test_collection_1"
		dimensions     = 4
	)

	vectors := [][]float32{
		{-0.2161688655614853, 0.4428754150867462, 0.12087928503751755, 0.38950398564338684},
		{-0.023337043821811676, 0.19466467201709747, -0.5630808472633362, 0.5578770637512207},
	}

	cfg := db.Config{
		Host: "localhost",
		Port: 19530,
	}
	ctx := context.Background()
	s, err := New(ctx, cfg)
	assert.NoError(t, err)

	_, err = s.CreateVectorStore(ctx, collectionName, dimensions)
	assert.NoError(t, err)

	err = s.InsertDocuments(ctx, collectionName, vectors)
	assert.NoError(t, err)

	err = s.DeleteVectorStore(ctx, collectionName)
	assert.NoError(t, err)
}
