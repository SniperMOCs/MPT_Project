package user

import (
	"context"
	"io"

	"github.com/google/uuid"
)

type UserService interface {
	Create(ctx context.Context, u *User) error
	GetByID(ctx context.Context, id uuid.UUID) (*User, error)
	GetByUsername(ctx context.Context, username string) (*User, error)
	SearchByFullName(ctx context.Context, searchQuery string) ([]*User, error)

	Update(ctx context.Context, id uuid.UUID, username string, fullname string, description string) error
	UpdatePassword(ctx context.Context, userID uuid.UUID, oldPassword string, newPassword string) error

	CheckPassword(u *User, password string) bool
	UpdateAvatar(ctx context.Context, id uuid.UUID, file io.Reader) error
}
