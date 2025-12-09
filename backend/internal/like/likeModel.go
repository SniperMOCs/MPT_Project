package like

import (
	"time"

	"github.com/google/uuid"
)

type Like struct {
	ID        uuid.UUID
	UserId    uuid.UUID
	ContentId uuid.UUID
	CreatedAt time.Time
}
