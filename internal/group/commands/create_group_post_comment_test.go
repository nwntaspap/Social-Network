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

	createErr error
}

func (s *createGroupPostCommentStub) CreatePostComment(_ context.Context, _ *group.PostComment) error {
	return s.createErr
}

func TestCreateGroupPostComment_WithValidImage(t *testing.T) {
	store := &groupCommentStorage{}
	h := NewCreateGroupPostCommentHandler(&createGroupPostCommentStub{}, store)

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
	h := NewCreateGroupPostCommentHandler(&createGroupPostCommentStub{}, &groupCommentStorage{})

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
	h := NewCreateGroupPostCommentHandler(&createGroupPostCommentStub{}, store)

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
	h := NewCreateGroupPostCommentHandler(&createGroupPostCommentStub{}, &groupCommentStorage{})

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

func TestCreateGroupPostComment_RepoError(t *testing.T) {
	h := NewCreateGroupPostCommentHandler(&createGroupPostCommentStub{createErr: errors.New("db fail")}, &groupCommentStorage{})

	_, err := h.Execute(context.Background(), CreateGroupPostCommentCommand{
		PostID:   "p1",
		AuthorID: "u1",
		Content:  "x",
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
