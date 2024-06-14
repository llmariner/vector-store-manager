package store

import (
	"fmt"

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

	// CollectionID is the ID of the collection in vector store.
	CollectionID int64 `gorm:"uniqueIndex"`

	Name string

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

	TenantID string

	OrganizationID string
	ProjectID      string `gorm:"index"`

	Version int
}

// CreateCollection creates a new collection.
func (s *S) CreateCollection(c *Collection) error {
	if err := s.db.Create(c).Error; err != nil {
		return err
	}
	return nil
}

// GetCollectionByCollectionID gets a collection.
func (s *S) GetCollectionByCollectionID(projectID string, collectionID int64) (*Collection, error) {
	var c Collection
	if err := s.db.Where("collection_id = ? AND project_id = ?", collectionID, projectID).Take(&c).Error; err != nil {
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

// ListCollectionsWithPagination finds collections with pagination. Collections are returned in the order of CollectionID.
func (s *S) ListCollectionsWithPagination(projectID string, afterID int64, order string, limit int) ([]*Collection, bool, error) {
	var cs []*Collection
	q := s.db.Where("project_id = ?", projectID)
	if order == "asc" {
		order = "collection_id ASC"
	} else {
		order = "collection_id DESC"
	}
	if afterID >= 0 {
		if order == "collection_id ASC" {
			q = q.Where("collection_id > ?", afterID)
		} else {
			q = q.Where("collection_id < ?", afterID)
		}
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
		Where("collection_id = ?", nc.CollectionID).
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
func (s *S) DeleteCollection(projectID string, collectionID int64) error {
	result := s.db.Unscoped().
		Where("collection_id = ?", collectionID).
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
