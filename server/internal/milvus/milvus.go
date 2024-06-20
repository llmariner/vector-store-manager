package milvus

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/llm-operator/common/pkg/db"
	"github.com/milvus-io/milvus-sdk-go/v2/client"
	"github.com/milvus-io/milvus-sdk-go/v2/entity"
)

const (
	defaultShardNum = 1
	vectorColName   = "vector"
)

// S wraps Milvus client.
type S struct {
	client client.Client
}

// New creates an active client connection to the Milvus server.
func New(ctx context.Context, cfg db.Config) (*S, error) {
	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	log.Printf("Connecting to Milvus: %s\n", addr)
	passwd := os.Getenv(cfg.PasswordEnvName)
	config := client.Config{
		Address:  addr,
		Username: cfg.Username,
		Password: passwd,
		DBName:   cfg.Database,
	}
	c, err := client.NewClient(ctx, config)
	if err != nil {
		return nil, err
	}
	log.Printf("Connected to Milvus\n")
	return &S{
		client: c,
	}, nil
}

// CreateVectorStore creates a new collection in milvus.
func (s *S) CreateVectorStore(ctx context.Context, name string, dimensions int) (int64, error) {
	schema := &entity.Schema{
		CollectionName: name,
		AutoID:         true,
		Fields: []*entity.Field{
			{
				Name:       "pk",
				DataType:   entity.FieldTypeInt64,
				AutoID:     true,
				PrimaryKey: true,
			},
			{
				Name:     vectorColName,
				DataType: entity.FieldTypeFloatVector,
				TypeParams: map[string]string{
					entity.TypeParamDim: strconv.Itoa(dimensions),
				},
			},
		},
	}

	err := s.client.CreateCollection(ctx, schema, defaultShardNum)
	if err != nil {
		return 0, err
	}

	c, err := s.client.DescribeCollection(ctx, name)
	if err != nil {
		return 0, err
	}
	log.Printf("Created collection: %+v", c)
	return c.ID, nil
}

// ListVectorStores lists collections in milvus.
func (s *S) ListVectorStores(ctx context.Context) ([]int64, error) {
	cs, err := s.client.ListCollections(ctx)
	if err != nil {
		return nil, err
	}
	var vss []int64
	for _, c := range cs {
		vss = append(vss, c.ID)
	}
	return vss, nil
}

// UpdateVectorStoreName updates a collection name in milvus.
func (s *S) UpdateVectorStoreName(ctx context.Context, oldName, newName string) error {
	return s.client.RenameCollection(ctx, oldName, newName)
}

// DeleteVectorStore deletes a collection in milvus.
func (s *S) DeleteVectorStore(ctx context.Context, name string) error {
	return s.client.DropCollection(ctx, name)
}

// InsertDocuments inserts documents into a collection in milvus.
func (s *S) InsertDocuments(ctx context.Context, name string, vectors [][]float32) error {
	vectorCol := entity.NewColumnFloatVector(vectorColName, len(vectors[0]), vectors)
	if _, err := s.client.Insert(ctx, name, "", vectorCol); err != nil {
		return err
	}
	return nil
}
