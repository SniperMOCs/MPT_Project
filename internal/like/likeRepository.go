package like

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type likeRepository struct {
	db *sql.DB
}

func NewLikeRepository(db *sql.DB) LikeRepository {
	return &likeRepository{db: db}
}

func (r *likeRepository) Create(ctx context.Context, userId uuid.UUID, contentId uuid.UUID) error {
	query := `
		INSERT INTO likes (id, user_id, content_id, created_at)
		VALUES ($1, $2, $3, $4)
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		uuid.New(),
		userId,
		contentId,
		time.Now(),
	)
	return err
}

func (r *likeRepository) Delete(ctx context.Context, userId uuid.UUID, contentId uuid.UUID) error {
	query := `
		DELETE FROM likes
		WHERE user_id = $1 AND content_id = $2
	`

	_, err := r.db.ExecContext(ctx, query, userId, contentId)
	return err
}

func (r *likeRepository) Exists(ctx context.Context, userId uuid.UUID, contentId uuid.UUID) (bool, error) {
	var exists bool

	err := r.db.QueryRowContext(ctx, `
		SELECT EXISTS(SELECT 1 FROM likes WHERE user_id = $1 AND content_id = $2)
	`, userId, contentId).Scan(&exists)

	return exists, err
}

func (r *likeRepository) GetLikesCount(ctx context.Context, contentId uuid.UUID) (int, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM likes WHERE content_id = $1
	`, contentId)

	var res int

	err := row.Scan(&res)

	if err != nil {
		return 0, err
	}

	return res, nil
}
