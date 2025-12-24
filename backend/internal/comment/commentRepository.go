package comment

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
)

type commentRepository struct {
	db *sql.DB
}

func NewCommentRepository(db *sql.DB) CommentRepository {
	return &commentRepository{db: db}
}

func (r *commentRepository) Create(ctx context.Context, c *Comment) error {
	query := `
		INSERT INTO comments (id, author_id, parent_id, content_id, text, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		c.ID,
		c.AuthorId,
		c.ParentId,
		c.ContentId,
		c.Text,
		c.CreatedAt,
	)
	return err
}

func (r *commentRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `
		DELETE FROM comments
		WHERE id = $1
	`

	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *commentRepository) GetByContentId(ctx context.Context, contentID uuid.UUID) ([]Comment, error) {
	query := `
		SELECT id, author_id, parent_id, content_id, text, created_at
		FROM comments
		WHERE content_id = $1
		ORDER BY created_at ASC
	`

	rows, err := r.db.QueryContext(ctx, query, contentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []Comment
	for rows.Next() {
		var c Comment

		if err := rows.Scan(
			&c.ID,
			&c.AuthorId,
			&c.ParentId,
			&c.ContentId,
			&c.Text,
			&c.CreatedAt,
		); err != nil {
			return nil, err
		}

		res = append(res, c)
	}

	return res, nil
}

func (r *commentRepository) GetById(ctx context.Context, id uuid.UUID) (*Comment, error) {
	query := `
		SELECT id, author_id, parent_id, content_id, text, created_at
		FROM comments
		WHERE id = $1
	`

	row := r.db.QueryRowContext(ctx, query, id)
	var c Comment

	err := row.Scan(
		&c.ID,
		&c.AuthorId,
		&c.ParentId,
		&c.ContentId,
		&c.Text,
		&c.CreatedAt,
	)

	if err != nil {
		return nil, err
	}
	return &c, nil
}
