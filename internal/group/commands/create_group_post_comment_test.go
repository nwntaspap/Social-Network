package commands

import (
	"context"
	"errors"
	"testing"

	"social-network/internal/group"
)

var (
	validGroupPNG  = []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, 0x00, 0x00}
	badGroupHeader = []byte{0x00, 0x01, 0x02, 0x03}
)

type groupCommentStorage struct {
	uploaded bool
	gotPath  string
}

func (m *groupCommentStorage) Upload(_ context.Context, _ []byte, path string) error {
	m.uploaded = true
	m.gotPath = path
	return nil
}

type createGroupPostCommentStub struct {
	group.Repository

	post      *group.Post
	postErr   error
	isMember  bool
	memberErr error
	createErr error
}

func (s *createGroupPostCommentStub) GetPostByID(_ context.Context, _ string) (*group.Post, error) {
	return s.post, s.postErr
}

func (s *createGroupPostCommentStub) IsMember(_ context.Context, _, _ string) (bool, error) {
	return s.isMember, s.memberErr
}

func (s *createGroupPostCommentStub) CreatePostComment(_ context.Context, _ *group.PostComment) error {
	return s.createErr
}

func memberStub() *createGroupPostCommentStub {
	return &createGroupPostCommentStub{
		post:     &group.Post{ID: "p1", GroupID: "g1"},
		isMember: true,
	}
}

func TestCreateGroupPostComment_WithValidImage(t *testing.T) {
	store := &groupCommentStorage{}
	h := NewCreateGroupPostCommentHandler(memberStub(), store)

	c, err := h.Execute(context.Background(), CreateGroupPostCommentCommand{
		PostID:        "p1",
		AuthorID:      "u1",
		Content:       "pic",
		ImageData:     validGroupPNG,
		ImageFileName: "pic.png",
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if c == nil {
		t.Fatal("expected comment, got nil")
	}
	if !store.uploaded {
		t.Error("expected image to be uploaded")
	}
	if c.ImagePath == "" {
		t.Error("expected ImagePath to be set")
	}
	if store.gotPath != "pic.png" {
		t.Errorf("upload path = %q, want bare filename pic.png", store.gotPath)
	}
}

func TestCreateGroupPostComment_InvalidImage(t *testing.T) {
	h := NewCreateGroupPostCommentHandler(memberStub(), &groupCommentStorage{})

	_, err := h.Execute(context.Background(), CreateGroupPostCommentCommand{
		PostID:        "p1",
		AuthorID:      "u1",
		Content:       "bad pic",
		ImageData:     badGroupHeader,
		ImageFileName: "bad.png",
	})
	if err == nil {
		t.Fatal("expected error for bad image, got nil")
	}
}

func TestCreateGroupPostComment_NoImage(t *testing.T) {
	store := &groupCommentStorage{}
	h := NewCreateGroupPostCommentHandler(memberStub(), store)

	c, err := h.Execute(context.Background(), CreateGroupPostCommentCommand{
		PostID:   "p1",
		AuthorID: "u1",
		Content:  "text only",
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if store.uploaded {
		t.Error("expected no upload for text-only comment")
	}
	if c.ImagePath != "" {
		t.Errorf("ImagePath = %q, want empty", c.ImagePath)
	}
}

func TestCreateGroupPostComment_Validation(t *testing.T) {
	h := NewCreateGroupPostCommentHandler(memberStub(), &groupCommentStorage{})

	if _, err := h.Execute(context.Background(), CreateGroupPostCommentCommand{AuthorID: "u1", Content: "x"}); err == nil {
		t.Error("expected error for missing post id")
	}
	if _, err := h.Execute(context.Background(), CreateGroupPostCommentCommand{PostID: "p1", Content: "x"}); err == nil {
		t.Error("expected error for missing author id")
	}
	if _, err := h.Execute(context.Background(), CreateGroupPostCommentCommand{PostID: "p1", AuthorID: "u1"}); err == nil {
		t.Error("expected error for missing content")
	}
}

func TestCreateGroupPostComment_PostNotFound(t *testing.T) {
	h := NewCreateGroupPostCommentHandler(
		&createGroupPostCommentStub{postErr: group.ErrPostNotFound}, &groupCommentStorage{},
	)

	_, err := h.Execute(context.Background(), CreateGroupPostCommentCommand{PostID: "missing", AuthorID: "u1", Content: "x"})
	if !errors.Is(err, group.ErrPostNotFound) {
		t.Errorf("Execute() error = %v, want %v", err, group.ErrPostNotFound)
	}
}

func TestCreateGroupPostComment_NonMemberDenied(t *testing.T) {
	h := NewCreateGroupPostCommentHandler(
		&createGroupPostCommentStub{
			post:     &group.Post{ID: "p1", GroupID: "g1"},
			isMember: false,
		}, &groupCommentStorage{},
	)

	_, err := h.Execute(context.Background(), CreateGroupPostCommentCommand{PostID: "p1", AuthorID: "outsider", Content: "x"})
	if !errors.Is(err, group.ErrNotMember) {
		t.Errorf("Execute() error = %v, want %v", err, group.ErrNotMember)
	}
}

func TestCreateGroupPostComment_RepoError(t *testing.T) {
	h := NewCreateGroupPostCommentHandler(&createGroupPostCommentStub{
		post:      &group.Post{ID: "p1", GroupID: "g1"},
		isMember:  true,
		createErr: errors.New("db fail"),
	}, &groupCommentStorage{})

	_, err := h.Execute(context.Background(), CreateGroupPostCommentCommand{
		PostID:   "p1",
		AuthorID: "u1",
		Content:  "x",
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
