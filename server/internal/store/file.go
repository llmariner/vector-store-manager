package store

import (
	"time"

	"gorm.io/gorm"
)

// FileStatus represents the status of a file.
type FileStatus string

// LastErrorCode represents the error code of the last error.
type LastErrorCode string

// ChunkingStrategyType represents the type of chunking strategy.
type ChunkingStrategyType string

const (
	// FileStatusInProgress represents the in_progress status.
	FileStatusInProgress FileStatus = "in_progress"
	// FileStatusCompleted represents the completed status.
	FileStatusCompleted FileStatus = "completed"
	// FileStatusFailed represents the failed status.
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
)

// File represents a file.
type File struct {
	gorm.Model

	VectorStoreID string `gorm:"uniqueIndex:idx_file_vector_store_id_file_id"`

	// FileID is the file ID.
	FileID string `gorm:"uniqueIndex:idx_file_vector_store_id_file_id"`

	// UsageBytes is the total vector store usage in bytes. Note that this may be different from the original file size.
	UsageBytes int64

	// TODO(guangrui): handle status update.
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
func (s *S) GetFileByFileID(vectorStoreID, fileID string) (*File, error) {
	var f File
	if err := s.db.Where("file_id = ?", fileID).
		Where("vector_store_id = ?", vectorStoreID).
		Take(&f).Error; err != nil {
		return nil, err
	}
	return &f, nil
}

// ListFiles lists files.
func (s *S) ListFiles(vectorStoreID string) ([]*File, error) {
	var fs []*File
	if err := s.db.
		Where("vector_store_id = ?", vectorStoreID).
		Order("file_id").Find(&fs).Error; err != nil {
		return nil, err
	}
	return fs, nil
}

// ListFilesWithPagination finds files with pagination. Files are returned in the order of created_at.
func (s *S) ListFilesWithPagination(
	vectorStoreID string,
	afterCreatedAt time.Time,
	afterID uint,
	order string,
	limit int,
) ([]*File, bool, error) {
	var fs []*File
	q := s.db.Where("vector_store_id = ?", vectorStoreID)
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
func (s *S) DeleteFile(vectorStoreID, fileID string) error {
	result := s.db.Unscoped().
		Where("file_id = ?", fileID).
		Where("vector_store_id = ?", vectorStoreID).
		Delete(&File{})
	if err := result.Error; err != nil {
		return err
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// DeleteAllFilesByVectorStoreID deletes all files of the collection.
func (s *S) DeleteAllFilesByVectorStoreID(vectorStoreID string) error {
	return DeleteAllFilesByVectorStoreIDInTransaction(s.db, vectorStoreID)
}

// DeleteAllFilesByVectorStoreIDInTransaction deletes all files of the collection.
func DeleteAllFilesByVectorStoreIDInTransaction(tx *gorm.DB, vectorStoreID string) error {
	if err := tx.Unscoped().
		Where("vector_store_id = ?", vectorStoreID).
		Delete(&File{}).Error; err != nil {
		return err
	}
	return nil
}
