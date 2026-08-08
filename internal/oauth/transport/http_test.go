package transport

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	coremiddleware "social-network/internal/core/middleware"
	"social-network/internal/oauth"
	"social-network/internal/oauth/commands"
)

const testFrontendURL = "http://localhost:3001/auth/callback"

// ─── Stubs ────────────────────────────────────────────────────────────────────

type stubRepo struct {
	providerID   string
	email        string
	createUserID string
	createCalls  int
}

func (r *stubRepo) GetUserByProviderID(_ context.Context, _ oauth.Provider, _ string) (string, error) {
	return r.providerID, nil
}

func (r *stubRepo) GetUserByEmail(_ context.Context, _ string) (string, error) {
	return r.email, nil
}

func (r *stubRepo) CreateOAuthUser(_ context.Context, _ *oauth.User) (string, error) {
	r.createCalls++
	return r.createUserID, nil
}

func (r *stubRepo) LinkOAuthProvider(_ context.Context, _ string, _ *oauth.User) error {
	return nil
}

func (r *stubRepo) GetOAuthProvider(_ context.Context, _ string, _ oauth.Provider) (*oauth.User, error) {
	return nil, errors.New("not implemented")
}

type stubStateVerifier struct {
	data commands.StateData
	err  error
}

func (s *stubStateVerifier) Verify(_ string) (commands.StateData, error) {
	return s.data, s.err
}

type stubStateGenerator struct {
	state string
}

func (s *stubStateGenerator) Generate(_ commands.StateData) (string, error) {
	return s.state, nil
}

type stubProviderClient struct {
	name        string
	authURL     string
	user        *oauth.User
	exchangeErr error
	infoErr     error
}

func (p *stubProviderClient) Name() string { return p.name }

func (p *stubProviderClient) GetAuthURL(state string) string {
	if p.authURL == "" {
		return "https://example.com/authorize?state=" + state
	}
	return p.authURL + "?state=" + state
}

func (p *stubProviderClient) ExchangeCode(_ context.Context, _ string) (string, error) {
	if p.exchangeErr != nil {
		return "", p.exchangeErr
	}
	return "access-token", nil
}

func (p *stubProviderClient) GetUserInfo(_ context.Context, _ string) (*oauth.User, error) {
	if p.infoErr != nil {
		return nil, p.infoErr
	}
	return p.user, nil
}

type stubSessionCreator struct {
	sess *oauth.Session
}

func (s *stubSessionCreator) CreateSession(_ context.Context, _ string) (*oauth.Session, error) {
	return s.sess, nil
}

type stubCookieSetter struct {
	called bool
	token  string
}

func (s *stubCookieSetter) SetCookies(_ http.ResponseWriter, session *oauth.Session) {
	s.called = true
	s.token = session.AccessToken
}

type realCookieSetter struct {
	cookies *coremiddleware.SessionCookies
}

func (s *realCookieSetter) SetCookies(w http.ResponseWriter, session *oauth.Session) {
	s.cookies.SetAccessCookie(w, session.AccessToken, session.ExpiresAt)
}

type stubProviderRegistry struct {
	providers map[string]oauth.ProviderClient
}

func (r *stubProviderRegistry) Get(name string) (oauth.ProviderClient, bool) {
	p, ok := r.providers[name]
	return p, ok
}

// registryProviderClient satisfies oauth.ProviderClient (used by InitiateHandler's registry).
type registryProviderClient struct {
	name    string
	authURL string
}

func (p *registryProviderClient) Name() string { return p.name }

func (p *registryProviderClient) GetAuthURL(state string) string {
	return p.authURL + "?state=" + state
}

func (p *registryProviderClient) ExchangeCode(_ context.Context, _ string) (string, error) {
	return "", nil
}

func (p *registryProviderClient) GetUserInfo(_ context.Context, _ string) (*oauth.ProviderUserInfo, error) {
	return nil, errors.New("not implemented")
}

// ─── Helpers ──────────────────────────────────────────────────────────────────

func newCallbackHandler(repo *stubRepo, verifier *stubStateVerifier, provider *stubProviderClient, sc *stubSessionCreator, name string) *commands.CallbackHandler {
	return commands.NewCallbackHandler(repo, verifier, provider, sc, name)
}

func newHandler(callbacks map[string]*commands.CallbackHandler, cookieSetter oauth.CookieSetter) *Handler {
	initiate := commands.NewInitiateHandler(&stubStateGenerator{state: "test-state"}, &stubProviderRegistry{})
	return NewHandler(initiate, callbacks, cookieSetter, func(_ *http.Request) (string, bool) { return "user-1", true }, testFrontendURL)
}

func doGet(handler *Handler, path string) *httptest.ResponseRecorder {
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)
	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, path, nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	return rec
}

