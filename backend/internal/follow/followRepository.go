package follow

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type followRepository struct {
	db *sql.DB
}

func NewFollowRepository(db *sql.DB) FollowRepository {
	return &followRepository{db: db}
}

func (r *followRepository) Create(ctx context.Context, followerId uuid.UUID, followingId uuid.UUID) error {
	query := `
		INSERT INTO follows (id, follower_id, following_id, created_at)
		VALUES ($1, $2, $3, $4)
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		uuid.New(),
		followerId,
		followingId,
		time.Now(),
	)
	return err
}

func (r *followRepository) Delete(ctx context.Context, followerId uuid.UUID, followingId uuid.UUID) error {
	query := `
		DELETE FROM follows
		WHERE follower_id = $1 AND following_id = $2
	`

	_, err := r.db.ExecContext(ctx, query, followerId, followingId)
	return err
}

func (r *followRepository) Exists(ctx context.Context, followerId uuid.UUID, followingId uuid.UUID) (bool, error) {
	var exists bool

	err := r.db.QueryRowContext(ctx, `
		SELECT EXISTS(SELECT 1 FROM follows WHERE follower_id = $1 AND following_id = $2)
	`, followerId, followingId).Scan(&exists)

	return exists, err
}

func (r *followRepository) GetFollowers(ctx context.Context, userId uuid.UUID) ([]Follow, error) {
	query := `
		SELECT id, follower_id, following_id, created_at
		FROM follows
		WHERE following_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []Follow
	for rows.Next() {
		var f Follow

		if err := rows.Scan(
			&f.ID,
			&f.FollowerId,
			&f.FollowingId,
			&f.CreatedAt,
		); err != nil {
			return nil, err
		}

		res = append(res, f)
	}

	return res, nil
}

func (r *followRepository) GetFollowing(ctx context.Context, userId uuid.UUID) ([]Follow, error) {
	query := `
		SELECT id, follower_id, following_id, created_at
		FROM follows
		WHERE follower_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []Follow
	for rows.Next() {
		var f Follow

		if err := rows.Scan(
			&f.ID,
			&f.FollowerId,
			&f.FollowingId,
			&f.CreatedAt,
		); err != nil {
			return nil, err
		}

		res = append(res, f)
	}

	return res, nil
}

func (r *followRepository) GetFollowersCount(ctx context.Context, userId uuid.UUID) (int, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM follows WHERE following_id = $1
	`, userId)

	var res int

	err := row.Scan(&res)

	if err != nil {
		return 0, err
	}

	return res, nil
}

func (r *followRepository) GetFollowingCount(ctx context.Context, userId uuid.UUID) (int, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM follows WHERE follower_id = $1
	`, userId)

	var res int

	err := row.Scan(&res)

	if err != nil {
		return 0, err
	}

	return res, nil
}
