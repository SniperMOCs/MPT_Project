package follow

import (
	"time"

	"github.com/google/uuid"
)

type Follow struct {
	ID          uuid.UUID
	FollowerId  uuid.UUID
	FollowingId uuid.UUID
	CreatedAt   time.Time
}
