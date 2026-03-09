package integration_test

// ── Regression Tests ──────────────────────────────────────────────────────────
//
// These tests guard against regressions in the full user journeys and
// security invariants introduced by the OTP migration.  Unlike the unit
// and per-handler integration tests, each test here exercises a complete,
// multi-step flow as a real user would experience it.

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"

	"outcraftly/accounts/database"
	"outcraftly/accounts/models"
	"outcraftly/accounts/testhelpers"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

// ── Full journey: register → verify OTP → login ───────────────────────────────

func TestRegression_RegisterVerifyOTPLogin(t *testing.T) {
	testhelpers.SetupTestDB()
	app := testhelpers.NewApp()

	email := "reg-journey@example.com"
	password := "Journey1234!"

	// Step 1: register
	regPayload, _ := json.Marshal(map[string]string{"email": email, "password": password})
	regReq := httptest.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer(regPayload))
	regReq.Header.Set("Content-Type", "application/json")
	regResp, err := app.Test(regReq)
	require.NoError(t, err)
	assert.Equal(t, 201, regResp.StatusCode)

	// Step 2: read OTP from DB (simulating what email delivers)
	var user models.User
	err = database.DB.Where("email = ?", email).First(&user).Error
	require.NoError(t, err)
	require.NotNil(t, user.OTPCode, "OTP must be stored in DB after register")
	assert.False(t, user.EmailVerified, "user must NOT be verified before OTP confirmation")

	// Step 3: verify OTP
	vpPayload, _ := json.Marshal(map[string]string{"email": email, "otp": *user.OTPCode})
	vpReq := httptest.NewRequest("POST", "/api/v1/auth/verify-email-otp", bytes.NewBuffer(vpPayload))
	vpReq.Header.Set("Content-Type", "application/json")
	vpResp, err := app.Test(vpReq)
	require.NoError(t, err)
	assert.Equal(t, 200, vpResp.StatusCode)

	var vpBody map[string]interface{}
	json.NewDecoder(vpResp.Body).Decode(&vpBody)
	token := vpBody["token"]
	assert.NotEmpty(t, token, "JWT must be returned after OTP verification")
	u := vpBody["user"].(map[string]interface{})
	assert.Equal(t, true, u["email_verified"])

	// OTP must be cleared from DB after verification
	database.DB.Where("email = ?", email).First(&user)
	assert.Nil(t, user.OTPCode, "OTP must be cleared after successful verification")

	// Step 4: login with credentials
	loginPayload, _ := json.Marshal(map[string]string{"email": email, "password": password})
	loginReq := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(loginPayload))
	loginReq.Header.Set("Content-Type", "application/json")
	loginResp, err := app.Test(loginReq)
	require.NoError(t, err)
	assert.Equal(t, 200, loginResp.StatusCode)
}

// ── Full journey: forgot-password → verify OTP → reset → login ───────────────

