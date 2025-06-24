package auth

import (
	"context"
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
)

var GithubOAuthConfig = &oauth2.Config{
	ClientID:     os.Getenv("GITHUB_CLIENT_ID"),
	ClientSecret: os.Getenv("GITHUB_CLIENT_SECRET"),
	Endpoint:     github.Endpoint,
	RedirectURL:  "http://localhost:8080/auth/github/callback",
	Scopes:       []string{"read:user", "user:email"},
}

func ExchangeCodeForToken(code string) (*oauth2.Token, error) {
	return GithubOAuthConfig.Exchange(context.Background(), code)
}
