package transport

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"social-network/internal/group"
	"social-network/internal/group/queries"
)

type mockListGroups struct {
	result *queries.ListGroupsResult
	err    error
}

func (m *mockListGroups) Resolve(_ context.Context, _ queries.ListGroupsQuery) (*queries.ListGroupsResult, error) {
	if m.result != nil {
		return m.result, nil
	}
	return &queries.ListGroupsResult{}, m.err
}

type mockGroupMembers struct {
	total int
	err   error
}

func (m *mockGroupMembers) Resolve(_ context.Context, _ queries.GetGroupMembersQuery) (*queries.GetGroupMembersResult, error) {
	if m.err != nil {
		return nil, m.err
	}
	return &queries.GetGroupMembersResult{Total: m.total}, nil
}

type mockGroupUserLookup struct {
	result *UserResult
}

func (m *mockGroupUserLookup) GetUserByID(_ context.Context, _ string) (*UserResult, error) {
	if m.result != nil {
		return m.result, nil
	}
	return &UserResult{ID: "u2", Username: "creator"}, nil
}

func newGroupTestHandler(extractUser UserExtractor, list ListGroupsResolver, members GetGroupMembersResolver) *Handler {
	return NewHandler(
		extractUser,
		&mockGroupUserLookup{},
		nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
		list,
		nil, nil, nil, nil,
		members,
	)
}

func TestListGroups_IncludesMembershipStatus(t *testing.T) {
	g := group.Group{ID: "g1", Title: "Go", CreatorID: "u2", MembershipStatus: "member", CreatedAt: time.Now()}
	h := newGroupTestHandler(
		func(_ *http.Request) (string, bool) { return "u1", true },
		&mockListGroups{result: &queries.ListGroupsResult{Groups: []group.Group{g}, Total: 1}},
		&mockGroupMembers{total: 3},
	)
	srv := httptest.NewServer(http.HandlerFunc(h.ListGroups))
	defer srv.Close()

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, srv.URL+"/api/groups", nil)
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

	if got, ok := d["membershipStatus"].(string); !ok || got != "member" {
		t.Errorf("membershipStatus = %#v, want %q", d["membershipStatus"], "member")
	}
	if got, ok := d["membersCount"].(float64); !ok || got != 3 {
		t.Errorf("membersCount = %#v, want 3", d["membersCount"])
	}
}
