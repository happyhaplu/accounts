package handlers

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"outcraftly/accounts/config"
	"outcraftly/accounts/database"
	"outcraftly/accounts/models"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// ─────────────────────────────────────────────────────────────────────────────
// GET /api/v1/products  (protected — any authenticated user)
// Returns all active products.
// ─────────────────────────────────────────────────────────────────────────────

func ListProducts(c *fiber.Ctx) error {
	var products []models.Product
	database.DB.
		Where("is_active = true").
		Order("created_at ASC").
		Find(&products)

	return c.JSON(fiber.Map{"products": products})
}

// ─────────────────────────────────────────────────────────────────────────────
// GET /api/v1/admin/products  (admin only — includes inactive products)
// ─────────────────────────────────────────────────────────────────────────────

func AdminListProducts(c *fiber.Ctx) error {
	var products []models.Product
	database.DB.Order("created_at ASC").Find(&products)
	return c.JSON(fiber.Map{"products": products})
}

// ─────────────────────────────────────────────────────────────────────────────
// POST /api/v1/admin/products  (admin only)
// Body: { "name": "my-product", "description": "..." }
// ─────────────────────────────────────────────────────────────────────────────

func CreateProduct(c *fiber.Ctx) error {
	type body struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		LogoURL     string `json:"logo_url"`
		BillingMode string `json:"billing_mode"`
	}
	req := new(body)
	if err := c.BodyParser(req); err != nil {
		return badRequest(c, "Invalid request body")
	}
	req.Name = strings.ToLower(strings.TrimSpace(req.Name))
	if req.Name == "" {
		return badRequest(c, "Product name is required")
	}

	// Reject if a product with this name already exists (including inactive).
	var existing models.Product
	if database.DB.Where("name = ?", req.Name).First(&existing).Error == nil {
		return c.Status(fiber.StatusConflict).JSON(errJSON("A product with that name already exists"))
	}

	billingMode := strings.TrimSpace(req.BillingMode)
	if billingMode != "external" {
		billingMode = "managed"
	}
	product := models.Product{
		Name:        req.Name,
		Description: strings.TrimSpace(req.Description),
		LogoURL:     strings.TrimSpace(req.LogoURL),
		BillingMode: billingMode,
		IsActive:    true,
	}
	if err := database.DB.Create(&product).Error; err != nil {
		return serverError(c, "Failed to create product")
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"product": product})
}

// ─────────────────────────────────────────────────────────────────────────────
// PATCH /api/v1/admin/products/:id  (admin only)
// Body: { "name": "...", "description": "...", "is_active": true/false }
// All fields are optional; only supplied fields are updated.
// ─────────────────────────────────────────────────────────────────────────────

func UpdateProduct(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return badRequest(c, "Invalid product ID")
	}

	var product models.Product
	if tx := database.DB.Where("id = ?", id.String()).First(&product); tx.Error != nil {
		return c.Status(fiber.StatusNotFound).JSON(errJSON("Product not found"))
	}

	type body struct {
		Name         *string  `json:"name"`
		Description  *string  `json:"description"`
		IsActive     *bool    `json:"is_active"`
		RedirectURLs []string `json:"redirect_urls"`
		LogoURL      *string  `json:"logo_url"`
		BillingMode  *string  `json:"billing_mode"`
	}
	req := new(body)
	if err := c.BodyParser(req); err != nil {
		return badRequest(c, "Invalid request body")
	}

	updates := map[string]interface{}{}
	if req.Name != nil {
		n := strings.ToLower(strings.TrimSpace(*req.Name))
		if n == "" {
			return badRequest(c, "Product name cannot be empty")
		}
		updates["name"] = n
	}
	if req.Description != nil {
		updates["description"] = strings.TrimSpace(*req.Description)
	}
	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
	}
	if req.LogoURL != nil {
		updates["logo_url"] = strings.TrimSpace(*req.LogoURL)
	}
	if req.BillingMode != nil {
		bm := strings.TrimSpace(*req.BillingMode)
		if bm == "external" || bm == "managed" {
			updates["billing_mode"] = bm
		}
	}

	if len(updates) > 0 {
		if err := database.DB.Model(&product).Updates(updates).Error; err != nil {
			return serverError(c, "Failed to update product")
		}
	}

	// redirect_urls uses gorm serializer:json — must be updated via struct field,
	// not a raw map value, so GORM correctly JSON-encodes the slice for the text column.
	if req.RedirectURLs != nil {
		product.RedirectURLs = models.StringArray(req.RedirectURLs)
		if err := database.DB.Model(&product).Select("redirect_urls").Updates(&product).Error; err != nil {
			return serverError(c, "Failed to update redirect URLs")
		}
	}

	if len(updates) == 0 && req.RedirectURLs == nil {
		return badRequest(c, "No fields provided to update")
	}

	// Re-fetch to return fresh data (GORM's Updates won't reload json-serialized fields).
	database.DB.Where("id = ?", id.String()).First(&product)
	return c.JSON(fiber.Map{"product": product})
}

