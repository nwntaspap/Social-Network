package commands

import (
	"context"
	"testing"
)

func TestCreateGroupHandler_Execute(t *testing.T) {
	ctx := context.Background()

	t.Run("returns error when user ID is empty", func(t *testing.T) {
		handler := NewCreateGroupHandler(nil)
		_, err := handler.Execute(ctx, CreateGroupCommand{
			UserID: "",
			Title:  "Test Group",
		})
		if err == nil {
			t.Error("expected error for empty user ID")
		}
	})

	t.Run("returns error when title is empty", func(t *testing.T) {
		handler := NewCreateGroupHandler(nil)
		_, err := handler.Execute(ctx, CreateGroupCommand{
			UserID: "user-123",
			Title:  "",
		})
		if err == nil {
			t.Error("expected error for empty title")
		}
	})

	t.Run("returns error when title is too long", func(t *testing.T) {
		handler := NewCreateGroupHandler(nil)
		longTitle := string(make([]byte, 101))
		_, err := handler.Execute(ctx, CreateGroupCommand{
			UserID: "user-123",
			Title:  longTitle,
		})
		if err == nil {
			t.Error("expected error for title exceeding 100 characters")
		}
	})

	t.Run("returns error when description is too long", func(t *testing.T) {
		handler := NewCreateGroupHandler(nil)
		longDesc := string(make([]byte, 501))
		_, err := handler.Execute(ctx, CreateGroupCommand{
			UserID:      "user-123",
			Title:       "Test Group",
			Description: longDesc,
		})
		if err == nil {
			t.Error("expected error for description exceeding 500 characters")
		}
	})
}
