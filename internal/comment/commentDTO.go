package comment

import "github.com/google/uuid"

type CommentDTO struct {
	ParentId  *uuid.UUID `json:"parent_id"`
	ContentId uuid.UUID  `json:"content_id"`
	Text      string     `json:"text"`
}
