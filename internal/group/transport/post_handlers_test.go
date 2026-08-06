package transport

import (
	"bytes"
	"context"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"social-network/internal/group"
	"social-network/internal/group/commands"
)

type mockCreateGroupPost struct {
	post    *group.Post
	err     error
	lastCmd commands.CreateGroupPostCommand
}

func (m *mockCreateGroupPost) Execute(_ context.Context, cmd commands.CreateGroupPostCommand) (*group.Post, error) {
	m.lastCmd = cmd
	if m.err != nil {
		return nil, m.err
	}
	return m.post, nil
}

type mockCastGroupPostVote struct {
	err     error
	lastCmd commands.CastGroupPostVoteCommand
}

func (m *mockCastGroupPostVote) Execute(_ context.Context, cmd commands.CastGroupPostVoteCommand) error {
	m.lastCmd = cmd
	return m.err
}

func TestVoteGroupPost_CapturesReaction(t *testing.T) {
	tests := []struct {
		name         string
		reaction     string
		wantReaction int
	}{
		{"like", `{"reactionType":1}`, 1},
		{"dislike", `{"reactionType":-1}`, -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockCastGroupPostVote{}
			h := NewHandler(
				func(_ *http.Request) (string, bool) { return "u1", true },
				&mockGroupUserLookup{},
				nil, nil, nil, nil, nil, nil, nil,
				mock,
				nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
			)
			srv := httptest.NewServer(groupRoutesMux(h))
			defer srv.Close()

			req, err := http.NewRequestWithContext(context.Background(), http.MethodPost,
				srv.URL+"/api/groups/posts/p1/vote", bytes.NewBufferString(tt.reaction))
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
			if mock.lastCmd.UserID != "u1" || mock.lastCmd.PostID != "p1" || mock.lastCmd.ReactionType != tt.wantReaction {
				t.Errorf("command = %+v, want u1/p1/%d", mock.lastCmd, tt.wantReaction)
			}
		})
	}
}

func TestVoteGroupPost_UnauthorizedWithoutUser(t *testing.T) {
	h := NewHandler(
		func(_ *http.Request) (string, bool) { return "", false },
		&mockGroupUserLookup{},
		nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
	)
	srv := httptest.NewServer(groupRoutesMux(h))
	defer srv.Close()

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost,
		srv.URL+"/api/groups/posts/p1/vote", bytes.NewBufferString(`{"reactionType":1}`))
	if err != nil {
		t.Fatalf("build request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", resp.StatusCode)
	}
}

func TestCreateGroupPost_CapturesCommandFromMultipart(t *testing.T) {
	mock := &mockCreateGroupPost{
		post: &group.Post{ID: "p1", GroupID: "g1", AuthorID: "u1", Title: "Hello", Content: "World"},
	}

	h := NewHandler(
		func(_ *http.Request) (string, bool) { return "u1", true },
		&mockGroupUserLookup{},
		nil, nil, nil, nil, nil,
		mock,
		nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
	)
	srv := httptest.NewServer(groupRoutesMux(h))
	defer srv.Close()

	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	_ = mw.WriteField("title", "Hello")
	_ = mw.WriteField("content", "World")
	img, err := mw.CreateFormFile("image", "pic.png")
	if err != nil {
		t.Fatalf("create form file: %v", err)
	}
	_, _ = img.Write([]byte("fake-png-bytes"))
	if err = mw.Close(); err != nil {
		t.Fatalf("close writer: %v", err)
	}

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, srv.URL+"/api/groups/g1/posts", &buf)
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

	if mock.lastCmd.GroupID != "g1" || mock.lastCmd.AuthorID != "u1" ||
		mock.lastCmd.Title != "Hello" || mock.lastCmd.Content != "World" {
		t.Errorf("command = %+v, want groupId=g1 author=u1 title=Hello content=World", mock.lastCmd)
	}
	if mock.lastCmd.ImageFileName != "pic.png" {
		t.Errorf("ImageFileName = %q, want pic.png", mock.lastCmd.ImageFileName)
	}
	if len(mock.lastCmd.ImageData) == 0 || !bytes.Equal(mock.lastCmd.ImageData, []byte("fake-png-bytes")) {
		t.Errorf("ImageData = %q, want fake-png-bytes", mock.lastCmd.ImageData)
	}
}

func TestCreateGroupPost_UnauthorizedWithoutUser(t *testing.T) {
	h := NewHandler(
		func(_ *http.Request) (string, bool) { return "", false },
		&mockGroupUserLookup{},
		nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
	)
	srv := httptest.NewServer(groupRoutesMux(h))
	defer srv.Close()

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, srv.URL+"/api/groups/g1/posts", bytes.NewBufferString(""))
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