// ─── Callback tests ───────────────────────────────────────────────────────────

func TestCallback_HappyPathNewUser_SetsCookieAndRedirects(t *testing.T) {
	repo := &stubRepo{providerID: "", email: "", createUserID: "new-user-1"}
	verifier := &stubStateVerifier{data: commands.StateData{Flow: "login", Provider: "github"}}
	provider := &stubProviderClient{
		name: "github",
		user: &oauth.User{ProviderID: "gh-1", Email: "gh@example.com", Username: "ghuser"},
	}
	expiresAt := time.Now().Add(time.Hour)
	sc := &stubSessionCreator{sess: &oauth.Session{AccessToken: "tok", ExpiresAt: expiresAt}}
	cookieSetter := &stubCookieSetter{}

	handler := newHandler(map[string]*commands.CallbackHandler{
		"github": newCallbackHandler(repo, verifier, provider, sc, "github"),
	}, cookieSetter)

	rec := doGet(handler, "/api/v1/auth/oauth/github/callback?code=abc&state=xyz")

	if rec.Code != http.StatusTemporaryRedirect {
		t.Fatalf("status = %d, want 307", rec.Code)
	}
	if got := rec.Header().Get("Location"); !strings.Contains(got, "http://localhost:3001/auth/callback") ||
		!strings.Contains(got, "flow=login&success=ok&provider=github") {
		t.Errorf("Location = %q, want success redirect for github", got)
	}
	if !cookieSetter.called {
		t.Fatal("cookie setter not called on successful login")
	}
	if cookieSetter.token != "tok" {
		t.Errorf("cookie token = %q, want tok", cookieSetter.token)
	}
}

func TestCallback_ExistingUser_DoesNotCreateNew(t *testing.T) {
	repo := &stubRepo{providerID: "existing-1", email: "", createUserID: ""}
	verifier := &stubStateVerifier{data: commands.StateData{Flow: "login", Provider: "github"}}
	provider := &stubProviderClient{name: "github", user: &oauth.User{ProviderID: "gh-1", Email: "gh@example.com"}}
	sc := &stubSessionCreator{sess: &oauth.Session{AccessToken: "tok", ExpiresAt: time.Now().Add(time.Hour)}}

	handler := newHandler(map[string]*commands.CallbackHandler{
		"github": newCallbackHandler(repo, verifier, provider, sc, "github"),
	}, &stubCookieSetter{})

	rec := doGet(handler, "/api/v1/auth/oauth/github/callback?code=abc&state=xyz")

	if rec.Code != http.StatusTemporaryRedirect {
		t.Fatalf("status = %d, want 307", rec.Code)
	}
	if got := rec.Header().Get("Location"); !strings.Contains(got, "success=ok") {
		t.Errorf("Location = %q, want success", got)
	}
}

func TestCallback_GoogleDispatchesToGoogleHandler(t *testing.T) {
	githubRepo := &stubRepo{providerID: "", email: "", createUserID: "gh-user"}
	googleRepo := &stubRepo{providerID: "", email: "", createUserID: "google-user"}
	verifier := &stubStateVerifier{data: commands.StateData{Flow: "login", Provider: "google"}}
	googleProvider := &stubProviderClient{name: "google", user: &oauth.User{ProviderID: "goog-1", Email: "g@example.com"}}
	googleSC := &stubSessionCreator{sess: &oauth.Session{AccessToken: "tok-g", ExpiresAt: time.Now().Add(time.Hour)}}

	handler := newHandler(map[string]*commands.CallbackHandler{
		"github": newCallbackHandler(githubRepo, verifier, &stubProviderClient{name: "github", user: &oauth.User{}}, &stubSessionCreator{sess: &oauth.Session{}}, "github"),
		"google": newCallbackHandler(googleRepo, verifier, googleProvider, googleSC, "google"),
	}, &stubCookieSetter{})

	rec := doGet(handler, "/api/v1/auth/oauth/google/callback?code=abc&state=xyz")

	if rec.Code != http.StatusTemporaryRedirect {
		t.Fatalf("status = %d, want 307", rec.Code)
	}
	if got := rec.Header().Get("Location"); !strings.Contains(got, "provider=google") {
		t.Errorf("Location = %q, want provider=google", got)
	}
	if googleRepo.createCalls != 1 {
		t.Fatalf("google repo CreateOAuthUser calls = %d, want 1", googleRepo.createCalls)
	}
	if githubRepo.createCalls != 0 {
		t.Fatalf("github repo CreateOAuthUser calls = %d, want 0", githubRepo.createCalls)
	}
}

