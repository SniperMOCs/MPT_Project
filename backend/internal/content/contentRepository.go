package content

import (
	"context"
	"database/sql"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

type contentRepository struct {
	db *sql.DB
}

func NewContentRepository(db *sql.DB) ContentRepository {
	return &contentRepository{db: db}
}

func (r *contentRepository) Create(ctx context.Context, c *Content) error {
	query := `
		INSERT INTO contents (id, author_id, content_type, description, media_urls, created_at, is_visible)
		VALUES($1, $2, $3, $4, $5, $6, $7)	
	`
	_, err := r.db.ExecContext(
		ctx,
		query,
		c.ID,
		c.AuthorId,
		c.ContentType,
		c.Description,
		pq.Array(c.MediaURLs),
		c.CreatedAt,
		c.IsVisible,
	)
	return err
}

func (r *contentRepository) GetByID(ctx context.Context, id uuid.UUID) (*Content, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, author_id, content_type, description, media_urls, created_at, is_visible
		FROM contents WHERE id=$1
	`, id)

	var c Content

	err := row.Scan(
		&c.ID,
		&c.AuthorId,
		&c.ContentType,
		&c.Description,
		pq.Array(&c.MediaURLs),
		&c.CreatedAt,
		&c.IsVisible,
	)

	if err != nil {
		return nil, err
	}

	return &c, nil
}

func (r *contentRepository) GetByType(ctx context.Context, ct ContentType) ([]Content, error) {
	query := `
		SELECT id, author_id, content_type, description, media_urls, created_at, is_visible
		FROM contents
		WHERE content_type = $1 AND is_visible = true
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, ct)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []Content
	for rows.Next() {
		var c Content

		if err := rows.Scan(
			&c.ID,
			&c.AuthorId,
			&c.ContentType,
			&c.Description,
			pq.Array(&c.MediaURLs),
			&c.CreatedAt,
			&c.IsVisible,
		); err != nil {
			return nil, err
		}

		res = append(res, c)
	}

	return res, nil
}

func (r *contentRepository) GetByAuthorAndType(ctx context.Context, authorId uuid.UUID, ct ContentType) ([]Content, error) {
	query := `
		SELECT id, author_id, content_type, description, media_urls, created_at, is_visible
		FROM contents
		WHERE content_type = $1 AND author_id = $2
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, ct, authorId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []Content
	for rows.Next() {
		var c Content

		if err := rows.Scan(
			&c.ID,
			&c.AuthorId,
			&c.ContentType,
			&c.Description,
			pq.Array(&c.MediaURLs),
			&c.CreatedAt,
			&c.IsVisible,
		); err != nil {
			return nil, err
		}

		res = append(res, c)
	}

	return res, nil
}

func (r *contentRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE contents
		SET is_visible = false
		WHERE id = $1
	`

	res, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *contentRepository) Update(ctx context.Context, description *string, isVisible *bool, mediaURLs *[]string, id uuid.UUID) error {
	query := "UPDATE contents SET "
	args := []any{}
	i := 1

	if isVisible != nil {
		query += "is_visible = $" + strconv.Itoa(i) + ", "
		args = append(args, isVisible)
		i++
	}

	if description != nil {
		query += "description = $" + strconv.Itoa(i) + ", "
		args = append(args, description)
		i++
	}

	if mediaURLs != nil {
		query += "media_urls = $" + strconv.Itoa(i) + ", "
		args = append(args, pq.Array(mediaURLs))
		i++
	}

	if len(args) == 0 {
		return nil
	}
	query = strings.TrimSuffix(query, ", ")
	query += " WHERE id = $" + strconv.Itoa(i)
	args = append(args, id)

	_, err := r.db.ExecContext(ctx, query, args...)
	return err
}
