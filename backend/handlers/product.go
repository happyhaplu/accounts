package handlers

import (
	"net/url"
	"os"
	"strings"
	"time"

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
// Body: { "name": "cold_email", "description": "..." }
// ─────────────────────────────────────────────────────────────────────────────

func CreateProduct(c *fiber.Ctx) error {
	type body struct {
		Name          string  `json:"name"`
		Description   string  `json:"description"`
		StripePriceID *string `json:"stripe_price_id"`
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

	product := models.Product{
		Name:          req.Name,
		Description:   strings.TrimSpace(req.Description),
		StripePriceID: req.StripePriceID,
		IsActive:      true,
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
	if tx := database.DB.First(&product, "id = ?", id); tx.Error != nil {
		return c.Status(fiber.StatusNotFound).JSON(errJSON("Product not found"))
	}

	type body struct {
		Name          *string `json:"name"`
		Description   *string `json:"description"`
		IsActive      *bool   `json:"is_active"`
		StripePriceID *string `json:"stripe_price_id"`
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
	if req.StripePriceID != nil {
		updates["stripe_price_id"] = *req.StripePriceID
	}

	if len(updates) == 0 {
		return badRequest(c, "No fields provided to update")
	}

	if err := database.DB.Model(&product).Updates(updates).Error; err != nil {
		return serverError(c, "Failed to update product")
	}

	// Re-fetch to return fresh data
	database.DB.First(&product, "id = ?", id)
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
// GET /api/v1/products/:name/launch  (protected — authenticated user)
//
// Checks the caller's active subscription, signs a 7-day launch token, and
// returns the full callback URL for the frontend to navigate to.
//
// Flow A — dashboard launch (no redirect_uri):
//   → picks the best URL from product.redirect_urls (https in prod / localhost in dev)
//
// Flow B — external product redirect (Google-style):
//   GET /api/v1/products/email-warmup/launch?redirect_uri=https://warmup.outcraftly.com/callback
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

	// Verify an active, non-expired subscription exists.
	var sub models.Subscription
	tx := database.DB.
		Where("workspace_id = ? AND product_id = ? AND status = 'active'", workspaceID, product.ID).
		First(&sub)
	if tx.Error != nil || !sub.IsAccessible() {
		return c.Status(fiber.StatusForbidden).JSON(errJSON("No active subscription for this product"))
	}

	// Sign a 7-day launch token containing identity + workspace.
	launchToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":          userID.String(),
		"email":        email,
		"workspace_id": workspaceID.String(),
		"exp":          time.Now().Add(7 * 24 * time.Hour).Unix(),
	})
	signed, err := launchToken.SignedString([]byte(os.Getenv("JWT_SECRET")))
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
		if !isAllowedRedirectURI(ru, product.RedirectURLs) {
			return c.Status(fiber.StatusBadRequest).JSON(errJSON(
				"redirect_uri is not in the list of allowed URLs for this product",
			))
		}
		redirectURL = ru
	} else if len(product.RedirectURLs) > 0 {
		appURL := os.Getenv("APP_URL")
		isProd := strings.HasPrefix(appURL, "https://")
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
		redirectURL = os.Getenv("APP_URL") + "/dashboard"
	}

	return c.JSON(fiber.Map{
		"redirect_url": redirectURL + "?token=" + signed,
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

// uriOrigin returns "scheme://host" (e.g. "https://warmup.outcraftly.com").
func uriOrigin(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil || u.Host == "" {
		return ""
	}
	return u.Scheme + "://" + u.Host
}

// ─────────────────────────────────────────────────────────────────────────────
// GET /api/v1/products/:name/check
// Called by external Outcraftly apps (e.g. warmup.outcraftly.com) to verify
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
		return []byte(os.Getenv("JWT_SECRET")), nil
	})
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
