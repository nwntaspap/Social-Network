package transport

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"social-network/internal/group"
	"social-network/internal/group/queries"
	"social-network/internal/platform/logger"
)

type mockPendingJoinRequests struct {
	result *queries.GetPendingJoinRequestsResult
	err    error
	query  queries.GetPendingJoinRequestsQuery
}

func (m *mockPendingJoinRequests) Resolve(_ context.Context, q queries.GetPendingJoinRequestsQuery) (*queries.GetPendingJoinRequestsResult, error) {
	m.query = q
	if m.err != nil {
		return nil, m.err
	}
	return m.result, nil
}

func newRequestsTestHandler(extractUser UserExtractor, pending GetPendingJoinRequestsResolver) *Handler {
	return NewHandler(
		extractUser,
		&mockGroupUserLookup{},
		nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
		pending,
		nil,
		nil,
		logger.New(io.Discard, logger.LevelOff),
	)
}

func requestsRoutesMux(h *Handler) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/groups/{groupId}/requests/pending", h.GetPendingJoinRequests)
	return mux
}

func TestGetPendingJoinRequests_ReturnsListWithRequester(t *testing.T) {
	jr := group.JoinRequest{ID: "jr1", GroupID: "g1", RequesterID: "u3", CreatedAt: time.Now()}
	g := group.Group{ID: "g1", Title: "Go Meetup", CreatorID: "u1"}

	pending := &mockPendingJoinRequests{
		result: &queries.GetPendingJoinRequestsResult{
			Requests: []queries.JoinRequestWithGroup{{Request: jr, Group: g}},
		},
	}
	h := newRequestsTestHandler(
		func(_ *http.Request) (string, bool) { return "u1", true },
		pending,
	)
	srv := httptest.NewServer(requestsRoutesMux(h))
	defer srv.Close()

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, srv.URL+"/api/groups/g1/requests/pending", nil)
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
	if pending.query.GroupID != "g1" || pending.query.UserID != "u1" {
		t.Errorf("query = %+v, want group=g1 user=u1", pending.query)
	}

	var body struct {
		Data []struct {
			ID      string `json:"id"`
			GroupID string `json:"groupId"`
			Group   struct {
				ID    string `json:"id"`
				Title string `json:"title"`
			} `json:"group"`
			RequesterID string `json:"requesterId"`
			Requester   struct {
				ID string `json:"id"`
			} `json:"requester"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}

	if len(body.Data) != 1 {
		t.Fatalf("len(data) = %d, want 1", len(body.Data))
	}
	d := body.Data[0]
	if d.ID != "jr1" || d.GroupID != "g1" {
		t.Errorf("unexpected request id/group: %s/%s", d.ID, d.GroupID)
	}
	if d.Group.ID != "g1" || d.Group.Title != "Go Meetup" {
		t.Errorf("group brief = %+v, want g1/Go Meetup", d.Group)
	}
	if d.RequesterID != "u3" {
		t.Errorf("requesterId = %q, want u3", d.RequesterID)
	}
	if d.Requester.ID == "" {
		t.Error("expected requester user object to be populated")
	}
}

func TestGetPendingJoinRequests_RequiresAuth(t *testing.T) {
	h := newRequestsTestHandler(
		func(_ *http.Request) (string, bool) { return "", false },
		&mockPendingJoinRequests{},
	)
	srv := httptest.NewServer(requestsRoutesMux(h))
	defer srv.Close()

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, srv.URL+"/api/groups/g1/requests/pending", nil)
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

func TestGetPendingJoinRequests_ForbiddenForNonAdmin(t *testing.T) {
	pending := &mockPendingJoinRequests{err: group.ErrNotAdmin}
	h := newRequestsTestHandler(
		func(_ *http.Request) (string, bool) { return "u2", true },
		pending,
	)
	srv := httptest.NewServer(requestsRoutesMux(h))
	defer srv.Close()

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, srv.URL+"/api/groups/g1/requests/pending", nil)
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

func TestGetPendingJoinRequests_ForwardsError(t *testing.T) {
	pending := &mockPendingJoinRequests{err: group.ErrJoinRequestNotFound}
	h := newRequestsTestHandler(
		func(_ *http.Request) (string, bool) { return "u1", true },
		pending,
	)
	srv := httptest.NewServer(requestsRoutesMux(h))
	defer srv.Close()

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, srv.URL+"/api/groups/g1/requests/pending", nil)
	if err != nil {
		t.Fatalf("build request: %v", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}
