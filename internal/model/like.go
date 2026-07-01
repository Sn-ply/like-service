package model

import (
	"time"

	"github.com/google/uuid"
)

type Like struct {
	PostID    uuid.UUID `db:"post_id" json:"post_id"`
	UserID    uuid.UUID `db:"user_id" json:"user_id"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}
