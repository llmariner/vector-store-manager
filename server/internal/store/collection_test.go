package store

import (
	"errors"
	"fmt"
	"testing"

	gerrors "github.com/llm-operator/common/pkg/gormlib/errors"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestCreateAndGetCollection(t *testing.T) {
	st, teardown := NewTest(t)
	defer teardown()

	collectionID := int64(1)
	collectionName := "collection0"
	project := "project0"
	_, err := st.GetCollectionByCollectionID(project, collectionID)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, gorm.ErrRecordNotFound))

	c := &Collection{
		CollectionID: collectionID,
		Name:         collectionName,
		Status:       CollectionStatusCompleted,
		ProjectID:    project,
	}
	err = st.CreateCollection(c)
	assert.NoError(t, err)

	got, err := st.GetCollectionByCollectionID(project, collectionID)
	assert.NoError(t, err)
	assert.Equal(t, collectionID, got.CollectionID)

	got, err = st.GetCollectionByName(project, collectionName)
	assert.NoError(t, err)
	assert.Equal(t, collectionID, got.CollectionID)

	// Different tenant.
	_, err = st.GetCollectionByCollectionID("unknow", collectionID)
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
			CollectionID: collectionID + int64(i),
			Name:         fmt.Sprintf("%s-%d", collectionName, i),
			Status:       CollectionStatusCompleted,
			ProjectID:    project,
		}
		err := st.CreateCollection(&c)
		assert.NoError(t, err)
	}
	for i := 0; i < 2; i++ {
		c := Collection{
			CollectionID: collectionID + int64(i) + int64(100),
			Name:         fmt.Sprintf("%s-%d", collectionName, 100+i),
			Status:       CollectionStatusCompleted,
			ProjectID:    "unknown",
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

	for i := 0; i < 10; i++ {
		c := Collection{
			CollectionID: collectionID + int64(i),
			Name:         fmt.Sprintf("%s-%d", collectionName, i),
			Status:       CollectionStatusCompleted,
			ProjectID:    project,
		}
		err := st.CreateCollection(&c)
		assert.NoError(t, err)
	}

	got, hasMore, err := st.ListCollectionsWithPagination(project, 100, "desc", 5)
	assert.NoError(t, err)
	assert.True(t, hasMore)
	assert.Len(t, got, 5)
	want := []int64{10, 9, 8, 7, 6}
	for i, c := range got {
		assert.Equal(t, want[i], c.CollectionID)
	}

	got, hasMore, err = st.ListCollectionsWithPagination(project, got[4].CollectionID, "", 2)
	assert.NoError(t, err)
	assert.True(t, hasMore)
	assert.Len(t, got, 2)
	want = []int64{5, 4}
	for i, c := range got {
		assert.Equal(t, want[i], c.CollectionID)
	}

	got, hasMore, err = st.ListCollectionsWithPagination(project, got[1].CollectionID, "", 3)
	assert.NoError(t, err)
	assert.False(t, hasMore)
	assert.Len(t, got, 3)
	want = []int64{3, 2, 1}
	for i, c := range got {
		assert.Equal(t, want[i], c.CollectionID)
	}

	got, hasMore, err = st.ListCollectionsWithPagination(project, 0, "asc", 3)
	assert.NoError(t, err)
	assert.True(t, hasMore)
	assert.Len(t, got, 3)
	want = []int64{1, 2, 3}
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
		CollectionID: collectionID,
		Name:         collectionName,
		Status:       CollectionStatusCompleted,
		ProjectID:    project,
	}
	err := st.CreateCollection(&c)
	assert.NoError(t, err)

	nc := c
	nc.Name = "new name"
	nc.Status = CollectionStatusExpired
	err = st.UpdateCollection(&nc)
	assert.NoError(t, err)

	got, err := st.GetCollectionByCollectionID(project, collectionID)
	assert.NoError(t, err)
	assert.Equal(t, nc.Name, got.Name)
	assert.Equal(t, nc.Status, got.Status)
}

func TestDeleteCollection(t *testing.T) {
	st, teardown := NewTest(t)
	defer teardown()

	const (
		collectionID   = int64(1)
		collectionName = "collection0"
		project        = "project0"
	)

	c := Collection{
		CollectionID: collectionID,
		Name:         collectionName,
		Status:       CollectionStatusCompleted,
		ProjectID:    project,
	}
	err := st.CreateCollection(&c)
	assert.NoError(t, err)

	err = st.DeleteCollection(project, collectionID)
	assert.NoError(t, err)

	_, err = st.GetCollectionByCollectionID(project, collectionID)
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
		CollectionID: int64(1),
		Name:         collectionName,
		Status:       CollectionStatusCompleted,
		ProjectID:    project,
	})
	assert.NoError(t, err)

	err = st.CreateCollection(&Collection{
		CollectionID: int64(2),
		Name:         collectionName,
		Status:       CollectionStatusCompleted,
		ProjectID:    project,
	})
	assert.Error(t, err)
	assert.True(t, gerrors.IsUniqueConstraintViolation(err))
}