// ─────────────────────────────────────────────────────────────────────────────
// DELETE /api/v1/admin/products/:id  (admin only)
// Deactivates a product (soft-delete — sets is_active = false).
// Existing subscriptions that reference this product are unaffected.
// ─────────────────────────────────────────────────────────────────────────────

func DeactivateProduct(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return badRequest(c, "Invalid product ID")
	}

	result := database.DB.
		Model(&models.Product{}).
		Where("id = ? AND is_active = true", id).
		Update("is_active", false)

	if result.Error != nil {
		return serverError(c, "Failed to deactivate product")
	}
	if result.RowsAffected == 0 {
		return c.Status(fiber.StatusNotFound).JSON(errJSON("Product not found or already inactive"))
	}
	return c.JSON(fiber.Map{"message": "Product deactivated"})
}

// ─────────────────────────────────────────────────────────────────────────────
// POST /api/v1/admin/products/:id/logo  (admin only)
// Accepts a multipart/form-data file field named "logo".
// Saves the image to ./uploads/logos/<product-id>.<ext> and stores the
// public path in product.logo_url so users see it in the waffle menu and
// dashboard. Replaces any previously uploaded logo for this product.
//
// Constraints: JPEG, PNG, SVG, WEBP, GIF only; max 2 MB.
// ─────────────────────────────────────────────────────────────────────────────

func UploadProductLogo(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return badRequest(c, "Invalid product ID")
	}

	var product models.Product
	if tx := database.DB.Where("id = ?", id.String()).First(&product); tx.Error != nil {
		return c.Status(fiber.StatusNotFound).JSON(errJSON("Product not found"))
	}

	file, err := c.FormFile("logo")
	if err != nil {
		return badRequest(c, "No logo file provided (field name: \"logo\")")
	}

	// Max 2 MB
	const maxSize = 2 * 1024 * 1024
	if file.Size > maxSize {
		return badRequest(c, "Logo must be under 2 MB")
	}

	// Determine extension from Content-Type header
	ct := file.Header.Get("Content-Type")
	allowedExt := map[string]string{
		"image/png":     ".png",
		"image/jpeg":    ".jpg",
		"image/jpg":     ".jpg",
		"image/svg+xml": ".svg",
		"image/webp":    ".webp",
		"image/gif":     ".gif",
	}
	ext, ok := allowedExt[ct]
	if !ok {
		// Fall back to filename extension when Content-Type is generic
		ext = strings.ToLower(filepath.Ext(file.Filename))
		validExts := map[string]bool{".png": true, ".jpg": true, ".jpeg": true, ".svg": true, ".webp": true, ".gif": true}
		if !validExts[ext] {
			return badRequest(c, fmt.Sprintf("Unsupported file type %q — use PNG, JPEG, SVG, WEBP or GIF", ct))
		}
		if ext == ".jpeg" {
			ext = ".jpg"
		}
	}

	// Ensure upload directory exists
	uploadDir := "./uploads/logos"
	if mkErr := os.MkdirAll(uploadDir, 0755); mkErr != nil {
		return serverError(c, "Failed to create uploads directory")
	}

	// Deterministic filename: <uuid>.<ext> — overwrites previous logo for this product
	filename := product.ID.String() + ext
	destPath := filepath.Join(uploadDir, filename)

	if saveErr := c.SaveFile(file, destPath); saveErr != nil {
		return serverError(c, "Failed to save logo file")
	}

	logoURL := "/uploads/logos/" + filename
	if dbErr := database.DB.Model(&product).Update("logo_url", logoURL).Error; dbErr != nil {
		return serverError(c, "Failed to update logo URL in database")
	}
	product.LogoURL = logoURL

	return c.JSON(fiber.Map{"product": product, "logo_url": logoURL})
}

// ─────────────────────────────────────────────────────────────────────────────
// DELETE /api/v1/admin/products/:id/permanent  (admin only)
// Permanently removes a product row from the database.
// ⚠  Any existing subscriptions referencing this product will have a dangling
//    foreign key — only call this when you are certain no live subscriptions
//    reference the product (deactivate first, confirm, then hard-delete).
// ─────────────────────────────────────────────────────────────────────────────

