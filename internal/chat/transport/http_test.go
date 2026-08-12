package transport

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"social-network/internal/chat"
	"social-network/internal/chat/commands"
	"social-network/internal/chat/queries"
	"social-network/internal/platform/logger"
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

type mockChatHistory struct {
	messages []*chat.Message
	err      error
}

func (m *mockChatHistory) Resolve(_ context.Context, _ queries.GetChatHistoryQuery) ([]*chat.Message, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.messages, nil
}

type mockChatUserLookup struct {
	result *UserResult
	err    error
}

func (m *mockChatUserLookup) GetUserByID(_ context.Context, _ string) (*UserResult, error) {
	if m.err != nil {
		return nil, m.err
	}
	if m.result != nil {
		return m.result, nil
	}
	return &UserResult{ID: "u1", Username: "alice"}, nil
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
		logger.New(io.Discard, logger.LevelOff),
		func(_ *http.Request) (string, bool) { return "me", true },
		&mockChatUserLookup{},
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
		logger.New(io.Discard, logger.LevelOff),
		func(_ *http.Request) (string, bool) { return "me", true },
		&mockChatUserLookup{},
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

func TestGetChatHistory_MatchesFrontendChatMessage(t *testing.T) {
	now := time.Date(2026, 8, 1, 12, 0, 0, 0, time.UTC)

	msg := &chat.Message{ID: 7, ChatID: "c1", SenderID: "u1", Content: "hello", CreatedAt: now}

	h := NewHandler(
		logger.New(io.Discard, logger.LevelOff),
		func(_ *http.Request) (string, bool) { return "u1", true },
		&mockChatUserLookup{result: &UserResult{ID: "u1", Username: "alice", FirstName: "Alice"}},
		&mockChatHistory{messages: []*chat.Message{msg}},
		&mockChatUsers{},
	)
	srv := httptest.NewServer(http.HandlerFunc(h.GetChatHistory))
	defer srv.Close()

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, srv.URL+"/api/chat/history?chatId=c1", nil)
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

	if got, ok := d["id"].(string); !ok || got != "7" {
		t.Errorf("id = %#v, want string %q", d["id"], "7")
	}
	if got, ok := d["senderId"].(string); !ok || got != "u1" {
		t.Errorf("senderId = %#v, want %q", d["senderId"], "u1")
	}
	if got, ok := d["content"].(string); !ok || got != "hello" {
		t.Errorf("content = %#v, want %q", d["content"], "hello")
	}
	if got, ok := d["type"].(string); !ok || got != "private" {
		t.Errorf("type = %#v, want %q", d["type"], "private")
	}
	if _, ok := d["createdAt"].(string); !ok {
		t.Errorf("createdAt = %#v, want string", d["createdAt"])
	}

	sender, ok := d["sender"].(map[string]any)
	if !ok {
		t.Fatalf("sender = %#v, want object", d["sender"])
	}
	if got, ok := sender["username"].(string); !ok || got != "alice" {
		t.Errorf("sender.username = %#v, want %q", sender["username"], "alice")
	}
	if got, ok := sender["firstName"].(string); !ok || got != "Alice" {
		t.Errorf("sender.firstName = %#v, want %q", sender["firstName"], "Alice")
	}

	for _, snake := range []string{"chat_id", "sender_id", "created_at"} {
		if _, ok := d[snake]; ok {
			t.Errorf("snake_case key %q should not be present", snake)
		}
	}
}

type mockStartChat struct {
	result commands.OpenPrivateChatResult
	err    error
}

func (m *mockStartChat) Execute(_ context.Context, _ commands.OpenPrivateChatCommand) (commands.OpenPrivateChatResult, error) {
	if m.err != nil {
		return commands.OpenPrivateChatResult{}, m.err
	}
	return m.result, nil
}

