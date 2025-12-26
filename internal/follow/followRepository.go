package follow

import (
	"context"
	"database/sql"
	"time"

	"github.com/SniperMOCs/MPT_Project/internal/user"
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

func (r *followRepository) GetFollowers(ctx context.Context, userId uuid.UUID) ([]*user.User, error) {
	query := `
		SELECT u.id, u.username, u.password, u.full_name, u.avatar_url, u.description, u.created_at
		FROM users u
		INNER JOIN follows f ON u.id = f.follower_id
		WHERE f.following_id = $1
		ORDER BY f.created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []*user.User
	for rows.Next() {
		var u user.User

		if err := rows.Scan(
			&u.ID,
			&u.Username,
			&u.Password,
			&u.FullName,
			&u.AvatarURL,
			&u.Description,
			&u.CreatedAt,
		); err != nil {
			return nil, err
		}

		res = append(res, &u)
	}

	return res, nil
}

func (r *followRepository) GetFollowing(ctx context.Context, userId uuid.UUID) ([]*user.User, error) {
	query := `
		SELECT u.id, u.username, u.password, u.full_name, u.avatar_url, u.description, u.created_at
		FROM users u
		INNER JOIN follows f ON u.id = f.following_id
		WHERE f.follower_id = $1
		ORDER BY f.created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []*user.User
	for rows.Next() {
		var u user.User

		if err := rows.Scan(
			&u.ID,
			&u.Username,
			&u.Password,
			&u.FullName,
			&u.AvatarURL,
			&u.Description,
			&u.CreatedAt,
		); err != nil {
			return nil, err
		}

		res = append(res, &u)
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