func PermanentDeleteProduct(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return badRequest(c, "Invalid product ID")
	}

	var product models.Product
	if tx := database.DB.Where("id = ?", id.String()).First(&product); tx.Error != nil {
		return c.Status(fiber.StatusNotFound).JSON(errJSON("Product not found"))
	}

	// Delete child rows first to satisfy FK constraints before removing the product.
	// subscriptions.product_id → products.id (fk_subscriptions_product)
	if err := database.DB.Where("product_id = ?", id).Delete(&models.Subscription{}).Error; err != nil {
		return serverError(c, "Failed to remove product subscriptions before delete")
	}

	if err := database.DB.Delete(&models.Product{}, "id = ?", id).Error; err != nil {
		return serverError(c, "Failed to delete product")
	}

	return c.JSON(fiber.Map{"message": "Product permanently deleted"})
}

// ─────────────────────────────────────────────────────────────────────────────
// POST /api/v1/admin/products/:id/regenerate-key  (admin only)
//
// Generates a brand-new API key for the product and saves it.
// The old key is immediately invalidated — inform the product team to update
// their ACCOUNTS_API_KEY env var.
// ───────────────────────────────────────────────────────────────────────────────

func RegenerateProductAPIKey(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return badRequest(c, "Invalid product ID")
	}

	var product models.Product
	if tx := database.DB.Where("id = ?", id.String()).First(&product); tx.Error != nil {
		return c.Status(fiber.StatusNotFound).JSON(errJSON("Product not found"))
	}

	newKey, err := models.GenerateAPIKey()
	if err != nil {
		return serverError(c, "Failed to generate API key")
	}

	if err := database.DB.Model(&product).Update("api_key", newKey).Error; err != nil {
		return serverError(c, "Failed to save API key")
	}

	product.APIKey = newKey
	return c.JSON(fiber.Map{"product": product})
}

// ───────────────────────────────────────────────────────────────────────────────
// POST /api/v1/products/verify  (public, authenticated by X-API-Key)
//
// Used by Gour products for server-to-server token verification.
// The product passes the JWT it received from the frontend and their API key;
// Accounts confirms the token is valid, checks the subscription, and returns
// the user identity so the product never needs to share the JWT_SECRET.
//
// Request
//   Header: X-API-Key: gour_ce_xxxx
//   Body:   { "token": "<launch-token-jwt>" }
//
// Response 200
//   { "valid": true, "user_id": "", "email": "", "workspace_id": "",
//     "subscribed": true }               <- subscribed = has active subscription
//
// Response 401  invalid or missing API key
// Response 422  token missing or malformed
// ───────────────────────────────────────────────────────────────────────────────

func VerifyToken(c *fiber.Ctx) error {
	// 1. Authenticate the calling product via its API key.
	apiKey := strings.TrimSpace(c.Get("X-API-Key"))
	if apiKey == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(errJSON("Missing X-API-Key header"))
	}

	var product models.Product
	if database.DB.Where("api_key = ? AND is_active = true", apiKey).First(&product).Error != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(errJSON("Invalid API key"))
	}

	// 2. Parse and validate the JWT token from the request body.
	type reqBody struct {
		Token string `json:"token"`
	}
	req := new(reqBody)
	if err := c.BodyParser(req); err != nil || strings.TrimSpace(req.Token) == "" {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(errJSON("Request body must contain { \"token\": \"...\" }"))
	}

	parsed, err := jwt.Parse(req.Token, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fiber.ErrUnauthorized
		}
		return []byte(config.Cfg.JWTSecret), nil
	},
		jwt.WithIssuer(config.Cfg.JWTIssuer),
		jwt.WithAudience(config.Cfg.JWTAudience),
	)
	if err != nil || !parsed.Valid {
		return c.JSON(fiber.Map{"valid": false, "reason": "token expired or invalid"})
	}

	claims, ok := parsed.Claims.(jwt.MapClaims)
	if !ok {
		return c.JSON(fiber.Map{"valid": false, "reason": "malformed claims"})
	}

	userID, _      := claims["sub"].(string)
	email, _       := claims["email"].(string)
	workspaceIDStr, _ := claims["workspace_id"].(string)

	// 3. Check whether this workspace has an active subscription to the calling product.
	subscribed := false
	if wid, err := uuid.Parse(workspaceIDStr); err == nil {
		var sub models.Subscription
		if database.DB.
			Where("workspace_id = ? AND product_id = ? AND status = 'active'", wid, product.ID).
			First(&sub).Error == nil {
			subscribed = sub.IsAccessible()
		}
	}

	return c.JSON(fiber.Map{
		"valid":        true,
		"user_id":      userID,
		"email":        email,
		"workspace_id": workspaceIDStr,
		"subscribed":   subscribed,
	})
}
//
// Checks the caller's active subscription, signs a 7-day launch token, and
// returns the full callback URL for the frontend to navigate to.
//
// Flow A — dashboard launch (no redirect_uri):
//   → picks the best URL from product.redirect_urls (https in prod / localhost in dev)
//
// Flow B — external product redirect (Google-style):
//   GET /api/v1/products/my-product/launch?redirect_uri=https://app.example.com/callback
//   → validates redirect_uri against product's allowed origins, then uses it
//
// Response: 200 { "redirect_url": "https://...?token=<jwt>", "token": "<jwt>" }
// ─────────────────────────────────────────────────────────────────────────────

