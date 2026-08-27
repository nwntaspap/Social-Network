package transport

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"social-network/internal/group"
	"social-network/internal/group/commands"
	"social-network/internal/group/queries"
	"social-network/internal/platform/logger"
)

type mockPendingInvitations struct {
	result *queries.GetPendingInvitationsResult
	err    error
	userID string
}

func (m *mockPendingInvitations) Resolve(_ context.Context, q queries.GetPendingInvitationsQuery) (*queries.GetPendingInvitationsResult, error) {
	m.userID = q.UserID
	if m.err != nil {
		return nil, m.err
	}
	return m.result, nil
}

type mockRespondInvite struct {
	lastGroupID string
	lastAccept  bool
	err         error
	result      commands.RespondInviteResult
}

func (m *mockRespondInvite) Execute(_ context.Context, cmd commands.RespondInviteCommand) (commands.RespondInviteResult, error) {
	m.lastGroupID = cmd.GroupID
	m.lastAccept = cmd.Accept
	return m.result, m.err
}

func newInvitationTestHandler(
	extractUser UserExtractor,
	respond RespondInviteExecutor,
	pending GetPendingInvitationsResolver,
) *Handler {
	return NewHandler(
		extractUser,
		&mockGroupUserLookup{},
		nil,
		nil,
		respond,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		pending,
		nil,
		nil,
		nil,
		nil,
		logger.New(io.Discard, logger.LevelOff),
	)
}

func invitationRoutesMux(h *Handler) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/groups/invitations/pending", h.GetPendingInvitations)
	mux.HandleFunc("POST /api/groups/{groupId}/invite/respond", h.RespondInvite)
	return mux
}

func TestGetPendingInvitations_ReturnsListWithGroup(t *testing.T) {
	inv := group.Invitation{ID: "i1", GroupID: "g1", InviterID: "u2", InviteeID: "u1", CreatedAt: time.Now()}
	g := group.Group{ID: "g1", Title: "Go Meetup", CreatorID: "u2"}

	pending := &mockPendingInvitations{
		result: &queries.GetPendingInvitationsResult{
			Invitations: []queries.InvitationWithGroup{{Invitation: inv, Group: g}},
		},
	}
	h := newInvitationTestHandler(
		func(_ *http.Request) (string, bool) { return "u1", true },
		nil,
		pending,
	)
	srv := httptest.NewServer(invitationRoutesMux(h))
	defer srv.Close()

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, srv.URL+"/api/groups/invitations/pending", nil)
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
	if pending.userID != "u1" {
		t.Errorf("userID forwarded = %q, want %q", pending.userID, "u1")
	}

	var body struct {
		Data []struct {
			ID      string `json:"id"`
			GroupID string `json:"groupId"`
			Group   struct {
				ID    string `json:"id"`
				Title string `json:"title"`
			} `json:"group"`
			InviterID string `json:"inviterId"`
			InviteeID string `json:"inviteeId"`
			Inviter   struct {
				ID string `json:"id"`
			} `json:"inviter"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}

	if len(body.Data) != 1 {
		t.Fatalf("len(data) = %d, want 1", len(body.Data))
	}
	d := body.Data[0]
	if d.ID != "i1" || d.GroupID != "g1" {
		t.Errorf("unexpected invitation id/group: %s/%s", d.ID, d.GroupID)
	}
	if d.Group.ID != "g1" || d.Group.Title != "Go Meetup" {
		t.Errorf("group brief = %+v, want g1/Go Meetup", d.Group)
	}
	if d.InviterID != "u2" || d.InviteeID != "u1" {
		t.Errorf("unexpected parties: inviter=%s invitee=%s", d.InviterID, d.InviteeID)
	}
	if d.Inviter.ID == "" {
		t.Error("expected inviter user object to be populated")
	}
}

func TestGetPendingInvitations_RequiresAuth(t *testing.T) {
	h := newInvitationTestHandler(
		func(_ *http.Request) (string, bool) { return "", false },
		nil,
		&mockPendingInvitations{},
	)
	srv := httptest.NewServer(invitationRoutesMux(h))
	defer srv.Close()

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, srv.URL+"/api/groups/invitations/pending", nil)
	if err != nil {
		t.Fatalf("build request: %v", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", resp.StatusCode)
	}
}

func TestRespondInvite_Accepts(t *testing.T) {
	respond := &mockRespondInvite{}
	h := newInvitationTestHandler(
		func(_ *http.Request) (string, bool) { return "u1", true },
		respond,
		nil,
	)
	srv := httptest.NewServer(invitationRoutesMux(h))
	defer srv.Close()

	req, err := http.NewRequestWithContext(
		context.Background(),
		http.MethodPost,
		srv.URL+"/api/groups/g1/invite/respond",
		strings.NewReader(`{"action":"accept"}`),
	)
	if err != nil {
		t.Fatalf("build request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	if respond.lastGroupID != "g1" {
		t.Errorf("groupID forwarded = %q, want %q", respond.lastGroupID, "g1")
	}
	if !respond.lastAccept {
		t.Error("expected accept=true to be forwarded")
	}
}

func TestRespondInvite_Declines(t *testing.T) {
	respond := &mockRespondInvite{}
	h := newInvitationTestHandler(
		func(_ *http.Request) (string, bool) { return "u1", true },
		respond,
		nil,
	)
	srv := httptest.NewServer(invitationRoutesMux(h))
	defer srv.Close()

	req, err := http.NewRequestWithContext(
		context.Background(),
		http.MethodPost,
		srv.URL+"/api/groups/g1/invite/respond",
		strings.NewReader(`{"action":"decline"}`),
	)
	if err != nil {
		t.Fatalf("build request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	if respond.lastAccept {
		t.Error("expected accept=false to be forwarded")
	}
}

func TestRespondInvite_Pending(t *testing.T) {
	respond := &mockRespondInvite{result: commands.RespondInvitePending}
	h := newInvitationTestHandler(
		func(_ *http.Request) (string, bool) { return "u1", true },
		respond,
		nil,
	)
	srv := httptest.NewServer(invitationRoutesMux(h))
	defer srv.Close()

	req, err := http.NewRequestWithContext(
		context.Background(),
		http.MethodPost,
		srv.URL+"/api/groups/g1/invite/respond",
		strings.NewReader(`{"action":"accept"}`),
	)
	if err != nil {
		t.Fatalf("build request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	var body struct {
		Data struct {
			Status string `json:"status"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if body.Data.Status != "pending" {
		t.Errorf("status = %q, want %q", body.Data.Status, "pending")
	}
}

func TestRespondInvite_ForwardsError(t *testing.T) {
	respond := &mockRespondInvite{err: group.ErrInvitationNotFound}
	h := newInvitationTestHandler(
		func(_ *http.Request) (string, bool) { return "u1", true },
		respond,
		nil,
	)
	srv := httptest.NewServer(invitationRoutesMux(h))
	defer srv.Close()

	req, err := http.NewRequestWithContext(
		context.Background(),
		http.MethodPost,
		srv.URL+"/api/groups/g1/invite/respond",
		strings.NewReader(`{"action":"accept"}`),
	)
	if err != nil {
		t.Fatalf("build request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}
