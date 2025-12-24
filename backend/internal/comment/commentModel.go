package comment

import (
	"time"

	"github.com/google/uuid"
)

type Comment struct {
	ID        uuid.UUID  `json:"id"`
	AuthorId  uuid.UUID  `json:"author_id"`
	ParentId  *uuid.UUID `json:"parent_id"`
	ContentId uuid.UUID  `json:"content_id"`
	Text      string     `json:"text"`
	CreatedAt time.Time  `json:"created_at"`
}
