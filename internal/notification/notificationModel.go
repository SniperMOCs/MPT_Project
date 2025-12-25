package notification

import (
	"time"

	"github.com/google/uuid"
)

type NotificationType string

const (
	Like    NotificationType = "LIKE"
	Comment NotificationType = "COMMENT"
	Follow  NotificationType = "FOLLOW"
)

type Notification struct {
	ID        uuid.UUID        `db:"id"`
	Recipient uuid.UUID        `db:"recipient_id"`
	Actor     uuid.UUID        `db:"actor_id"`
	Type      NotificationType `db:"type"`
	TargetId  uuid.UUID        `db:"target_id"`
	CreatedAt time.Time        `db:"created_at"`
	Read      bool             `db:"read"`
}
