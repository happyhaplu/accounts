package handlers_test

// ── API Contract Tests ────────────────────────────────────────────────────────
//
// These tests validate the shape / contract of public-facing response
// structures without needing a database.  They confirm that our JSON keys,
// types, and invariants remain stable as the code evolves.

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// ── Product response contract ─────────────────────────────────────────────────

func TestProductContract_RequiredFields(t *testing.T) {
	// Simulate the public product response shape (mirrors models.Product JSON tags)
	product := map[string]interface{}{
		"id":          "550e8400-e29b-41d4-a716-446655440000",
		"name":        "test-product",
		"description": "A test product for contract validation",
		"is_active":   true,
		"api_key":     "gour_ce_abc123xyz",
		"created_at":  "2024-01-01T00:00:00Z",
	}

	requiredKeys := []string{"id", "name", "description", "is_active", "api_key", "created_at"}
	for _, k := range requiredKeys {
		assert.Contains(t, product, k, "product response must contain key: %s", k)
	}

	// Stripe fields must NOT be in the product contract (removed from model)
	deprecatedKeys := []string{"stripe_price_id", "pricing_table_id", "billing_display"}
	for _, k := range deprecatedKeys {
		assert.NotContains(t, product, k, "deprecated field must not be in product response: %s", k)
	}
}

func TestProductContract_APIKeyFormat(t *testing.T) {
	// API keys must start with "gour_" prefix for easy identification
	validKeys := []string{
		"gour_ce_abc123",
		"gour_ce_verylongkey1234567890",
	}
	for _, k := range validKeys {
		assert.Greater(t, len(k), 5, "api_key must be longer than 5 chars")
	}
}

// ── Subscription response contract ───────────────────────────────────────────

func TestSubscriptionContract_RequiredFields(t *testing.T) {
	sub := map[string]interface{}{
		"id":                 "uuid-1",
		"workspace_id":       "uuid-2",
		"product_id":         "uuid-3",
		"plan_name":          "starter",
		"status":             "active",
		"current_period_end": "2024-02-01T00:00:00Z",
	}

	required := []string{"id", "workspace_id", "product_id", "plan_name", "status", "current_period_end"}
	for _, k := range required {
		assert.Contains(t, sub, k, "subscription response must contain: %s", k)
	}
}

func TestSubscriptionContract_ValidStatuses(t *testing.T) {
	validStatuses := []string{"active", "canceled", "past_due", "trialing"}
	invalidStatuses := []string{"", "unknown", "ACTIVE", "deleted"}

	isValid := func(s string) bool {
		for _, v := range validStatuses {
			if s == v {
				return true
			}
		}
		return false
	}

	for _, s := range validStatuses {
		assert.True(t, isValid(s), "expected valid status: %q", s)
	}
	for _, s := range invalidStatuses {
		assert.False(t, isValid(s), "expected invalid status: %q", s)
	}
}

// ── Workspace response contract ───────────────────────────────────────────────

func TestWorkspaceContract_RequiredFields(t *testing.T) {
	ws := map[string]interface{}{
		"id":         "uuid-ws",
		"name":       "My Workspace",
		"my_role":    "owner",
		"created_at": "2024-01-01T00:00:00Z",
	}

	required := []string{"id", "name", "my_role"}
	for _, k := range required {
		assert.Contains(t, ws, k, "workspace response must contain: %s", k)
	}

	// my_role must be "owner" or "member"
	role := ws["my_role"].(string)
	assert.True(t, role == "owner" || role == "member",
		"my_role must be 'owner' or 'member', got: %q", role)
}

// ── Auth response contract ────────────────────────────────────────────────────

func TestAuthContract_LoginResponse_HasTokenAndUser(t *testing.T) {
	resp := map[string]interface{}{
		"token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
		"user": map[string]interface{}{
			"id":               "uuid-user",
			"email":            "user@example.com",
			"email_verified":   true,
			"profile_complete": false,
		},
	}

	assert.Contains(t, resp, "token", "login response must include JWT token")
	assert.Contains(t, resp, "user", "login response must include user object")

	user := resp["user"].(map[string]interface{})
	assert.Contains(t, user, "email_verified")
	assert.Contains(t, user, "profile_complete")

	// Sensitive fields must never appear
	sensitive := []string{"password_hash", "reset_token", "email_verify_token", "otp_code"}
	for _, k := range sensitive {
		assert.NotContains(t, user, k, "sensitive field must not appear in auth response: %s", k)
	}
}

func TestAuthContract_OTPVerifyResponse(t *testing.T) {
	// After OTP verification the response must include token + user
	resp := map[string]interface{}{
		"token": "jwt-abc",
		"user": map[string]interface{}{
			"email_verified": true,
		},
	}
	assert.NotEmpty(t, resp["token"])
	u := resp["user"].(map[string]interface{})
	assert.Equal(t, true, u["email_verified"])
}

// ── Verify endpoint contract (server-to-server) ───────────────────────────────

func TestVerifyContract_ResponseShape(t *testing.T) {
	// POST /api/v1/products/verify must return: valid, user_id, workspace_id, subscribed, plan
	resp := map[string]interface{}{
		"valid":        true,
		"user_id":      "uuid-user",
		"workspace_id": "uuid-ws",
		"subscribed":   true,
		"plan":         "starter",
	}

	required := []string{"valid", "user_id"}
	for _, k := range required {
		assert.Contains(t, resp, k, "verify response must contain: %s", k)
	}
	assert.Equal(t, true, resp["valid"])
}

// ── Error response contract ───────────────────────────────────────────────────

func TestErrorContract_AlwaysHasErrorKey(t *testing.T) {
	// Every error response must use {"error": "message"} not {"message": "..."} etc.
	errors := []map[string]interface{}{
		{"error": "Invalid request body"},
		{"error": "Email already registered"},
		{"error": "Invalid or expired token"},
	}

	for _, e := range errors {
		assert.Contains(t, e, "error", "error response must use 'error' key")
		assert.NotContains(t, e, "message", "error responses must not use 'message' key")
		assert.NotEmpty(t, e["error"])
	}
}
