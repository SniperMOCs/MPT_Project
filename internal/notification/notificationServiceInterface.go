package notification

import (
	"context"

	"github.com/google/uuid"
)

type NotificationService interface {
	Notify(ctx context.Context, actor uuid.UUID, recipient uuid.UUID, target uuid.UUID, notitype NotificationType) (*NotificationResponse, error)
	GetByRecipient(ctx context.Context, userId uuid.UUID) ([]NotificationResponse, error)
	MarkAsRead(ctx context.Context, userId uuid.UUID) error
	GetUnreadCount(ctx context.Context, userId uuid.UUID) (int, error)
}
