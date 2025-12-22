package user

import (
	"context"

	"github.com/google/uuid"
)

type UserService interface {
	Create(ctx context.Context, u *User) error
	GetByID(ctx context.Context, id uuid.UUID) (*User, error)
	GetByUsername(ctx context.Context, username string) (*User, error)

	Update(ctx context.Context, userID uuid.UUID, dto UserDTO) error
	UpdatePassword(ctx context.Context, userID uuid.UUID, dto UpdatePasswordDTO) error

	CheckPassword(u *User, password string) bool
	UpdateAvatar(ctx context.Context, id uuid.UUID, avatarURL string) error
}
