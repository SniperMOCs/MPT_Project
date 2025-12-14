package user

import (
	"context"

	"github.com/google/uuid"
)

type UserRepository interface {
	Create(ctx context.Context, u *User) error
	GetByID(ctx context.Context, id uuid.UUID) (*User, error)
	GetByUsername(ctx context.Context, username string) (*User, error)
	ExistsByUsername(ctx context.Context, username string) (bool, error)
	Update(ctx context.Context, userID uuid.UUID, dto UserDTO) error
	UpdatePassword(ctx context.Context, userID uuid.UUID, passwordHash string) error
}
