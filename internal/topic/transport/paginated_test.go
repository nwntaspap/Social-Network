package transport

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"social-network/internal/topic"
	"social-network/internal/topic/queries"
)

type paginatedEnvelope struct {
	Data struct {
		Data       []map[string]any `json:"data"`
		Page       int              `json:"page"`
		PageSize   int              `json:"pageSize"`
		TotalCount int              `json:"totalCount"`
		TotalPages int              `json:"totalPages"`
	} `json:"data"`
}

func decodePaginated(t *testing.T, resp *http.Response) paginatedEnvelope {
	t.Helper()
	var body paginatedEnvelope
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	return body
}

func assertFlatPaginated(t *testing.T, body paginatedEnvelope, wantItems, wantPage, wantPageSize, wantTotal, wantTotalPages int) {
	t.Helper()
	if len(body.Data.Data) != wantItems {
		t.Errorf("len(data.data) = %d, want %d", len(body.Data.Data), wantItems)
	}
	if body.Data.Page != wantPage {
		t.Errorf("page = %d, want %d", body.Data.Page, wantPage)
	}
	if body.Data.PageSize != wantPageSize {
		t.Errorf("pageSize = %d, want %d", body.Data.PageSize, wantPageSize)
	}
	if body.Data.TotalCount != wantTotal {
		t.Errorf("totalCount = %d, want %d", body.Data.TotalCount, wantTotal)
	}
	if body.Data.TotalPages != wantTotalPages {
		t.Errorf("totalPages = %d, want %d", body.Data.TotalPages, wantTotalPages)
	}
}

func TestGetFeed_MatchesFrontendPaginatedResponse(t *testing.T) {
	h := newTestHandler(extractUserOK, &mockGetFeed{
		result: &queries.GetFeedResult{Topics: []topic.Topic{*testTopic()}, Total: 1},
	})
	srv := httptest.NewServer(mux(h))
	defer srv.Close()

	resp := doReq(t, srv, http.MethodGet, "/api/posts/feed?page=2&limit=10")
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	assertFlatPaginated(t, decodePaginated(t, resp), 1, 2, 10, 1, 1)
}

func TestGetUserTopics_MatchesFrontendPaginatedResponse(t *testing.T) {
	h := newTestHandler(extractUserOK, &mockGetByUser{
		result: &queries.GetTopicsByUserResult{Topics: []topic.Topic{*testTopic()}, Total: 1},
	})
	srv := httptest.NewServer(mux(h))
	defer srv.Close()

	resp := doReq(t, srv, http.MethodGet, "/api/posts/user?userId=user-1&page=2&limit=10")
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	assertFlatPaginated(t, decodePaginated(t, resp), 1, 2, 10, 1, 1)
}

func TestGetGroupTopics_MatchesFrontendPaginatedResponse(t *testing.T) {
	h := newTestHandler(extractUserOK, &mockGetByGroup{
		result: &queries.GetTopicsByGroupResult{Topics: []topic.Topic{*testTopic()}, Total: 1},
	})
	srv := httptest.NewServer(mux(h))
	defer srv.Close()

	resp := doReq(t, srv, http.MethodGet, "/api/posts/group?groupId=g1&page=2&limit=10")
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	assertFlatPaginated(t, decodePaginated(t, resp), 1, 2, 10, 1, 1)
}
