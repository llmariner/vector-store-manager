package store

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestCreateAndGetFile(t *testing.T) {
	st, teardown := NewTest(t)
	defer teardown()

	fileID := "file0"
	vectorStoreID := "vector_store_0"
	project := "project0"
	_, err := st.GetFileByFileID(project, vectorStoreID, fileID)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, gorm.ErrRecordNotFound))

	f := &File{
		FileID:        fileID,
		VectorStoreID: vectorStoreID,
		Status:        FileStatusCompleted,
		ProjectID:     project,
	}
	err = st.CreateFile(f)
	assert.NoError(t, err)

	got, err := st.GetFileByFileID(project, vectorStoreID, fileID)
	assert.NoError(t, err)
	assert.Equal(t, fileID, got.FileID)

	// Different tenant.
	_, err = st.GetFileByFileID("unknow", vectorStoreID, fileID)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, gorm.ErrRecordNotFound))
}

func TestCreateAndListFiles(t *testing.T) {
	st, teardown := NewTest(t)
	defer teardown()

	const (
		fileID        = "fileID"
		vectorStoreID = "vector_store_0"
		project       = "project0"
	)

	for i := 0; i < 3; i++ {
		f := File{
			FileID:        fmt.Sprintf("%s%d", fileID, i),
			VectorStoreID: vectorStoreID,
			Status:        FileStatusCompleted,
			ProjectID:     project,
		}
		err := st.CreateFile(&f)
		assert.NoError(t, err)
	}
	for i := 0; i < 2; i++ {
		f := File{
			FileID:        fmt.Sprintf("%s%d", fileID, i),
			VectorStoreID: "unknown",
			Status:        FileStatusCompleted,
			ProjectID:     project,
		}
		err := st.CreateFile(&f)
		assert.NoError(t, err)
	}

	got, err := st.ListFiles(project, vectorStoreID)
	assert.NoError(t, err)
	assert.Len(t, got, 3)
}

func TestListFilesWithPagination(t *testing.T) {
	st, teardown := NewTest(t)
	defer teardown()

	const (
		fileID        = "file"
		vectorStoreID = "vs0"
		project       = "project0"
	)

	for i := 0; i < 10; i++ {
		f := File{
			FileID:        fmt.Sprintf("%s%d", fileID, i),
			VectorStoreID: vectorStoreID,
			Status:        FileStatusCompleted,
			ProjectID:     project,
		}
		err := st.CreateFile(&f)
		assert.NoError(t, err)
	}

	got, hasMore, err := st.ListFilesWithPagination(project, vectorStoreID, "zzz", "desc", 5)
	assert.NoError(t, err)
	assert.True(t, hasMore)
	assert.Len(t, got, 5)
	want := []int64{9, 8, 7, 6, 5}
	for i, f := range got {
		assert.Equal(t, fmt.Sprintf("%s%d", fileID, want[i]), f.FileID)
	}

	got, hasMore, err = st.ListFilesWithPagination(project, vectorStoreID, got[4].FileID, "", 2)
	assert.NoError(t, err)
	assert.True(t, hasMore)
	assert.Len(t, got, 2)
	want = []int64{4, 3}
	for i, f := range got {
		assert.Equal(t, fmt.Sprintf("%s%d", fileID, want[i]), f.FileID)
	}

	got, hasMore, err = st.ListFilesWithPagination(project, vectorStoreID, got[1].FileID, "", 3)
	assert.NoError(t, err)
	assert.False(t, hasMore)
	assert.Len(t, got, 3)
	want = []int64{2, 1, 0}
	for i, f := range got {
		assert.Equal(t, fmt.Sprintf("%s%d", fileID, want[i]), f.FileID)
	}

	got, hasMore, err = st.ListFilesWithPagination(project, vectorStoreID, "", "asc", 3)
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
		project       = "project0"
	)

	f := File{
		FileID:        fileID,
		VectorStoreID: vectorStoreID,
		Status:        FileStatusCompleted,
		ProjectID:     project,
	}
	err := st.CreateFile(&f)
	assert.NoError(t, err)

	err = st.DeleteFile(project, vectorStoreID, fileID)
	assert.NoError(t, err)

	_, err = st.GetFileByFileID(project, vectorStoreID, fileID)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, gorm.ErrRecordNotFound))
}
