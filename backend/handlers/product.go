package handlers

import (
	"strings"

	"outcraftly/accounts/database"
	"outcraftly/accounts/models"

	"github.com/gofiber/fiber/v2"
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
