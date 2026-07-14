package queries

import (
	"context"
	"errors"
	"testing"
)

func TestAreConnectedResolver_Connected(t *testing.T) {
	repo := &mockRepo{areConnectedResult: true}
	r := NewAreConnectedResolver(repo)

	result, err := r.Resolve(context.Background(), AreConnectedQuery{UserID: "user-1", TargetID: "user-2"})
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	if !result {
		t.Errorf("result = false, want true")
	}
}

func TestAreConnectedResolver_NotConnected(t *testing.T) {
	repo := &mockRepo{areConnectedResult: false}
	r := NewAreConnectedResolver(repo)

	result, err := r.Resolve(context.Background(), AreConnectedQuery{UserID: "user-1", TargetID: "user-3"})
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	if result {
		t.Errorf("result = true, want false")
	}
}

func TestAreConnectedResolver_Error(t *testing.T) {
	repo := &mockRepo{areConnectedErr: errors.New("db down")}
	r := NewAreConnectedResolver(repo)

	_, err := r.Resolve(context.Background(), AreConnectedQuery{UserID: "user-1", TargetID: "user-2"})
	if err == nil {
		t.Fatal("Resolve() expected error, got nil")
	}
}
