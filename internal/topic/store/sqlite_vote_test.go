package store

import (
	"context"
	"errors"
	"testing"

	"social-network/internal/topic"
)

func TestCastVote(t *testing.T) {
	s := setupTopicStore(t)

	top := &topic.Topic{UserID: "u1", Title: "Vote", Content: "x"}
	_ = s.CreateTopic(context.Background(), top, nil)

	if err := s.CastVote(context.Background(), "u2", top.ID, 1); err != nil {
		t.Fatalf("CastVote: %v", err)
	}

	vc, err := s.GetVoteCounts(context.Background(), top.ID)
	if err != nil {
		t.Fatalf("GetVoteCounts: %v", err)
	}
	if vc.Upvotes != 1 || vc.Downvotes != 0 || vc.Score != 1 {
		t.Errorf("votes = %+v, want {1 0 1}", vc)
	}
}

func TestCastVote_Toggle(t *testing.T) {
	s := setupTopicStore(t)

	top := &topic.Topic{UserID: "u1", Title: "Toggle", Content: "x"}
	_ = s.CreateTopic(context.Background(), top, nil)

	_ = s.CastVote(context.Background(), "u2", top.ID, 1)
	_ = s.CastVote(context.Background(), "u2", top.ID, -1)

	vc, _ := s.GetVoteCounts(context.Background(), top.ID)
	if vc.Upvotes != 0 || vc.Downvotes != 1 || vc.Score != -1 {
		t.Errorf("after toggle: votes = %+v, want {0 1 -1}", vc)
	}
}

func TestDeleteVote(t *testing.T) {
	s := setupTopicStore(t)

	top := &topic.Topic{UserID: "u1", Title: "DelVote", Content: "x"}
	_ = s.CreateTopic(context.Background(), top, nil)
	_ = s.CastVote(context.Background(), "u2", top.ID, 1)

	if err := s.DeleteVote(context.Background(), "u2", top.ID); err != nil {
		t.Fatalf("DeleteVote: %v", err)
	}

	vc, _ := s.GetVoteCounts(context.Background(), top.ID)
	if vc.Upvotes != 0 {
		t.Errorf("UpvoteCount = %d, want 0", vc.Upvotes)
	}
}

func TestGetVoteCounts(t *testing.T) {
	s := setupTopicStore(t)

	top := &topic.Topic{UserID: "u1", Title: "Counts", Content: "x"}
	_ = s.CreateTopic(context.Background(), top, nil)
	_ = s.CastVote(context.Background(), "u1", top.ID, 1)
	_ = s.CastVote(context.Background(), "u2", top.ID, -1)

	vc, err := s.GetVoteCounts(context.Background(), top.ID)
	if err != nil {
		t.Fatalf("GetVoteCounts: %v", err)
	}
	if vc.Upvotes != 1 || vc.Downvotes != 1 || vc.Score != 0 {
		t.Errorf("votes = %+v, want {1 1 0}", vc)
	}
}

func TestCastVote_NonVisibleTopic(t *testing.T) {
	s := setupTopicStore(t)

	priv := &topic.Topic{UserID: "u1", Title: "Private", Content: "x", Visibility: topic.VisibilityPrivate}
	_ = s.CreateTopic(context.Background(), priv, []string{"u2"})

	if err := s.CastVote(context.Background(), "u3", priv.ID, 1); !errors.Is(err, topic.ErrTopicNotFound) {
		t.Errorf("non-allowed voter: err = %v, want ErrTopicNotFound", err)
	}
	if err := s.CastVote(context.Background(), "u2", priv.ID, 1); err != nil {
		t.Fatalf("allowed voter: %v", err)
	}
	if err := s.CastVote(context.Background(), "u1", priv.ID, -1); err != nil {
		t.Fatalf("owner voter: %v", err)
	}
}