func TestStartChat_ReturnsChat(t *testing.T) {
	h := NewHandlerWithStart(
		logger.New(io.Discard, logger.LevelOff),
		func(_ *http.Request) (string, bool) { return "me", true },
		&mockChatUserLookup{},
		nil,
		&mockChatUsers{},
		&mockStartChat{result: commands.OpenPrivateChatResult{Chat: &chat.Chat{ID: "c1", UserOneID: "me", UserTwoID: "u2"}}},
	)
	srv := httptest.NewServer(http.HandlerFunc(h.StartChat))
	defer srv.Close()

	body := bytes.NewBufferString(`{"userId":"u2"}`)
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, srv.URL+"/api/chat/start", body)
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

	var payload struct {
		Data *chat.Chat `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if payload.Data == nil || payload.Data.ID != "c1" {
		t.Fatalf("data = %#v, want chat c1", payload.Data)
	}
}

func TestStartChat_GateRejection(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want int
	}{
		{"not connected", commands.ErrNotConnected, http.StatusForbidden},
		{"cannot message", commands.ErrCannotMessage, http.StatusForbidden},
		{"missing user", errors.New("boom"), http.StatusInternalServerError},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			h := NewHandlerWithStart(
				logger.New(io.Discard, logger.LevelOff),
				func(_ *http.Request) (string, bool) { return "me", true },
				&mockChatUserLookup{},
				nil,
				&mockChatUsers{},
				&mockStartChat{err: tc.err},
			)
			srv := httptest.NewServer(http.HandlerFunc(h.StartChat))
			defer srv.Close()

			body := bytes.NewBufferString(`{"userId":"u2"}`)
			req, _ := http.NewRequestWithContext(context.Background(), http.MethodPost, srv.URL+"/api/chat/start", body)
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Fatalf("request: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != tc.want {
				t.Fatalf("expected %d, got %d", tc.want, resp.StatusCode)
			}
		})
	}
}

func TestStartChat_BadPayload(t *testing.T) {
	cases := []struct {
		name string
		body string
	}{
		{"empty user id", `{"userId":""}`},
		{"self chat", `{"userId":"me"}`},
		{"invalid json", `{not json`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			h := NewHandlerWithStart(
				logger.New(io.Discard, logger.LevelOff),
				func(_ *http.Request) (string, bool) { return "me", true },
				&mockChatUserLookup{},
				nil,
				&mockChatUsers{},
				&mockStartChat{},
			)
			srv := httptest.NewServer(http.HandlerFunc(h.StartChat))
			defer srv.Close()

			body := bytes.NewBufferString(tc.body)
			req, _ := http.NewRequestWithContext(context.Background(), http.MethodPost, srv.URL+"/api/chat/start", body)
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Fatalf("request: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusBadRequest {
				t.Fatalf("expected 400, got %d", resp.StatusCode)
			}
		})
	}
}

func TestGetChatHistory_RejectsNonParticipant(t *testing.T) {
	h := NewHandler(
		logger.New(io.Discard, logger.LevelOff),
		func(_ *http.Request) (string, bool) { return "intruder", true },
		&mockChatUserLookup{},
		&mockChatHistory{err: chat.ErrNotParticipant},
		&mockChatUsers{},
	)
	srv := httptest.NewServer(http.HandlerFunc(h.GetChatHistory))
	defer srv.Close()

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, srv.URL+"/api/chat/history?chatId=c1", nil)
	if err != nil {
		t.Fatalf("build request: %v", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", resp.StatusCode)
	}
}

func TestGetChatHistory_RejectsDisconnected(t *testing.T) {
	h := NewHandler(
		logger.New(io.Discard, logger.LevelOff),
		func(_ *http.Request) (string, bool) { return "u1", true },
		&mockChatUserLookup{},
		&mockChatHistory{err: chat.ErrNotConnected},
		&mockChatUsers{},
	)
	srv := httptest.NewServer(http.HandlerFunc(h.GetChatHistory))
	defer srv.Close()

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, srv.URL+"/api/chat/history?chatId=c1", nil)
	if err != nil {
		t.Fatalf("build request: %v", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "You are not connected to this user") {
		t.Fatalf("expected connection error message, got %s", body)
	}
}

func TestStartChat_Unauthenticated(t *testing.T) {
	h := NewHandlerWithStart(
		logger.New(io.Discard, logger.LevelOff),
		func(_ *http.Request) (string, bool) { return "", false },
		&mockChatUserLookup{},
		nil,
		&mockChatUsers{},
		&mockStartChat{},
	)
	srv := httptest.NewServer(http.HandlerFunc(h.StartChat))
	defer srv.Close()

	body := strings.NewReader(`{"userId":"u2"}`)
	req, _ := http.NewRequestWithContext(context.Background(), http.MethodPost, srv.URL+"/api/chat/start", body)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", resp.StatusCode)
	}
}
