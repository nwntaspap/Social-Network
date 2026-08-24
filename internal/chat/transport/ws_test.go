package transport

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"social-network/internal/chat"
	"social-network/internal/chat/commands"
	"social-network/internal/chat/queries"
	"social-network/internal/core/realtime"
)

type mockWSSend struct {
	result commands.SendPrivateMessageResult
	err    error
}

func (m *mockWSSend) Execute(_ context.Context, _ commands.SendPrivateMessageCommand) (commands.SendPrivateMessageResult, error) {
	if m.err != nil {
		return commands.SendPrivateMessageResult{}, m.err
	}
	return m.result, nil
}

type mockWSGetChat struct {
	chat *chat.Chat
	err  error
}

func (m *mockWSGetChat) GetChat(_ context.Context, _ string) (*chat.Chat, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.chat, nil
}

type mockWSMarkRead struct {
	calls int
	err   error
}

func (m *mockWSMarkRead) MarkAsRead(_ context.Context, _, _ string, _ int) error {
	m.calls++
	return m.err
}

type mockWSTypingGate struct {
	err error
}

func (m *mockWSTypingGate) Validate(_ context.Context, _, _ string) error {
	return m.err
}

func newTestWSHandler(t *testing.T, hub *realtime.Hub, c *chat.Chat) (*WSHandler, *mockWSSend, *mockWSMarkRead) {
	t.Helper()
	send := &mockWSSend{}
	markRead := &mockWSMarkRead{}
	history := &mockChatHistory{messages: []*chat.Message{{ID: 1, ChatID: "c1", SenderID: "u2", Content: "hi"}}}
	return NewWSHandler(hub, send, history, markRead, &mockWSGetChat{chat: c}, &mockWSTypingGate{}), send, markRead
}

func readEnvelope(t *testing.T, client *realtime.Client) realtime.Envelope {
	t.Helper()
	select {
	case raw := <-client.SendChannel():
		var env realtime.Envelope
		if err := json.Unmarshal(raw, &env); err != nil {
			t.Fatalf("unmarshal envelope: %v", err)
		}
		return env
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for envelope")
		return realtime.Envelope{}
	}
}

// drainStatus discards the isOnlineStatus.update broadcast sent on register.
func drainStatus(t *testing.T, client *realtime.Client) {
	t.Helper()
	env := readEnvelope(t, client)
	if env.Type != realtime.TypeIsOnlineStatus {
		t.Fatalf("expected isOnlineStatus.update broadcast, got %s", env.Type)
	}
}

func TestWS_PingRespondsPong(t *testing.T) {
	hub := realtime.NewHub()
	h, _, _ := newTestWSHandler(t, hub, nil)
	client := realtime.NewClient("u1", hub, nil)

	env := realtime.Envelope{Type: realtime.TypePing, RequestID: "r1"}
	h.handlePing(client, env)

	reply := readEnvelope(t, client)
	if reply.Type != realtime.TypePong || reply.RequestID != "r1" {
		t.Fatalf("reply = %#v, want pong r1", reply)
	}
}

func TestWS_SendDeliversToBothParticipants(t *testing.T) {
	hub := realtime.NewHub()
	h, send, _ := newTestWSHandler(t, hub, &chat.Chat{ID: "c1", UserOneID: "u1", UserTwoID: "u2"})

	now := time.Now()
	send.result = commands.SendPrivateMessageResult{
		Message:     &chat.Message{ID: 1, ChatID: "c1", SenderID: "u1", Content: "hello", CreatedAt: now},
		RecipientID: "u2",
	}

	senderClient := realtime.NewClient("u1", hub, nil)
	receiverClient := realtime.NewClient("u2", hub, nil)
	hub.Register(senderClient)
	hub.Register(receiverClient)

	drainStatus(t, senderClient)
	drainStatus(t, receiverClient)
	drainStatus(t, senderClient) // receiver's register broadcast

	payload, _ := json.Marshal(realtime.SendPayload{ChatID: "c1", Content: "hello", ClientMessageID: "cm1"})
	h.handleSend(senderClient, realtime.Envelope{Type: realtime.TypeChatSend, Payload: payload})

	senderReply := readEnvelope(t, senderClient)
	receiverReply := readEnvelope(t, receiverClient)

	if senderReply.Type != realtime.TypeChatMessage {
		t.Fatalf("sender reply type = %s, want chat.message", senderReply.Type)
	}
	if receiverReply.Type != realtime.TypeChatMessage {
		t.Fatalf("receiver reply type = %s, want chat.message", receiverReply.Type)
	}

	var msg realtime.MessagePayload
	if err := json.Unmarshal(receiverReply.Payload, &msg); err != nil {
		t.Fatalf("unmarshal payload: %v", err)
	}
	if msg.SenderID != "u1" || msg.Content != "hello" {
		t.Fatalf("message = %#v", msg)
	}
}

