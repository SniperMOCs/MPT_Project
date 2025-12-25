package chat

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID
	Title     string
	IsGroup   bool
	CreatedAt time.Time
}
