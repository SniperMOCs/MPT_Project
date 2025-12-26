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
		INSERT INTO contents (id, author_id, description, media_urls, created_at, is_visible)
		VALUES($1, $2, $3, $4, $5, $6)	
	`
	_, err := r.db.ExecContext(
		ctx,
		query,
		c.ID,
		c.AuthorId,
		c.Description,
		pq.Array(c.MediaURLs),
		c.CreatedAt,
		c.IsVisible,
	)
	return err
}

func (r *contentRepository) GetByID(ctx context.Context, id uuid.UUID) (*ContentResponse, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT c.id, c.description, c.media_urls, c.created_at, c.is_visible, c.author_id, u.username, u.avatar_url
		FROM contents c
		INNER JOIN users u ON u.id = c.author_id
		WHERE c.id=$1
	`, id)

	var c ContentResponse

	err := row.Scan(
		&c.ID,
		&c.Description,
		pq.Array(&c.MediaURLs),
		&c.CreatedAt,
		&c.IsVisible,
		&c.Author.ID,
		&c.Author.Username,
		&c.Author.AvatarURL,
	)

	if err != nil {
		return nil, err
	}

	return &c, nil
}

func (r *contentRepository) GetByAuthor(ctx context.Context, authorId uuid.UUID) ([]ContentResponse, error) {
	query := `
		SELECT c.id,c.description, c.media_urls, c.created_at, c.is_visible, c.author_id, u.username, u.avatar_url
		FROM contents c
		INNER JOIN users u ON u.id = c.author_id
		WHERE author_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, authorId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []ContentResponse
	for rows.Next() {
		var c ContentResponse

		if err := rows.Scan(
			&c.ID,
			&c.Description,
			pq.Array(&c.MediaURLs),
			&c.CreatedAt,
			&c.IsVisible,
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

func (r *contentRepository) GetByFollowing(ctx context.Context, userId uuid.UUID) ([]ContentResponse, error) {
	query := `
		SELECT c.id, c.description, c.media_urls, c.created_at, c.is_visible, c.author_id, u.username, u.avatar_url
		FROM contents c
		INNER JOIN follows f ON c.author_id = f.following_id
		INNER JOIN users u ON c.author_id = u.id
		WHERE f.follower_id = $1 AND c.is_visible = true
		ORDER BY c.created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []ContentResponse
	for rows.Next() {
		var c ContentResponse

		if err := rows.Scan(
			&c.ID,
			&c.Description,
			pq.Array(&c.MediaURLs),
			&c.CreatedAt,
			&c.IsVisible,
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
