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

	collectionID := int64(1)

	for i := 0; i < 3; i++ {
		cm := CollectionMetadata{
			CollectionID: collectionID,
			Key:          fmt.Sprintf("key%d", i),
			Value:        fmt.Sprintf("value%d", i),
		}
		err := st.CreateCollectionMetadata(&cm)
		assert.NoError(t, err)
	}
	for i := 0; i < 2; i++ {
		cm := CollectionMetadata{
			CollectionID: collectionID + int64(100),
			Key:          fmt.Sprintf("key%d", i),
			Value:        fmt.Sprintf("value%d", i),
		}
		err := st.CreateCollectionMetadata(&cm)
		assert.NoError(t, err)
	}

	got, err := st.ListCollectionMetadataByCollectionID(collectionID)
	assert.NoError(t, err)
	assert.Len(t, got, 3)
}

func TestUpdateCollectionMetadata(t *testing.T) {
	st, teardown := NewTest(t)
	defer teardown()

	collectionID := int64(1)
	cm := CollectionMetadata{
		CollectionID: collectionID,
		Key:          "key0",
		Value:        "value0",
	}
	err := st.CreateCollectionMetadata(&cm)
	assert.NoError(t, err)

	ncm := cm
	ncm.Value = "new value"
	err = st.UpdateCollectionMetadata(&ncm)
	assert.NoError(t, err)

	got, err := st.ListCollectionMetadataByCollectionID(collectionID)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(got))
	assert.Equal(t, ncm.Key, got[0].Key)
	assert.Equal(t, ncm.Value, got[0].Value)
}

func TestDeleteCollectionMetadata(t *testing.T) {
	st, teardown := NewTest(t)
	defer teardown()

	collectionID := int64(1)
	for i := 0; i < 3; i++ {
		cm := CollectionMetadata{
			CollectionID: collectionID,
			Key:          fmt.Sprintf("key%d", i),
			Value:        fmt.Sprintf("value%d", i),
		}
		err := st.CreateCollectionMetadata(&cm)
		assert.NoError(t, err)
	}

	err := st.DeleteCollectionMetadata(collectionID, "key0")
	assert.NoError(t, err)

	got, err := st.ListCollectionMetadataByCollectionID(collectionID)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(got))

	err = st.DeleteCollectionMetadataByCollectionID(collectionID)
	assert.NoError(t, err)

	err = st.DeleteCollectionMetadata(collectionID, "key0")
	assert.True(t, errors.Is(err, gorm.ErrRecordNotFound))

	got, err = st.ListCollectionMetadataByCollectionID(collectionID)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(got))
}
