package follow

import (
	"testing"
)

func TestFollow_Fields(t *testing.T) {
	f := Follow{
		FollowerID: "user-1",
		FolloweeID: "user-2",
	}

	if f.FollowerID != "user-1" {
		t.Errorf("FollowerID = %q, want %q", f.FollowerID, "user-1")
	}
	if f.FolloweeID != "user-2" {
		t.Errorf("FolloweeID = %q, want %q", f.FolloweeID, "user-2")
	}
}

func TestRequest_Fields(t *testing.T) {
	req := Request{
		FollowerID: "user-1",
		FolloweeID: "user-2",
	}

	if req.FollowerID != "user-1" {
		t.Errorf("FollowerID = %q, want %q", req.FollowerID, "user-1")
	}
	if req.FolloweeID != "user-2" {
		t.Errorf("FolloweeID = %q, want %q", req.FolloweeID, "user-2")
	}
}
