// Package auth handles GitHub OAuth authentication.
package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
)

// GitHubUser is the subset of GitHub's user API response we care about.
type GitHubUser struct {
	ID        int64  `json:"id"`
	Login     string `json:"login"`
	AvatarURL string `json:"avatar_url"`
}

// GitHubOAuth holds the OAuth config and provides helpers.
type GitHubOAuth struct {
	config *oauth2.Config
}

// NewGitHubOAuth creates a GitHub OAuth handler. Credentials come from
// environment variables: GITHUB_CLIENT_ID, GITHUB_CLIENT_SECRET, and
// optionally GITHUB_REDIRECT_URL.
func NewGitHubOAuth() (*GitHubOAuth, error) {
	clientID := os.Getenv("GITHUB_CLIENT_ID")
	clientSecret := os.Getenv("GITHUB_CLIENT_SECRET")
	if clientID == "" || clientSecret == "" {
		return nil, fmt.Errorf("GITHUB_CLIENT_ID and GITHUB_CLIENT_SECRET must be set")
	}

	redirectURL := os.Getenv("GITHUB_REDIRECT_URL")
	if redirectURL == "" {
		redirectURL = "http://localhost:8080/api/auth/callback"
	}

	config := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Scopes:       []string{"read:user"},
		Endpoint:     github.Endpoint,
	}

	return &GitHubOAuth{config: config}, nil
}

// AuthCodeURL returns the URL to redirect the user to for GitHub authorization.
func (g *GitHubOAuth) AuthCodeURL(state string) string {
	return g.config.AuthCodeURL(state)
}

// Exchange converts an authorization code into an OAuth token.
func (g *GitHubOAuth) Exchange(ctx context.Context, code string) (*oauth2.Token, error) {
	return g.config.Exchange(ctx, code)
}

// FetchUser uses an OAuth token to fetch the authenticated GitHub user's profile.
func (g *GitHubOAuth) FetchUser(ctx context.Context, token *oauth2.Token) (*GitHubUser, error) {
	client := g.config.Client(ctx, token)
	resp, err := client.Get("https://api.github.com/user")
	if err != nil {
		return nil, fmt.Errorf("fetch github user: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("github API returned %d", resp.StatusCode)
	}

	var user GitHubUser
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, fmt.Errorf("decode github user: %w", err)
	}
	return &user, nil
}

// Client returns an HTTP client authenticated with the given token.
func (g *GitHubOAuth) Client(ctx context.Context, token *oauth2.Token) *http.Client {
	return g.config.Client(ctx, token)
}
