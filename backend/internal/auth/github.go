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
	RedirectURL:  os.Getenv("GITHUB_REDIRECT_URL"),
	Scopes:       []string{"read:user", "user:email"},
}

func ExchangeCodeForToken(code string) (*oauth2.Token, error) {
	return GithubOAuthConfig.Exchange(context.Background(), code)
}
