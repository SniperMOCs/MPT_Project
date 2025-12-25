package chat

import (
	"time"

	"github.com/google/uuid"
)

type ChatUser struct {
	ID       uuid.UUID
	ChatId   uuid.UUID
	UserId   uuid.UUID
	JoinedAt time.Time
}
