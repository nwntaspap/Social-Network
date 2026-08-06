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
	result    *queries.ListGroupsResult
	err       error
	lastQuery queries.ListGroupsQuery
}

func (m *mockListGroups) Resolve(_ context.Context, q queries.ListGroupsQuery) (*queries.ListGroupsResult, error) {
	m.lastQuery = q
	if m.result != nil {
		return m.result, nil
	}
	return &queries.ListGroupsResult{}, m.err
}

type mockGroupMembers struct {
	members []group.Member
	total   int
	err     error
}

func (m *mockGroupMembers) Resolve(_ context.Context, _ queries.GetGroupMembersQuery) (*queries.GetGroupMembersResult, error) {
	if m.err != nil {
		return nil, m.err
	}
	return &queries.GetGroupMembersResult{Members: m.members, Total: m.total}, nil
}

type mockGetGroupFeed struct {
	result *queries.GetGroupFeedResult
	err    error
}

func (m *mockGetGroupFeed) Resolve(_ context.Context, _ queries.GetGroupFeedQuery) (*queries.GetGroupFeedResult, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.result, nil
}

type mockGetGroupPostComments struct {
	result *queries.GetGroupPostCommentsResult
	err    error
}

func (m *mockGetGroupPostComments) Resolve(_ context.Context, _ queries.GetGroupPostCommentsQuery) (*queries.GetGroupPostCommentsResult, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.result, nil
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

func newGroupTestHandler(extractUser UserExtractor, list ListGroupsResolver, members GetGroupMembersResolver, feed GetGroupFeedResolver, comments GetGroupPostCommentsResolver) *Handler {
	return NewHandler(
		extractUser,
		&mockGroupUserLookup{},
		nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
		list,
		nil,
		feed,
		nil,
		comments,
		members,
	)
}

func TestListGroups_IncludesMembershipStatus(t *testing.T) {
	g := group.Group{ID: "g1", Title: "Go", CreatorID: "u2", MembershipStatus: "member", CreatedAt: time.Now()}
	h := newGroupTestHandler(
		func(_ *http.Request) (string, bool) { return "u1", true },
		&mockListGroups{result: &queries.ListGroupsResult{Groups: []group.Group{g}, Total: 1}},
		&mockGroupMembers{total: 3},
		nil, nil,
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

	if got, ok := d["membershipStatus"].(string); !ok || got != "member" {
		t.Errorf("membershipStatus = %#v, want %q", d["membershipStatus"], "member")
	}
	if got, ok := d["membersCount"].(float64); !ok || got != 3 {
		t.Errorf("membersCount = %#v, want 3", d["membersCount"])
	}
}

func TestListGroups_ForwardsQueryParam(t *testing.T) {
	mock := &mockListGroups{result: &queries.ListGroupsResult{}}
	h := newGroupTestHandler(
		func(_ *http.Request) (string, bool) { return "u1", true },
		mock,
		&mockGroupMembers{},
		nil, nil,
	)
	srv := httptest.NewServer(http.HandlerFunc(h.ListGroups))
	defer srv.Close()

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, srv.URL+"/api/groups?query=go&page=2", nil)
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

	if mock.lastQuery.Query != "go" {
		t.Errorf("query forwarded = %q, want %q", mock.lastQuery.Query, "go")
	}
	if mock.lastQuery.Page != 2 {
		t.Errorf("page forwarded = %d, want 2", mock.lastQuery.Page)
	}
	if mock.lastQuery.UserID != "u1" {
		t.Errorf("userID forwarded = %q, want %q", mock.lastQuery.UserID, "u1")
	}
}

func TestListGroups_MatchesFrontendPaginatedResponse(t *testing.T) {
	g := group.Group{ID: "g1", Title: "Go", CreatorID: "u2", MembershipStatus: "member", CreatedAt: time.Now()}
	h := newGroupTestHandler(
		func(_ *http.Request) (string, bool) { return "u1", true },
		&mockListGroups{result: &queries.ListGroupsResult{Groups: []group.Group{g}, Total: 1}},
		&mockGroupMembers{total: 3},
		nil, nil,
	)
	srv := httptest.NewServer(http.HandlerFunc(h.ListGroups))
	defer srv.Close()

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, srv.URL+"/api/groups?query=go", nil)
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
		Data struct {
			Data       []map[string]any `json:"data"`
			Page       int              `json:"page"`
			PageSize   int              `json:"pageSize"`
			TotalCount int              `json:"totalCount"`
			TotalPages int              `json:"totalPages"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}

	if len(body.Data.Data) != 1 {
		t.Errorf("len(data.data) = %d, want 1", len(body.Data.Data))
	}
	if body.Data.Page != 1 {
		t.Errorf("page = %d, want 1", body.Data.Page)
	}
	if body.Data.PageSize != 20 {
		t.Errorf("pageSize = %d, want 20 (default limit)", body.Data.PageSize)
	}
	if body.Data.TotalCount != 1 {
		t.Errorf("totalCount = %d, want 1", body.Data.TotalCount)
	}
	if body.Data.TotalPages != 1 {
		t.Errorf("totalPages = %d, want 1", body.Data.TotalPages)
	}
}

type flatPaginatedBody struct {
	Data struct {
		Data       []map[string]any `json:"data"`
		Page       int              `json:"page"`
		PageSize   int              `json:"pageSize"`
		TotalCount int              `json:"totalCount"`
		TotalPages int              `json:"totalPages"`
	} `json:"data"`
}

func decodeFlatPaginated(t *testing.T, resp *http.Response) flatPaginatedBody {
	t.Helper()
	var body flatPaginatedBody
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	return body
}

func groupRoutesMux(h *Handler) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/groups/{groupId}/members", h.GetGroupMembers)
	mux.HandleFunc("GET /api/groups/{groupId}/posts", h.GetGroupFeed)
	mux.HandleFunc("POST /api/groups/{groupId}/posts", h.CreateGroupPost)
	mux.HandleFunc("POST /api/groups/posts/{postId}/vote", h.VoteGroupPost)
	mux.HandleFunc("GET /api/groups/{groupId}/posts/{postId}/comments", h.GetGroupPostComments)
	mux.HandleFunc("POST /api/groups/posts/{postId}/comments", h.CreateGroupPostComment)
	return mux
}

func TestGetGroupMembers_MatchesFrontendPaginatedResponse(t *testing.T) {
	member := group.Member{GroupID: "g1", UserID: "u2", Role: group.RoleMember, JoinedAt: time.Now()}
	h := newGroupTestHandler(
		func(_ *http.Request) (string, bool) { return "u1", true },
		&mockListGroups{result: &queries.ListGroupsResult{}},
		&mockGroupMembers{members: []group.Member{member}, total: 1},
		nil, nil,
	)
	srv := httptest.NewServer(groupRoutesMux(h))
	defer srv.Close()

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, srv.URL+"/api/groups/g1/members?page=2", nil)
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

	body := decodeFlatPaginated(t, resp)
	if len(body.Data.Data) != 1 {
		t.Errorf("len(data.data) = %d, want 1", len(body.Data.Data))
	}
	if body.Data.Page != 2 {
		t.Errorf("page = %d, want 2", body.Data.Page)
	}
	if body.Data.PageSize != 20 {
		t.Errorf("pageSize = %d, want 20 (default limit)", body.Data.PageSize)
	}
	if body.Data.TotalCount != 1 {
		t.Errorf("totalCount = %d, want 1", body.Data.TotalCount)
	}
	if body.Data.TotalPages != 1 {
		t.Errorf("totalPages = %d, want 1", body.Data.TotalPages)
	}
}

func TestGetGroupFeed_MatchesFrontendPaginatedResponse(t *testing.T) {
	post := group.Post{ID: "p1", GroupID: "g1", AuthorID: "u2", Title: "Hello", Content: "World", CreatedAt: time.Now()}
	h := newGroupTestHandler(
		func(_ *http.Request) (string, bool) { return "u1", true },
		&mockListGroups{result: &queries.ListGroupsResult{}},
		&mockGroupMembers{},
		&mockGetGroupFeed{result: &queries.GetGroupFeedResult{Posts: []group.Post{post}, Total: 1}},
		&mockGetGroupPostComments{result: &queries.GetGroupPostCommentsResult{}},
	)
	srv := httptest.NewServer(groupRoutesMux(h))
	defer srv.Close()

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, srv.URL+"/api/groups/g1/posts?page=2", nil)
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

	body := decodeFlatPaginated(t, resp)
	if len(body.Data.Data) != 1 {
		t.Errorf("len(data.data) = %d, want 1", len(body.Data.Data))
	}
	if body.Data.Page != 2 {
		t.Errorf("page = %d, want 2", body.Data.Page)
	}
	if body.Data.PageSize != 20 {
		t.Errorf("pageSize = %d, want 20 (default limit)", body.Data.PageSize)
	}
	if body.Data.TotalCount != 1 {
		t.Errorf("totalCount = %d, want 1", body.Data.TotalCount)
	}
	if body.Data.TotalPages != 1 {
		t.Errorf("totalPages = %d, want 1", body.Data.TotalPages)
	}
}

func TestGetGroupPostComments_MatchesFrontendPaginatedResponse(t *testing.T) {
	comment := group.PostComment{ID: "c1", PostID: "p1", AuthorID: "u2", Content: "Nice", CreatedAt: time.Now()}
	h := newGroupTestHandler(
		func(_ *http.Request) (string, bool) { return "u1", true },
		&mockListGroups{result: &queries.ListGroupsResult{}},
		&mockGroupMembers{},
		nil,
		&mockGetGroupPostComments{result: &queries.GetGroupPostCommentsResult{Comments: []group.PostComment{comment}, Total: 1}},
	)
	srv := httptest.NewServer(groupRoutesMux(h))
	defer srv.Close()

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, srv.URL+"/api/groups/g1/posts/p1/comments?page=2", nil)
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

	body := decodeFlatPaginated(t, resp)
	if len(body.Data.Data) != 1 {
		t.Errorf("len(data.data) = %d, want 1", len(body.Data.Data))
	}
	if body.Data.Page != 2 {
		t.Errorf("page = %d, want 2", body.Data.Page)
	}
	if body.Data.PageSize != 20 {
		t.Errorf("pageSize = %d, want 20 (default limit)", body.Data.PageSize)
	}
	if body.Data.TotalCount != 1 {
		t.Errorf("totalCount = %d, want 1", body.Data.TotalCount)
	}
	if body.Data.TotalPages != 1 {
		t.Errorf("totalPages = %d, want 1", body.Data.TotalPages)
	}
}
