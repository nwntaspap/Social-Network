package github

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestProvider_Name(t *testing.T) {
	p := NewProvider("id", "secret", "http://callback", []string{"user:email"})
	if got := p.Name(); got != "github" {
		t.Errorf("Name() = %q, want %q", got, "github")
	}
}

func TestProvider_GetAuthURL(t *testing.T) {
	p := NewProvider("myid", "mysecret", "http://localhost:8080/callback", []string{"user:email"})
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
	// Scope is URL-encoded as user%3Aemail
	if !containsStr(url, "scope=user") {
		t.Errorf("URL missing scope, got: %s", url)
	}
}

func TestProvider_ExchangeCode_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}

		var body map[string]string
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode body: %v", err)
		}
		if body["code"] != "test-code" {
			t.Errorf("expected code=test-code, got %s", body["code"])
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(tokenResponse{
			AccessToken: "gho_test_token_123",
			TokenType:   "bearer",
			Scope:       "user:email",
		})
	}))
	defer server.Close()

	p := NewProvider("id", "secret", "http://localhost:8080/callback", []string{"user:email"})
	p.SetBaseURL(server.URL)

	token, err := p.ExchangeCode(context.Background(), "test-code")
	if err != nil {
		t.Fatalf("ExchangeCode() error = %v", err)
	}
	if token != "gho_test_token_123" { //nolint:gosec // Test assertion, not a credential
		t.Errorf("token = %q, want %q", token, "gho_test_token_123")
	}
}

func TestProvider_ExchangeCode_EmptyToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(tokenResponse{
			AccessToken: "",
			TokenType:   "bearer",
		})
	}))
	defer server.Close()

	p := NewProvider("id", "secret", "http://localhost:8080/callback", []string{"user:email"})
	p.SetBaseURL(server.URL)

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
		_ = json.NewEncoder(w).Encode(gitHubUser{
			Login:     "testuser",
			Email:     "test@example.com",
			Name:      "Test User",
			AvatarURL: "https://avatar.example.com/test",
			ID:        12345,
		})
	}))
	defer server.Close()

	p := NewProvider("id", "secret", "http://localhost:8080/callback", []string{"user:email"})
	p.SetBaseURL(server.URL)

	user, err := p.fetchUser(context.Background(), "test_token")
	if err != nil {
		t.Fatalf("fetchUser() error = %v", err)
	}
	if user.Login != "testuser" {
		t.Errorf("Login = %q, want %q", user.Login, "testuser")
	}
	if user.Email != "test@example.com" {
		t.Errorf("Email = %q, want %q", user.Email, "test@example.com")
	}
	if user.ID != 12345 {
		t.Errorf("ID = %d, want %d", user.ID, 12345)
	}
}

func TestProvider_GetUserInfo_EmailFromSecondaryEndpoint(t *testing.T) {
	// Server that returns empty email on /user, then returns email on /user/emails
	mux := http.NewServeMux()
	callCount := 0

	mux.HandleFunc("/user", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(gitHubUser{
			Login:     "testuser",
			Email:     "", // empty email triggers secondary fetch
			Name:      "Test User",
			AvatarURL: "https://avatar.example.com/test",
			ID:        12345,
		})
	})

	mux.HandleFunc("/user/emails", func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]gitHubEmail{
			{Email: "primary@example.com", Primary: true, Verified: true},
			{Email: "secondary@example.com", Primary: false, Verified: true},
		})
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	p := NewProvider("id", "secret", "http://localhost:8080/callback", []string{"user:email"})
	p.SetBaseURL(server.URL)

	user, err := p.fetchUser(context.Background(), "test_token")
	if err != nil {
		t.Fatalf("fetchUser() error = %v", err)
	}
	if user.Email != "primary@example.com" {
		t.Errorf("Email = %q, want %q", user.Email, "primary@example.com")
	}
	if callCount != 1 {
		t.Errorf("expected 1 call to /user/emails, got %d", callCount)
	}
}

func TestProvider_GetUserInfo_NoVerifiedEmail(t *testing.T) {
	mux := http.NewServeMux()

	mux.HandleFunc("/user", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(gitHubUser{
			Login: "testuser",
			Email: "",
			ID:    12345,
		})
	})

	mux.HandleFunc("/user/emails", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]gitHubEmail{
			{Email: "unverified@example.com", Primary: true, Verified: false},
		})
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	p := NewProvider("id", "secret", "http://localhost:8080/callback", []string{"user:email"})
	p.SetBaseURL(server.URL)

	_, err := p.fetchUser(context.Background(), "test_token")
	if err == nil {
		t.Fatal("expected error for no verified email")
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
