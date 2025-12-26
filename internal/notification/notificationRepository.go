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

func (r *notificationRepository) GetByRecipient(ctx context.Context, userId uuid.UUID) ([]NotificationResponse, error) {
	query := `
		SELECT n.id, n.recipient_id, n.target_id, n.type, n.created_at, n.read, n.actor_id, u.username, u.avatar_url
		FROM notifications n
		INNER JOIN users u ON u.id = n.actor_id
		WHERE n.recipient_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []NotificationResponse
	for rows.Next() {
		var n NotificationResponse

		if err := rows.Scan(
			&n.ID,
			&n.Recipient,
			&n.TargetId,
			&n.Type,
			&n.CreatedAt,
			&n.Read,
			&n.Actor.ID,
			&n.Actor.Username,
			&n.Actor.AvatarURL,
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
