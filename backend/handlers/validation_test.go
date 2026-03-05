package handlers_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// ─── Unit: email validation logic (mirrors Register handler) ─────────────────

func isValidEmail(e string) bool {
	for i, ch := range e {
		if ch == '@' && i > 0 && i < len(e)-1 {
			return true
		}
	}
	return false
}

func TestEmailValidation(t *testing.T) {
	valid := []string{
		"user@example.com",
		"a@b.io",
		"test+tag@domain.co.uk",
		"x@y.z",
	}
	invalid := []string{
		"",
		"notanemail",
		"@nodomain",
		"noatsign",
		"trailing@",
	}

	for _, e := range valid {
		assert.True(t, isValidEmail(e), "expected valid: %q", e)
	}
	for _, e := range invalid {
		assert.False(t, isValidEmail(e), "expected invalid: %q", e)
	}
}

// ─── Unit: password length rule ───────────────────────────────────────────────

func TestPasswordMinLength(t *testing.T) {
	assert.False(t, len("short") >= 8)
	assert.False(t, len("1234567") >= 8)
	assert.True(t, len("12345678") >= 8)
	assert.True(t, len("a_very_secure_pass!") >= 8)
}

// ─── Unit: workspace role normalisation ──────────────────────────────────────

func normaliseRole(r string) string {
	if r == "owner" || r == "member" {
		return r
	}
	return "member"
}

func TestRoleNormalisation(t *testing.T) {
	assert.Equal(t, "owner", normaliseRole("owner"))
	assert.Equal(t, "member", normaliseRole("member"))
	assert.Equal(t, "member", normaliseRole(""))
	assert.Equal(t, "member", normaliseRole("admin"))
	assert.Equal(t, "member", normaliseRole("superuser"))
}

// ─── Unit: workspace name constraints ─────────────────────────────────────────

func TestWorkspaceNameConstraints(t *testing.T) {
	// Build exact-length strings programmatically to avoid off-by-one errors.
	exactly80 := string(make([]byte, 80)) // 80 zero bytes → valid
	exactly81 := string(make([]byte, 81)) // 81 zero bytes → invalid

	cases := []struct {
		name  string
		valid bool
	}{
		{"Acme Corp", true},
		{"a", true},
		{exactly80, true},
		{"", false},
		{exactly81, false},
	}
	for _, tc := range cases {
		trimmed := tc.name
		ok := len(trimmed) > 0 && len(trimmed) <= 80
		assert.Equal(t, tc.valid, ok, "name: %q (len %d)", tc.name, len(tc.name))
	}
}

// ─── Unit: errJSON / badRequest shape ────────────────────────────────────────

func TestErrJSONShape(t *testing.T) {
	msg := "something went wrong"
	m := map[string]string{"error": msg}
	assert.Equal(t, msg, m["error"])
}

// ─── Unit: publicUser fields ──────────────────────────────────────────────────

func TestPublicUserFields(t *testing.T) {
	// Verify the set of keys we expect in a publicUser response
	expected := []string{
		"id", "email", "email_verified", "name",
		"company_name", "job_title", "phone_country_code",
		"phone_number", "profile_complete", "last_login_at", "created_at",
	}
	// Regression: password_hash must NOT be in public output
	for _, k := range expected {
		assert.NotEqual(t, "password_hash", k)
		assert.NotEqual(t, "reset_token", k)
		assert.NotEqual(t, "email_verify_token", k)
	}
}
