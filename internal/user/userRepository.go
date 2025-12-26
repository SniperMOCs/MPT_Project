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

func (r *userRepository) Update(ctx context.Context, id uuid.UUID, username string, fullname string, description string) error {
	query := "UPDATE users SET "
	args := []interface{}{}
	i := 1

	if username != "" {
		query += "username = $" + strconv.Itoa(i) + ", "
		args = append(args, username)
		i++
	}

	if fullname != "" {
		query += "full_name = $" + strconv.Itoa(i) + ", "
		args = append(args, fullname)
		i++
	}

	if description != "" {
		query += "description = $" + strconv.Itoa(i) + ", "
		args = append(args, description)
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
	query := `UPDATE users SET avatar_url = $1 WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, avatarURL, id)
	return err
}

func (r *userRepository) SearchByFullName(ctx context.Context, searchQuery string) ([]*User, error) {
	words := strings.Fields(strings.TrimSpace(searchQuery))
	if len(words) == 0 {
		return []*User{}, nil
	}

	query := `
		SELECT id, username, full_name, avatar_url, description, created_at
		FROM users
		WHERE (`

	args := []any{}
	argIndex := 1

	for i, word := range words {
		if i > 0 {
			query += " OR "
		}
		query += "(COALESCE(full_name, '') ILIKE $" + strconv.Itoa(argIndex) + " OR username ILIKE $" + strconv.Itoa(argIndex+1) + ")"
		args = append(args, "%"+word+"%", "%"+word+"%")
		argIndex += 2
	}
	query += ") ORDER BY username"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*User
	for rows.Next() {
		var u User
		err := rows.Scan(
			&u.ID,
			&u.Username,
			&u.FullName,
			&u.AvatarURL,
			&u.Description,
			&u.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		users = append(users, &u)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}
