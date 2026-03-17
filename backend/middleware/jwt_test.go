package middleware_test

import (
	"fmt"
	"net/http/httptest"
	"testing"
	"time"

	"outcraftly/accounts/config"
	"outcraftly/accounts/middleware"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

const testSecret = "test-jwt-secret-for-unit-tests"

func makeToken(t *testing.T, sub, email string, exp time.Time) string {
	t.Helper()
	claims := jwt.MapClaims{
		"sub":   sub,
		"email": email,
		"role":  "user",
		"iss":   "test-issuer",
		"aud":   "test-aud",
		"exp":   exp.Unix(),
		"iat":   time.Now().Unix(),
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	s, err := tok.SignedString([]byte(testSecret))
	if err != nil {
		t.Fatalf("makeToken: %v", err)
	}
	return s
}

func setupApp() *fiber.App {
	config.Cfg = &config.Config{
		JWTSecret:   testSecret,
		JWTIssuer:   "test-issuer",
		JWTAudience: "test-aud",
	}
	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	app.Use(middleware.Protected())
	app.Get("/ping", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"userID": c.Locals("userID"),
			"email":  c.Locals("email"),
		})
	})
	return app
}

// ─── unit: Protected middleware ───────────────────────────────────────────────

func TestProtected_MissingHeader(t *testing.T) {
	app := setupApp()
	req := httptest.NewRequest("GET", "/ping", nil)
	resp, _ := app.Test(req)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestProtected_MalformedHeader(t *testing.T) {
	app := setupApp()
	for _, bad := range []string{"token abc", "Basic abc", "Bearer", ""} {
		req := httptest.NewRequest("GET", "/ping", nil)
		req.Header.Set("Authorization", bad)
		resp, _ := app.Test(req)
		assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode, "header: %q", bad)
	}
}

func TestProtected_ExpiredToken(t *testing.T) {
	app := setupApp()
	tok := makeToken(t, "uid-1", "a@b.com", time.Now().Add(-time.Hour))
	req := httptest.NewRequest("GET", "/ping", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	resp, _ := app.Test(req)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestProtected_WrongSecret(t *testing.T) {
	app := setupApp()
	claims := jwt.MapClaims{"sub": "uid-1", "email": "a@b.com", "exp": time.Now().Add(time.Hour).Unix()}
	tok, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte("wrong-secret"))
	req := httptest.NewRequest("GET", "/ping", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	resp, _ := app.Test(req)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestProtected_ValidToken(t *testing.T) {
	app := setupApp()
	uid := "550e8400-e29b-41d4-a716-446655440000"
	email := "test@example.com"
	tok := makeToken(t, uid, email, time.Now().Add(time.Hour))
	req := httptest.NewRequest("GET", "/ping", nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", tok))
	resp, _ := app.Test(req)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestProtected_SetLocals(t *testing.T) {
	app := setupApp()
	uid := "550e8400-e29b-41d4-a716-446655440001"
	email := "locals@example.com"
	tok := makeToken(t, uid, email, time.Now().Add(time.Hour))
	req := httptest.NewRequest("GET", "/ping", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	resp, _ := app.Test(req)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}
