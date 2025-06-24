package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	chi "github.com/go-chi/chi/v5"
)

func TestGetToken_CookieOnly(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	req.AddCookie(&http.Cookie{
		Name:  "webhook_token",
		Value: "abc123",
	})
	rr := httptest.NewRecorder()

	token, ok := GetToken(rr, req)
	if !ok || token != "abc123" {
		t.Errorf("expected token 'abc123', got '%s'", token)
	}
}

func TestGetToken_URLParamMismatch(t *testing.T) {
	req := httptest.NewRequest("GET", "/logs/wrongtoken", nil)
	req.AddCookie(&http.Cookie{
		Name:  "webhook_token",
		Value: "abc123",
	})
	rr := httptest.NewRecorder()

	// Set fake route param manually
	req = req.WithContext(setURLParam(req.Context(), "token", "wrongtoken"))

	_, ok := GetToken(rr, req)
	if ok {
		t.Error("expected token mismatch to return false")
	}
}

func setURLParam(ctx context.Context, key, val string) context.Context {
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add(key, val)
	return context.WithValue(ctx, chi.RouteCtxKey, rctx)
}
