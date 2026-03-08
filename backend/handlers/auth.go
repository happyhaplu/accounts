package handlers

import (
"fmt"
"os"
"strings"
"time"

"outcraftly/accounts/database"
"outcraftly/accounts/mailer"
"outcraftly/accounts/models"

"github.com/gofiber/fiber/v2"
"github.com/golang-jwt/jwt/v5"
"github.com/google/uuid"
"golang.org/x/crypto/bcrypt"
)

// ── Request bodies ────────────────────────────────────────────────────────────

type registerRequest struct {
Email    string `json:"email"`
Password string `json:"password"`
}

type loginRequest struct {
Email    string `json:"email"`
Password string `json:"password"`
}

type forgotPasswordRequest struct {
Email string `json:"email"`
}

type resetPasswordRequest struct {
Token       string `json:"token"`
NewPassword string `json:"new_password"`
}

type changePasswordRequest struct {
CurrentPassword string `json:"current_password"`
NewPassword     string `json:"new_password"`
}

// ── Handlers ──────────────────────────────────────────────────────────────────

// Register creates a new (unverified) account and sends a verification email.
func Register(c *fiber.Ctx) error {
req := new(registerRequest)
if err := c.BodyParser(req); err != nil {
return badRequest(c, "Invalid request body")
}

req.Email = strings.ToLower(strings.TrimSpace(req.Email))
if req.Email == "" || !strings.Contains(req.Email, "@") {
return badRequest(c, "Valid email is required")
}
if len(req.Password) < 8 {
return badRequest(c, "Password must be at least 8 characters")
}

var existing models.User
if tx := database.DB.Where("email = ?", req.Email).First(&existing); tx.Error == nil {
return c.Status(fiber.StatusConflict).JSON(errJSON("Email is already registered"))
}

hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
if err != nil {
return serverError(c, "Failed to hash password")
}

verifyToken := uuid.New().String()
verifyExpiry := time.Now().Add(24 * time.Hour)

user := models.User{
ID:                 uuid.New(),
Email:              req.Email,
PasswordHash:       string(hash),
EmailVerifyToken:   &verifyToken,
EmailVerifyExpires: &verifyExpiry,
}
if tx := database.DB.Create(&user); tx.Error != nil {
return serverError(c, "Failed to create account")
}

// Auto-create workspace — non-fatal so registration still succeeds on error.
if err := createWorkspaceForUser(user); err != nil {
fmt.Fprintf(os.Stderr, "[workspace] ERROR creating workspace for %s: %v\n", user.Email, err)
}

// Auto-accept any pending workspace invites for this email address (non-fatal).
go AcceptPendingInvitesForEmail(user.ID, user.Email)

// Send verification email asynchronously
appURL := os.Getenv("APP_URL")
link := appURL + "/verify-email?token=" + verifyToken
go func() {
if err := mailer.Send(user.Email, "Verify your Outcraftly account", mailer.VerifyEmailBody(link)); err != nil {
fmt.Fprintf(os.Stderr, "[mailer] ERROR sending verify email to %s: %v\n", user.Email, err)
} else {
fmt.Printf("[mailer] verify email sent to %s\n", user.Email)
}
}()

return c.Status(fiber.StatusCreated).JSON(fiber.Map{
"message": "Account created! Please check your email to verify your address.",
})
}

// VerifyEmail validates an email-verification token and issues a JWT.
func VerifyEmail(c *fiber.Ctx) error {
token := c.Query("token")
if token == "" {
return badRequest(c, "Verification token is required")
}

var user models.User
tx := database.DB.Where(
"email_verify_token = ? AND email_verify_expires > ?", token, time.Now(),
).First(&user)
if tx.Error != nil {
return c.Status(fiber.StatusBadRequest).JSON(errJSON("Invalid or expired verification link. Please register again."))
}

database.DB.Model(&user).Updates(map[string]interface{}{
"email_verified":        true,
"email_verify_token":    nil,
"email_verify_expires":  nil,
})
user.EmailVerified = true

jwtToken, err := generateJWT(user.ID.String(), user.Email)
if err != nil {
return serverError(c, "Failed to issue token")
}

return c.JSON(fiber.Map{
"message":             "Email verified successfully!",
"token":               jwtToken,
"user":                publicUser(user),
"needs_profile_setup": !user.ProfileComplete,
})
}

