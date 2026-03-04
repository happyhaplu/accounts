package handlers

import (
	"time"

	"outcraftly/accounts/database"
	"outcraftly/accounts/models"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// ─────────────────────────────────────────────────────────────────────────────
// POST /api/v1/workspaces/:id/subscriptions  (owner only)
// createSubscription — create or renew a subscription for a workspace+product.
// Body: { "product_id": "uuid", "plan_name": "starter", "period_days": 30 }
// If a subscription for this workspace+product already exists it is replaced
// (status reset to "active", period extended) — idempotent for renewals.
// ─────────────────────────────────────────────────────────────────────────────

func CreateSubscription(c *fiber.Ctx) error {
	uid, _ := uuid.Parse(c.Locals("userID").(string))
	wsID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return badRequest(c, "Invalid workspace ID")
	}

	// Only workspace owners may create subscriptions.
	if !isWorkspaceOwner(uid, wsID) {
		return c.Status(fiber.StatusForbidden).JSON(errJSON("Only workspace owners can manage subscriptions"))
	}

	type body struct {
		ProductID  string `json:"product_id"`
		PlanName   string `json:"plan_name"`
		PeriodDays int    `json:"period_days"`
	}
	req := new(body)
	if err := c.BodyParser(req); err != nil {
		return badRequest(c, "Invalid request body")
	}

	productID, err := uuid.Parse(req.ProductID)
	if err != nil {
		return badRequest(c, "Invalid product_id")
	}
	if req.PlanName == "" {
		return badRequest(c, "plan_name is required")
	}
	if req.PeriodDays <= 0 {
		req.PeriodDays = 30 // default billing period
	}

	// Verify the product exists and is active.
	var product models.Product
	if database.DB.Where("id = ? AND is_active = true", productID).First(&product).Error != nil {
		return c.Status(fiber.StatusNotFound).JSON(errJSON("Product not found or inactive"))
	}

	periodEnd := time.Now().Add(time.Duration(req.PeriodDays) * 24 * time.Hour)

	// Upsert: revoke any existing subscription for this workspace+product first.
	database.DB.
		Model(&models.Subscription{}).
		Where("workspace_id = ? AND product_id = ?", wsID, productID).
		Updates(map[string]interface{}{"status": "canceled"})

	sub := models.Subscription{
		WorkspaceID:      wsID,
		ProductID:        productID,
		PlanName:         req.PlanName,
		Status:           "active",
		CurrentPeriodEnd: periodEnd,
	}
	if err := database.DB.Create(&sub).Error; err != nil {
		return serverError(c, "Failed to create subscription")
	}

	database.DB.Preload("Product").First(&sub, "id = ?", sub.ID)
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"subscription": sub})
}

// ─────────────────────────────────────────────────────────────────────────────
// GET /api/v1/workspaces/:id/subscriptions  (owner only)
// listSubscriptions — all subscriptions for a workspace.
// ─────────────────────────────────────────────────────────────────────────────

func ListSubscriptions(c *fiber.Ctx) error {
	uid, _ := uuid.Parse(c.Locals("userID").(string))
	wsID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return badRequest(c, "Invalid workspace ID")
	}

	if !isWorkspaceMember(uid, wsID) {
		return c.Status(fiber.StatusForbidden).JSON(errJSON("Access denied"))
	}

	var subs []models.Subscription
	database.DB.
		Preload("Product").
		Where("workspace_id = ?", wsID).
		Order("created_at DESC").
		Find(&subs)

	// Annotate expired-but-not-yet-marked subscriptions on the fly.
	now := time.Now()
	for i := range subs {
		if subs[i].Status == "active" && now.After(subs[i].CurrentPeriodEnd) {
			subs[i].Status = "expired"
			database.DB.Model(&subs[i]).Update("status", "expired")
		}
	}

	return c.JSON(fiber.Map{"subscriptions": subs})
}

// ─────────────────────────────────────────────────────────────────────────────
// GET /api/v1/workspaces/:id/subscriptions/access?product_id=uuid  (any member)
// checkAccess — returns whether the workspace has active access to a product.
// ─────────────────────────────────────────────────────────────────────────────

func CheckAccess(c *fiber.Ctx) error {
	uid, _ := uuid.Parse(c.Locals("userID").(string))
	wsID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return badRequest(c, "Invalid workspace ID")
	}
	productID, err := uuid.Parse(c.Query("product_id"))
	if err != nil {
		return badRequest(c, "Invalid or missing product_id query param")
	}

	if !isWorkspaceMember(uid, wsID) {
		return c.Status(fiber.StatusForbidden).JSON(errJSON("Access denied"))
	}

	var sub models.Subscription
	tx := database.DB.
		Where("workspace_id = ? AND product_id = ? AND status = 'active'", wsID, productID).
		First(&sub)

	hasAccess := tx.Error == nil && sub.IsAccessible()
	return c.JSON(fiber.Map{"has_access": hasAccess})
}

// ─────────────────────────────────────────────────────────────────────────────
// DELETE /api/v1/workspaces/:id/subscriptions/:subID  (owner only)
// cancelSubscription — marks a subscription as canceled.
// ─────────────────────────────────────────────────────────────────────────────

func CancelSubscription(c *fiber.Ctx) error {
	uid, _ := uuid.Parse(c.Locals("userID").(string))
	wsID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return badRequest(c, "Invalid workspace ID")
	}
	subID, err := uuid.Parse(c.Params("subID"))
	if err != nil {
		return badRequest(c, "Invalid subscription ID")
	}

	if !isWorkspaceOwner(uid, wsID) {
		return c.Status(fiber.StatusForbidden).JSON(errJSON("Only workspace owners can cancel subscriptions"))
	}

	result := database.DB.
		Model(&models.Subscription{}).
		Where("id = ? AND workspace_id = ? AND status = 'active'", subID, wsID).
		Update("status", "canceled")

	if result.Error != nil {
		return serverError(c, "Failed to cancel subscription")
	}
	if result.RowsAffected == 0 {
		return c.Status(fiber.StatusNotFound).JSON(errJSON("Active subscription not found"))
	}
	return c.JSON(fiber.Map{"message": "Subscription canceled"})
}

// ─────────────────────────────────────────────────────────────────────────────
// Admin: GET /api/v1/admin/subscriptions?workspace_id=&product_id=&status=
// ─────────────────────────────────────────────────────────────────────────────

func AdminListSubscriptions(c *fiber.Ctx) error {
	q := database.DB.Preload("Workspace").Preload("Product").Order("created_at DESC")

	if wsID := c.Query("workspace_id"); wsID != "" {
		q = q.Where("workspace_id = ?", wsID)
	}
	if prodID := c.Query("product_id"); prodID != "" {
		q = q.Where("product_id = ?", prodID)
	}
	if status := c.Query("status"); status != "" {
		q = q.Where("status = ?", status)
	}

	var subs []models.Subscription
	q.Find(&subs)
	return c.JSON(fiber.Map{"subscriptions": subs})
}

// ─────────────────────────────────────────────────────────────────────────────
// Helpers
// ─────────────────────────────────────────────────────────────────────────────

func isWorkspaceOwner(userID, wsID uuid.UUID) bool {
	var m models.WorkspaceMember
	return database.DB.Where("workspace_id = ? AND user_id = ? AND role = 'owner'", wsID, userID).
		First(&m).Error == nil
}

func isWorkspaceMember(userID, wsID uuid.UUID) bool {
	var m models.WorkspaceMember
	return database.DB.Where("workspace_id = ? AND user_id = ?", wsID, userID).
		First(&m).Error == nil
}
