package commands

import (
	"context"
	"testing"

	"social-network/internal/group"
)

type groupPostImageRepoStub struct {
	group.Repository
}

func (s *groupPostImageRepoStub) IsMember(_ context.Context, _, _ string) (bool, error) {
	return true, nil
}

func (s *groupPostImageRepoStub) CreatePost(_ context.Context, _ *group.Post) error { return nil }

type groupPostImageStorageStub struct{}

func (*groupPostImageStorageStub) Upload(_ context.Context, _ []byte, _ string) error {
	return nil
}

func TestCreateGroupPost_RejectsInvalidImageHeader(t *testing.T) {
	h := NewCreateGroupPostHandler(&groupPostImageRepoStub{}, &groupPostImageStorageStub{})

	_, err := h.Execute(context.Background(), CreateGroupPostCommand{
		GroupID:       "g1",
		AuthorID:      "u1",
		Title:         "T",
		Content:       "C",
		ImageData:     []byte{0x00, 0x01, 0x02, 0x03},
		ImageFileName: "evil.png",
	})
	if err == nil {
		t.Fatal("expected error for invalid image header")
	}
}
