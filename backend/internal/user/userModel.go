package user

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID          uuid.UUID
	Username    string
	FullName    string
	Password    string
	AvatarURL   string
	Description string
	CreatedAt   time.Time
}
