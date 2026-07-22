package google

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestProvider_Name(t *testing.T) {
	p := NewProvider("id", "secret", "http://callback", "https://oauth2.googleapis.com/token", []string{"email"})
	if got := p.Name(); got != "google" {
		t.Errorf("Name() = %q, want %q", got, "google")
	}
}

func TestProvider_GetAuthURL(t *testing.T) {
	p := NewProvider("myid", "mysecret", "http://localhost:8080/callback", "https://oauth2.googleapis.com/token", []string{"email", "profile"})
	url := p.GetAuthURL("test-state")

	if url == "" {
		t.Fatal("GetAuthURL() returned empty string")
	}
	if !containsStr(url, "client_id=myid") {
		t.Errorf("URL missing client_id, got: %s", url)
	}
	if !containsStr(url, "state=test-state") {
		t.Errorf("URL missing state, got: %s", url)
	}
	if !containsStr(url, "response_type=code") {
		t.Errorf("URL missing response_type, got: %s", url)
	}
}

func TestProvider_ExchangeCode_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(tokenResponse{
			AccessToken: "ya29.test_token_123",
			TokenType:   "Bearer",
			ExpiresIn:   3600,
		})
	}))
	defer server.Close()

	p := NewProvider("id", "secret", "http://localhost:8080/callback", server.URL, []string{"email"})

	token, err := p.ExchangeCode(context.Background(), "test-code")
	if err != nil {
		t.Fatalf("ExchangeCode() error = %v", err)
	}
	if token != "ya29.test_token_123" { //nolint:gosec // Test assertion, not a credential
		t.Errorf("token = %q, want %q", token, "ya29.test_token_123")
	}
}

func TestProvider_ExchangeCode_EmptyToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(tokenResponse{
			AccessToken: "",
			TokenType:   "Bearer",
		})
	}))
	defer server.Close()

	p := NewProvider("id", "secret", "http://localhost:8080/callback", server.URL, []string{"email"})

	_, err := p.ExchangeCode(context.Background(), "test-code")
	if err == nil {
		t.Fatal("expected error for empty token")
	}
}

func TestProvider_GetUserInfo_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test_token" {
			t.Errorf("expected Bearer test_token, got %s", r.Header.Get("Authorization"))
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(googleUser{
			ID:            "google-123",
			Email:         "test@example.com",
			Name:          "Test User",
			Picture:       "https://avatar.example.com/test",
			VerifiedEmail: true,
		})
	}))
	defer server.Close()

	p := NewProvider("id", "secret", "http://localhost:8080/callback", "https://oauth2.googleapis.com/token", []string{"email"})
	p.SetBaseURL(server.URL)

	user, err := p.fetchUser(context.Background(), "test_token")
	if err != nil {
		t.Fatalf("fetchUser() error = %v", err)
	}
	if user.ID != "google-123" {
		t.Errorf("ID = %q, want %q", user.ID, "google-123")
	}
	if user.Email != "test@example.com" {
		t.Errorf("Email = %q, want %q", user.Email, "test@example.com")
	}
	if !user.VerifiedEmail {
		t.Error("expected VerifiedEmail to be true")
	}
}

func TestProvider_GetUserInfo_EmailNotVerified(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(googleUser{
			ID:            "google-456",
			Email:         "test@example.com",
			Name:          "Test User",
			VerifiedEmail: false,
		})
	}))
	defer server.Close()

	p := NewProvider("id", "secret", "http://localhost:8080/callback", "https://oauth2.googleapis.com/token", []string{"email"})
	p.SetBaseURL(server.URL)

	_, err := p.fetchUser(context.Background(), "test_token")
	if err == nil {
		t.Fatal("expected error for unverified email")
	}
}

func containsStr(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstr(s, substr))
}

func containsSubstr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
