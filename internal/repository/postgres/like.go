package postgres

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/snaply/like-service/internal/repository"
)

type likeRepo struct {
	db *sqlx.DB
}

func NewLikeRepository(db *sqlx.DB) repository.LikeRepository {
	return &likeRepo{db: db}
}

func (r *likeRepo) Create(ctx context.Context, postID, userID uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO likes (post_id, user_id)
		VALUES ($1, $2)
		ON CONFLICT (post_id, user_id) DO NOTHING`,
		postID, userID)
	return err
}

func (r *likeRepo) Delete(ctx context.Context, postID, userID uuid.UUID) error {
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM likes WHERE post_id = $1 AND user_id = $2`,
		postID, userID)
	return err
}

func (r *likeRepo) CountsAndStatus(ctx context.Context, postIDs []uuid.UUID, callerID uuid.UUID) (map[uuid.UUID]repository.Summary, error) {
	result := make(map[uuid.UUID]repository.Summary, len(postIDs))
	for _, id := range postIDs {
		result[id] = repository.Summary{}
	}
	if len(postIDs) == 0 {
		return result, nil
	}

	countQuery, countArgs, err := sqlx.In(
		`SELECT post_id, COUNT(*) AS count FROM likes WHERE post_id IN (?) GROUP BY post_id`,
		postIDs,
	)
	if err != nil {
		return nil, err
	}
	countQuery = r.db.Rebind(countQuery)

	type countRow struct {
		PostID uuid.UUID `db:"post_id"`
		Count  int       `db:"count"`
	}
	var countRows []countRow
	if err := r.db.SelectContext(ctx, &countRows, countQuery, countArgs...); err != nil {
		return nil, err
	}
	for _, row := range countRows {
		summary := result[row.PostID]
		summary.Count = row.Count
		result[row.PostID] = summary
	}

	likedQuery, likedArgs, err := sqlx.In(
		`SELECT post_id FROM likes WHERE user_id = ? AND post_id IN (?)`,
		callerID, postIDs,
	)
	if err != nil {
		return nil, err
	}
	likedQuery = r.db.Rebind(likedQuery)

	var likedIDs []uuid.UUID
	if err := r.db.SelectContext(ctx, &likedIDs, likedQuery, likedArgs...); err != nil {
		return nil, err
	}
	for _, id := range likedIDs {
		summary := result[id]
		summary.Liked = true
		result[id] = summary
	}

	return result, nil
}

type likerRow struct {
	ID        uuid.UUID `db:"id"`
	CreatedAt time.Time `db:"created_at"`
}

func (r *likeRepo) ListLikers(ctx context.Context, postID uuid.UUID, cursor *repository.Cursor, limit int) ([]uuid.UUID, *repository.Cursor, error) {
	var (
		query string
		args  []interface{}
	)
	base := `SELECT user_id AS id, created_at FROM likes WHERE post_id = $1`
	if cursor == nil {
		query = base + ` ORDER BY created_at ASC, user_id ASC LIMIT $2`
		args = []interface{}{postID, limit + 1}
	} else {
		query = base + ` AND (created_at, user_id) > ($2, $3) ORDER BY created_at ASC, user_id ASC LIMIT $4`
		args = []interface{}{postID, cursor.CreatedAt, cursor.ID, limit + 1}
	}

	rows := []*likerRow{}
	if err := r.db.SelectContext(ctx, &rows, query, args...); err != nil {
		return nil, nil, err
	}

	var nextCursor *repository.Cursor
	if len(rows) > limit {
		last := rows[limit-1]
		nextCursor = &repository.Cursor{CreatedAt: last.CreatedAt, ID: last.ID}
		rows = rows[:limit]
	}

	ids := make([]uuid.UUID, 0, len(rows))
	for _, row := range rows {
		ids = append(ids, row.ID)
	}

	return ids, nextCursor, nil
}
