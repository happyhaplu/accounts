package handlers

import (
	"crypto/rand"
	"fmt"
	"math/big"
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
	"gorm.io/gorm"
)

// ── Request bodies ────────────────────────────────────────────────────────────

type registerRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginRequest struct {
	Email       string `json:"email"`
	Password    string `json:"password"`
	RedirectURI string `json:"redirect_uri"`
}

type otpRequest struct {
	Email string `json:"email"`
	OTP   string `json:"otp"`
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

// ── Helpers ───────────────────────────────────────────────────────────────────

// generateOTP returns a cryptographically-secure 6-digit numeric code.
func generateOTP() string {
	n, _ := rand.Int(rand.Reader, big.NewInt(1000000))
	return fmt.Sprintf("%06d", n.Int64())
}

// storeOTP persists a fresh OTP (and purpose) on the user row and returns it.
func storeOTP(user *models.User, purpose string) string {
	otp := generateOTP()
	expiry := time.Now().Add(10 * time.Minute)
	database.DB.Model(user).Updates(map[string]interface{}{
		"otp_code":    otp,
		"otp_expires": expiry,
		"otp_purpose": purpose,
	})
	return otp
}

// ── Handlers ──────────────────────────────────────────────────────────────────

// Register creates a new (unverified) account and sends an OTP.
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

	otp := generateOTP()
	otpExpiry := time.Now().Add(10 * time.Minute)
	purpose := "email_verify"

	user := models.User{
		ID:           uuid.New(),
		Email:        req.Email,
		PasswordHash: string(hash),
		OTPCode:      &otp,
		OTPExpires:   &otpExpiry,
		OTPPurpose:   &purpose,
	}
	if tx := database.DB.Create(&user); tx.Error != nil {
		return serverError(c, "Failed to create account")
	}

	// Auto-create workspace — non-fatal.
	if err := createWorkspaceForUser(user); err != nil {
		fmt.Fprintf(os.Stderr, "[workspace] ERROR creating workspace for %s: %v\n", user.Email, err)
	}

	// Auto-accept any pending workspace invites for this email address (non-fatal).
	go AcceptPendingInvitesForEmail(user.ID, user.Email)

	// Send OTP email asynchronously.
	go func() {
		if err := mailer.Send(user.Email, "Your Outcraftly verification code", mailer.OTPBody(otp, "email_verify")); err != nil {
			fmt.Fprintf(os.Stderr, "[mailer] ERROR sending OTP to %s: %v\n", user.Email, err)
		} else {
			fmt.Printf("[mailer] OTP sent to %s\n", user.Email)
		}
	}()

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Account created! Enter the 6-digit code sent to your email.",
	})
}

// VerifyEmailOTP validates the 6-digit OTP and issues a JWT.
func VerifyEmailOTP(c *fiber.Ctx) error {
	req := new(otpRequest)
	if err := c.BodyParser(req); err != nil {
		return badRequest(c, "Invalid request body")
	}
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))
	req.OTP = strings.TrimSpace(req.OTP)
	if req.Email == "" || req.OTP == "" {
		return badRequest(c, "Email and code are required")
	}

	var user models.User
	if tx := database.DB.Where("email = ?", req.Email).First(&user); tx.Error != nil {
		return badRequest(c, "Invalid or expired code")
	}

	// Already verified — issue token directly (idempotent).
	if user.EmailVerified {
		jwtToken, err := generateJWT(user.ID.String(), user.Email)
		if err != nil {
			return serverError(c, "Failed to issue token")
		}
		return c.JSON(fiber.Map{
			"message":             "Email already verified",
			"token":               jwtToken,
			"user":                publicUser(user),
			"needs_profile_setup": !user.ProfileComplete,
		})
	}

	if user.OTPCode == nil || *user.OTPCode != req.OTP {
		return badRequest(c, "Invalid or expired code")
	}
	if user.OTPExpires == nil || time.Now().After(*user.OTPExpires) {
		return badRequest(c, "Code has expired — please request a new one")
	}
	if user.OTPPurpose == nil || *user.OTPPurpose != "email_verify" {
		return badRequest(c, "Invalid code")
	}

	database.DB.Model(&user).Updates(map[string]interface{}{
		"email_verified": true,
		"otp_code":       nil,
		"otp_expires":    nil,
		"otp_purpose":    nil,
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

// ResendVerification sends a fresh OTP to the given email.
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
	genericOK := fiber.Map{"message": "If that email is registered and unverified, a new code has been sent."}

	var user models.User
	if tx := database.DB.Where("email = ?", req.Email).First(&user); tx.Error != nil {
		return c.JSON(genericOK)
	}
	if user.EmailVerified {
		return c.JSON(genericOK)
	}

	otp := storeOTP(&user, "email_verify")
	go func() {
		if err := mailer.Send(user.Email, "Your Outcraftly verification code", mailer.OTPBody(otp, "email_verify")); err != nil {
			fmt.Fprintf(os.Stderr, "[mailer] ERROR resending OTP to %s: %v\n", user.Email, err)
		} else {
			fmt.Printf("[mailer] OTP resent to %s\n", user.Email)
		}
	}()

	return c.JSON(genericOK)
}

// Login authenticates a user and returns a JWT.
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

	// If a redirect_uri was supplied, sign a 7-day launch token and return the
	// full redirect URL so the frontend can navigate directly to the target app.
	if req.RedirectURI != "" {
		var member models.WorkspaceMember
		if database.DB.Where("user_id = ? AND role = 'owner'", user.ID).
			Order("joined_at ASC").First(&member).Error == nil {

			launchToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
				"sub":          user.ID.String(),
				"email":        user.Email,
				"workspace_id": member.WorkspaceID.String(),
				"exp":          time.Now().Add(7 * 24 * time.Hour).Unix(),
			})
			if signed, err := launchToken.SignedString([]byte(os.Getenv("JWT_SECRET"))); err == nil {
				return c.JSON(fiber.Map{
					"message":      "Login successful",
					"token":        jwtToken,
					"user":         publicUser(user),
					"redirect_url": req.RedirectURI + "?token=" + signed,
				})
			}
		}
		// Workspace not found or token sign failed — fall through to normal response.
	}

	return c.JSON(fiber.Map{
		"message":             "Login successful",
		"token":               jwtToken,
		"user":                publicUser(user),
		"needs_profile_setup": !user.ProfileComplete,
	})
}

