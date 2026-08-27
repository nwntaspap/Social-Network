package integration_test

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"social-network/internal/bootstrap"
	"social-network/internal/config"
	coreserver "social-network/internal/core/server"
	"social-network/internal/pkg/path"
	"social-network/internal/platform/database"
)

func TestNotificationFlow(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// --- core backend server ---
	cfg := &config.ServerConfig{
		Host: "localhost",
		Port: "0",
		Database: config.DatabaseConfig{
			Driver:         "sqlite3",
			Path:           ":memory:",
			Pragma:         "_foreign_keys=on",
			MigrateOnStart: true,
		},
		SessionManager: config.SessionManagerConfig{
			DefaultExpiry:      24 * time.Hour,
			AccessCookieName:   "access_token",
			RefreshCookieName:  "refresh_token",
			CookiePath:         "/",
			SecureCookie:       false,
			HTTPOnlyCookie:     true,
			SameSite:           "Lax",
			MaxSessionsPerUser: 5,
			CleanupInterval:    3600,
			SessionIDLength:    32,
			EnablePersistence:  true,
		},
	}

	db, err := database.NewDB(database.Config{
		Driver: cfg.Database.Driver,
		Path:   cfg.Database.Path,
		Pragma: cfg.Database.Pragma,
	})
	if err != nil {
		t.Fatalf("new db: %v", err)
	}
	defer db.Close()

	resolver := path.NewResolver()
	migrator := database.NewMigrator(db, resolver.GetPath("db/migrations"))
	if err := migrator.Up(ctx); err != nil {
		t.Fatalf("migrations: %v", err)
	}

	app := bootstrap.Bootstrap(db, cfg)

	srv := coreserver.New(
		cfg,
		coreserver.WithAuth(app.SessionStore, cfg.SessionManager.AccessCookieName),
		coreserver.WithHandlers(&coreserver.AllHandlers{
			User:    app.User,
			Follow:  app.Follow,
			Chat:    app.Chat,
			Comment: app.Comment,
			Topic:   app.Topic,
			Group:   app.Group,
			Event:   app.Event,
			OAuth:   app.OAuth,
		}),
	)
	httpSrv := httptest.NewServer(srv.Handler())
	defer httpSrv.Close()

	// --- notifications service ---
	notifURL, stopNotif := startNotificationService(t, ctx, httpSrv.URL+"/api/v1")
	t.Cleanup(stopNotif)

	client := &http.Client{
		Timeout: 10 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	api := httpSrv.URL + "/api/v1"
	ts := time.Now().Unix()
	aliceEmail := fmt.Sprintf("alice-%d@test.com", ts)
	bobEmail := fmt.Sprintf("bob-%d@test.com", ts)
	aliceNick := fmt.Sprintf("alice-%d", ts)
	bobNick := fmt.Sprintf("bob-%d", ts)

	// Step 1: Register alice
	t.Log("registering alice")
	aliceID := registerUser(t, client, api, aliceEmail, "password123", "Alice", "Smith", aliceNick)
	if aliceID == "" {
		t.Fatal("alice ID is empty")
	}
	t.Logf("alice registered: %s", aliceID)

	// Step 2: Login as alice
	t.Log("logging in as alice")
	aliceToken := loginUser(t, client, api, aliceEmail, "password123")
	if aliceToken == "" {
		t.Fatal("alice token is empty")
	}
	t.Logf("alice token: %s", aliceToken)

	// Step 3: Alice opens notification stream (SSE on notif service)
	t.Log("alice opening notification stream on notifications service")
	notificationCh := make(chan string, 10)
	streamCtx, streamCancel := context.WithCancel(ctx)
	defer streamCancel()
	go streamNotifications(t, streamCtx, notifURL, aliceToken, notificationCh)

	// Give SSE connection time to establish and auth to validate
	time.Sleep(time.Second)

	// Step 4: Alice creates a post
	t.Log("alice creating a post")
	aliceCookie := &http.Cookie{Name: "access_token", Value: aliceToken}
	postID := createPost(t, client, api, aliceCookie, "Test Post Title", "Test post content", "public")
	if postID == 0 {
		t.Fatal("post ID is 0")
	}
	t.Logf("post created: %d", postID)

	// Step 5: Register bob
	t.Log("registering bob")
	bobID := registerUser(t, client, api, bobEmail, "password456", "Bob", "Jones", bobNick)
	if bobID == "" {
		t.Fatal("bob ID is empty")
	}
	t.Logf("bob registered: %s", bobID)

	// Step 6: Login as bob
	t.Log("logging in as bob")
	bobToken := loginUser(t, client, api, bobEmail, "password456")
	if bobToken == "" {
		t.Fatal("bob token is empty")
	}
	t.Logf("bob token: %s", bobToken)

	// Step 7: Bob likes alice's post
	t.Log("bob liking alice's post")
	bobCookie := &http.Cookie{Name: "access_token", Value: bobToken}
	likePost(t, client, api, bobCookie, postID)

	// Step 8: Wait for notification and check (skip SSE meta events)
	t.Log("waiting for notification on alice's stream")
NotifLoop:
	for {
		select {
		case notif := <-notificationCh:
			t.Logf("SSE event received: %s", notif)
			// SSE meta events have lowercase "type" field
			var meta struct {
				Type string `json:"type"`
			}
			json.Unmarshal([]byte(notif), &meta)
			if meta.Type == "connected" || meta.Type == "unread_count" {
				t.Logf("skipping meta event: %s", meta.Type)
				continue
			}
			// Notification payload has capitalized "Type" — skip own events
			var notifObj struct {
				Type string `json:"Type"`
			}
			json.Unmarshal([]byte(notif), &notifObj)
			if notifObj.Type == "post_created" {
				t.Logf("skipping own post_created notification")
				continue
			}
			if !strings.Contains(notifObj.Type, "like") {
				t.Errorf("expected like notification, got type=%q: %s", notifObj.Type, notif)
			}
			break NotifLoop
		case <-time.After(15 * time.Second):
			t.Fatal("timed out waiting for notification via SSE")
		}
	}
}

// startNotificationService builds the notifications microservice and starts it
// as a subprocess. It returns the base URL and a stop function.
func startNotificationService(t *testing.T, ctx context.Context, backendURL string) (string, func()) {
	t.Helper()

	resolver := path.NewResolver()
	notifDir := filepath.Join(resolver.GetPath("."), "services/notifications")
	binPath := filepath.Join(os.TempDir(), fmt.Sprintf("notif-svc-%d", time.Now().UnixNano()))

	// Build the binary
	t.Log("building notifications service...")
	build := exec.CommandContext(ctx, "go", "build", "-o", binPath, "./cmd/server")
	build.Dir = notifDir
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build notifications service: %v\n%s", err, string(out))
	}

	// Find a free port
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("find free port: %v", err)
	}
	port := listener.Addr().(*net.TCPAddr).Port
	listener.Close()

	// Create temp dir for notifications DB
	tmpDir, err := os.MkdirTemp("", "notif-test-*")
	if err != nil {
		t.Fatalf("temp dir: %v", err)
	}

	notifURL := fmt.Sprintf("http://127.0.0.1:%d/api/v1/notifications", port)

	cmd := exec.CommandContext(ctx, binPath)
	cmd.Dir = notifDir
	cmd.Env = append(
		os.Environ(),
		"NOTIFICATIONS_HOST=127.0.0.1",
		fmt.Sprintf("NOTIFICATIONS_PORT=%d", port),
		"NOTIFICATIONS_BACKEND_URL="+backendURL,
		"NOTIFICATIONS_DB_PATH="+filepath.Join(tmpDir, "notifications.db"),
		"NOTIFICATIONS_READ_TIMEOUT=10",
		"NOTIFICATIONS_WRITE_TIMEOUT=20",
		"NOTIFICATIONS_IDLE_TIMEOUT=30",
	)
	// Forward stderr so we see any startup errors
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		os.RemoveAll(tmpDir)
		os.Remove(binPath)
		t.Fatalf("start notifications service: %v", err)
	}

	stop := func() {
		cmd.Process.Kill()
		cmd.Wait()
		os.RemoveAll(tmpDir)
		os.Remove(binPath)
	}

	// Wait for the service to be ready (poll SSE endpoint — expect 401 without auth)
	t.Log("waiting for notifications service to be ready...")
	ready := false
	pollCtx, pollCancel := context.WithTimeout(ctx, 10*time.Second)
	defer pollCancel()
	for {
		select {
		case <-pollCtx.Done():
			stop()
			t.Fatal("notifications service did not become ready in time")
		default:
		}
		req, _ := http.NewRequestWithContext(pollCtx, http.MethodGet, notifURL+"/stream", nil)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			time.Sleep(200 * time.Millisecond)
			continue
		}
		resp.Body.Close()
		if resp.StatusCode == http.StatusUnauthorized {
			ready = true
			break
		}
		// Also accept 404 or other non-connection-refused responses
		if resp.StatusCode != 0 {
			ready = true
			break
		}
		time.Sleep(200 * time.Millisecond)
	}
	if !ready {
		stop()
		t.Fatal("notifications service not ready")
	}
	t.Logf("notifications service ready at %s", notifURL)

	return notifURL, stop
}

