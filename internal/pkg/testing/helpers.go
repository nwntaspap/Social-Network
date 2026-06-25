package testhelpers

import (
	"testing"

	"social-network/internal/domain/topic"
	"social-network/internal/domain/user"
)

func AssertUserMatch(t *testing.T, got, want *user.User) {
	t.Helper()

	if want == nil {
		if got != nil {
			t.Error("Handle() expected nil user, got non-nil")
		}
		return
	}

	if got == nil {
		t.Fatalf("Handle() got nil user, want user with ID %s", want.ID)
		return
	}

	CompareUserFields(t, got, want)
}

func CompareUserFields(t *testing.T, got, want *user.User) {
	t.Helper()

	if got.ID != want.ID {
		t.Errorf("Handle() got ID = %v, want %v", got.ID, want.ID)
	}
	if got.Nickname != want.Nickname {
		t.Errorf("Handle() got Username = %v, want %v", got.Nickname, want.Nickname)
	}
	if got.Email != want.Email {
		t.Errorf("Handle() got Email = %v, want %v", got.Email, want.Email)
	}
	if got.Password != want.Password {
		t.Errorf("Handle() got Password = %v, want %v", got.Password, want.Password)
	}
}

func AssertTopicMatch(t *testing.T, got, want *topic.Topic) {
	t.Helper()

	if want == nil {
		if got != nil {
			t.Error("Handle() expected nil topic, got non-nil")
		}
		return
	}

	if got == nil {
		t.Fatalf("Handle() got nil topic, want topic with ID %d", want.ID)
		return
	}

	CompareCreateTopicFields(t, got, want)
}

func CompareCreateTopicFields(t *testing.T, got, want *topic.Topic) {
	t.Helper()

	if got.ID != want.ID {
		t.Errorf("Handle() got ID = %v, want %v", got.ID, want.ID)
	}
	if got.Title != want.Title {
		t.Errorf("Handle() got Title = %v, want %v", got.Title, want.Title)
	}
	if got.Content != want.Content {
		t.Errorf("Handle() got Content = %v, want %v", got.Content, want.Content)
	}
	if got.ImagePath != want.ImagePath {
		t.Errorf("Handle() got ImagePath = %v, want %v", got.ImagePath, want.ImagePath)
	}
}