func LaunchProduct(c *fiber.Ctx) error {
	userID, _ := uuid.Parse(c.Locals("userID").(string))
	email := c.Locals("email").(string)
	name := strings.ToLower(c.Params("name"))

	// Resolve product by name slug.
	var product models.Product
	if database.DB.Where("name = ? AND is_active = true", name).First(&product).Error != nil {
		return c.Status(fiber.StatusNotFound).JSON(errJSON("Product not found"))
	}

	// Find the user's owner workspace (oldest first if multiple).
	var member models.WorkspaceMember
	if database.DB.Where("user_id = ? AND role = 'owner'", userID).
		Order("joined_at ASC").First(&member).Error != nil {
		return c.Status(fiber.StatusBadRequest).JSON(errJSON("No workspace found for user"))
	}
	workspaceID := member.WorkspaceID

	// Check subscription — only enforce for "managed" billing products.
	// Products with billing_mode = "external" handle their own billing
	// and must always receive a launch token regardless of subscription status.
	if product.BillingMode != "external" {
		var sub models.Subscription
		tx := database.DB.
			Where("workspace_id = ? AND product_id = ? AND status = 'active'", workspaceID, product.ID).
			First(&sub)
		if tx.Error != nil || !sub.IsAccessible() {
			return c.Status(fiber.StatusForbidden).JSON(errJSON("No active subscription for this product"))
		}
	}

	// Sign a 7-day launch token containing identity + workspace.
	launchToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":          userID.String(),
		"email":        email,
		"workspace_id": workspaceID.String(),
		"role":         c.Locals("role"),
		"iss":          config.Cfg.JWTIssuer,
		"aud":          config.Cfg.JWTAudience,
		"exp":          time.Now().Add(7 * 24 * time.Hour).Unix(),
		"iat":          time.Now().Unix(),
	})
	signed, err := launchToken.SignedString([]byte(config.Cfg.JWTSecret))
	if err != nil {
		return serverError(c, "Failed to generate launch token")
	}

	// ── Determine the redirect URL ──────────────────────────────────────────
	//
	// Priority:
	//   1. redirect_uri query param  — caller-supplied, validated against allowed origins
	//   2. DB redirect_urls          — pick prod/dev URL based on APP_URL prefix
	//   3. Fallback                  — send the user to the accounts dashboard
	//
	var redirectURL string

	if ru := strings.TrimSpace(c.Query("redirect_uri")); ru != "" {
		// Security: only accept redirect_uri values whose scheme+host matches one
		// of the product's configured redirect URLs. Prevents open-redirect attacks.
		if !isAllowedRedirectURI(ru, []string(product.RedirectURLs)) {
			return c.Status(fiber.StatusBadRequest).JSON(errJSON(
				"redirect_uri is not in the list of allowed URLs for this product",
			))
		}
		redirectURL = ru
	} else if len(product.RedirectURLs) > 0 {
		isProd := strings.HasPrefix(config.Cfg.AccountsBaseURL, "https://")
		for _, u := range product.RedirectURLs {
			if isProd && strings.HasPrefix(u, "https://") {
				redirectURL = u
				break
			}
			if !isProd && strings.Contains(u, "localhost") {
				redirectURL = u
				break
			}
		}
		if redirectURL == "" {
			redirectURL = product.RedirectURLs[0]
		}
	} else {
		// No redirect URLs configured anywhere — send back to dashboard.
		redirectURL = config.Cfg.AccountsBaseURL + "/dashboard"
	}

	// Use & if the redirect URL already contains query params.
	sep := "?"
	if strings.Contains(redirectURL, "?") {
		sep = "&"
	}

	return c.JSON(fiber.Map{
		"redirect_url": redirectURL + sep + "token=" + signed,
		"token":        signed,
	})
}

