package localstorage

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
)

func TestUpload_AcceptsBareFilename(t *testing.T) {
	s := NewLocalStorage()
	err := s.Upload(context.Background(), []byte("data"), "pic.png")
	if err != nil {
		t.Fatalf("Upload(bare filename) error = %v, want nil", err)
	}
	_ = s.Delete(context.Background(), "pic.png")
}

func TestUpload_RejectsAbsolutePath(t *testing.T) {
	s := NewLocalStorage()
	err := s.Upload(context.Background(), []byte("data"), filepath.Join("/static/images/uploads", "pic.png"))
	if !errors.Is(err, ErrInvalidPath) {
		t.Fatalf("Upload(absolute path) error = %v, want ErrInvalidPath", err)
	}
}