func registerUser(t *testing.T, client *http.Client, api, email, password, firstName, lastName, nickname string) string {
	t.Helper()

	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	fields := map[string]string{
		"email":       email,
		"password":    password,
		"firstName":   firstName,
		"lastName":    lastName,
		"nickname":    nickname,
		"dateOfBirth": "2000-01-01T00:00:00Z",
	}
	for k, v := range fields {
		if err := w.WriteField(k, v); err != nil {
			t.Fatalf("write field %s: %v", k, err)
		}
	}
	if err := w.Close(); err != nil {
		t.Fatalf("close multipart writer: %v", err)
	}

	resp, err := client.Post(api+"/register", w.FormDataContentType(), &buf)
	if err != nil {
		t.Fatalf("register %s: %v", email, err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		t.Fatalf("register %s: status=%d body=%s", email, resp.StatusCode, string(respBody))
	}

	var result struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		t.Fatalf("register %s: unmarshal: %v body=%s", email, err, string(respBody))
	}
	return result.Data.ID
}

func loginUser(t *testing.T, client *http.Client, api, identifier, password string) string {
	t.Helper()

	payload := map[string]string{
		"identifier": identifier,
		"password":   password,
	}
	body, _ := json.Marshal(payload)

	resp, err := client.Post(api+"/login/email", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("login %s: %v", identifier, err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("login %s: status=%d body=%s", identifier, resp.StatusCode, string(respBody))
	}

	var result struct {
		Data struct {
			Token string `json:"token"`
		} `json:"data"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		t.Fatalf("login %s: unmarshal: %v body=%s", identifier, err, string(respBody))
	}
	if result.Data.Token == "" {
		t.Fatalf("login %s: empty token in response", identifier)
	}
	return result.Data.Token
}

func createPost(t *testing.T, client *http.Client, api string, cookie *http.Cookie, title, content, privacy string) int {
	t.Helper()

	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	if err := w.WriteField("title", title); err != nil {
		t.Fatalf("write title field: %v", err)
	}
	if err := w.WriteField("content", content); err != nil {
		t.Fatalf("write content field: %v", err)
	}
	if privacy != "" {
		if err := w.WriteField("privacy", privacy); err != nil {
			t.Fatalf("write privacy field: %v", err)
		}
	}
	if err := w.Close(); err != nil {
		t.Fatalf("close multipart: %v", err)
	}

	req, err := http.NewRequest(http.MethodPost, api+"/topics/create", &buf)
	if err != nil {
		t.Fatalf("create post request: %v", err)
	}
	req.Header.Set("Content-Type", w.FormDataContentType())
	req.AddCookie(cookie)

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("create post do: %v", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		t.Fatalf("create post: status=%d body=%s", resp.StatusCode, string(respBody))
	}

	var result struct {
		Data struct {
			ID string `json:"id"`
		}
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		t.Fatalf("create post unmarshal: %v body=%s", err, string(respBody))
	}
	id, err := strconv.Atoi(result.Data.ID)
	if err != nil {
		t.Fatalf("create post id parse: %v body=%s", err, string(respBody))
	}
	return id
}

func likePost(t *testing.T, client *http.Client, api string, cookie *http.Cookie, postID int) {
	t.Helper()

	body := bytes.NewReader([]byte(`{"reactionType":1}`))
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf(api+"/topics/vote?id=%d", postID), body)
	if err != nil {
		t.Fatalf("like post request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(cookie)

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("like post do: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		t.Fatalf("like post: status=%d body=%s", resp.StatusCode, string(respBody))
	}
}

func streamNotifications(t *testing.T, ctx context.Context, baseURL, token string, ch chan<- string) {
	t.Helper()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+"/stream", nil)
	if err != nil {
		t.Logf("stream request error: %v", err)
		return
	}
	req.AddCookie(&http.Cookie{Name: "access_token", Value: token})

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Logf("stream do error: %v", err)
		return
	}
	defer resp.Body.Close()

	t.Logf("stream response status: %d", resp.StatusCode)

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Logf("stream response body: %s", string(body))
		return
	}

	reader := bufio.NewReader(resp.Body)
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		line, err := reader.ReadString('\n')
		if err != nil {
			t.Logf("stream read error: %v", err)
			return
		}

		line = strings.TrimSpace(line)

		if after, ok := strings.CutPrefix(line, "data:"); ok {
			data := strings.TrimSpace(after)
			select {
			case ch <- data:
			default:
			}
		}
	}
}
