package consumer

import (
	"testing"

	"social-network/services/notifications/internal/store"
)

func TestEventEnvelope_ToNotification_MapsType(t *testing.T) {
	tests := []struct {
		eventType string
		want      string
	}{
		{"post.liked", "like"},
		{"comment.voted", "like"},
		{"post.unliked", "unlike"},
		{"comment.vote.deleted", "unlike"},
		{"follow.requested", "follow_request"},
		{"follow.accepted", "follow_accept"},
		{"follow.declined", "follow_decline"},
		{"comment.created", "comment"},
		{"post.created", "post"},
		{"unknown.event", "unknown.event"},
	}

	for _, tt := range tests {
		env := &EventEnvelope{Type: tt.eventType}
		n := env.ToNotification()
		if n.Type != tt.want {
			t.Errorf("EventType %q: got %q, want %q", tt.eventType, n.Type, tt.want)
		}
	}
}

func TestEventEnvelope_ToNotification_MapsFields(t *testing.T) {
	env := &EventEnvelope{
		Type:         "post.liked",
		RecipientID:  "u1",
		ActorID:      "u2",
		ActorName:    "Bob",
		ActorAvatar:  "/avatars/bob.png",
		ResourceType: "post",
		ResourceID:   42,
		ContentText:  "Bob liked your post",
		ImageURL:     "/posts/42.jpg",
	}

	n := env.ToNotification()
	want := &store.Notification{
		RecipientID:  "u1",
		Type:         "like",
		ResourceType: "post",
		ResourceID:   42,
		ActorID:      "u2",
		ActorName:    "Bob",
		ActorAvatar:  "/avatars/bob.png",
		ContentText:  "Bob liked your post",
		ImageURL:     "/posts/42.jpg",
	}

	if n.RecipientID != want.RecipientID {
		t.Errorf("RecipientID = %q, want %q", n.RecipientID, want.RecipientID)
	}
	if n.Type != want.Type {
		t.Errorf("Type = %q, want %q", n.Type, want.Type)
	}
	if n.ResourceType != want.ResourceType {
		t.Errorf("ResourceType = %q, want %q", n.ResourceType, want.ResourceType)
	}
	if n.ResourceID != want.ResourceID {
		t.Errorf("ResourceID = %d, want %d", n.ResourceID, want.ResourceID)
	}
	if n.ActorID != want.ActorID {
		t.Errorf("ActorID = %q, want %q", n.ActorID, want.ActorID)
	}
	if n.ActorName != want.ActorName {
		t.Errorf("ActorName = %q, want %q", n.ActorName, want.ActorName)
	}
	if n.ActorAvatar != want.ActorAvatar {
		t.Errorf("ActorAvatar = %q, want %q", n.ActorAvatar, want.ActorAvatar)
	}
	if n.ContentText != want.ContentText {
		t.Errorf("ContentText = %q, want %q", n.ContentText, want.ContentText)
	}
	if n.ImageURL != want.ImageURL {
		t.Errorf("ImageURL = %q, want %q", n.ImageURL, want.ImageURL)
	}
}