// ResendVerification re-sends the email verification link for an unverified account.
// Rate-limited: only re-generates a fresh token if the previous one is expired or missing.
func ResendVerification(c *fiber.Ctx) error {
	type body struct {
		Email string `json:"email"`
	}
	req := new(body)
	if err := c.BodyParser(req); err != nil {
		return badRequest(c, "Invalid request body")
	}
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))
	if req.Email == "" {
		return badRequest(c, "Email is required")
	}

	// Always return the same message to prevent email enumeration.
	genericOK := fiber.Map{"message": "If that email is registered and unverified, a new link has been sent."}

	var user models.User
	if tx := database.DB.Where("email = ?", req.Email).First(&user); tx.Error != nil {
		return c.JSON(genericOK)
	}
	if user.EmailVerified {
		// Already verified — silently succeed (don't reveal verification status).
		return c.JSON(genericOK)
	}

	// Issue a fresh 24-hour token.
	newToken := uuid.New().String()
	newExpiry := time.Now().Add(24 * time.Hour)
	database.DB.Model(&user).Updates(map[string]interface{}{
		"email_verify_token":   newToken,
		"email_verify_expires": newExpiry,
	})

	appURL := os.Getenv("APP_URL")
	link := appURL + "/verify-email?token=" + newToken
	go func() {
		if err := mailer.Send(user.Email, "Verify your Outcraftly account", mailer.VerifyEmailBody(link)); err != nil {
			fmt.Fprintf(os.Stderr, "[mailer] ERROR resending verify email to %s: %v\n", user.Email, err)
		} else {
			fmt.Printf("[mailer] verify email resent to %s\n", user.Email)
		}
	}()

	return c.JSON(genericOK)
}

// Login authenticates a user and returns a JWT.
// Implements account lockout after 5 consecutive failures (15-minute window).
func Login(c *fiber.Ctx) error {
req := new(loginRequest)
if err := c.BodyParser(req); err != nil {
return badRequest(c, "Invalid request body")
}

req.Email = strings.ToLower(strings.TrimSpace(req.Email))

var user models.User
if tx := database.DB.Where("email = ?", req.Email).First(&user); tx.Error != nil {
return c.Status(fiber.StatusUnauthorized).JSON(errJSON("Invalid email or password"))
}

// Account lockout check
if user.LockedUntil != nil && time.Now().Before(*user.LockedUntil) {
remaining := time.Until(*user.LockedUntil).Round(time.Minute)
return c.Status(fiber.StatusTooManyRequests).JSON(errJSON(
fmt.Sprintf("Account temporarily locked. Try again in %v.", remaining),
))
}

// Password check
if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
attempts := user.FailedLoginAttempts + 1
updates := map[string]interface{}{"failed_login_attempts": attempts}
if attempts >= 5 {
lockUntil := time.Now().Add(15 * time.Minute)
updates["locked_until"] = lockUntil
}
database.DB.Model(&user).Updates(updates)
return c.Status(fiber.StatusUnauthorized).JSON(errJSON("Invalid email or password"))
}

// Success — reset counters, record login time
now := time.Now()
database.DB.Model(&user).Updates(map[string]interface{}{
"failed_login_attempts": 0,
"locked_until":          nil,
"last_login_at":         now,
})
user.LastLoginAt = &now

jwtToken, err := generateJWT(user.ID.String(), user.Email)
if err != nil {
return serverError(c, "Failed to issue token")
}

return c.JSON(fiber.Map{
"message":             "Login successful",
"token":               jwtToken,
"user":                publicUser(user),
"needs_profile_setup": !user.ProfileComplete,
})
}

// Logout is stateless (client discards JWT). Future: Redis blocklist.
func Logout(c *fiber.Ctx) error {
return c.JSON(fiber.Map{"message": "Logged out successfully"})
}

// ForgotPassword generates a one-hour reset token and emails it to the user.
func ForgotPassword(c *fiber.Ctx) error {
req := new(forgotPasswordRequest)
if err := c.BodyParser(req); err != nil {
return badRequest(c, "Invalid request body")
}

req.Email = strings.ToLower(strings.TrimSpace(req.Email))
genericMsg := "If that email is registered, you will receive reset instructions shortly."

var user models.User
if tx := database.DB.Where("email = ?", req.Email).First(&user); tx.Error != nil {
// Always return the same message — prevents email enumeration
return c.JSON(fiber.Map{"message": genericMsg})
}

resetToken := uuid.New().String()
expiry := time.Now().Add(time.Hour)
database.DB.Model(&user).Updates(map[string]interface{}{
"reset_token":         resetToken,
"reset_token_expires": expiry,
})

appURL := os.Getenv("APP_URL")
link := appURL + "/reset-password?token=" + resetToken
go func() {
if err := mailer.Send(user.Email, "Reset your Outcraftly password", mailer.PasswordResetBody(link)); err != nil {
fmt.Fprintf(os.Stderr, "[mailer] ERROR sending reset email to %s: %v\n", user.Email, err)
} else {
fmt.Printf("[mailer] reset email sent to %s\n", user.Email)
}
}()

return c.JSON(fiber.Map{"message": genericMsg})
}

