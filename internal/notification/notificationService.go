package notification

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type notificationService struct {
	repo NotificationRepository
	hub  *WebSocketHub
}

func NewNotificationService(repo NotificationRepository, hub *WebSocketHub) NotificationService {
	return &notificationService{
		repo: repo,
		hub:  hub,
	}
}

func (s *notificationService) Notify(ctx context.Context, actor uuid.UUID, recipient uuid.UUID, target uuid.UUID, notitype NotificationType) (*NotificationResponse, error) {
	if actor == recipient {
		return nil, nil
	}
	n := &Notification{
		ID:        uuid.New(),
		Recipient: recipient,
		Actor:     actor,
		Type:      notitype,
		TargetId:  target,
		CreatedAt: time.Now(),
		Read:      false,
	}
	nr, err := s.repo.Create(ctx, n)
	if err != nil {
		return nil, err
	}
	s.hub.SendToUser(recipient, nr)
	return nr, nil
}

func (s *notificationService) GetByRecipient(ctx context.Context, userId uuid.UUID) ([]NotificationResponse, error) {
	return s.repo.GetByRecipient(ctx, userId)
}

func (s *notificationService) MarkAsRead(ctx context.Context, userId uuid.UUID) error {
	return s.repo.MarkAsRead(ctx, userId)
}

func (s *notificationService) GetUnreadCount(ctx context.Context, userId uuid.UUID) (int, error) {
	return s.repo.GetUnreadCount(ctx, userId)
}
