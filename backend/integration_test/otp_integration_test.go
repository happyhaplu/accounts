package integration_test

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

// ── Helpers ───────────────────────────────────────────────────────────────────

// insertUnverifiedWithOTP inserts an unverified user with a known OTP directly into the DB.
func insertUnverifiedWithOTP(email, otp, purpose string, expires time.Time) models.User {
	hash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.MinCost)
	user := models.User{
		ID:            uuid.New(),
		Email:         email,
		PasswordHash:  string(hash),
		EmailVerified: false,
		OTPCode:       &otp,
		OTPExpires:    &expires,
		OTPPurpose:    &purpose,
	}
	database.DB.Create(&user)
	return user
}

// ── VerifyEmailOTP ────────────────────────────────────────────────────────────

func TestVerifyEmailOTP_Success(t *testing.T) {
	testhelpers.SetupTestDB()
	app := testhelpers.NewApp()

	otp := "123456"
	expires := time.Now().Add(10 * time.Minute)
	insertUnverifiedWithOTP("otp-verify@example.com", otp, "email_verify", expires)

	payload, _ := json.Marshal(map[string]string{"email": "otp-verify@example.com", "otp": otp})
	req := httptest.NewRequest("POST", "/api/v1/auth/verify-email-otp", bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var body map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&body)
	assert.NotEmpty(t, body["token"])
	user := body["user"].(map[string]interface{})
	assert.Equal(t, true, user["email_verified"])
}

