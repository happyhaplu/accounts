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
	"github.com/google/uuid"
)

// ─────────────────────────────────────────────────────────────────────────────
// POST /api/v1/workspaces/:id/invites
// Send an invite email to any address; no account required.
// Only workspace owners may invite.
// ─────────────────────────────────────────────────────────────────────────────

func SendInvite(c *fiber.Ctx) error {
	uid, _ := uuid.Parse(c.Locals("userID").(string))
	wsID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return badRequest(c, "Invalid workspace ID")
	}

	// Must be an owner of this workspace.
	var requester models.WorkspaceMember
	if tx := database.DB.Where("workspace_id = ? AND user_id = ?", wsID, uid).First(&requester); tx.Error != nil {
		return c.Status(fiber.StatusForbidden).JSON(errJSON("Workspace not found or access denied"))
	}
	if requester.Role != "owner" {
		return c.Status(fiber.StatusForbidden).JSON(errJSON("Only workspace owners can send invites"))
	}

	type body struct {
		Email string `json:"email"`
		Role  string `json:"role"`
	}
	req := new(body)
	if err := c.BodyParser(req); err != nil {
		return badRequest(c, "Invalid request body")
	}
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))
	if req.Email == "" {
		return badRequest(c, "Email is required")
	}
	if req.Role != "owner" && req.Role != "member" {
		req.Role = "member"
	}

	// If this email already belongs to an active member, reject.
	var existingUser models.User
	if database.DB.Where("email = ?", req.Email).First(&existingUser).Error == nil {
		var mem models.WorkspaceMember
		if database.DB.Where("workspace_id = ? AND user_id = ?", wsID, existingUser.ID).First(&mem).Error == nil {
			return c.Status(fiber.StatusConflict).JSON(errJSON("This person is already a workspace member"))
		}
	}

	// Invalidate any existing pending invite for the same email+workspace.
	database.DB.
		Model(&models.WorkspaceInvite{}).
		Where("workspace_id = ? AND email = ? AND status = 'pending'", wsID, req.Email).
		Update("status", "revoked")

	// Load workspace and inviter info for the email.
	var ws models.Workspace
	database.DB.First(&ws, "id = ?", wsID)
	var inviter models.User
	database.DB.First(&inviter, "id = ?", uid)

	// Create invite record.
	token := uuid.New().String()
	invite := models.WorkspaceInvite{
		WorkspaceID: wsID,
		InvitedBy:   uid,
		Email:       req.Email,
		Role:        req.Role,
		Token:       token,
		Status:      "pending",
		ExpiresAt:   time.Now().Add(7 * 24 * time.Hour),
	}
	if err := database.DB.Create(&invite).Error; err != nil {
		return serverError(c, "Failed to create invite")
	}

	// Send invite email asynchronously.
	appURL := os.Getenv("APP_URL")
	link := appURL + "/invite?token=" + token
	go func() {
		body := mailer.WorkspaceInviteBody(ws.Name, inviter.Name, inviter.Email, req.Role, link)
		if err := mailer.Send(req.Email, "You're invited to join "+ws.Name+" on Gour", body); err != nil {
			fmt.Fprintf(os.Stderr, "[mailer] ERROR sending invite to %s: %v\n", req.Email, err)
		} else {
			fmt.Printf("[mailer] invite email sent to %s\n", req.Email)
		}
	}()

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Invite sent to " + req.Email,
		"invite": fiber.Map{
			"id":         invite.ID,
			"email":      invite.Email,
			"role":       invite.Role,
			"status":     invite.Status,
			"expires_at": invite.ExpiresAt,
			"created_at": invite.CreatedAt,
		},
	})
}

// ─────────────────────────────────────────────────────────────────────────────
// GET /api/v1/workspaces/:id/invites
// List pending invites for a workspace (owners only).
// ─────────────────────────────────────────────────────────────────────────────

func ListInvites(c *fiber.Ctx) error {
	uid, _ := uuid.Parse(c.Locals("userID").(string))
	wsID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return badRequest(c, "Invalid workspace ID")
	}

	var requester models.WorkspaceMember
	if tx := database.DB.Where("workspace_id = ? AND user_id = ?", wsID, uid).First(&requester); tx.Error != nil {
		return c.Status(fiber.StatusForbidden).JSON(errJSON("Workspace not found or access denied"))
	}
	if requester.Role != "owner" {
		return c.Status(fiber.StatusForbidden).JSON(errJSON("Only owners can view invites"))
	}

	var invites []models.WorkspaceInvite
	database.DB.
		Where("workspace_id = ? AND status = 'pending' AND expires_at > ?", wsID, time.Now()).
		Order("created_at DESC").
		Find(&invites)

	out := make([]fiber.Map, len(invites))
	for i, inv := range invites {
		out[i] = fiber.Map{
			"id":         inv.ID,
			"email":      inv.Email,
			"role":       inv.Role,
			"status":     inv.Status,
			"expires_at": inv.ExpiresAt,
			"created_at": inv.CreatedAt,
		}
	}
	return c.JSON(fiber.Map{"invites": out})
}

// ─────────────────────────────────────────────────────────────────────────────
// DELETE /api/v1/workspaces/:id/invites/:inviteID
// Revoke a pending invite (owner only).
// ─────────────────────────────────────────────────────────────────────────────

