package notification

import (
	"time"

	"github.com/SniperMOCs/MPT_Project/internal/content"
	"github.com/google/uuid"
)

type NotificationType string

const (
	Like    NotificationType = "LIKE"
	Comment NotificationType = "COMMENT"
	Mention NotificationType = "MENTION"
	Follow  NotificationType = "FOLLOW"
)

type Notification struct {
	ID        uuid.UUID
	UserId    uuid.UUID
	ContentId uuid.UUID
	NotificationType
	ContentType content.ContentType
	CreatedAt   time.Time
}
