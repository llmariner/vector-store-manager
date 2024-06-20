package store

import (
	"gorm.io/gorm"
)

// FileStatus represents the status of a file.
type FileStatus string

// LastCodeError represents the error code of the last error.
type LastErrorCode string

// ChunkingStrategyType represents the type of chunking strategy.
type ChunkingStrategyType string

const (
	// FileStatusInProgress represents the in_progress status.
	FileStatusInProgress FileStatus = "in_progress"
	// FileStatusCompleted represents the completed status.
	FileStatusCompleted FileStatus = "completed"
	// FileStatusExpired represents the failed status.
	FileStatusFailed FileStatus = "failed"
	// FileStatusCancelled represents the cancelled status.
	FileStatusCancelled FileStatus = "cancelled"

	// LastErrorCodeNone represents no error.
	LastErrorCodeNone LastErrorCode = ""
	// LastErrorCodeServerError represents a server error.
	LastErrorCodeServerError LastErrorCode = "server_error"
	// LastErrorCodeRateLimitExceeded represents a rate limit exceeded error.
	LastErrorCodeRateLimitExceeded LastErrorCode = "rate_limit_exceeded"

	// ChunkingStrategyTypeAuto represents the auto chunking strategy.
	ChunkingStrategyTypeAuto ChunkingStrategyType = "auto"
	// ChunkingStrategyTypeStatic represents the static chunking strategy.
	ChunkingStrategyTypeStatic ChunkingStrategyType = "static"
	// ChunkingStrategyTypeOther represents the other chunking strategy.
	ChunkingStrategyTypeOther ChunkingStrategyType = "other"
)

// File represents a file.
type File struct {
	gorm.Model

	// FileID is the file ID.
	FileID string `gorm:"uniqueIndex:idx_file_file_id_vector_store_id"`

	TenantID       string
	OrganizationID string
	ProjectID      string `gorm:"index:idx_file_project_id"`

	VectorStoreID string `gorm:"uniqueIndex:idx_file_file_id_vector_store_id"`

	// UsageBytes is the total vector store usage in bytes. Note that this may be different from the original file size.
	UsageBytes int64

	Status FileStatus

	LastErrorCode    LastErrorCode
	LastErrorMessage string

	ChunkingStrategyType ChunkingStrategyType
	MaxChunkSizeTokens   int64
	ChunkOverlapTokens   int64

	Version int
}

// CreateFile creates a new file.
func (s *S) CreateFile(f *File) error {
	if err := s.db.Create(f).Error; err != nil {
		return err
	}
	return nil
}

// GetFileByFileID gets a file.
func (s *S) GetFileByFileID(projectID, vectorStoreID, fileID string) (*File, error) {
	var f File
	if err := s.db.Where("file_id = ?", fileID).
		Where("vector_store_id = ?", vectorStoreID).
		Where("project_id = ?", projectID).
		Take(&f).Error; err != nil {
		return nil, err
	}
	return &f, nil
}

// ListFiles lists files.
func (s *S) ListFiles(projectID, vectorStoreID string) ([]*File, error) {
	var fs []*File
	if err := s.db.Where("project_id = ?", projectID).
		Where("vector_store_id = ?", vectorStoreID).
		Order("file_id").Find(&fs).Error; err != nil {
		return nil, err
	}
	return fs, nil
}

// ListFilesWithPagination finds files with pagination. Files are returned in the order of created_at.
func (s *S) ListFilesWithPagination(projectID, vectorStoreID, afterID, order string, limit int) ([]*File, bool, error) {
	var fs []*File
	q := s.db.Where("project_id = ? AND vector_store_id = ?", projectID, vectorStoreID)
	if order == "asc" {
		order = "created_at ASC"
	} else {
		order = "created_at DESC"
	}
	if afterID != "" {
		if order == "file_id ASC" {
			q = q.Where("file_id > ?", afterID)
		} else {
			q = q.Where("file_id < ?", afterID)
		}
	}
	if err := q.Order(order).Limit(limit + 1).Find(&fs).Error; err != nil {
		return nil, false, err
	}

	var hasMore bool
	if len(fs) > limit {
		fs = fs[:limit]
		hasMore = true
	}
	return fs, hasMore, nil
}

// DeleteFile deletes the file.
func (s *S) DeleteFile(projectID, vectorStoreID, fileID string) error {
	result := s.db.Unscoped().
		Where("file_id = ?", fileID).
		Where("vector_store_id = ?", vectorStoreID).
		Where("project_id = ?", projectID).
		Delete(&File{})
	if err := result.Error; err != nil {
		return err
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