func TestWS_SendRejectsNonParticipant(t *testing.T) {
	hub := realtime.NewHub()
	h, _, _ := newTestWSHandler(t, hub, &chat.Chat{ID: "c1", UserOneID: "u1", UserTwoID: "u2"})
	client := realtime.NewClient("u9", hub, nil)

	payload, _ := json.Marshal(realtime.SendPayload{ChatID: "c1", Content: "hello"})
	h.handleSend(client, realtime.Envelope{Type: realtime.TypeChatSend, Payload: payload})

	reply := readEnvelope(t, client)
	if reply.Type != realtime.TypeError {
		t.Fatalf("reply type = %s, want error", reply.Type)
	}
}

func TestWS_SendGateError(t *testing.T) {
	hub := realtime.NewHub()
	h, send, _ := newTestWSHandler(t, hub, &chat.Chat{ID: "c1", UserOneID: "u1", UserTwoID: "u2"})
	send.err = commands.ErrNotConnected

	client := realtime.NewClient("u1", hub, nil)
	payload, _ := json.Marshal(realtime.SendPayload{ChatID: "c1", Content: "hello"})
	h.handleSend(client, realtime.Envelope{Type: realtime.TypeChatSend, Payload: payload})

	reply := readEnvelope(t, client)
	if reply.Type != realtime.TypeError {
		t.Fatalf("reply type = %s, want error", reply.Type)
	}
}

func TestWS_HistoryReturnsResult(t *testing.T) {
	hub := realtime.NewHub()
	h, _, _ := newTestWSHandler(t, hub, nil)
	client := realtime.NewClient("u1", hub, nil)

	payload, _ := json.Marshal(realtime.HistoryPayload{ChatID: "c1"})
	h.handleHistory(client, realtime.Envelope{Type: realtime.TypeChatHistory, RequestID: "r2", Payload: payload})

	reply := readEnvelope(t, client)
	if reply.Type != realtime.TypeHistoryResult || reply.RequestID != "r2" {
		t.Fatalf("reply = %#v, want history_result r2", reply)
	}
	var messages []chat.Message
	if err := json.Unmarshal(reply.Payload, &messages); err != nil {
		t.Fatalf("unmarshal messages: %v", err)
	}
	if len(messages) != 1 || messages[0].Content != "hi" {
		t.Fatalf("messages = %#v", messages)
	}
}

func TestWS_MarkRead(t *testing.T) {
	hub := realtime.NewHub()
	h, _, markRead := newTestWSHandler(t, hub, &chat.Chat{ID: "c1", UserOneID: "u1", UserTwoID: "u2"})
	client := realtime.NewClient("u1", hub, nil)

	payload, _ := json.Marshal(realtime.MarkReadPayload{ChatID: "c1", UpToMessageID: 5})
	h.handleMarkRead(client, realtime.Envelope{Type: realtime.TypeMarkRead, Payload: payload})

	if markRead.calls != 1 {
		t.Fatalf("markRead.calls = %d, want 1", markRead.calls)
	}
}

func TestWS_MarkReadRejectsNonParticipant(t *testing.T) {
	hub := realtime.NewHub()
	h, _, markRead := newTestWSHandler(t, hub, &chat.Chat{ID: "c1", UserOneID: "u1", UserTwoID: "u2"})
	client := realtime.NewClient("u9", hub, nil)

	payload, _ := json.Marshal(realtime.MarkReadPayload{ChatID: "c1", UpToMessageID: 5})
	h.handleMarkRead(client, realtime.Envelope{Type: realtime.TypeMarkRead, Payload: payload})

	reply := readEnvelope(t, client)
	if reply.Type != realtime.TypeError {
		t.Fatalf("reply type = %s, want error", reply.Type)
	}
	if markRead.calls != 0 {
		t.Fatalf("markRead.calls = %d, want 0", markRead.calls)
	}
}

func TestWS_OpenMarksRead(t *testing.T) {
	hub := realtime.NewHub()
	last := 5
	h, _, markRead := newTestWSHandler(t, hub, &chat.Chat{ID: "c1", UserOneID: "u1", UserTwoID: "u2", LastMessageID: &last})
	client := realtime.NewClient("u1", hub, nil)

	payload, _ := json.Marshal(realtime.ChatOpenClosePayload{ChatID: "c1"})
	h.handleOpen(client, realtime.Envelope{Type: realtime.TypeChatOpen, Payload: payload})

	if markRead.calls != 1 {
		t.Fatalf("markRead.calls = %d, want 1", markRead.calls)
	}
	if len(hub.GetObserversForChat("c1", "u2")) == 0 {
		t.Fatal("expected client to observe the opened chat")
	}
}

