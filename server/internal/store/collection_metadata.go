package store

import (
	"fmt"

	"gorm.io/gorm"
)

// CollectionMetadata represents the metadata of a collection.
type CollectionMetadata struct {
	gorm.Model

	// VectorStoreID is the ID of the vector store.
	VectorStoreID string `gorm:"uniqueIndex:idx_collectionmeta_vsid_key"`

	Key   string `gorm:"uniqueIndex:idx_collectionmeta_vsid_key"`
	Value string

	Version int
}

// CreateCollectionMetadata creates a new collection metadata.
func (s *S) CreateCollectionMetadata(cm *CollectionMetadata) error {
	if err := s.db.Create(cm).Error; err != nil {
		return err
	}
	return nil
}

// ListCollectionMetadataByVectorStoreID lists metadata of a collections.
func (s *S) ListCollectionMetadataByVectorStoreID(vectorStoreID string) ([]*CollectionMetadata, error) {
	var cms []*CollectionMetadata
	if err := s.db.Where("vector_store_id = ?", vectorStoreID).Order("key").Find(&cms).Error; err != nil {
		return nil, err
	}
	return cms, nil
}

// UpdateCollectionMetadata updates the metadata of a collection.
func (s *S) UpdateCollectionMetadata(cm *CollectionMetadata) error {
	result := s.db.Model(&CollectionMetadata{}).
		Where("id = ?", cm.ID).
		Where("version = ?", cm.Version).
		Where("key = ?", cm.Key).
		Updates(map[string]interface{}{
			"value":   cm.Value,
			"version": cm.Version + 1,
		})
	if err := result.Error; err != nil {
		return err
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("update collection metadata: %w", ErrConcurrentUpdate)
	}
	return nil
}

// DeleteCollectionMetadata deletes the metadata for the collection.
func (s *S) DeleteCollectionMetadata(id uint) error {
	result := s.db.Unscoped().
		Where("id = ?", id).
		Delete(&CollectionMetadata{})
	if err := result.Error; err != nil {
		return err
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// DeleteCollectionMetadatasByVectorStoreID deletes all metadata of the collection.
func (s *S) DeleteCollectionMetadatasByVectorStoreID(vectorStoreID string) error {
	if err := s.db.Unscoped().
		Where("vector_store_id = ?", vectorStoreID).
		Delete(&CollectionMetadata{}).Error; err != nil {
		return err
	}
	return nil
}
