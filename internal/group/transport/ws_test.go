package transport

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"social-network/internal/core/realtime"
	"social-network/internal/group"
	"social-network/internal/group/commands"
	"social-network/internal/group/queries"
)

type mockGroupSend struct {
	result commands.SendGroupMessageResult
	err    error
}

func (m *mockGroupSend) Execute(_ context.Context, _ commands.SendGroupMessageCommand) (commands.SendGroupMessageResult, error) {
	if m.err != nil {
		return commands.SendGroupMessageResult{}, m.err
	}
	return m.result, nil
}

type mockGroupChatResolver struct {
	result *queries.GetGroupChatResult
	err    error
}

func (m *mockGroupChatResolver) Resolve(_ context.Context, _ queries.GetGroupChatQuery) (*queries.GetGroupChatResult, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.result, nil
}

type mockGroupMemberIDs struct {
	ids []string
	err error
}

func (m *mockGroupMemberIDs) Resolve(_ context.Context, _ queries.ListGroupMemberIDsQuery) ([]string, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.ids, nil
}

func newTestGroupWSHandler(hub *realtime.Hub, send *mockGroupSend, history *mockGroupChatResolver, memberIDs *mockGroupMemberIDs) *GroupWSHandler {
	return NewGroupWSHandler(hub, send, history, memberIDs)
}

func TestGroupWS_SendBroadcastsToMembers(t *testing.T) {
	hub := realtime.NewHub()
	send := &mockGroupSend{result: commands.SendGroupMessageResult{
		GroupID: "g1",
		Message: &group.ChatMessage{ID: "m1", GroupID: "g1", SenderID: "u1", Content: "hello"},
	}}
	history := &mockGroupChatResolver{result: &queries.GetGroupChatResult{}}
	memberIDs := &mockGroupMemberIDs{ids: []string{"u1", "u2", "u3"}}
	h := newTestGroupWSHandler(hub, send, history, memberIDs)

	u1 := realtime.NewClient("u1", hub, nil)
	u2 := realtime.NewClient("u2", hub, nil)
	u3 := realtime.NewClient("u3", hub, nil)
	hub.Register(u1)
	hub.Register(u2)
	hub.Register(u3)

	drainStatusN(t, u1, 3)
	drainStatusN(t, u2, 2)
	drainStatusN(t, u3, 1)

	payload, _ := json.Marshal(realtime.GroupChatSendPayload{GroupID: "g1", Content: "hello"})
	h.handleSend(u1, realtime.Envelope{Type: realtime.TypeGroupChatSend, Payload: payload})

	for name, c := range map[string]*realtime.Client{"u1": u1, "u2": u2, "u3": u3} {
		env := readGroupEnvelope(t, c)
		if env.Type != realtime.TypeGroupChatMessage {
			t.Fatalf("%s got %s, want group_chat.message", name, env.Type)
		}
		var msg realtime.GroupChatMessagePayload
		if err := json.Unmarshal(env.Payload, &msg); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		if msg.GroupID != "g1" || msg.Content != "hello" || msg.SenderID != "u1" {
			t.Fatalf("%s message = %#v", name, msg)
		}
	}
}

func TestGroupWS_SendRejectsNonMember(t *testing.T) {
	hub := realtime.NewHub()
	send := &mockGroupSend{err: group.ErrNotMember}
	h := newTestGroupWSHandler(hub, send, &mockGroupChatResolver{}, &mockGroupMemberIDs{})
	client := realtime.NewClient("u9", hub, nil)

	payload, _ := json.Marshal(realtime.GroupChatSendPayload{GroupID: "g1", Content: "hello"})
	h.handleSend(client, realtime.Envelope{Type: realtime.TypeGroupChatSend, Payload: payload})

	env := readGroupEnvelope(t, client)
	if env.Type != realtime.TypeError {
		t.Fatalf("got %s, want error", env.Type)
	}
}

func TestGroupWS_HistoryReturnsMessages(t *testing.T) {
	hub := realtime.NewHub()
	history := &mockGroupChatResolver{result: &queries.GetGroupChatResult{
		Messages: []group.ChatMessage{{ID: "m1", GroupID: "g1", SenderID: "u2", Content: "hi"}},
	}}
	h := newTestGroupWSHandler(hub, &mockGroupSend{}, history, &mockGroupMemberIDs{})
	client := realtime.NewClient("u1", hub, nil)

	payload, _ := json.Marshal(realtime.GroupChatHistoryPayload{GroupID: "g1", Limit: 10})
	h.handleHistory(client, realtime.Envelope{Type: realtime.TypeGroupChatHistory, RequestID: "r1", Payload: payload})

	env := readGroupEnvelope(t, client)
	if env.Type != realtime.TypeGroupChatHistResult || env.RequestID != "r1" {
		t.Fatalf("got %s/%s, want group_chat.history_result r1", env.Type, env.RequestID)
	}
	var messages []group.ChatMessage
	if err := json.Unmarshal(env.Payload, &messages); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(messages) != 1 || messages[0].Content != "hi" {
		t.Fatalf("messages = %#v", messages)
	}
}

func readGroupEnvelope(t *testing.T, client *realtime.Client) realtime.Envelope {
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

func drainStatusN(t *testing.T, client *realtime.Client, n int) {
	t.Helper()
	for range n {
		env := readGroupEnvelope(t, client)
		if env.Type != realtime.TypeIsOnlineStatus {
			t.Fatalf("expected isOnlineStatus.update, got %s", env.Type)
		}
	}
}

var (
	_ SendGroupMessageExecutor = (*mockGroupSend)(nil)
	_ GetGroupChatResolver     = (*mockGroupChatResolver)(nil)
	_ GroupMemberIDsResolver   = (*mockGroupMemberIDs)(nil)
)
