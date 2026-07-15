package user

import (
	"testing"
	"time"
)

func TestUserFields(t *testing.T) {
	now := time.Date(2025, 1, 15, 10, 30, 0, 0, time.UTC)
	dob := time.Date(2000, 6, 15, 0, 0, 0, 0, time.UTC)
	avatar := "/avatars/test.jpg"

	u := User{
		ID:           "user-1",
		Email:        "alice@example.com",
		PasswordHash: "hashed_password_123",
		FirstName:    "Alice",
		LastName:     "Smith",
		DateOfBirth:  dob,
		Nickname:     "alice",
		AboutMe:      "Hello, I am Alice",
		AvatarPath:   avatar,
		IsPrivate:    true,
		CreatedAt:    now,
	}

	if u.ID != "user-1" {
		t.Errorf("ID = %q, want %q", u.ID, "user-1")
	}
	if u.Email != "alice@example.com" {
		t.Errorf("Email = %q, want %q", u.Email, "alice@example.com")
	}
	if u.PasswordHash != "hashed_password_123" {
		t.Errorf("PasswordHash = %q, want %q", u.PasswordHash, "hashed_password_123")
	}
	if u.FirstName != "Alice" {
		t.Errorf("FirstName = %q, want %q", u.FirstName, "Alice")
	}
	if u.LastName != "Smith" {
		t.Errorf("LastName = %q, want %q", u.LastName, "Smith")
	}
	if !u.DateOfBirth.Equal(dob) {
		t.Errorf("DateOfBirth = %v, want %v", u.DateOfBirth, dob)
	}
	if u.Nickname != "alice" {
		t.Errorf("Nickname = %q, want %q", u.Nickname, "alice")
	}
	if u.AboutMe != "Hello, I am Alice" {
		t.Errorf("AboutMe = %q, want %q", u.AboutMe, "Hello, I am Alice")
	}
	if u.AvatarPath != avatar {
		t.Errorf("AvatarPath = %q, want %q", u.AvatarPath, avatar)
	}
	if !u.IsPrivate {
		t.Errorf("IsPrivate = %v, want %v", u.IsPrivate, true)
	}
	if !u.CreatedAt.Equal(now) {
		t.Errorf("CreatedAt = %v, want %v", u.CreatedAt, now)
	}
}

func TestUserZeroValues(t *testing.T) {
	var u User

	if u.ID != "" {
		t.Errorf("zero ID = %q, want empty", u.ID)
	}
	if u.Email != "" {
		t.Errorf("zero Email = %q, want empty", u.Email)
	}
	if u.PasswordHash != "" {
		t.Errorf("zero PasswordHash = %q, want empty", u.PasswordHash)
	}
	if u.FirstName != "" {
		t.Errorf("zero FirstName = %q, want empty", u.FirstName)
	}
	if u.LastName != "" {
		t.Errorf("zero LastName = %q, want empty", u.LastName)
	}
	if !u.DateOfBirth.IsZero() {
		t.Errorf("zero DateOfBirth = %v, want zero", u.DateOfBirth)
	}
	if u.Nickname != "" {
		t.Errorf("zero Nickname = %q, want empty", u.Nickname)
	}
	if u.AboutMe != "" {
		t.Errorf("zero AboutMe = %q, want empty", u.AboutMe)
	}
	if u.AvatarPath != "" {
		t.Errorf("zero AvatarPath = %q, want empty", u.AvatarPath)
	}
	if u.IsPrivate {
		t.Errorf("zero IsPrivate = %v, want %v", u.IsPrivate, false)
	}
	if !u.CreatedAt.IsZero() {
		t.Errorf("zero CreatedAt = %v, want zero", u.CreatedAt)
	}
}

var _ Repository = Repository(nil)
