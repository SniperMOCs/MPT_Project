package notification

import (
	"context"

	"github.com/google/uuid"
)

type NotificationRepository interface {
	Create(ctx context.Context, n *Notification) error
	GetByRecipient(ctx context.Context, userId uuid.UUID) ([]Notification, error)
	MarkAsRead(ctx context.Context, userId uuid.UUID) error
	GetUnreadCount(ctx context.Context, userId uuid.UUID) (int, error)
}