// isAllowedRedirectURI returns true when redirectURI's scheme+host is permitted.
//
// It checks two sources (union):
//   1. The product's DB-configured redirect_urls  (per-product whitelist)
//   2. The ALLOWED_REDIRECT_ORIGINS env var       (global whitelist, comma-separated)
//
// This means localhost:3000 works in dev without needing to be in the DB, and
// production URLs work without re-seeding on every deploy.
func isAllowedRedirectURI(redirectURI string, productURLs []string) bool {
	target := uriOrigin(redirectURI)
	if target == "" {
		return false
	}
	// Check product-level DB URLs.
	for _, u := range productURLs {
		if uriOrigin(u) == target {
			return true
		}
	}
	// Check AUTH_REDIRECT_URIS env allowlist.
	for _, u := range config.Cfg.AuthRedirectURIs {
		if uriOrigin(u) == target {
			return true
		}
	}
	// Check global env-var allowlist.
	return isGloballyAllowedOrigin(redirectURI)
}

// isGloballyAllowedOrigin checks redirectURI's origin against:
//   1. All product redirect_urls in the database (via the CORS cache)
//   2. The ALLOWED_REDIRECT_ORIGINS env var (legacy fallback / dev overrides)
//   3. The ALLOW_ORIGINS env var (static origins like the accounts frontend)
//
// This means adding a new product with its redirect_urls in the admin API
// automatically allows redirects to that origin — zero env changes.
func isGloballyAllowedOrigin(redirectURI string) bool {
	target := uriOrigin(redirectURI)
	if target == "" {
		return false
	}

	// Check the dynamic product-origin cache + static origins.
	if database.IsAllowedOrigin(target) {
		return true
	}

	// Legacy fallback: ALLOWED_REDIRECT_ORIGINS env var.
	raw := os.Getenv("ALLOWED_REDIRECT_ORIGINS")
	if raw == "" {
		return false
	}
	for _, entry := range strings.Split(raw, ",") {
		entry = strings.TrimSpace(entry)
		if uriOrigin(entry) == target || strings.TrimRight(entry, "/") == target {
			return true
		}
	}
	return false
}

// uriOrigin returns "scheme://host" (e.g. "https://app.example.com").
func uriOrigin(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil || u.Host == "" {
		return ""
	}
	return u.Scheme + "://" + u.Host
}

// ─────────────────────────────────────────────────────────────────────────────
// GET /api/v1/products/:name/check
// Called by external Gour apps to verify
// whether the bearer of a launch token has an active subscription.
// Header: Authorization: Bearer <launch-token>
//
// 200 { "active": true,  "status": "active", "plan_name": "pro" }
// 200 { "active": false }
// 401 invalid / expired token
// ─────────────────────────────────────────────────────────────────────────────

func CheckProductSubscription(c *fiber.Ctx) error {
	name := strings.ToLower(c.Params("name"))

	// Extract and verify the Bearer token.
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(errJSON("Missing Authorization header"))
	}
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
		return c.Status(fiber.StatusUnauthorized).JSON(errJSON("Authorization header must be: Bearer <token>"))
	}

	token, err := jwt.Parse(parts[1], func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fiber.ErrUnauthorized
		}
		return []byte(config.Cfg.JWTSecret), nil
	},
		jwt.WithIssuer(config.Cfg.JWTIssuer),
		jwt.WithAudience(config.Cfg.JWTAudience),
	)
	if err != nil || !token.Valid {
		return c.Status(fiber.StatusUnauthorized).JSON(errJSON("Invalid or expired token"))
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(errJSON("Invalid token claims"))
	}

	workspaceIDStr, _ := claims["workspace_id"].(string)
	workspaceID, err := uuid.Parse(workspaceIDStr)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(errJSON("Invalid workspace_id in token"))
	}

	// Lookup product.
	var product models.Product
	if database.DB.Where("name = ? AND is_active = true", name).First(&product).Error != nil {
		return c.Status(fiber.StatusNotFound).JSON(errJSON("Product not found"))
	}

	// Lookup subscription.
	var sub models.Subscription
	if database.DB.
		Where("workspace_id = ? AND product_id = ? AND status = 'active'", workspaceID, product.ID).
		First(&sub).Error != nil {
		return c.JSON(fiber.Map{"active": false})
	}

	return c.JSON(fiber.Map{
		"active":    sub.IsAccessible(),
		"status":    sub.Status,
		"plan_name": sub.PlanName,
	})
}
