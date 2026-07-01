package service

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/snaply/like-service/internal/repository"
	"go.uber.org/zap"
)

type Summary struct {
	PostID uuid.UUID `json:"post_id"`
	Count  int       `json:"count"`
	Liked  bool      `json:"liked"`
}

type Page struct {
	UserIDs    []uuid.UUID
	NextCursor string
}

type LikeService interface {
	Like(ctx context.Context, postID, userID uuid.UUID) error
	Unlike(ctx context.Context, postID, userID uuid.UUID) error
	Batch(ctx context.Context, postIDs []uuid.UUID, callerID uuid.UUID) ([]Summary, error)
	Likers(ctx context.Context, postID uuid.UUID, cursor string, limit int) (*Page, error)
}

type likeService struct {
	likes repository.LikeRepository
	log   *zap.Logger
}

func NewLikeService(likes repository.LikeRepository, log *zap.Logger) LikeService {
	return &likeService{likes: likes, log: log}
}

func (s *likeService) Like(ctx context.Context, postID, userID uuid.UUID) error {
	if err := s.likes.Create(ctx, postID, userID); err != nil {
		return fmt.Errorf("creating like: %w", err)
	}
	return nil
}

func (s *likeService) Unlike(ctx context.Context, postID, userID uuid.UUID) error {
	if err := s.likes.Delete(ctx, postID, userID); err != nil {
		return fmt.Errorf("deleting like: %w", err)
	}
	return nil
}

func (s *likeService) Batch(ctx context.Context, postIDs []uuid.UUID, callerID uuid.UUID) ([]Summary, error) {
	summaries, err := s.likes.CountsAndStatus(ctx, postIDs, callerID)
	if err != nil {
		return nil, fmt.Errorf("fetching like summaries: %w", err)
	}

	result := make([]Summary, 0, len(postIDs))
	for _, id := range postIDs {
		s := summaries[id]
		result = append(result, Summary{PostID: id, Count: s.Count, Liked: s.Liked})
	}
	return result, nil
}

func (s *likeService) Likers(ctx context.Context, postID uuid.UUID, cursorStr string, limit int) (*Page, error) {
	if limit <= 0 || limit > 50 {
		limit = 20
	}

	var cur *repository.Cursor
	if cursorStr != "" {
		decoded, err := decodeCursor(cursorStr)
		if err != nil {
			return nil, fmt.Errorf("invalid cursor: %w", err)
		}
		cur = decoded
	}

	ids, nextCur, err := s.likes.ListLikers(ctx, postID, cur, limit)
	if err != nil {
		return nil, fmt.Errorf("listing likers: %w", err)
	}

	page := &Page{UserIDs: ids}
	if page.UserIDs == nil {
		page.UserIDs = []uuid.UUID{}
	}
	if nextCur != nil {
		page.NextCursor, err = encodeCursor(nextCur)
		if err != nil {
			s.log.Warn("failed to encode cursor", zap.Error(err))
		}
	}
	return page, nil
}

type cursorPayload struct {
	CreatedAt time.Time `json:"ca"`
	ID        string    `json:"id"`
}

func encodeCursor(c *repository.Cursor) (string, error) {
	b, err := json.Marshal(cursorPayload{CreatedAt: c.CreatedAt, ID: c.ID.String()})
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(b), nil
}

func decodeCursor(s string) (*repository.Cursor, error) {
	b, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return nil, err
	}
	var p cursorPayload
	if err := json.Unmarshal(b, &p); err != nil {
		return nil, err
	}
	id, err := uuid.Parse(p.ID)
	if err != nil {
		return nil, err
	}
	return &repository.Cursor{CreatedAt: p.CreatedAt, ID: id}, nil
}
