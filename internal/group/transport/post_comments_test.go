package transport

import (
	"bytes"
	"context"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"social-network/internal/group"
	"social-network/internal/group/commands"
	"social-network/internal/group/queries"
)

type mockCreateGroupPostComment struct {
	result   *group.PostComment
	err      error
	lastCmd  commands.CreateGroupPostCommentCommand
	executed bool
}

func (m *mockCreateGroupPostComment) Execute(_ context.Context, cmd commands.CreateGroupPostCommentCommand) (*group.PostComment, error) {
	m.executed = true
	m.lastCmd = cmd
	return m.result, m.err
}

func TestCreateGroupPostComment_WithImage(t *testing.T) {
	created := &group.PostComment{ID: "c1", PostID: "p1", AuthorID: "u1", Content: "pic", ImagePath: "/static/images/uploads/pic.png", CreatedAt: time.Now()}
	mock := &mockCreateGroupPostComment{result: created}
	h := newGroupTestHandler(
		func(_ *http.Request) (string, bool) { return "u1", true },
		&mockListGroups{result: &queries.ListGroupsResult{}},
		&mockGroupMembers{},
		nil,
		&mockGetGroupPostComments{result: &queries.GetGroupPostCommentsResult{}},
	)
	h.createGroupPostComment = mock
	srv := httptest.NewServer(groupRoutesMux(h))
	defer srv.Close()

	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	_ = mw.WriteField("content", "pic")
	pw, err := mw.CreateFormFile("image", "pic.png")
	if err != nil {
		t.Fatalf("create form file: %v", err)
	}
	if _, werr := pw.Write([]byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}); werr != nil {
		t.Fatalf("write image: %v", werr)
	}
	if cerr := mw.Close(); cerr != nil {
		t.Fatalf("close writer: %v", cerr)
	}

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, srv.URL+"/api/groups/posts/p1/comments", &buf)
	if err != nil {
		t.Fatalf("build request: %v", err)
	}
	req.Header.Set("Content-Type", mw.FormDataContentType())
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d", resp.StatusCode)
	}
	if !mock.executed {
		t.Fatal("expected command to be executed")
	}
	if mock.lastCmd.PostID != "p1" {
		t.Errorf("PostID = %q, want p1", mock.lastCmd.PostID)
	}
	if mock.lastCmd.ImageFileName != "pic.png" {
		t.Errorf("ImageFileName = %q, want pic.png", mock.lastCmd.ImageFileName)
	}
	if len(mock.lastCmd.ImageData) == 0 {
		t.Error("expected ImageData to be non-empty")
	}

	var body struct {
		Data map[string]any `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if got, ok := body.Data["imageUrl"].(string); !ok || got != "/static/images/uploads/pic.png" {
		t.Errorf("imageUrl = %#v, want /static/images/uploads/pic.png", body.Data["imageUrl"])
	}
}

func TestCreateGroupPostComment_NoImage(t *testing.T) {
	created := &group.PostComment{ID: "c1", PostID: "p1", AuthorID: "u1", Content: "text only", CreatedAt: time.Now()}
	mock := &mockCreateGroupPostComment{result: created}
	h := newGroupTestHandler(
		func(_ *http.Request) (string, bool) { return "u1", true },
		&mockListGroups{result: &queries.ListGroupsResult{}},
		&mockGroupMembers{},
		nil,
		&mockGetGroupPostComments{result: &queries.GetGroupPostCommentsResult{}},
	)
	h.createGroupPostComment = mock
	srv := httptest.NewServer(groupRoutesMux(h))
	defer srv.Close()

	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	_ = mw.WriteField("content", "text only")
	if err := mw.Close(); err != nil {
		t.Fatalf("close writer: %v", err)
	}

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, srv.URL+"/api/groups/posts/p1/comments", &buf)
	if err != nil {
		t.Fatalf("build request: %v", err)
	}
	req.Header.Set("Content-Type", mw.FormDataContentType())
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d", resp.StatusCode)
	}
	if len(mock.lastCmd.ImageData) != 0 || mock.lastCmd.ImageFileName != "" {
		t.Errorf("expected no image, got data=%d name=%q", len(mock.lastCmd.ImageData), mock.lastCmd.ImageFileName)
	}
}

func TestCreateGroupPostComment_Unauthorized(t *testing.T) {
	h := newGroupTestHandler(
		func(_ *http.Request) (string, bool) { return "", false },
		&mockListGroups{result: &queries.ListGroupsResult{}},
		&mockGroupMembers{},
		nil,
		&mockGetGroupPostComments{result: &queries.GetGroupPostCommentsResult{}},
	)
	srv := httptest.NewServer(groupRoutesMux(h))
	defer srv.Close()

	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	_ = mw.WriteField("content", "x")
	_ = mw.Close()

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, srv.URL+"/api/groups/posts/p1/comments", &buf)
	if err != nil {
		t.Fatalf("build request: %v", err)
	}
	req.Header.Set("Content-Type", mw.FormDataContentType())
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", resp.StatusCode)
	}
}
