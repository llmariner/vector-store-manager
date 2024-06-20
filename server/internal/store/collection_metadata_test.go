package store

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestCreateAndListCollectionMetadata(t *testing.T) {
	st, teardown := NewTest(t)
	defer teardown()

	const (
		vectorStoreID = "vs1"
	)

	for i := 0; i < 3; i++ {
		cm := CollectionMetadata{
			VectorStoreID: vectorStoreID,
			Key:           fmt.Sprintf("key%d", i),
			Value:         fmt.Sprintf("value%d", i),
		}
		err := st.CreateCollectionMetadata(&cm)
		assert.NoError(t, err)
	}
	for i := 0; i < 2; i++ {
		cm := CollectionMetadata{
			VectorStoreID: fmt.Sprintf("different-%d", i),
			Key:           fmt.Sprintf("key%d", i),
			Value:         fmt.Sprintf("value%d", i),
		}
		err := st.CreateCollectionMetadata(&cm)
		assert.NoError(t, err)
	}

	got, err := st.ListCollectionMetadataByVectorStoreID(vectorStoreID)
	assert.NoError(t, err)
	assert.Len(t, got, 3)
}

func TestUpdateCollectionMetadata(t *testing.T) {
	st, teardown := NewTest(t)
	defer teardown()

	const (
		vectorStoreID = "vs1"
	)

	cm := CollectionMetadata{
		VectorStoreID: vectorStoreID,
		Key:           "key0",
		Value:         "value0",
	}
	err := st.CreateCollectionMetadata(&cm)
	assert.NoError(t, err)

	ncm := cm
	ncm.Value = "new value"
	err = st.UpdateCollectionMetadata(&ncm)
	assert.NoError(t, err)

	got, err := st.ListCollectionMetadataByVectorStoreID(vectorStoreID)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(got))
	assert.Equal(t, ncm.Key, got[0].Key)
	assert.Equal(t, ncm.Value, got[0].Value)
}

func TestDeleteCollectionMetadata(t *testing.T) {
	st, teardown := NewTest(t)
	defer teardown()

	const (
		vectorStoreID = "vs0"
	)
	var cms []CollectionMetadata
	for i := 0; i < 3; i++ {
		cm := CollectionMetadata{
			VectorStoreID: vectorStoreID,
			Key:           fmt.Sprintf("key%d", i),
			Value:         fmt.Sprintf("value%d", i),
		}
		err := st.CreateCollectionMetadata(&cm)
		assert.NoError(t, err)

		cms = append(cms, cm)
	}

	err := st.DeleteCollectionMetadata(cms[0].ID)
	assert.NoError(t, err)

	got, err := st.ListCollectionMetadataByVectorStoreID(vectorStoreID)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(got))

	err = st.DeleteAllCollectionMetadatasByVectorStoreID(vectorStoreID)
	assert.NoError(t, err)

	err = st.DeleteCollectionMetadata(cms[1].ID)
	assert.True(t, errors.Is(err, gorm.ErrRecordNotFound))

	got, err = st.ListCollectionMetadataByVectorStoreID(vectorStoreID)
	assert.NoError(t, err)
	assert.Empty(t, got)
}
