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

type mockListMyGroups struct {
	result *queries.ListMyGroupsResult
	err    error
}

func (m *mockListMyGroups) Resolve(_ context.Context, _ queries.ListMyGroupsQuery) (*queries.ListMyGroupsResult, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.result, nil
}

type mockGetGroupPresence struct {
	result *queries.GetGroupPresenceResult
	err    error
}

func (m *mockGetGroupPresence) Resolve(_ context.Context, _ queries.GetGroupPresenceQuery) (*queries.GetGroupPresenceResult, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.result, nil
}

func presenceTestHandler(listMyGroups ListMyGroupsResolver, presence GetGroupPresenceResolver) *Handler {
	return NewHandler(
		func(_ *http.Request) (string, bool) { return "u1", true },
		&mockGroupUserLookup{},
		nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		&mockGroupMembers{total: 5},
		nil,
		nil,
		nil,
		listMyGroups,
		presence,
		logger.New(io.Discard, logger.LevelOff),
	)
}

func presenceRoutesMux(h *Handler) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/groups/mine", h.ListMyGroups)
	mux.HandleFunc("GET /api/groups/{groupId}/presence", h.GetGroupPresence)
	return mux
}

func doGet(t *testing.T, url string) *http.Response {
	t.Helper()
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, url, nil)
	if err != nil {
		t.Fatalf("build request: %v", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	return resp
}

func TestListMyGroups_ReturnsMemberGroups(t *testing.T) {
	g := group.Group{ID: "g1", Title: "Go", CreatorID: "u2", MembershipStatus: "member", CreatedAt: time.Now()}
	h := presenceTestHandler(
		&mockListMyGroups{result: &queries.ListMyGroupsResult{Groups: []group.Group{g}, Total: 1}},
		nil,
	)
	srv := httptest.NewServer(presenceRoutesMux(h))
	defer srv.Close()

	resp := doGet(t, srv.URL+"/api/groups/mine")
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var body struct {
		Data struct {
			Groups []map[string]any `json:"data"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(body.Data.Groups) != 1 {
		t.Fatalf("len(data) = %d, want 1", len(body.Data.Groups))
	}
	d := body.Data.Groups[0]
	if d["id"] != "g1" || d["title"] != "Go" {
		t.Errorf("group = %#v, want g1/Go", d)
	}
	if count, ok := d["membersCount"].(float64); !ok || count != 5 {
		t.Errorf("membersCount = %v, want 5", d["membersCount"])
	}
	if d["membershipStatus"] != "member" {
		t.Errorf("membershipStatus = %v, want member", d["membershipStatus"])
	}
}

func TestGetGroupPresence_ReturnsCounts(t *testing.T) {
	h := presenceTestHandler(
		nil,
		&mockGetGroupPresence{result: &queries.GetGroupPresenceResult{
			GroupID: "g1",
			Total:   3,
			Online:  2,
			Members: []queries.GroupPresenceMember{
				{ID: "u1", IsOnline: true},
				{ID: "u2", IsOnline: true},
				{ID: "u3", IsOnline: false},
			},
		}},
	)
	srv := httptest.NewServer(presenceRoutesMux(h))
	defer srv.Close()

	resp := doGet(t, srv.URL+"/api/groups/g1/presence")
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var body struct {
		Data struct {
			GroupID string `json:"groupId"`
			Total   int    `json:"total"`
			Online  int    `json:"online"`
			Members []struct {
				ID       string `json:"id"`
				IsOnline bool   `json:"isOnline"`
			} `json:"members"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body.Data.GroupID != "g1" || body.Data.Total != 3 || body.Data.Online != 2 {
		t.Errorf("presence = %+v, want g1/3/2", body.Data)
	}
	if len(body.Data.Members) != 3 || !body.Data.Members[0].IsOnline || body.Data.Members[2].IsOnline {
		t.Errorf("members = %+v, want u1/u2 online, u3 offline", body.Data.Members)
	}
}

func TestGetGroupPresence_RejectsNonMember(t *testing.T) {
	h := presenceTestHandler(
		nil,
		&mockGetGroupPresence{err: group.ErrNotMember},
	)
	srv := httptest.NewServer(presenceRoutesMux(h))
	defer srv.Close()

	resp := doGet(t, srv.URL+"/api/groups/g1/presence")
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", resp.StatusCode)
	}
}