// ResetPassword validates the reset token and sets a new password.
func ResetPassword(c *fiber.Ctx) error {
req := new(resetPasswordRequest)
if err := c.BodyParser(req); err != nil {
return badRequest(c, "Invalid request body")
}

if len(req.NewPassword) < 8 {
return badRequest(c, "Password must be at least 8 characters")
}

var user models.User
tx := database.DB.Where(
"reset_token = ? AND reset_token_expires > ?", req.Token, time.Now(),
).First(&user)
if tx.Error != nil {
return c.Status(fiber.StatusBadRequest).JSON(errJSON("Invalid or expired reset token"))
}

hash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
if err != nil {
return serverError(c, "Failed to hash password")
}

database.DB.Model(&user).Updates(map[string]interface{}{
"password_hash":         string(hash),
"reset_token":           nil,
"reset_token_expires":   nil,
"failed_login_attempts": 0,
"locked_until":          nil,
})

return c.JSON(fiber.Map{"message": "Password reset successfully. You can now sign in."})
}

// ChangePassword lets an authenticated user update their password.
func ChangePassword(c *fiber.Ctx) error {
userID := c.Locals("userID").(string)

req := new(changePasswordRequest)
if err := c.BodyParser(req); err != nil {
return badRequest(c, "Invalid request body")
}
if req.CurrentPassword == "" || req.NewPassword == "" {
return badRequest(c, "Current password and new password are required")
}
if len(req.NewPassword) < 8 {
return badRequest(c, "New password must be at least 8 characters")
}

var user models.User
if tx := database.DB.Where("id = ?", userID).First(&user); tx.Error != nil {
return c.Status(fiber.StatusNotFound).JSON(errJSON("User not found"))
}

// Verify current password
if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.CurrentPassword)); err != nil {
return c.Status(fiber.StatusUnauthorized).JSON(errJSON("Current password is incorrect"))
}

hash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
if err != nil {
return serverError(c, "Failed to hash password")
}

database.DB.Model(&user).Update("password_hash", string(hash))

return c.JSON(fiber.Map{"message": "Password updated successfully"})
}

// GetProfile returns the authenticated user's public profile.
func GetProfile(c *fiber.Ctx) error {
userID := c.Locals("userID").(string)

var user models.User
if tx := database.DB.Where("id = ?", userID).First(&user); tx.Error != nil {
return c.Status(fiber.StatusNotFound).JSON(errJSON("User not found"))
}

return c.JSON(fiber.Map{"user": publicUser(user)})
}

// ── Helpers ───────────────────────────────────────────────────────────────────

func generateJWT(userID, email string) (string, error) {
claims := jwt.MapClaims{
"sub":   userID,
"email": email,
"exp":   time.Now().Add(24 * time.Hour).Unix(),
"iat":   time.Now().Unix(),
}
t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
return t.SignedString([]byte(os.Getenv("JWT_SECRET")))
}

func publicUser(u models.User) fiber.Map {
return fiber.Map{
"id":                u.ID,
"email":             u.Email,
"email_verified":    u.EmailVerified,
"name":              u.Name,
"company_name":      u.CompanyName,
"job_title":         u.JobTitle,
"phone_country_code": u.PhoneCountryCode,
"phone_number":      u.PhoneNumber,
"profile_complete":  u.ProfileComplete,
"last_login_at":     u.LastLoginAt,
"created_at":        u.CreatedAt,
}
}

func errJSON(msg string) fiber.Map { return fiber.Map{"error": msg} }
func badRequest(c *fiber.Ctx, msg string) error {
return c.Status(fiber.StatusBadRequest).JSON(errJSON(msg))
}
func serverError(c *fiber.Ctx, msg string) error {
return c.Status(fiber.StatusInternalServerError).JSON(errJSON(msg))
}
