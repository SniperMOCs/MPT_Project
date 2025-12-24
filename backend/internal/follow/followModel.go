package follow

import (
	"time"

	"github.com/google/uuid"
)

type Follow struct {
	ID          uuid.UUID `json:"id"`
	FollowerId  uuid.UUID `json:"follower_id"`
	FollowingId uuid.UUID `json:"following_id"`
	CreatedAt   time.Time `json:"created_at"`
}