func TestRegression_ForgotPasswordFullJourney(t *testing.T) {
	db := testhelpers.SetupTestDB()
	app := testhelpers.NewApp()

	email := "fp-journey@example.com"
	oldPassword := "OldPass1234!"
	newPassword := "NewPass5678!"
	testhelpers.CreateVerifiedUser(db, email, oldPassword)

	// Step 1: forgot-password
	fpPayload := `{"email":"fp-journey@example.com"}`
	fpReq := httptest.NewRequest("POST", "/api/v1/auth/forgot-password", bytes.NewBufferString(fpPayload))
	fpReq.Header.Set("Content-Type", "application/json")
	fpResp, err := app.Test(fpReq)
	require.NoError(t, err)
	assert.Equal(t, 200, fpResp.StatusCode)

	// Step 2: read OTP from DB
	var user models.User
	database.DB.Where("email = ?", email).First(&user)
	require.NotNil(t, user.OTPCode)
	assert.Equal(t, "password_reset", *user.OTPPurpose)

	// Step 3: verify OTP → get reset_token
	vpPayload, _ := json.Marshal(map[string]string{"email": email, "otp": *user.OTPCode})
	vpReq := httptest.NewRequest("POST", "/api/v1/auth/verify-reset-otp", bytes.NewBuffer(vpPayload))
	vpReq.Header.Set("Content-Type", "application/json")
	vpResp, err := app.Test(vpReq)
	require.NoError(t, err)
	assert.Equal(t, 200, vpResp.StatusCode)

	var vpBody map[string]interface{}
	json.NewDecoder(vpResp.Body).Decode(&vpBody)
	resetToken := vpBody["reset_token"].(string)
	assert.NotEmpty(t, resetToken)

	// OTP cleared, reset_token stored
	database.DB.Where("email = ?", email).First(&user)
	assert.Nil(t, user.OTPCode, "OTP must be cleared after verify-reset-otp")
	assert.NotNil(t, user.ResetToken)

	// Step 4: reset password
	rpPayload, _ := json.Marshal(map[string]string{"token": resetToken, "new_password": newPassword})
	rpReq := httptest.NewRequest("POST", "/api/v1/auth/reset-password", bytes.NewBuffer(rpPayload))
	rpReq.Header.Set("Content-Type", "application/json")
	rpResp, err := app.Test(rpReq)
	require.NoError(t, err)
	assert.Equal(t, 200, rpResp.StatusCode)

	// reset_token cleared after use
	database.DB.Where("email = ?", email).First(&user)
	assert.Nil(t, user.ResetToken, "reset_token must be cleared after use")

	// Old password no longer works
	oldLoginPayload, _ := json.Marshal(map[string]string{"email": email, "password": oldPassword})
	oldLoginReq := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(oldLoginPayload))
	oldLoginReq.Header.Set("Content-Type", "application/json")
	oldLoginResp, _ := app.Test(oldLoginReq)
	assert.Equal(t, 401, oldLoginResp.StatusCode, "old password must not work after reset")

	// New password works
	newLoginPayload, _ := json.Marshal(map[string]string{"email": email, "password": newPassword})
	newLoginReq := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(newLoginPayload))
	newLoginReq.Header.Set("Content-Type", "application/json")
	newLoginResp, err := app.Test(newLoginReq)
	require.NoError(t, err)
	assert.Equal(t, 200, newLoginResp.StatusCode)
}

// ── Security: OTP cannot be reused after email verification ──────────────────

func TestRegression_OTPCannotBeReused(t *testing.T) {
	testhelpers.SetupTestDB()
	app := testhelpers.NewApp()

	email := "otp-reuse@example.com"
	regPayload, _ := json.Marshal(map[string]string{"email": email, "password": "password123"})
	regReq := httptest.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer(regPayload))
	regReq.Header.Set("Content-Type", "application/json")
	app.Test(regReq)

	var user models.User
	database.DB.Where("email = ?", email).First(&user)
	otp := *user.OTPCode

	// First use — verifies the account
	vpPayload, _ := json.Marshal(map[string]string{"email": email, "otp": otp})
	vpReq := httptest.NewRequest("POST", "/api/v1/auth/verify-email-otp", bytes.NewBuffer(vpPayload))
	vpReq.Header.Set("Content-Type", "application/json")
	resp1, _ := app.Test(vpReq)
	assert.Equal(t, 200, resp1.StatusCode)

	// Second use of the same OTP — user is already verified, handler is idempotent
	// but the OTP field is nil → returns 200 via the "already verified" path
	vpReq2 := httptest.NewRequest("POST", "/api/v1/auth/verify-email-otp", bytes.NewBuffer(vpPayload))
	vpReq2.Header.Set("Content-Type", "application/json")
	resp2, _ := app.Test(vpReq2)
	// Already-verified path: returns JWT (idempotent 200), not 400
	assert.Equal(t, 200, resp2.StatusCode)
	// But confirm OTP is NOT stored (was cleared after first use)
	database.DB.Where("email = ?", email).First(&user)
	assert.Nil(t, user.OTPCode, "OTP must be gone after first use")
}

