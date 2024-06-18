package store

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

// CollectionStatus represents the status of a collection.
type CollectionStatus string

// ExpiresAfterAnchor represents the anchor for the expiration time.
type ExpiresAfterAnchor string

const (
	// CollectionStatusExpired represents the expired status.
	CollectionStatusExpired CollectionStatus = "expired"
	// CollectionStatusInProgress represents the in_progress status.
	CollectionStatusInProgress CollectionStatus = "in_progress"
	// CollectionStatusCompleted represents the completed status.
	CollectionStatusCompleted CollectionStatus = "completed"

	// ExpiresAfterAnchorLastActiveAt represents the anchor for the expiration time based on the last_active_at time.
	ExpiresAfterAnchorLastActiveAt ExpiresAfterAnchor = "last_active_at"
)

// Collection represents a collection.
type Collection struct {
	gorm.Model

	// VectorStoreID is the ID of the vector store that is externally visible in the API.
	// This is also used as the name of the Milvus collection.
	VectorStoreID string `gorm:"uniqueIndex"`

	// CollectionID is the ID of the Milvus collection.
	CollectionID int64 `gorm:"uniqueIndex"`

	TenantID       string
	OrganizationID string
	ProjectID      string `gorm:"uniqueIndex:idx_collection_project_id_name"`

	Name string `gorm:"uniqueIndex:idx_collection_project_id_name"`

	// UsageBytes is the total number of bytes used by the files in the vector store.
	UsageBytes int64

	FileCountsInProgress int64
	FileCountsCompleted  int64
	FileCountsFailed     int64
	FileCountsCancelled  int64
	FileCountsTotal      int64

	Status CollectionStatus

	Anchor ExpiresAfterAnchor
	// ExpiresAfterDays is the number of days the anchor time for when the vector store will expire.
	ExpiresAfterDays int64
	// ExpiresAt is the Unix timestamp (in seconds) for when the vector store will expire.
	ExpiresAt int64

	// LastActiveAt is the Unix timestamp (in seconds) for when the vector store was last active.
	LastActiveAt int64

	EmbeddingModel      string
	EmbeddingDimensions int

	Version int
}

// CreateCollection creates a new collection.
func (s *S) CreateCollection(c *Collection) error {
	if err := s.db.Create(c).Error; err != nil {
		return err
	}
	return nil
}

// GetCollectionByVectorStoreID gets a collection.
func (s *S) GetCollectionByVectorStoreID(projectID string, vectorStoreID string) (*Collection, error) {
	var c Collection
	if err := s.db.Where("vector_store_id = ? AND project_id = ?", vectorStoreID, projectID).Take(&c).Error; err != nil {
		return nil, err
	}
	return &c, nil
}

// GetCollectionByName gets a collection.
func (s *S) GetCollectionByName(projectID, name string) (*Collection, error) {
	var c Collection
	if err := s.db.Where("name = ? AND project_id = ?", name, projectID).Take(&c).Error; err != nil {
		return nil, err
	}
	return &c, nil
}

// ListCollections lists collections.
func (s *S) ListCollections(projectID string) ([]*Collection, error) {
	var cs []*Collection
	if err := s.db.Where("project_id = ?", projectID).Order("collection_id").Find(&cs).Error; err != nil {
		return nil, err
	}
	return cs, nil
}

// ListCollectionsWithPagination finds collections with pagination. Collections are returned in the order of VectorStoreID.
func (s *S) ListCollectionsWithPagination(
	projectID string,
	afterCreatedAt time.Time,
	afterID uint,
	order string,
	limit int,
) ([]*Collection, bool, error) {
	var cs []*Collection
	q := s.db.Where("project_id = ?", projectID)
	isAsc := order == "asc"

	if afterCreatedAt != (time.Time{}) {
		if isAsc {
			q = q.Where("(created_at > ? OR (created_at = ? AND id > ?))", afterCreatedAt, afterCreatedAt, afterID)
		} else {
			q = q.Where("(created_at < ? OR (created_at = ? AND id < ?))", afterCreatedAt, afterCreatedAt, afterID)
		}
	}

	if isAsc {
		order = "created_at ASC, id ASC"
	} else {
		order = "created_at DESC, id DESC"
	}

	if err := q.Order(order).Limit(limit + 1).Find(&cs).Error; err != nil {
		return nil, false, err
	}

	var hasMore bool
	if len(cs) > limit {
		cs = cs[:limit]
		hasMore = true
	}
	return cs, hasMore, nil
}

// UpdateCollection updates the collection.
func (s *S) UpdateCollection(nc *Collection) error {
	result := s.db.Model(&Collection{}).
		Where("id = ?", nc.ID).
		Where("version = ?", nc.Version).
		Updates(map[string]interface{}{
			"name":               nc.Name,
			"status":             nc.Status,
			"expires_after_days": nc.ExpiresAfterDays,
			"expires_at":         nc.ExpiresAt,
			"version":            nc.Version + 1,
		})
	if err := result.Error; err != nil {
		return err
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("update collection: %w", ErrConcurrentUpdate)
	}
	return nil
}

// DeleteCollection deletes the collection.
func (s *S) DeleteCollection(projectID string, vectorStoreID string) error {
	result := s.db.Unscoped().
		Where("vector_store_id = ?", vectorStoreID).
		Where("project_id = ?", projectID).
		Delete(&Collection{})
	if err := result.Error; err != nil {
		return err
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
