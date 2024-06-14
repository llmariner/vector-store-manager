package store

import (
	"fmt"

	"gorm.io/gorm"
)

// CollectionMetadata represents the metadata of a collection.
type CollectionMetadata struct {
	gorm.Model

	// CollectionID is the ID of the collection in vector store.
	CollectionID int64 `gorm:"uniqueIndex:idx_collectionmeta_collection_id_key"`

	Key   string `gorm:"uniqueIndex:idx_collectionmeta_collection_id_key"`
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

// ListCollectionMetadataByCollectionID lists metadata of a collections.
func (s *S) ListCollectionMetadataByCollectionID(collectionID int64) ([]*CollectionMetadata, error) {
	var cms []*CollectionMetadata
	if err := s.db.Where("collection_id = ?", collectionID).Order("key").Find(&cms).Error; err != nil {
		return nil, err
	}
	return cms, nil
}

// UpdateCollectionMetadata updates the metadata of a collection.
func (s *S) UpdateCollectionMetadata(cm *CollectionMetadata) error {
	result := s.db.Model(&CollectionMetadata{}).
		Where("collection_id = ?", cm.CollectionID).
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
func (s *S) DeleteCollectionMetadata(collectionID int64, key string) error {
	result := s.db.Unscoped().
		Where("collection_id = ?", collectionID).
		Where("key = ?", key).
		Delete(&CollectionMetadata{})
	if err := result.Error; err != nil {
		return err
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// DeleteCollectionMetadataByCollectionID deletes all metadata of the collection.
func (s *S) DeleteCollectionMetadataByCollectionID(collectionID int64) error {
	if err := s.db.Unscoped().
		Where("collection_id = ?", collectionID).
		Delete(&CollectionMetadata{}).Error; err != nil {
		return err
	}
	return nil
}
