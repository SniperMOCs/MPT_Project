package user

import (
	"context"
	"database/sql"
	"strconv"
	"strings"

	"github.com/google/uuid"
)

type userRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(ctx context.Context, u *User) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO users (id, username, password, created_at)
		VALUES ($1, $2, $3, $4)
	`,
		u.ID,
		u.Username,
		u.Password,
		u.CreatedAt,
	)

	return err

}

func (r *userRepository) GetByID(ctx context.Context, id uuid.UUID) (*User, error) {
	var u User

	row := r.db.QueryRow(`
		SELECT id, username, password, full_name, avatar_url, description, created_at
		FROM users WHERE id=$1
	`, id)

	err := row.Scan(
		&u.ID,
		&u.Username,
		&u.Password,
		&u.FullName,
		&u.AvatarURL,
		&u.Description,
		&u.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &u, nil
}

func (r *userRepository) GetByUsername(ctx context.Context, username string) (*User, error) {
	var u User

	row := r.db.QueryRowContext(ctx, `
		SELECT id, username, password, full_name, avatar_url, description, created_at
		FROM users WHERE username=$1
	`, username)

	err := row.Scan(
		&u.ID,
		&u.Username,
		&u.Password,
		&u.FullName,
		&u.AvatarURL,
		&u.Description,
		&u.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &u, nil
}

func (r *userRepository) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	var exists bool

	err := r.db.QueryRowContext(ctx, `
		SELECT EXISTS(SELECT 1 FROM users WHERE username=$1)
	`, username).Scan(&exists)

	return exists, err
}

func (r *userRepository) Update(ctx context.Context, id uuid.UUID, dto UserDTO) error {
	query := "UPDATE users SET "
	args := []interface{}{}
	i := 1

	if dto.Username != nil {
		query += "username = $" + strconv.Itoa(i) + ", "
		args = append(args, *dto.Username)
		i++
	}

	if dto.FullName != nil {
		query += "full_name = $" + strconv.Itoa(i) + ", "
		args = append(args, *dto.FullName)
		i++
	}

	if dto.Description != nil {
		query += "description = $" + strconv.Itoa(i) + ", "
		args = append(args, *dto.Description)
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

func (r *userRepository) UpdatePassword(ctx context.Context, id uuid.UUID, passwordHash string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE users SET password = $1 WHERE id = $2
	`, passwordHash, id)

	return err
}

func (r *userRepository) UpdateAvatar(ctx context.Context, id uuid.UUID, avatarURL string) error {
	query := "UPDATE users SET avatar_url = $1 WHERE id = $2"
	args := []interface{}{}

	if avatarURL != "" {
		args = append(args, avatarURL)
	}
	args = append(args, id)
	_, err := r.db.ExecContext(ctx, query, args...)
	return err
}
