package store

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestCreateAndGetFile(t *testing.T) {
	st, teardown := NewTest(t)
	defer teardown()

	fileID := "file0"
	vectorStoreID := "vector_store_0"
	_, err := st.GetFileByFileID(vectorStoreID, fileID)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, gorm.ErrRecordNotFound))

	f := &File{
		FileID:        fileID,
		VectorStoreID: vectorStoreID,
		Status:        FileStatusCompleted,
	}
	err = st.CreateFile(f)
	assert.NoError(t, err)

	got, err := st.GetFileByFileID(vectorStoreID, fileID)
	assert.NoError(t, err)
	assert.Equal(t, fileID, got.FileID)

	// Different vector store.
	_, err = st.GetFileByFileID("different", fileID)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, gorm.ErrRecordNotFound))
}

func TestCreateAndListFiles(t *testing.T) {
	st, teardown := NewTest(t)
	defer teardown()

	const (
		fileID        = "fileID"
		vectorStoreID = "vector_store_0"
	)

	for i := 0; i < 3; i++ {
		f := File{
			FileID:        fmt.Sprintf("%s%d", fileID, i),
			VectorStoreID: vectorStoreID,
			Status:        FileStatusCompleted,
		}
		err := st.CreateFile(&f)
		assert.NoError(t, err)
	}
	for i := 0; i < 2; i++ {
		f := File{
			FileID:        fmt.Sprintf("%s%d", fileID, i),
			VectorStoreID: "unknown",
			Status:        FileStatusCompleted,
		}
		err := st.CreateFile(&f)
		assert.NoError(t, err)
	}

	got, err := st.ListFiles(vectorStoreID)
	assert.NoError(t, err)
	assert.Len(t, got, 3)
}

func TestListFilesWithPagination(t *testing.T) {
	st, teardown := NewTest(t)
	defer teardown()

	const (
		fileID        = "file"
		vectorStoreID = "vs0"
	)

	for i := 0; i < 10; i++ {
		f := File{
			FileID:        fmt.Sprintf("%s%d", fileID, i),
			VectorStoreID: vectorStoreID,
			Status:        FileStatusCompleted,
		}
		err := st.CreateFile(&f)
		assert.NoError(t, err)
	}

	got, hasMore, err := st.ListFilesWithPagination(vectorStoreID, time.Time{}, 0, "desc", 5)
	assert.NoError(t, err)
	assert.True(t, hasMore)
	assert.Len(t, got, 5)
	want := []int64{9, 8, 7, 6, 5}
	for i, f := range got {
		assert.Equal(t, fmt.Sprintf("%s%d", fileID, want[i]), f.FileID)
	}

	got, hasMore, err = st.ListFilesWithPagination(vectorStoreID, got[4].CreatedAt, got[4].ID, "", 2)
	assert.NoError(t, err)
	assert.True(t, hasMore)
	assert.Len(t, got, 2)
	want = []int64{4, 3}
	for i, f := range got {
		assert.Equal(t, fmt.Sprintf("%s%d", fileID, want[i]), f.FileID)
	}

	got, hasMore, err = st.ListFilesWithPagination(vectorStoreID, got[1].CreatedAt, got[1].ID, "", 3)
	assert.NoError(t, err)
	assert.False(t, hasMore)
	assert.Len(t, got, 3)
	want = []int64{2, 1, 0}
	for i, f := range got {
		assert.Equal(t, fmt.Sprintf("%s%d", fileID, want[i]), f.FileID)
	}

	got, hasMore, err = st.ListFilesWithPagination(vectorStoreID, time.Time{}, 0, "asc", 3)
	assert.NoError(t, err)
	assert.True(t, hasMore)
	assert.Len(t, got, 3)
	want = []int64{0, 1, 2}
	for i, f := range got {
		assert.Equal(t, fmt.Sprintf("%s%d", fileID, want[i]), f.FileID)
	}
}

func TestDeleteFile(t *testing.T) {
	st, teardown := NewTest(t)
	defer teardown()

	const (
		fileID        = "file0"
		vectorStoreID = "vs0"
	)

	f := File{
		FileID:        fileID,
		VectorStoreID: vectorStoreID,
		Status:        FileStatusCompleted,
	}
	err := st.CreateFile(&f)
	assert.NoError(t, err)

	err = st.DeleteFile(vectorStoreID, fileID)
	assert.NoError(t, err)

	_, err = st.GetFileByFileID(vectorStoreID, fileID)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, gorm.ErrRecordNotFound))
}

func TestDeleteAllFilesByVectorStoreID(t *testing.T) {
	st, teardown := NewTest(t)
	defer teardown()

	const (
		fileID         = "file0"
		vectorStoreID0 = "vs0"
		vectorStoreID1 = "vs1"
	)

	f0 := File{
		FileID:        fileID,
		VectorStoreID: vectorStoreID0,
	}
	err := st.CreateFile(&f0)
	assert.NoError(t, err)

	f1 := File{
		FileID:        fileID,
		VectorStoreID: vectorStoreID1,
	}
	err = st.CreateFile(&f1)
	assert.NoError(t, err)

	err = st.DeleteAllFilesByVectorStoreID(vectorStoreID0)
	assert.NoError(t, err)

	_, err = st.GetFileByFileID(vectorStoreID0, fileID)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, gorm.ErrRecordNotFound))

	_, err = st.GetFileByFileID(vectorStoreID0, fileID)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, gorm.ErrRecordNotFound))

	_, err = st.GetFileByFileID(vectorStoreID1, fileID)
	assert.NoError(t, err)
}