func RevokeInvite(c *fiber.Ctx) error {
	uid, _ := uuid.Parse(c.Locals("userID").(string))
	wsID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return badRequest(c, "Invalid workspace ID")
	}
	inviteID, err := uuid.Parse(c.Params("inviteID"))
	if err != nil {
		return badRequest(c, "Invalid invite ID")
	}

	var requester models.WorkspaceMember
	if tx := database.DB.Where("workspace_id = ? AND user_id = ?", wsID, uid).First(&requester); tx.Error != nil {
		return c.Status(fiber.StatusForbidden).JSON(errJSON("Workspace not found or access denied"))
	}
	if requester.Role != "owner" {
		return c.Status(fiber.StatusForbidden).JSON(errJSON("Only owners can revoke invites"))
	}

	result := database.DB.
		Model(&models.WorkspaceInvite{}).
		Where("id = ? AND workspace_id = ? AND status = 'pending'", inviteID, wsID).
		Update("status", "revoked")

	if result.Error != nil {
		return serverError(c, "Failed to revoke invite")
	}
	if result.RowsAffected == 0 {
		return c.Status(fiber.StatusNotFound).JSON(errJSON("Invite not found or already resolved"))
	}
	return c.JSON(fiber.Map{"message": "Invite revoked"})
}

// ─────────────────────────────────────────────────────────────────────────────
// GET /api/v1/invites/preview?token=xxx  (public)
// Returns invite details so the frontend can render the accept screen.
// ─────────────────────────────────────────────────────────────────────────────

func PreviewInvite(c *fiber.Ctx) error {
	token := c.Query("token")
	if token == "" {
		return badRequest(c, "Token is required")
	}

	var invite models.WorkspaceInvite
	if tx := database.DB.
		Preload("Workspace").
		Preload("InviterUser").
		Where("token = ? AND status = 'pending'", token).
		First(&invite); tx.Error != nil {
		return c.Status(fiber.StatusNotFound).JSON(errJSON("Invite not found or already used"))
	}
	if time.Now().After(invite.ExpiresAt) {
		return c.Status(fiber.StatusGone).JSON(errJSON("This invite has expired"))
	}

	return c.JSON(fiber.Map{
		"invite": fiber.Map{
			"id":             invite.ID,
			"email":          invite.Email,
			"role":           invite.Role,
			"workspace_name": invite.Workspace.Name,
			"invited_by":     invite.InviterUser.Name,
			"expires_at":     invite.ExpiresAt,
		},
	})
}

// ─────────────────────────────────────────────────────────────────────────────
// POST /api/v1/invites/accept  (protected — caller must be logged in)
// Accepts an invite for the currently authenticated user.
// ─────────────────────────────────────────────────────────────────────────────

func AcceptInvite(c *fiber.Ctx) error {
	uid, _ := uuid.Parse(c.Locals("userID").(string))

	type body struct {
		Token string `json:"token"`
	}
	req := new(body)
	if err := c.BodyParser(req); err != nil {
		return badRequest(c, "Invalid request body")
	}
	if req.Token == "" {
		return badRequest(c, "Token is required")
	}

	var invite models.WorkspaceInvite
	if tx := database.DB.
		Preload("Workspace").
		Where("token = ? AND status = 'pending'", req.Token).
		First(&invite); tx.Error != nil {
		return c.Status(fiber.StatusNotFound).JSON(errJSON("Invite not found or already used"))
	}
	if time.Now().After(invite.ExpiresAt) {
		return c.Status(fiber.StatusGone).JSON(errJSON("This invite has expired"))
	}

	// Verify the logged-in user's email matches the invite.
	var user models.User
	database.DB.First(&user, "id = ?", uid)
	if strings.ToLower(user.Email) != strings.ToLower(invite.Email) {
		return c.Status(fiber.StatusForbidden).JSON(errJSON(
			"This invite was sent to " + invite.Email + ". Please log in with that account.",
		))
	}

	// Guard against already being a member.
	var existing models.WorkspaceMember
	if database.DB.Where("workspace_id = ? AND user_id = ?", invite.WorkspaceID, uid).First(&existing).Error == nil {
		// Already a member — just mark invite accepted and return success.
		database.DB.Model(&invite).Update("status", "accepted")
		return c.JSON(fiber.Map{
			"message":        "You are already a member of this workspace",
			"workspace_name": invite.Workspace.Name,
		})
	}

	// Add user as workspace member.
	member := models.WorkspaceMember{
		WorkspaceID: invite.WorkspaceID,
		UserID:      uid,
		Role:        invite.Role,
	}
	if err := database.DB.Create(&member).Error; err != nil {
		return serverError(c, "Failed to join workspace")
	}

	// Mark invite accepted.
	database.DB.Model(&invite).Update("status", "accepted")

	return c.JSON(fiber.Map{
		"message":        "You have joined " + invite.Workspace.Name,
		"workspace_name": invite.Workspace.Name,
		"role":           invite.Role,
	})
}

// ─────────────────────────────────────────────────────────────────────────────
// AcceptPendingInvitesForEmail — called from Register after user is created.
// Finds any pending invites for the new user's email and auto-accepts them.
// ─────────────────────────────────────────────────────────────────────────────

func AcceptPendingInvitesForEmail(userID uuid.UUID, email string) {
	var invites []models.WorkspaceInvite
	database.DB.
		Where("email = ? AND status = 'pending' AND expires_at > ?", strings.ToLower(email), time.Now()).
		Find(&invites)

	for _, inv := range invites {
		// Skip if already a member (shouldn't happen but be safe).
		var existing models.WorkspaceMember
		if database.DB.Where("workspace_id = ? AND user_id = ?", inv.WorkspaceID, userID).First(&existing).Error == nil {
			database.DB.Model(&inv).Update("status", "accepted")
			continue
		}
		member := models.WorkspaceMember{
			WorkspaceID: inv.WorkspaceID,
			UserID:      userID,
			Role:        inv.Role,
		}
		if database.DB.Create(&member).Error == nil {
			database.DB.Model(&inv).Update("status", "accepted")
		}
	}
}