func TestVerifyEmailOTP_WrongOTP(t *testing.T) {
	testhelpers.SetupTestDB()
	app := testhelpers.NewApp()

	expires := time.Now().Add(10 * time.Minute)
	insertUnverifiedWithOTP("otp-wrong@example.com", "999999", "email_verify", expires)

	payload, _ := json.Marshal(map[string]string{"email": "otp-wrong@example.com", "otp": "000000"})
	req := httptest.NewRequest("POST", "/api/v1/auth/verify-email-otp", bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")

	resp, _ := app.Test(req)
	assert.Equal(t, 400, resp.StatusCode)
}

func TestVerifyEmailOTP_ExpiredOTP(t *testing.T) {
	testhelpers.SetupTestDB()
	app := testhelpers.NewApp()

	otp := "654321"
	expired := time.Now().Add(-1 * time.Minute) // expired 1 min ago
	insertUnverifiedWithOTP("otp-expired@example.com", otp, "email_verify", expired)

	payload, _ := json.Marshal(map[string]string{"email": "otp-expired@example.com", "otp": otp})
	req := httptest.NewRequest("POST", "/api/v1/auth/verify-email-otp", bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")

	resp, _ := app.Test(req)
	assert.Equal(t, 400, resp.StatusCode)

	var body map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&body)
	// Handler returns specific expired message
	assert.Contains(t, body["error"].(string), "expired")
}

func TestVerifyEmailOTP_WrongPurpose(t *testing.T) {
	// A password_reset OTP must NOT be accepted for email_verify
	testhelpers.SetupTestDB()
	app := testhelpers.NewApp()

	otp := "112233"
	expires := time.Now().Add(10 * time.Minute)
	insertUnverifiedWithOTP("otp-purpose@example.com", otp, "password_reset", expires) // wrong purpose

	payload, _ := json.Marshal(map[string]string{"email": "otp-purpose@example.com", "otp": otp})
	req := httptest.NewRequest("POST", "/api/v1/auth/verify-email-otp", bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")

	resp, _ := app.Test(req)
	assert.Equal(t, 400, resp.StatusCode)
}

func TestVerifyEmailOTP_AlreadyVerified_IsIdempotent(t *testing.T) {
	// A user who is already verified should receive a fresh JWT (idempotent)
	db := testhelpers.SetupTestDB()
	app := testhelpers.NewApp()
	_, _ = testhelpers.CreateVerifiedUser(db, "otp-idempotent@example.com", "password123")

	// Any OTP (doesn't matter — already verified path short-circuits)
	payload, _ := json.Marshal(map[string]string{"email": "otp-idempotent@example.com", "otp": "000000"})
	req := httptest.NewRequest("POST", "/api/v1/auth/verify-email-otp", bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var body map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&body)
	assert.NotEmpty(t, body["token"])
}

func TestVerifyEmailOTP_MissingFields(t *testing.T) {
	testhelpers.SetupTestDB()
	app := testhelpers.NewApp()

	cases := []map[string]string{
		{"email": "", "otp": "123456"},
		{"email": "x@y.com", "otp": ""},
		{},
	}
	for _, c := range cases {
		payload, _ := json.Marshal(c)
		req := httptest.NewRequest("POST", "/api/v1/auth/verify-email-otp", bytes.NewBuffer(payload))
		req.Header.Set("Content-Type", "application/json")
		resp, _ := app.Test(req)
		assert.Equal(t, 400, resp.StatusCode)
	}
}

// ── ResendVerification ────────────────────────────────────────────────────────

func TestResendVerification_UnverifiedUser(t *testing.T) {
	testhelpers.SetupTestDB()
	app := testhelpers.NewApp()

	otp := "555555"
	expires := time.Now().Add(10 * time.Minute)
	insertUnverifiedWithOTP("resend@example.com", otp, "email_verify", expires)

	payload := `{"email":"resend@example.com"}`
	req := httptest.NewRequest("POST", "/api/v1/auth/resend-verification", bytes.NewBufferString(payload))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var body map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&body)
	assert.Contains(t, body["message"].(string), "code")
}

func TestResendVerification_AlreadyVerified_SameResponse(t *testing.T) {
	// Must return same 200 to prevent enumeration
	db := testhelpers.SetupTestDB()
	app := testhelpers.NewApp()
	testhelpers.CreateVerifiedUser(db, "resend-verified@example.com", "password123")

	payload := `{"email":"resend-verified@example.com"}`
	req := httptest.NewRequest("POST", "/api/v1/auth/resend-verification", bytes.NewBufferString(payload))
	req.Header.Set("Content-Type", "application/json")

	resp, _ := app.Test(req)
	assert.Equal(t, 200, resp.StatusCode)
}

func TestResendVerification_UnknownEmail_SameResponse(t *testing.T) {
	testhelpers.SetupTestDB()
	app := testhelpers.NewApp()

	payload := `{"email":"nobody@example.com"}`
	req := httptest.NewRequest("POST", "/api/v1/auth/resend-verification", bytes.NewBufferString(payload))
	req.Header.Set("Content-Type", "application/json")

	resp, _ := app.Test(req)
	assert.Equal(t, 200, resp.StatusCode) // anti-enumeration: always 200
}

func TestResendVerification_FreshOTPOverwritesOldOTP(t *testing.T) {
	// After resend, the old OTP is gone — old OTP must no longer work
	testhelpers.SetupTestDB()
	app := testhelpers.NewApp()

	oldOTP := "111111"
	expires := time.Now().Add(10 * time.Minute)
	insertUnverifiedWithOTP("resend-overwrite@example.com", oldOTP, "email_verify", expires)

	// Resend (generates new OTP in DB)
	resendPayload := `{"email":"resend-overwrite@example.com"}`
	resendReq := httptest.NewRequest("POST", "/api/v1/auth/resend-verification", bytes.NewBufferString(resendPayload))
	resendReq.Header.Set("Content-Type", "application/json")
	app.Test(resendReq)

	// Try old OTP — must fail (new OTP replaced it)
	payload, _ := json.Marshal(map[string]string{"email": "resend-overwrite@example.com", "otp": oldOTP})
	req := httptest.NewRequest("POST", "/api/v1/auth/verify-email-otp", bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)
	assert.Equal(t, 400, resp.StatusCode)
}

// ── VerifyResetOTP ────────────────────────────────────────────────────────────

func TestVerifyResetOTP_Success(t *testing.T) {
	db := testhelpers.SetupTestDB()
	app := testhelpers.NewApp()
	testhelpers.CreateVerifiedUser(db, "reset-otp@example.com", "password123")

	// Trigger forgot-password to store OTP in DB
	fpPayload := `{"email":"reset-otp@example.com"}`
	fpReq := httptest.NewRequest("POST", "/api/v1/auth/forgot-password", bytes.NewBufferString(fpPayload))
	fpReq.Header.Set("Content-Type", "application/json")
	app.Test(fpReq)

	// Read the OTP straight from DB (bypass email)
	var user models.User
	database.DB.Where("email = ?", "reset-otp@example.com").First(&user)
	require.NotNil(t, user.OTPCode)

	payload, _ := json.Marshal(map[string]string{"email": "reset-otp@example.com", "otp": *user.OTPCode})
	req := httptest.NewRequest("POST", "/api/v1/auth/verify-reset-otp", bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var body map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&body)
	assert.NotEmpty(t, body["reset_token"])
}

func TestVerifyResetOTP_WrongOTP(t *testing.T) {
	db := testhelpers.SetupTestDB()
	app := testhelpers.NewApp()
	testhelpers.CreateVerifiedUser(db, "reset-wrong@example.com", "password123")

	fpPayload := `{"email":"reset-wrong@example.com"}`
	fpReq := httptest.NewRequest("POST", "/api/v1/auth/forgot-password", bytes.NewBufferString(fpPayload))
	fpReq.Header.Set("Content-Type", "application/json")
	app.Test(fpReq)

	payload, _ := json.Marshal(map[string]string{"email": "reset-wrong@example.com", "otp": "000000"})
	req := httptest.NewRequest("POST", "/api/v1/auth/verify-reset-otp", bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")

	resp, _ := app.Test(req)
	assert.Equal(t, 400, resp.StatusCode)
}

func TestVerifyResetOTP_ExpiredOTP(t *testing.T) {
	testhelpers.SetupTestDB()
	app := testhelpers.NewApp()

	otp := "888888"
	expired := time.Now().Add(-2 * time.Minute)
	// Insert a verified user with an expired reset OTP manually
	hash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.MinCost)
	purpose := "password_reset"
	user := models.User{
		ID:            uuid.New(),
		Email:         "reset-expired@example.com",
		PasswordHash:  string(hash),
		EmailVerified: true,
		OTPCode:       &otp,
		OTPExpires:    &expired,
		OTPPurpose:    &purpose,
	}
	database.DB.Create(&user)

	payload, _ := json.Marshal(map[string]string{"email": "reset-expired@example.com", "otp": otp})
	req := httptest.NewRequest("POST", "/api/v1/auth/verify-reset-otp", bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")

	resp, _ := app.Test(req)
	assert.Equal(t, 400, resp.StatusCode)

	var body map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&body)
	assert.Contains(t, body["error"].(string), "expired")
}

func TestVerifyResetOTP_WrongPurpose(t *testing.T) {
	// email_verify OTP must NOT work for password reset
	testhelpers.SetupTestDB()
	app := testhelpers.NewApp()

	otp := "777777"
	expires := time.Now().Add(10 * time.Minute)
	purpose := "email_verify" // wrong purpose
	hash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.MinCost)
	user := models.User{
		ID:            uuid.New(),
		Email:         "reset-purpose@example.com",
		PasswordHash:  string(hash),
		EmailVerified: true,
		OTPCode:       &otp,
		OTPExpires:    &expires,
		OTPPurpose:    &purpose,
	}
	database.DB.Create(&user)

	payload, _ := json.Marshal(map[string]string{"email": "reset-purpose@example.com", "otp": otp})
	req := httptest.NewRequest("POST", "/api/v1/auth/verify-reset-otp", bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")

	resp, _ := app.Test(req)
	assert.Equal(t, 400, resp.StatusCode)
}

// ── ResetPassword ─────────────────────────────────────────────────────────────

func TestResetPassword_Success(t *testing.T) {
	db := testhelpers.SetupTestDB()
	app := testhelpers.NewApp()
	testhelpers.CreateVerifiedUser(db, "reset-pw@example.com", "oldpassword1")

	// Forgot → verify OTP → get reset_token → reset
	fpReq := httptest.NewRequest("POST", "/api/v1/auth/forgot-password",
		bytes.NewBufferString(`{"email":"reset-pw@example.com"}`))
	fpReq.Header.Set("Content-Type", "application/json")
	app.Test(fpReq)

	var user models.User
	database.DB.Where("email = ?", "reset-pw@example.com").First(&user)
	require.NotNil(t, user.OTPCode)

	// Verify OTP
	vpPayload, _ := json.Marshal(map[string]string{"email": "reset-pw@example.com", "otp": *user.OTPCode})
	vpReq := httptest.NewRequest("POST", "/api/v1/auth/verify-reset-otp", bytes.NewBuffer(vpPayload))
	vpReq.Header.Set("Content-Type", "application/json")
	vpResp, _ := app.Test(vpReq)
	var vpBody map[string]interface{}
	json.NewDecoder(vpResp.Body).Decode(&vpBody)
	resetToken := vpBody["reset_token"].(string)

	// Reset password
	rpPayload, _ := json.Marshal(map[string]string{"token": resetToken, "new_password": "NewSecure123!"})
	rpReq := httptest.NewRequest("POST", "/api/v1/auth/reset-password", bytes.NewBuffer(rpPayload))
	rpReq.Header.Set("Content-Type", "application/json")
	rpResp, err := app.Test(rpReq)
	require.NoError(t, err)
	assert.Equal(t, 200, rpResp.StatusCode)

	var rpBody map[string]interface{}
	json.NewDecoder(rpResp.Body).Decode(&rpBody)
	assert.Contains(t, rpBody["message"].(string), "Password reset")
}

func TestResetPassword_WrongToken(t *testing.T) {
	testhelpers.SetupTestDB()
	app := testhelpers.NewApp()

	payload := `{"token":"invalid-token-xyz","new_password":"NewPass12345!"}`
	req := httptest.NewRequest("POST", "/api/v1/auth/reset-password", bytes.NewBufferString(payload))
	req.Header.Set("Content-Type", "application/json")

	resp, _ := app.Test(req)
	assert.Equal(t, 400, resp.StatusCode)
}

func TestResetPassword_ExpiredToken(t *testing.T) {
	testhelpers.SetupTestDB()
	app := testhelpers.NewApp()

	// Insert user with already-expired reset token
	hash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.MinCost)
	tok := "expired-reset-token-abc"
	pastExpiry := time.Now().Add(-5 * time.Minute)
	user := models.User{
		ID:                uuid.New(),
		Email:             "reset-tok-expired@example.com",
		PasswordHash:      string(hash),
		EmailVerified:     true,
		ResetToken:        &tok,
		ResetTokenExpires: &pastExpiry,
	}
	database.DB.Create(&user)

	payload, _ := json.Marshal(map[string]string{"token": tok, "new_password": "NewPass12345!"})
	req := httptest.NewRequest("POST", "/api/v1/auth/reset-password", bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")

	resp, _ := app.Test(req)
	assert.Equal(t, 400, resp.StatusCode)
}

func TestResetPassword_TokenClearedAfterUse(t *testing.T) {
	// After successful reset, the token must be gone (can't reuse)
	db := testhelpers.SetupTestDB()
	app := testhelpers.NewApp()
	testhelpers.CreateVerifiedUser(db, "reset-onceonly@example.com", "oldpassword1")

	fpReq := httptest.NewRequest("POST", "/api/v1/auth/forgot-password",
		bytes.NewBufferString(`{"email":"reset-onceonly@example.com"}`))
	fpReq.Header.Set("Content-Type", "application/json")
	app.Test(fpReq)

	var user models.User
	database.DB.Where("email = ?", "reset-onceonly@example.com").First(&user)
	require.NotNil(t, user.OTPCode)

	vpPayload, _ := json.Marshal(map[string]string{"email": "reset-onceonly@example.com", "otp": *user.OTPCode})
	vpReq := httptest.NewRequest("POST", "/api/v1/auth/verify-reset-otp", bytes.NewBuffer(vpPayload))
	vpReq.Header.Set("Content-Type", "application/json")
	vpResp, _ := app.Test(vpReq)
	var vpBody map[string]interface{}
	json.NewDecoder(vpResp.Body).Decode(&vpBody)
	resetToken := vpBody["reset_token"].(string)

	// First use — should succeed
	rp1, _ := json.Marshal(map[string]string{"token": resetToken, "new_password": "NewPass111!"})
	r1 := httptest.NewRequest("POST", "/api/v1/auth/reset-password", bytes.NewBuffer(rp1))
	r1.Header.Set("Content-Type", "application/json")
	resp1, _ := app.Test(r1)
	assert.Equal(t, 200, resp1.StatusCode)

	// Second use with same token — must fail
	rp2, _ := json.Marshal(map[string]string{"token": resetToken, "new_password": "NewPass222!"})
	r2 := httptest.NewRequest("POST", "/api/v1/auth/reset-password", bytes.NewBuffer(rp2))
	r2.Header.Set("Content-Type", "application/json")
	resp2, _ := app.Test(r2)
	assert.Equal(t, 400, resp2.StatusCode)
}

func TestResetPassword_ShortNewPassword(t *testing.T) {
	testhelpers.SetupTestDB()
	app := testhelpers.NewApp()

	payload := `{"token":"any-token","new_password":"short"}`
	req := httptest.NewRequest("POST", "/api/v1/auth/reset-password", bytes.NewBufferString(payload))
	req.Header.Set("Content-Type", "application/json")

	resp, _ := app.Test(req)
	assert.Equal(t, 400, resp.StatusCode)
}

func TestResetPassword_CanLoginWithNewPassword(t *testing.T) {
	// Full flow: reset password → login with new password succeeds
	db := testhelpers.SetupTestDB()
	app := testhelpers.NewApp()
	testhelpers.CreateVerifiedUser(db, "reset-login@example.com", "oldpassword1")

	// Forgot → verify OTP → reset
	fpReq := httptest.NewRequest("POST", "/api/v1/auth/forgot-password",
		bytes.NewBufferString(`{"email":"reset-login@example.com"}`))
	fpReq.Header.Set("Content-Type", "application/json")
	app.Test(fpReq)

	var user models.User
	database.DB.Where("email = ?", "reset-login@example.com").First(&user)
	vpPayload, _ := json.Marshal(map[string]string{"email": "reset-login@example.com", "otp": *user.OTPCode})
	vpReq := httptest.NewRequest("POST", "/api/v1/auth/verify-reset-otp", bytes.NewBuffer(vpPayload))
	vpReq.Header.Set("Content-Type", "application/json")
	vpResp, _ := app.Test(vpReq)
	var vpBody map[string]interface{}
	json.NewDecoder(vpResp.Body).Decode(&vpBody)

	rpPayload, _ := json.Marshal(map[string]string{"token": vpBody["reset_token"].(string), "new_password": "UpdatedPass99!"})
	rpReq := httptest.NewRequest("POST", "/api/v1/auth/reset-password", bytes.NewBuffer(rpPayload))
	rpReq.Header.Set("Content-Type", "application/json")
	app.Test(rpReq)

	// Login with new password
	loginPayload, _ := json.Marshal(map[string]string{"email": "reset-login@example.com", "password": "UpdatedPass99!"})
	loginReq := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(loginPayload))
	loginReq.Header.Set("Content-Type", "application/json")
	loginResp, err := app.Test(loginReq)
	require.NoError(t, err)
	assert.Equal(t, 200, loginResp.StatusCode)

	var loginBody map[string]interface{}
	json.NewDecoder(loginResp.Body).Decode(&loginBody)
	assert.NotEmpty(t, loginBody["token"])
}
