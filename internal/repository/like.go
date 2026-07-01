package repository

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

var ErrNotFound = errors.New("not found")

type Cursor struct {
	CreatedAt time.Time
	ID        uuid.UUID
}

type Summary struct {
	Count int
	Liked bool
}

type LikeRepository interface {
	Create(ctx context.Context, postID, userID uuid.UUID) error
	Delete(ctx context.Context, postID, userID uuid.UUID) error
	// CountsAndStatus returns, for each requested post ID, the total like count and whether
	// callerID has liked it. Always returns an entry for every requested ID (zero value if unliked).
	CountsAndStatus(ctx context.Context, postIDs []uuid.UUID, callerID uuid.UUID) (map[uuid.UUID]Summary, error)
	ListLikers(ctx context.Context, postID uuid.UUID, cursor *Cursor, limit int) ([]uuid.UUID, *Cursor, error)
}
