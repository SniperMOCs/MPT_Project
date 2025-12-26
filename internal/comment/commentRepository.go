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

func (r *commentRepository) GetByContentId(ctx context.Context, contentID uuid.UUID) ([]CommentResponse, error) {
	query := `
		SELECT c.id, c.parent_id, c.content_id, c.text, c.created_at, c.author_id, u.username, u.avatar_url
		FROM comments c
		INNER JOIN users u ON u.id = c.author_id 
		WHERE content_id = $1
		ORDER BY created_at ASC
	`

	rows, err := r.db.QueryContext(ctx, query, contentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []CommentResponse
	for rows.Next() {
		var c CommentResponse

		if err := rows.Scan(
			&c.ID,
			&c.ParentId,
			&c.ContentId,
			&c.Text,
			&c.CreatedAt,
			&c.Author.ID,
			&c.Author.Username,
			&c.Author.AvatarURL,
		); err != nil {
			return nil, err
		}

		res = append(res, c)
	}

	return res, nil
}

func (r *commentRepository) GetById(ctx context.Context, id uuid.UUID) (*Comment, error) {
	query := `
		SELECT c.id, c.parent_id, c.content_id, c.text, c.created_at, c.author_id
		FROM comments c
		WHERE id = $1
	`

	row := r.db.QueryRowContext(ctx, query, id)
	var c Comment

	err := row.Scan(
		&c.ID,
		&c.ParentId,
		&c.ContentId,
		&c.Text,
		&c.CreatedAt,
		&c.AuthorId,
	)

	if err != nil {
		return nil, err
	}
	return &c, nil
}