func TestWS_OpenWithoutLastMessageSkipsMarkRead(t *testing.T) {
	hub := realtime.NewHub()
	h, _, markRead := newTestWSHandler(t, hub, &chat.Chat{ID: "c1", UserOneID: "u1", UserTwoID: "u2"})
	client := realtime.NewClient("u1", hub, nil)

	payload, _ := json.Marshal(realtime.ChatOpenClosePayload{ChatID: "c1"})
	h.handleOpen(client, realtime.Envelope{Type: realtime.TypeChatOpen, Payload: payload})

	if markRead.calls != 0 {
		t.Fatalf("markRead.calls = %d, want 0", markRead.calls)
	}
}

func TestWS_OpenRejectsNonParticipant(t *testing.T) {
	hub := realtime.NewHub()
	last := 3
	h, _, markRead := newTestWSHandler(t, hub, &chat.Chat{ID: "c1", UserOneID: "u1", UserTwoID: "u2", LastMessageID: &last})
	client := realtime.NewClient("u9", hub, nil)

	payload, _ := json.Marshal(realtime.ChatOpenClosePayload{ChatID: "c1"})
	h.handleOpen(client, realtime.Envelope{Type: realtime.TypeChatOpen, Payload: payload})

	reply := readEnvelope(t, client)
	if reply.Type != realtime.TypeError {
		t.Fatalf("reply type = %s, want error", reply.Type)
	}
	if markRead.calls != 0 {
		t.Fatalf("markRead.calls = %d, want 0", markRead.calls)
	}
}

func TestWS_TypingSuppressedWhenNotConnected(t *testing.T) {
	hub := realtime.NewHub()
	h, _, _ := newTestWSHandler(t, hub, &chat.Chat{ID: "c1", UserOneID: "u1", UserTwoID: "u2"})
	h.typingGate = &mockWSTypingGate{err: commands.ErrNotConnected}

	typer := realtime.NewClient("u1", hub, nil)
	observer := realtime.NewClient("u2", hub, nil)
	// Both have the chat open; observer would otherwise receive the indicator.
	hub.OpenChat(typer, "c1")
	hub.OpenChat(observer, "c1")

	payload, _ := json.Marshal(realtime.ChatTypingPayload{ChatID: "c1"})
	h.handleTyping(typer, realtime.Envelope{Type: realtime.TypeTyping, Payload: payload})

	// The typer gets an error reply; the observer must receive nothing.
	reply := readEnvelope(t, typer)
	if reply.Type != realtime.TypeError {
		t.Fatalf("reply type = %s, want error", reply.Type)
	}
	select {
	case msg := <-observer.SendChannel():
		t.Fatalf("observer received typing indicator: %s", msg)
	case <-time.After(100 * time.Millisecond):
	}
}

func TestWS_TypingDeliversWhenConnected(t *testing.T) {
	hub := realtime.NewHub()
	h, _, _ := newTestWSHandler(t, hub, &chat.Chat{ID: "c1", UserOneID: "u1", UserTwoID: "u2"})
	h.typingGate = &mockWSTypingGate{}

	typer := realtime.NewClient("u1", hub, nil)
	observer := realtime.NewClient("u2", hub, nil)
	hub.OpenChat(typer, "c1")
	hub.OpenChat(observer, "c1")

	payload, _ := json.Marshal(realtime.ChatTypingPayload{ChatID: "c1"})
	h.handleTyping(typer, realtime.Envelope{Type: realtime.TypeTyping, Payload: payload})

	reply := readEnvelope(t, observer)
	if reply.Type != realtime.TypeIsTyping {
		t.Fatalf("reply type = %s, want chat.is_typing", reply.Type)
	}
}

func TestWS_HistoryError(t *testing.T) {
	hub := realtime.NewHub()
	h := NewWSHandler(hub, &mockWSSend{}, &mockChatHistory{err: chat.ErrNotParticipant}, &mockWSMarkRead{}, &mockWSGetChat{}, &mockWSTypingGate{})
	client := realtime.NewClient("u1", hub, nil)

	payload, _ := json.Marshal(realtime.HistoryPayload{ChatID: "c1"})
	h.handleHistory(client, realtime.Envelope{Type: realtime.TypeChatHistory, Payload: payload})

	reply := readEnvelope(t, client)
	if reply.Type != realtime.TypeError {
		t.Fatalf("reply type = %s, want error", reply.Type)
	}
}

var (
	_ WSSendExecutor = (*mockWSSend)(nil)
	_ WSChatGetter   = (*mockWSGetChat)(nil)
	_ WSMarkAsRead   = (*mockWSMarkRead)(nil)
	_ WSTypingGate   = (*mockWSTypingGate)(nil)
	_                = queries.GetChatHistoryQuery{}
)