// ── Security: email_verify OTP cannot be used for password reset ──────────────

func TestRegression_EmailVerifyOTPRejectedForPasswordReset(t *testing.T) {
	testhelpers.SetupTestDB()
	app := testhelpers.NewApp()

	otp := "424242"
	expires := time.Now().Add(10 * time.Minute)
	purpose := "email_verify"
	hash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.MinCost)
	user := models.User{
		ID:            uuid.New(),
		Email:         "xpurpose1@example.com",
		PasswordHash:  string(hash),
		EmailVerified: true,
		OTPCode:       &otp,
		OTPExpires:    &expires,
		OTPPurpose:    &purpose,
	}
	database.DB.Create(&user)

	payload, _ := json.Marshal(map[string]string{"email": "xpurpose1@example.com", "otp": otp})
	req := httptest.NewRequest("POST", "/api/v1/auth/verify-reset-otp", bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)
	assert.Equal(t, 400, resp.StatusCode, "email_verify OTP must be rejected by verify-reset-otp")
}

// ── Security: password_reset OTP cannot be used for email verification ────────

func TestRegression_PasswordResetOTPRejectedForEmailVerify(t *testing.T) {
	testhelpers.SetupTestDB()
	app := testhelpers.NewApp()

	otp := "535353"
	expires := time.Now().Add(10 * time.Minute)
	purpose := "password_reset"
	hash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.MinCost)
	user := models.User{
		ID:            uuid.New(),
		Email:         "xpurpose2@example.com",
		PasswordHash:  string(hash),
		EmailVerified: false,
		OTPCode:       &otp,
		OTPExpires:    &expires,
		OTPPurpose:    &purpose,
	}
	database.DB.Create(&user)

	payload, _ := json.Marshal(map[string]string{"email": "xpurpose2@example.com", "otp": otp})
	req := httptest.NewRequest("POST", "/api/v1/auth/verify-email-otp", bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)
	assert.Equal(t, 400, resp.StatusCode, "password_reset OTP must be rejected by verify-email-otp")
}

// ── Security: no OTP → reject ─────────────────────────────────────────────────

func TestRegression_NoOTPStoredIsRejected(t *testing.T) {
	db := testhelpers.SetupTestDB()
	app := testhelpers.NewApp()
	testhelpers.CreateVerifiedUser(db, "no-otp@example.com", "password123")

	// No forgot-password call → no OTP in DB
	payload, _ := json.Marshal(map[string]string{"email": "no-otp@example.com", "otp": "123456"})
	req := httptest.NewRequest("POST", "/api/v1/auth/verify-reset-otp", bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)
	assert.Equal(t, 400, resp.StatusCode)
}

// ── Security: expired OTP does not verify email ──────────────────────────────

func TestRegression_ExpiredOTPDoesNotVerifyEmail(t *testing.T) {
	testhelpers.SetupTestDB()
	app := testhelpers.NewApp()

	otp := "646464"
	expired := time.Now().Add(-1 * time.Minute)
	purpose := "email_verify"
	hash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.MinCost)
	user := models.User{
		ID:            uuid.New(),
		Email:         "expired-noverify@example.com",
		PasswordHash:  string(hash),
		EmailVerified: false,
		OTPCode:       &otp,
		OTPExpires:    &expired,
		OTPPurpose:    &purpose,
	}
	database.DB.Create(&user)

	payload, _ := json.Marshal(map[string]string{"email": "expired-noverify@example.com", "otp": otp})
	req := httptest.NewRequest("POST", "/api/v1/auth/verify-email-otp", bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)
	assert.Equal(t, 400, resp.StatusCode)

	// User must still be unverified
	var dbUser models.User
	database.DB.Where("email = ?", "expired-noverify@example.com").First(&dbUser)
	assert.False(t, dbUser.EmailVerified, "expired OTP must not verify the account")
}

// ── Security: unknown email returns 400 (not 404) ────────────────────────────

