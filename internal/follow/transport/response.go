package transport

import (
	"time"

	"social-network/internal/follow"
)

type FollowRequestResponse struct {
	ID          string      `json:"id"`
	RequesterID string      `json:"requesterId"`
	Requester   *UserResult `json:"requester"`
	TargetID    string      `json:"targetId"`
	Target      *UserResult `json:"target"`
	Status      string      `json:"status"`
	CreatedAt   string      `json:"createdAt"`
}

func toFollowRequestResponse(r follow.Request, requester, target *UserResult) FollowRequestResponse {
	return FollowRequestResponse{
		ID:          r.FollowerID + ":" + r.FolloweeID,
		RequesterID: r.FollowerID,
		Requester:   requester,
		TargetID:    r.FolloweeID,
		Target:      target,
		Status:      "pending",
		CreatedAt:   r.CreatedAt.Format(time.RFC3339),
	}
}
