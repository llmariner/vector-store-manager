package milvus

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/llmariner/common/pkg/db"
	"github.com/milvus-io/milvus-sdk-go/v2/client"
	"github.com/milvus-io/milvus-sdk-go/v2/entity"
)

const (
	defaultShardNum                             = 1
	vectorColName                               = "vector"
	primaryKeyColName                           = "pk"
	fileIDColName                               = "fileID"
	textColName                                 = "text"
	maxVarCharLength                            = 4096 * 4 // maxMaxChunkSizeTokens * charactersPerToken
	defaultMetricType         entity.MetricType = entity.L2
	defaultIvfFlatNList                         = 128
	defaultIvfFlatSearchParam                   = 16
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
				Name:       primaryKeyColName,
				DataType:   entity.FieldTypeInt64,
				AutoID:     true,
				PrimaryKey: true,
			},
			{
				Name:     fileIDColName,
				DataType: entity.FieldTypeVarChar,
				TypeParams: map[string]string{
					entity.TypeParamMaxLength: strconv.Itoa(maxVarCharLength),
				},
			},
			{
				Name:     textColName,
				DataType: entity.FieldTypeVarChar,
				TypeParams: map[string]string{
					entity.TypeParamMaxLength: strconv.Itoa(maxVarCharLength),
				},
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

	// TODO(guangrui): Experiment with other type of indexing and tuning.
	// Refer to https://milvus.io/docs/index.md#Indexes-supported-in-Milvus
	idx, err := entity.NewIndexIvfFlat(defaultMetricType, defaultIvfFlatNList)
	if err != nil {
		return 0, fmt.Errorf("new ivf flat index: %s", err)
	}
	if err := s.client.CreateIndex(ctx, name, vectorColName, idx, false); err != nil {
		return 0, fmt.Errorf("create index:: %s", err)
	}

	c, err := s.client.DescribeCollection(ctx, name)
	if err != nil {
		return 0, err
	}
	log.Printf("Created collection: %+v\n", c)

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
func (s *S) InsertDocuments(ctx context.Context, name string, files, texts []string, vectors [][]float32) error {
	vectorCol := entity.NewColumnFloatVector(vectorColName, len(vectors[0]), vectors)
	fileCol := entity.NewColumnVarChar(fileIDColName, files)
	textCol := entity.NewColumnVarChar(textColName, texts)
	if _, err := s.client.Insert(ctx, name, "" /* partitionName */, vectorCol, fileCol, textCol); err != nil {
		return err
	}
	return nil
}

// DeleteDocuments deletes documents from a collection in milvus by fileID.
func (s *S) DeleteDocuments(ctx context.Context, collectionName, fileID string) error {
	if err := s.client.LoadCollection(ctx, collectionName, false); err != nil {
		return fmt.Errorf("load collection: %s", err)
	}
	defer func() {
		if err := s.client.ReleaseCollection(ctx, collectionName); err != nil {
			log.Printf("Failed to release collection: %s\n", err)
		}
	}()

	expr := fmt.Sprintf("%s like \"%s\"", fileIDColName, fileID)
	return s.client.Delete(ctx, collectionName, "" /* partitionName */, expr)
}

// Search searches for the documents with similar vectors in milvus. The texts of the matched documents are returned.
func (s *S) Search(ctx context.Context, collectionName string, vectors []float32, numDocuments int) ([]string, error) {
	if err := s.client.LoadCollection(ctx, collectionName, false); err != nil {
		return nil, fmt.Errorf("load collection: %s", err)
	}
	defer func() {
		if err := s.client.ReleaseCollection(ctx, collectionName); err != nil {
			log.Printf("Failed to release collection: %s\n", err)
		}
	}()

	sp, err := entity.NewIndexIvfFlatSearchParam(defaultIvfFlatSearchParam)
	if err != nil {
		return nil, err
	}

	vs := []entity.Vector{entity.FloatVector(vectors)}
	results, err := s.client.Search(
		ctx,
		collectionName,
		nil, /* partitions */
		"",  /* expr */
		[]string{primaryKeyColName, fileIDColName, textColName},
		vs,
		vectorColName,
		defaultMetricType,
		numDocuments,
		sp,
	)
	if err != nil {
		return nil, err
	}

	var res []string
	for _, r := range results {
		// TODO(guangrui): Investigate the case when ResultCount is 0.
		if r.ResultCount == 0 {
			continue
		}
		texts, ok := r.Fields.GetColumn(textColName).(*entity.ColumnVarChar)
		if !ok {
			return nil, fmt.Errorf("%s column missing", textColName)
		}
		res = append(res, texts.Data()...)
	}
	return res, nil
}