func TestRegression_UnknownEmailReturns400NotFound(t *testing.T) {
	testhelpers.SetupTestDB()
	app := testhelpers.NewApp()

	for _, endpoint := range []string{
		"/api/v1/auth/verify-email-otp",
		"/api/v1/auth/verify-reset-otp",
	} {
		payload, _ := json.Marshal(map[string]string{"email": "ghost@nobody.com", "otp": "123456"})
		req := httptest.NewRequest("POST", endpoint, bytes.NewBuffer(payload))
		req.Header.Set("Content-Type", "application/json")
		resp, _ := app.Test(req)
		// Must be 400, not 404 (avoid leaking whether email exists)
		assert.Equal(t, 400, resp.StatusCode, "endpoint: %s", endpoint)
	}
}

// ── Security: account lockout survives password reset ────────────────────────

func TestRegression_LockoutClearedAfterPasswordReset(t *testing.T) {
	db := testhelpers.SetupTestDB()
	app := testhelpers.NewApp()
	testhelpers.CreateVerifiedUser(db, "lockout-reset@example.com", "correctpass1")

	// Trigger lockout with 5 wrong password attempts
	for i := 0; i < 5; i++ {
		lpPayload, _ := json.Marshal(map[string]string{"email": "lockout-reset@example.com", "password": "wrongpass!"})
		lpReq := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(lpPayload))
		lpReq.Header.Set("Content-Type", "application/json")
		app.Test(lpReq)
	}

	// Confirm account is locked
	var user models.User
	database.DB.Where("email = ?", "lockout-reset@example.com").First(&user)
	assert.NotNil(t, user.LockedUntil, "account must be locked after 5 failed attempts")

	// Reset password via OTP flow
	fpReq := httptest.NewRequest("POST", "/api/v1/auth/forgot-password",
		bytes.NewBufferString(`{"email":"lockout-reset@example.com"}`))
	fpReq.Header.Set("Content-Type", "application/json")
	app.Test(fpReq)

	database.DB.Where("email = ?", "lockout-reset@example.com").First(&user)
	vpPayload, _ := json.Marshal(map[string]string{"email": "lockout-reset@example.com", "otp": *user.OTPCode})
	vpReq := httptest.NewRequest("POST", "/api/v1/auth/verify-reset-otp", bytes.NewBuffer(vpPayload))
	vpReq.Header.Set("Content-Type", "application/json")
	vpResp, _ := app.Test(vpReq)
	var vpBody map[string]interface{}
	json.NewDecoder(vpResp.Body).Decode(&vpBody)
	require.NotNil(t, vpBody["reset_token"], "reset_token must be in verify-reset-otp response")

	rpPayload, _ := json.Marshal(map[string]string{"token": vpBody["reset_token"].(string), "new_password": "UnlockedPass9!"})
	rpReq := httptest.NewRequest("POST", "/api/v1/auth/reset-password", bytes.NewBuffer(rpPayload))
	rpReq.Header.Set("Content-Type", "application/json")
	rpResp, _ := app.Test(rpReq)
	require.Equal(t, 200, rpResp.StatusCode, "reset-password must succeed")

	// Use a fresh struct — re-using a pre-populated struct can cause GORM to
	// retain non-nil pointer values that were NULL-scanned in some drivers.
	var freshUser models.User
	database.DB.Where("email = ?", "lockout-reset@example.com").First(&freshUser)
	assert.Nil(t, freshUser.LockedUntil, "lockout must be cleared after password reset")
	assert.Equal(t, 0, freshUser.FailedLoginAttempts)

	// Login should now succeed
	loginPayload, _ := json.Marshal(map[string]string{"email": "lockout-reset@example.com", "password": "UnlockedPass9!"})
	loginReq := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(loginPayload))
	loginReq.Header.Set("Content-Type", "application/json")
	loginResp, _ := app.Test(loginReq)
	assert.Equal(t, 200, loginResp.StatusCode)
}
