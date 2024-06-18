package store

import (
	"errors"
	"fmt"
	"testing"
	"time"

	gerrors "github.com/llm-operator/common/pkg/gormlib/errors"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestCreateAndGetCollection(t *testing.T) {
	st, teardown := NewTest(t)
	defer teardown()

	const (
		vectorStoreID  = "vs0"
		collectionID   = int64(1)
		collectionName = "collection0"
		project        = "project0"
	)

	_, err := st.GetCollectionByVectorStoreID(project, vectorStoreID)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, gorm.ErrRecordNotFound))

	c := &Collection{
		VectorStoreID: vectorStoreID,
		CollectionID:  collectionID,
		Name:          collectionName,
		Status:        CollectionStatusCompleted,
		ProjectID:     project,
	}
	err = st.CreateCollection(c)
	assert.NoError(t, err)

	got, err := st.GetCollectionByVectorStoreID(project, vectorStoreID)
	assert.NoError(t, err)
	assert.Equal(t, collectionID, got.CollectionID)

	got, err = st.GetCollectionByName(project, collectionName)
	assert.NoError(t, err)
	assert.Equal(t, collectionID, got.CollectionID)

	// Different project.
	_, err = st.GetCollectionByVectorStoreID("different", vectorStoreID)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, gorm.ErrRecordNotFound))
}

func TestCreateAndListJobs(t *testing.T) {
	st, teardown := NewTest(t)
	defer teardown()

	const (
		collectionID   = int64(1)
		collectionName = "collection"
		project        = "project0"
	)

	for i := 0; i < 3; i++ {
		c := Collection{
			VectorStoreID: fmt.Sprintf("vs-%d", i),
			CollectionID:  collectionID + int64(i),
			Name:          fmt.Sprintf("%s-%d", collectionName, i),
			Status:        CollectionStatusCompleted,
			ProjectID:     project,
		}
		err := st.CreateCollection(&c)
		assert.NoError(t, err)
	}
	for i := 0; i < 2; i++ {
		c := Collection{
			VectorStoreID: fmt.Sprintf("vs-%d", i+3),
			CollectionID:  collectionID + int64(i) + int64(100),
			Name:          fmt.Sprintf("%s-%d", collectionName, 100+i),
			Status:        CollectionStatusCompleted,
			ProjectID:     "unknown",
		}
		err := st.CreateCollection(&c)
		assert.NoError(t, err)
	}

	got, err := st.ListCollections(project)
	assert.NoError(t, err)
	assert.Len(t, got, 3)
}

func TestListCollectionsWithPagination(t *testing.T) {
	st, teardown := NewTest(t)
	defer teardown()

	const (
		collectionID   = int64(1)
		collectionName = "collection"
		project        = "project0"
	)

	var cs []Collection
	for i := 0; i < 10; i++ {
		c := Collection{
			VectorStoreID: fmt.Sprintf("vs-%d", i),
			CollectionID:  collectionID + int64(i),
			Name:          fmt.Sprintf("%s-%d", collectionName, i),
			Status:        CollectionStatusCompleted,
			ProjectID:     project,
		}
		err := st.CreateCollection(&c)
		assert.NoError(t, err)

		cs = append(cs, c)
	}

	got, hasMore, err := st.ListCollectionsWithPagination(project, time.Time{}, 0, "desc", 5)
	assert.NoError(t, err)
	assert.True(t, hasMore)
	assert.Len(t, got, 5)
	want := []int64{
		cs[9].CollectionID,
		cs[8].CollectionID,
		cs[7].CollectionID,
		cs[6].CollectionID,
		cs[5].CollectionID,
	}
	for i, c := range got {
		assert.Equal(t, want[i], c.CollectionID)
	}

	got, hasMore, err = st.ListCollectionsWithPagination(project, got[4].CreatedAt, got[4].ID, "", 2)
	assert.NoError(t, err)
	assert.True(t, hasMore)
	assert.Len(t, got, 2)
	want = []int64{
		cs[4].CollectionID,
		cs[3].CollectionID,
	}
	for i, c := range got {
		assert.Equal(t, want[i], c.CollectionID)
	}

	got, hasMore, err = st.ListCollectionsWithPagination(project, got[1].CreatedAt, got[1].ID, "", 3)
	assert.NoError(t, err)
	assert.False(t, hasMore)
	assert.Len(t, got, 3)
	want = []int64{
		cs[2].CollectionID,
		cs[1].CollectionID,
		cs[0].CollectionID,
	}
	for i, c := range got {
		assert.Equal(t, want[i], c.CollectionID)
	}

	got, hasMore, err = st.ListCollectionsWithPagination(project, time.Time{}, 0, "asc", 3)
	assert.NoError(t, err)
	assert.True(t, hasMore)
	assert.Len(t, got, 3)
	want = []int64{
		cs[0].CollectionID,
		cs[1].CollectionID,
		cs[2].CollectionID,
	}
	for i, c := range got {
		assert.Equal(t, want[i], c.CollectionID)
	}
}

func TestUpdateCollection(t *testing.T) {
	st, teardown := NewTest(t)
	defer teardown()

	const (
		collectionID   = int64(1)
		collectionName = "collection0"
		project        = "project0"
	)

	c := Collection{
		VectorStoreID: "vs0",
		CollectionID:  collectionID,
		Name:          collectionName,
		Status:        CollectionStatusCompleted,
		ProjectID:     project,
	}
	err := st.CreateCollection(&c)
	assert.NoError(t, err)

	nc := c
	nc.Name = "new name"
	nc.Status = CollectionStatusExpired
	err = st.UpdateCollection(&nc)
	assert.NoError(t, err)

	got, err := st.GetCollectionByVectorStoreID(project, c.VectorStoreID)
	assert.NoError(t, err)
	assert.Equal(t, nc.Name, got.Name)
	assert.Equal(t, nc.Status, got.Status)
}

func TestDeleteCollection(t *testing.T) {
	st, teardown := NewTest(t)
	defer teardown()

	const (
		vectorStoreID  = "vs0"
		collectionID   = int64(1)
		collectionName = "collection0"
		project        = "project0"
	)

	c := Collection{
		VectorStoreID: vectorStoreID,
		CollectionID:  collectionID,
		Name:          collectionName,
		Status:        CollectionStatusCompleted,
		ProjectID:     project,
	}
	err := st.CreateCollection(&c)
	assert.NoError(t, err)

	err = st.DeleteCollection(project, vectorStoreID)
	assert.NoError(t, err)

	_, err = st.GetCollectionByVectorStoreID(project, vectorStoreID)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, gorm.ErrRecordNotFound))
}

func TestCreateCollection_SameName(t *testing.T) {
	st, teardown := NewTest(t)
	defer teardown()

	const (
		collectionName = "collection0"
		project        = "project0"
	)

	err := st.CreateCollection(&Collection{
		VectorStoreID: "vs0",
		CollectionID:  int64(1),
		Name:          collectionName,
		Status:        CollectionStatusCompleted,
		ProjectID:     project,
	})
	assert.NoError(t, err)

	err = st.CreateCollection(&Collection{
		VectorStoreID: "vs1",
		CollectionID:  int64(2),
		Name:          collectionName,
		Status:        CollectionStatusCompleted,
		ProjectID:     project,
	})
	assert.Error(t, err)
	assert.True(t, gerrors.IsUniqueConstraintViolation(err))
}