func TestCallback_SetsRealAccessCookieHeader(t *testing.T) {
	repo := &stubRepo{providerID: "", email: "", createUserID: "new-user-1"}
	verifier := &stubStateVerifier{data: commands.StateData{Flow: "login", Provider: "github"}}
	provider := &stubProviderClient{name: "github", user: &oauth.User{ProviderID: "gh-1", Email: "gh@example.com"}}
	expiresAt := time.Now().Add(time.Hour)
	sc := &stubSessionCreator{sess: &oauth.Session{AccessToken: "tok", ExpiresAt: expiresAt}}

	cookies := coremiddleware.NewSessionCookies(coremiddleware.CookieConfig{
		Name:     "access_token",
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
	})

	handler := newHandler(map[string]*commands.CallbackHandler{
		"github": newCallbackHandler(repo, verifier, provider, sc, "github"),
	}, &realCookieSetter{cookies: cookies})

	rec := doGet(handler, "/api/v1/auth/oauth/github/callback?code=abc&state=xyz")

	setCookie := rec.Header().Get("Set-Cookie")
	if setCookie == "" {
		t.Fatal("expected Set-Cookie header on successful callback")
	}
	if !strings.HasPrefix(setCookie, "access_token=tok") {
		t.Errorf("Set-Cookie = %q, want access_token=tok", setCookie)
	}
}

func TestCallback_CodeMissing(t *testing.T) {
	handler := newHandler(map[string]*commands.CallbackHandler{
		"github": newCallbackHandler(&stubRepo{}, &stubStateVerifier{}, &stubProviderClient{name: "github"}, &stubSessionCreator{}, "github"),
	}, &stubCookieSetter{})

	rec := doGet(handler, "/api/v1/auth/oauth/github/callback")

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500 for missing code", rec.Code)
	}
}

func TestCallback_EmailAlreadyExists_RedirectsWithError(t *testing.T) {
	repo := &stubRepo{providerID: "", email: "existing@example.com"}
	verifier := &stubStateVerifier{data: commands.StateData{Flow: "login", Provider: "github"}}
	provider := &stubProviderClient{name: "github", user: &oauth.User{ProviderID: "gh-1", Email: "existing@example.com"}}

	handler := newHandler(map[string]*commands.CallbackHandler{
		"github": newCallbackHandler(repo, verifier, provider, &stubSessionCreator{}, "github"),
	}, &stubCookieSetter{})

	rec := doGet(handler, "/api/v1/auth/oauth/github/callback?code=abc&state=xyz")

	if rec.Code != http.StatusTemporaryRedirect {
		t.Fatalf("status = %d, want 307", rec.Code)
	}
	if got := rec.Header().Get("Location"); !strings.Contains(got, "error=email_exists") {
		t.Errorf("Location = %q, want error=email_exists", got)
	}
}

func TestCallback_UnknownProvider(t *testing.T) {
	handler := newHandler(map[string]*commands.CallbackHandler{
		"github": newCallbackHandler(&stubRepo{}, &stubStateVerifier{}, &stubProviderClient{name: "github"}, &stubSessionCreator{}, "github"),
	}, &stubCookieSetter{})

	rec := doGet(handler, "/api/v1/auth/oauth/unknown/callback?code=abc&state=xyz")

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404 for unknown provider", rec.Code)
	}
}

// ─── Initiate tests ───────────────────────────────────────────────────────────

func TestInitiate_GithubRedirectsToAuthorize(t *testing.T) {
	providers := &stubProviderRegistry{providers: map[string]oauth.ProviderClient{
		"github": &registryProviderClient{name: "github", authURL: "https://github.com/login/oauth/authorize"},
	}}
	initiate := commands.NewInitiateHandler(&stubStateGenerator{state: "test-state"}, providers)
	handler := NewHandler(initiate, nil, nil, func(_ *http.Request) (string, bool) { return "", true }, testFrontendURL)

	rec := doGet(handler, "/api/v1/auth/oauth/github/init")

	if rec.Code != http.StatusTemporaryRedirect {
		t.Fatalf("status = %d, want 307", rec.Code)
	}
	loc := rec.Header().Get("Location")
	if !strings.HasPrefix(loc, "https://github.com/login/oauth/authorize") {
		t.Errorf("Location = %q, want github authorize URL", loc)
	}
	if !strings.Contains(loc, "state=test-state") {
		t.Errorf("Location = %q, want embedded state", loc)
	}
}

func TestInitiate_UnknownProvider(t *testing.T) {
	initiate := commands.NewInitiateHandler(&stubStateGenerator{state: "test-state"}, &stubProviderRegistry{})
	handler := NewHandler(initiate, nil, nil, func(_ *http.Request) (string, bool) { return "", true }, testFrontendURL)

	rec := doGet(handler, "/api/v1/auth/oauth/unknown/init")

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500 for unknown provider", rec.Code)
	}
}
