package integration_test

// ── Smoke Tests ───────────────────────────────────────────────────────────────
//
// These tests verify that every top-level route responds with the expected
// HTTP status on a fresh boot — they don't test logic, only reachability.
// A failing smoke test signals that a route was broken or removed.

import (
	"net/http/httptest"
	"testing"

	"outcraftly/accounts/testhelpers"

	"github.com/stretchr/testify/assert"
)

// ── Public endpoints — should always be reachable ─────────────────────────────

func TestSmoke_HealthCheck(t *testing.T) {
	testhelpers.SetupTestDB()
	app := testhelpers.NewApp()

	req := httptest.NewRequest("GET", "/api/v1/health", nil)
	resp, _ := app.Test(req)
	assert.Equal(t, 200, resp.StatusCode)
}

func TestSmoke_PublicConfig(t *testing.T) {
	testhelpers.SetupTestDB()
	app := testhelpers.NewApp()

	req := httptest.NewRequest("GET", "/api/v1/config", nil)
	resp, _ := app.Test(req)
	// May return 200 even with empty stripe key
	assert.Less(t, resp.StatusCode, 500)
}

func TestSmoke_AuthEndpoints_NoBody_Return400or422(t *testing.T) {
	testhelpers.SetupTestDB()
	app := testhelpers.NewApp()

	endpoints := []string{
		"/api/v1/auth/register",
		"/api/v1/auth/login",
		"/api/v1/auth/forgot-password",
		"/api/v1/auth/verify-email-otp",
		"/api/v1/auth/verify-reset-otp",
		"/api/v1/auth/reset-password",
	}
	for _, ep := range endpoints {
		req := httptest.NewRequest("POST", ep, nil)
		req.Header.Set("Content-Type", "application/json")
		resp, _ := app.Test(req)
		// Bad input → 400/422, NOT 404/500
		assert.True(t, resp.StatusCode == 400 || resp.StatusCode == 422,
			"expected 400 or 422 for %s, got %d", ep, resp.StatusCode)
	}
}

// ── Protected endpoints — must return 401 without token ──────────────────────

func TestSmoke_ProtectedEndpoints_Require401(t *testing.T) {
	testhelpers.SetupTestDB()
	app := testhelpers.NewApp()

	cases := []struct {
		method string
		path   string
	}{
		{"GET", "/api/v1/profile"},
		{"POST", "/api/v1/profile"},
		{"POST", "/api/v1/auth/change-password"},
		{"GET", "/api/v1/workspaces"},
		{"POST", "/api/v1/workspaces"},
		{"GET", "/api/v1/products"},
	}
	for _, tc := range cases {
		req := httptest.NewRequest(tc.method, tc.path, nil)
		resp, _ := app.Test(req)
		assert.Equal(t, 401, resp.StatusCode,
			"expected 401 for %s %s, got %d", tc.method, tc.path, resp.StatusCode)
	}
}

// ── Admin endpoints — must return 403/503 without the admin secret ────────────

func TestSmoke_AdminEndpoints_Require403(t *testing.T) {
	testhelpers.SetupTestDB()
	app := testhelpers.NewApp()

	cases := []struct {
		method string
		path   string
	}{
		{"GET", "/api/v1/admin/products"},
		{"POST", "/api/v1/admin/products"},
		{"GET", "/api/v1/admin/subscriptions"},
		{"GET", "/api/v1/admin/billing"},
		{"GET", "/api/v1/admin/users"},
		{"GET", "/api/v1/admin/workspaces"},
	}
	for _, tc := range cases {
		req := httptest.NewRequest(tc.method, tc.path, nil)
		resp, _ := app.Test(req)
		assert.Equal(t, 403, resp.StatusCode,
			"expected 403 for %s %s, got %d", tc.method, tc.path, resp.StatusCode)
	}
}

// ── 404 for unknown routes ────────────────────────────────────────────────────

func TestSmoke_UnknownRoute_Returns404(t *testing.T) {
	testhelpers.SetupTestDB()
	app := testhelpers.NewApp()

	req := httptest.NewRequest("GET", "/api/v1/does-not-exist", nil)
	resp, _ := app.Test(req)
	assert.Equal(t, 404, resp.StatusCode)
}
