package notification

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
)

type notificationRepository struct {
	db *sql.DB
}

func NewNotificationRepository(db *sql.DB) NotificationRepository {
	return &notificationRepository{db: db}
}

func (r *notificationRepository) Create(ctx context.Context, n *Notification) error {
	query := `
		INSERT INTO notifications (id, recipient_id, target_id, type, created_at, actor_id, read)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		n.ID,
		n.Recipient,
		n.TargetId,
		n.Type,
		n.CreatedAt,
		n.Actor,
		n.Read,
	)
	return err
}

func (r *notificationRepository) GetByRecipient(ctx context.Context, userId uuid.UUID) ([]Notification, error) {
	query := `
		SELECT id, recipient_id, target_id, type, created_at, actor_id, read
		FROM notifications
		WHERE recipient_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []Notification
	for rows.Next() {
		var n Notification

		if err := rows.Scan(
			&n.ID,
			&n.Recipient,
			&n.TargetId,
			&n.Type,
			&n.CreatedAt,
			&n.Actor,
			&n.Read,
		); err != nil {
			return nil, err
		}

		res = append(res, n)
	}

	return res, nil
}

func (r *notificationRepository) MarkAsRead(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE notifications
		SET read = true
		WHERE recipient_id = $1
	`

	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *notificationRepository) GetUnreadCount(ctx context.Context, userId uuid.UUID) (int, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM notifications WHERE recipient_id = $1 AND read = false
	`, userId)

	var res int

	err := row.Scan(&res)

	if err != nil {
		return 0, err
	}

	return res, nil
}
