package github

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"social-network/pkg/oauth"
	"social-network/pkg/oauth/httpclient"
)

const (
	tokenURL     = "https://github.com/login/oauth/access_token" //nolint:gosec // OAuth endpoint URL, not a credential
	userURL      = "https://api.github.com/user"
	userEmailURL = "https://api.github.com/user/emails"
)

type Provider struct {
	clientID     string
	clientSecret string
	redirectURL  string
	scopes       []string
	tokenURL     string
	userURL      string
	userEmailURL string
}

func NewProvider(clientID, clientSecret, redirectURL string, scopes []string) *Provider {
	return &Provider{
		clientID:     clientID,
		clientSecret: clientSecret,
		redirectURL:  redirectURL,
		scopes:       scopes,
		tokenURL:     tokenURL,
		userURL:      userURL,
		userEmailURL: userEmailURL,
	}
}

// SetBaseURL overrides the default GitHub API URLs for testing.
func (p *Provider) SetBaseURL(baseURL string) {
	p.tokenURL = baseURL + "/login/oauth/access_token"
	p.userURL = baseURL + "/user"
	p.userEmailURL = baseURL + "/user/emails"
}

type tokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	Scope       string `json:"scope"`
}

type gitHubUser struct {
	Login     string `json:"login"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	AvatarURL string `json:"avatar_url"`
	ID        int    `json:"id"`
}

type gitHubEmail struct {
	Email      string `json:"email"`
	Visibility string `json:"visibility"`
	Primary    bool   `json:"primary"`
	Verified   bool   `json:"verified"`
}

func (p *Provider) Name() string {
	return "github"
}

func (p *Provider) GetAuthURL(state string) string {
	params := url.Values{}
	params.Add("client_id", p.clientID)
	params.Add("redirect_uri", p.redirectURL)
	params.Add("scope", strings.Join(p.scopes, " "))
	params.Add("state", state)

	return "https://github.com/login/oauth/authorize?" + params.Encode()
}

func (p *Provider) ExchangeCode(ctx context.Context, code string) (string, error) {
	body := map[string]string{
		"client_id":     p.clientID,
		"client_secret": p.clientSecret,
		"code":          code,
		"redirect_uri":  p.redirectURL,
	}

	headers := map[string]string{
		"Accept":       "application/json",
		"Content-Type": "application/json",
	}

	client := httpclient.NewClient()
	respBody, err := client.Post(ctx, p.tokenURL, headers, body)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrFailedToExchangeCode, err)
	}

	var tokenResp tokenResponse
	err = json.Unmarshal(respBody, &tokenResp)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrFailedToParseToken, err)
	}

	if tokenResp.AccessToken == "" {
		return "", ErrTokenNotFound
	}

	return tokenResp.AccessToken, nil
}

func (p *Provider) GetUserInfo(ctx context.Context, accessToken string) (*oauth.ProviderUserInfo, error) {
	user, err := p.fetchUser(ctx, accessToken)
	if err != nil {
		return nil, err
	}

	return &oauth.ProviderUserInfo{
		ProviderID: strconv.Itoa(user.ID),
		Email:      user.Email,
		Username:   user.Login,
		Name:       user.Name,
		AvatarURL:  user.AvatarURL,
	}, nil
}

func (p *Provider) fetchUser(ctx context.Context, accessToken string) (*gitHubUser, error) {
	headers := map[string]string{
		"Authorization": "Bearer " + accessToken,
		"Accept":        "application/json",
	}

	client := httpclient.NewClient()
	respBody, err := client.Get(ctx, p.userURL, headers)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrFailedToGetUser, err)
	}

	var user gitHubUser
	err = json.Unmarshal(respBody, &user)
	if err != nil {
		return nil, ErrFailedToParseUser
	}

	if user.Email == "" {
		email, err := p.getPrimaryEmail(ctx, accessToken)
		if err != nil {
			return nil, fmt.Errorf("%w: %w", ErrFailedToGetPrimaryEmail, err)
		}
		user.Email = email
	}

	return &user, nil
}

func (p *Provider) getPrimaryEmail(ctx context.Context, accessToken string) (string, error) {
	headers := map[string]string{
		"Authorization": "Bearer " + accessToken,
		"Accept":        "application/json",
	}

	client := httpclient.NewClient()
	respBody, err := client.Get(ctx, p.userEmailURL, headers)
	if err != nil {
		return "", fmt.Errorf("failed to get emails: %w", err)
	}

	var emails []gitHubEmail
	err = json.Unmarshal(respBody, &emails)
	if err != nil {
		return "", fmt.Errorf("failed to parse emails response: %w", err)
	}

	for _, email := range emails {
		if email.Primary && email.Verified {
			return email.Email, nil
		}
	}

	for _, email := range emails {
		if email.Verified {
			return email.Email, nil
		}
	}

	return "", ErrNoVerifiedEmailFound
}