// Logout is stateless (client discards JWT).
func Logout(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"message": "Logged out successfully"})
}

// ForgotPassword generates a 6-digit OTP and emails it to the user.
func ForgotPassword(c *fiber.Ctx) error {
	req := new(forgotPasswordRequest)
	if err := c.BodyParser(req); err != nil {
		return badRequest(c, "Invalid request body")
	}

	req.Email = strings.ToLower(strings.TrimSpace(req.Email))
	genericMsg := "If that email is registered, you will receive a reset code shortly."

	var user models.User
	if tx := database.DB.Where("email = ?", req.Email).First(&user); tx.Error != nil {
		return c.JSON(fiber.Map{"message": genericMsg})
	}

	otp := storeOTP(&user, "password_reset")
	go func() {
		if err := mailer.Send(user.Email, "Your Outcraftly password reset code", mailer.OTPBody(otp, "password_reset")); err != nil {
			fmt.Fprintf(os.Stderr, "[mailer] ERROR sending reset OTP to %s: %v\n", user.Email, err)
		} else {
			fmt.Printf("[mailer] reset OTP sent to %s\n", user.Email)
		}
	}()

	return c.JSON(fiber.Map{"message": genericMsg})
}

// VerifyResetOTP validates the password-reset OTP and returns a short-lived reset token.
func VerifyResetOTP(c *fiber.Ctx) error {
	req := new(otpRequest)
	if err := c.BodyParser(req); err != nil {
		return badRequest(c, "Invalid request body")
	}
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))
	req.OTP = strings.TrimSpace(req.OTP)
	if req.Email == "" || req.OTP == "" {
		return badRequest(c, "Email and code are required")
	}

	var user models.User
	if tx := database.DB.Where("email = ?", req.Email).First(&user); tx.Error != nil {
		return badRequest(c, "Invalid or expired code")
	}

	if user.OTPCode == nil || *user.OTPCode != req.OTP {
		return badRequest(c, "Invalid or expired code")
	}
	if user.OTPExpires == nil || time.Now().After(*user.OTPExpires) {
		return badRequest(c, "Code has expired — please request a new one")
	}
	if user.OTPPurpose == nil || *user.OTPPurpose != "password_reset" {
		return badRequest(c, "Invalid code")
	}

	// Issue a 15-minute reset token and clear the OTP.
	resetToken := uuid.New().String()
	expiry := time.Now().Add(15 * time.Minute)
	database.DB.Model(&user).Updates(map[string]interface{}{
		"reset_token":         resetToken,
		"reset_token_expires": expiry,
		"otp_code":            nil,
		"otp_expires":         nil,
		"otp_purpose":         nil,
	})

	return c.JSON(fiber.Map{
		"message":     "Code verified. Please set your new password.",
		"reset_token": resetToken,
	})
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

	// Use gorm.Expr("NULL") to force zero-value writes on all drivers (SQLite, Postgres).
	database.DB.Model(&user).Updates(map[string]interface{}{
		"password_hash":         string(hash),
		"reset_token":           gorm.Expr("NULL"),
		"reset_token_expires":   gorm.Expr("NULL"),
		"failed_login_attempts": 0,
		"locked_until":          gorm.Expr("NULL"),
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
		"id":                  u.ID,
		"email":               u.Email,
		"email_verified":      u.EmailVerified,
		"name":                u.Name,
		"company_name":        u.CompanyName,
		"job_title":           u.JobTitle,
		"phone_country_code":  u.PhoneCountryCode,
		"phone_number":        u.PhoneNumber,
		"profile_complete":    u.ProfileComplete,
		"last_login_at":       u.LastLoginAt,
		"created_at":          u.CreatedAt,
	}
}

func errJSON(msg string) fiber.Map { return fiber.Map{"error": msg} }
func badRequest(c *fiber.Ctx, msg string) error {
	return c.Status(fiber.StatusBadRequest).JSON(errJSON(msg))
}
func serverError(c *fiber.Ctx, msg string) error {
	return c.Status(fiber.StatusInternalServerError).JSON(errJSON(msg))
}
