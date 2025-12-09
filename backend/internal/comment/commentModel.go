package comment

import (
	"time"

	"github.com/google/uuid"
)

type Comment struct {
	ID        uuid.UUID
	AuthorId  uuid.UUID
	ParentId  uuid.UUID
	ContentId uuid.UUID
	Text      string
	CreatedAt time.Time
}
