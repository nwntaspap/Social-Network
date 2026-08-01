package transport

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"social-network/internal/chat"
	"social-network/internal/chat/queries"
)

type mockChatUsers struct {
	conversations []queries.Conversation
	err           error
}

func (m *mockChatUsers) Resolve(_ context.Context, _ queries.GetChatUsersRequest) ([]queries.Conversation, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.conversations, nil
}

func TestGetConversations_MatchesFrontendChat(t *testing.T) {
	now := time.Date(2026, 8, 1, 12, 0, 0, 0, time.UTC)

	conv := queries.Conversation{
		ID:            "c1",
		OtherUser:     chat.UserRef{ID: "u1", Nickname: "Alice", AvatarURL: "/img/a.png"},
		UnreadCount:   2,
		IsOnline:      true,
		LastMessageAt: &now,
		CreatedAt:     now,
	}

	h := NewHandler(
		func(_ *http.Request) (string, bool) { return "me", true },
		nil,
		&mockChatUsers{conversations: []queries.Conversation{conv}},
	)
	srv := httptest.NewServer(http.HandlerFunc(h.GetConversations))
	defer srv.Close()

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, srv.URL+"/api/chat/users", nil)
	if err != nil {
		t.Fatalf("build request: %v", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var body struct {
		Data []map[string]any `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(body.Data) != 1 {
		t.Fatalf("len(data) = %d, want 1", len(body.Data))
	}
	d := body.Data[0]

	if got, ok := d["id"].(string); !ok || got != "c1" {
		t.Errorf("id = %#v, want %q", d["id"], "c1")
	}
	if got, ok := d["unreadCount"].(float64); !ok || got != 2 {
		t.Errorf("unreadCount = %#v, want 2", d["unreadCount"])
	}
	if _, ok := d["createdAt"].(string); !ok {
		t.Errorf("createdAt = %#v, want string", d["createdAt"])
	}
	if _, ok := d["user_id"]; ok {
		t.Error("snake_case user_id should not be present")
	}
	if _, ok := d["unread_count"]; ok {
		t.Error("snake_case unread_count should not be present")
	}

	participants, ok := d["participants"].([]any)
	if !ok || len(participants) != 1 {
		t.Fatalf("participants = %#v, want 1 entry", d["participants"])
	}
	p0, ok := participants[0].(map[string]any)
	if !ok {
		t.Fatalf("participant 0 = %#v, want object", participants[0])
	}

	if got, ok := p0["id"].(string); !ok || got != "u1" {
		t.Errorf("participant id = %#v, want %q", p0["id"], "u1")
	}
	if got, ok := p0["username"].(string); !ok || got != "Alice" {
		t.Errorf("username = %#v, want %q", p0["username"], "Alice")
	}
	if got, ok := p0["avatarUrl"].(string); !ok || got != "/img/a.png" {
		t.Errorf("avatarUrl = %#v, want %q", p0["avatarUrl"], "/img/a.png")
	}
	if got, ok := p0["isOnline"].(bool); !ok || !got {
		t.Errorf("isOnline = %#v, want true", p0["isOnline"])
	}
	if _, ok := p0["lastMessageAt"].(string); !ok {
		t.Errorf("lastMessageAt = %#v, want string", p0["lastMessageAt"])
	}
	if _, ok := p0["nickname"]; ok {
		t.Error("nickname should be renamed to username")
	}
}

func TestGetConversations_MethodNotAllowed(t *testing.T) {
	h := NewHandler(
		func(_ *http.Request) (string, bool) { return "me", true },
		nil,
		&mockChatUsers{},
	)
	srv := httptest.NewServer(http.HandlerFunc(h.GetConversations))
	defer srv.Close()

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, srv.URL+"/api/chat/users", nil)
	if err != nil {
		t.Fatalf("build request: %v", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", resp.StatusCode)
	}
}
