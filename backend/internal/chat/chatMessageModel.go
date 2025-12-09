package chat

import (
	"time"

	"github.com/google/uuid"
)

type ChatMessage struct {
	ID        uuid.UUID
	ChatId    uuid.UUID
	SenderId  uuid.UUID
	Content   string
	CreatedAt time.Time
}
