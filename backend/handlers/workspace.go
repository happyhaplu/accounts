package handlers

import (
	"fmt"
	"strings"

	"outcraftly/accounts/database"
	"outcraftly/accounts/models"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// ─────────────────────────────────────────────────────────────────────────────
// Internal helper — auto-called after registration
// ─────────────────────────────────────────────────────────────────────────────

// createWorkspaceForUser creates a default workspace and adds the user as owner.
// Called inside the Register handler immediately after the user row is saved.
func createWorkspaceForUser(user models.User) error {
	// Derive a readable workspace name from the e-mail prefix.
	name := user.Email
	if idx := strings.Index(name, "@"); idx > 0 {
		name = name[:idx]
	}
	if len(name) > 0 {
		name = strings.ToUpper(name[:1]) + name[1:]
	}
	name += "'s Workspace"
	return provisionWorkspace(user.ID, name)
}

// provisionWorkspace creates a Workspace row and adds ownerID as "owner" member.
func provisionWorkspace(ownerID uuid.UUID, name string) error {
	ws := models.Workspace{
		Name:    name,
		OwnerID: ownerID,
	}
	if err := database.DB.Create(&ws).Error; err != nil {
		return fmt.Errorf("create workspace row: %w", err)
	}
	member := models.WorkspaceMember{
		WorkspaceID: ws.ID,
		UserID:      ownerID,
		Role:        "owner",
	}
	if err := database.DB.Create(&member).Error; err != nil {
		return fmt.Errorf("create workspace member: %w", err)
	}
	return nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Serialisation helpers
// ─────────────────────────────────────────────────────────────────────────────

func memberJSON(m models.WorkspaceMember) fiber.Map {
	return fiber.Map{
		"id":           m.ID,
		"workspace_id": m.WorkspaceID,
		"role":         m.Role,
		"joined_at":    m.JoinedAt,
		"user": fiber.Map{
			"id":        m.User.ID,
			"email":     m.User.Email,
			"name":      m.User.Name,
			"job_title": m.User.JobTitle,
		},
	}
}

func workspaceSummary(ws models.Workspace, myRole string) fiber.Map {
	return fiber.Map{
		"id":          ws.ID,
		"name":        ws.Name,
		"owner_id":    ws.OwnerID,
		"my_role":     myRole,
		"created_at":  ws.CreatedAt,
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// GET /api/v1/workspaces
// List all workspaces the caller belongs to.
// ─────────────────────────────────────────────────────────────────────────────

func ListWorkspaces(c *fiber.Ctx) error {
	uid, _ := uuid.Parse(c.Locals("userID").(string))

	var memberships []models.WorkspaceMember
	database.DB.Where("user_id = ?", uid).Find(&memberships)

	if len(memberships) == 0 {
		// Lazy-create default workspace for pre-migration accounts.
		var user models.User
		if database.DB.First(&user, "id = ?", uid).Error != nil {
			return c.Status(fiber.StatusNotFound).JSON(errJSON("User not found"))
		}
		if err := createWorkspaceForUser(user); err != nil {
			return serverError(c, "Failed to initialise workspace")
		}
		database.DB.Where("user_id = ?", uid).Find(&memberships)
	}

	// Collect workspace IDs, then fetch them all.
	wsIDs := make([]uuid.UUID, len(memberships))
	roleByWs := make(map[uuid.UUID]string, len(memberships))
	for i, m := range memberships {
		wsIDs[i] = m.WorkspaceID
		roleByWs[m.WorkspaceID] = m.Role
	}

	var workspaces []models.Workspace
	database.DB.Where("id IN ?", wsIDs).Find(&workspaces)

	out := make([]fiber.Map, len(workspaces))
	for i, ws := range workspaces {
		out[i] = workspaceSummary(ws, roleByWs[ws.ID])
	}

	return c.JSON(fiber.Map{"workspaces": out})
}

// ─────────────────────────────────────────────────────────────────────────────
// POST /api/v1/workspaces
// Create a new workspace; caller becomes its owner.
// ─────────────────────────────────────────────────────────────────────────────

func CreateWorkspace(c *fiber.Ctx) error {
	uid, _ := uuid.Parse(c.Locals("userID").(string))

	type body struct {
		Name string `json:"name"`
	}
	req := new(body)
	if err := c.BodyParser(req); err != nil {
		return badRequest(c, "Invalid request body")
	}
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		return badRequest(c, "Workspace name is required")
	}
	if len(req.Name) > 80 {
		return badRequest(c, "Workspace name must be 80 characters or fewer")
	}

	if err := provisionWorkspace(uid, req.Name); err != nil {
		return serverError(c, "Failed to create workspace")
	}

	// Return the newly created workspace + membership.
	var mem models.WorkspaceMember
	database.DB.Joins("JOIN workspaces ON workspaces.id = workspace_members.workspace_id").
		Where("workspace_members.user_id = ? AND workspaces.name = ? AND workspace_members.role = 'owner'", uid, req.Name).
		Order("workspace_members.joined_at DESC").
		First(&mem)

	var ws models.Workspace
	database.DB.First(&ws, "id = ?", mem.WorkspaceID)

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message":   "Workspace created",
		"workspace": workspaceSummary(ws, "owner"),
	})
}

// ─────────────────────────────────────────────────────────────────────────────
// GET /api/v1/workspaces/:id
// Return a single workspace with members. Caller must be a member.
// ─────────────────────────────────────────────────────────────────────────────

func GetWorkspace(c *fiber.Ctx) error {
	uid, _ := uuid.Parse(c.Locals("userID").(string))

	// Resolve which workspace to load.
	// Priority: URL param :id → else the caller's first (oldest) workspace.
	var myMembership models.WorkspaceMember

	if paramID := c.Params("id"); paramID != "" {
		wsID, err := uuid.Parse(paramID)
		if err != nil {
			return badRequest(c, "Invalid workspace ID")
		}
		if tx := database.DB.Where("workspace_id = ? AND user_id = ?", wsID, uid).First(&myMembership); tx.Error != nil {
			return c.Status(fiber.StatusForbidden).JSON(errJSON("Workspace not found or access denied"))
		}
	} else {
		// Legacy / default path — find any membership for this user.
		if tx := database.DB.Where("user_id = ?", uid).First(&myMembership); tx.Error != nil {
			// Lazy-create for pre-migration accounts.
			var user models.User
			if database.DB.First(&user, "id = ?", uid).Error != nil {
				return c.Status(fiber.StatusNotFound).JSON(errJSON("User not found"))
			}
			if err := createWorkspaceForUser(user); err != nil {
				return serverError(c, "Failed to create workspace")
			}
			if database.DB.Where("user_id = ?", uid).First(&myMembership).Error != nil {
				return serverError(c, "Workspace created but could not be retrieved")
			}
		}
	}

	var ws models.Workspace
	if tx := database.DB.
		Preload("Members.User").
		Where("id = ?", myMembership.WorkspaceID).
		First(&ws); tx.Error != nil {
		return c.Status(fiber.StatusNotFound).JSON(errJSON("Workspace not found"))
	}

	membersOut := make([]fiber.Map, len(ws.Members))
	for i, m := range ws.Members {
		membersOut[i] = memberJSON(m)
	}

	return c.JSON(fiber.Map{
		"workspace": fiber.Map{
			"id":         ws.ID,
			"name":       ws.Name,
			"owner_id":   ws.OwnerID,
			"my_role":    myMembership.Role,
			"members":    membersOut,
			"created_at": ws.CreatedAt,
		},
	})
}

// ─────────────────────────────────────────────────────────────────────────────
// POST /api/v1/workspaces/:id/members
// Add a member to a specific workspace. Only owners may call this.
// ─────────────────────────────────────────────────────────────────────────────

func AddMember(c *fiber.Ctx) error {
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
		return c.Status(fiber.StatusForbidden).JSON(errJSON("Only workspace owners can add members"))
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

	var target models.User
	if tx := database.DB.Where("email = ?", req.Email).First(&target); tx.Error != nil {
		return c.Status(fiber.StatusNotFound).JSON(errJSON("No account found with that email"))
	}

	var existing models.WorkspaceMember
	if tx := database.DB.
		Where("workspace_id = ? AND user_id = ?", wsID, target.ID).
		First(&existing); tx.Error == nil {
		return c.Status(fiber.StatusConflict).JSON(errJSON("User is already a workspace member"))
	}

	newMem := models.WorkspaceMember{
		WorkspaceID: wsID,
		UserID:      target.ID,
		Role:        req.Role,
	}
	if err := database.DB.Create(&newMem).Error; err != nil {
		return serverError(c, "Failed to add member")
	}
	database.DB.Preload("User").First(&newMem, "id = ?", newMem.ID)

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Member added successfully",
		"member":  memberJSON(newMem),
	})
}

// ─────────────────────────────────────────────────────────────────────────────
// DELETE /api/v1/workspaces/:id/members/:userID
// Remove a non-owner member. Only workspace owners may call this.
// ─────────────────────────────────────────────────────────────────────────────

func RemoveMember(c *fiber.Ctx) error {
	uid, _ := uuid.Parse(c.Locals("userID").(string))
	wsID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return badRequest(c, "Invalid workspace ID")
	}
	tid, err := uuid.Parse(c.Params("userID"))
	if err != nil {
		return badRequest(c, "Invalid user ID")
	}
	if uid == tid {
		return badRequest(c, "You cannot remove yourself from the workspace")
	}

	var requester models.WorkspaceMember
	if tx := database.DB.Where("workspace_id = ? AND user_id = ?", wsID, uid).First(&requester); tx.Error != nil {
		return c.Status(fiber.StatusForbidden).JSON(errJSON("Workspace not found or access denied"))
	}
	if requester.Role != "owner" {
		return c.Status(fiber.StatusForbidden).JSON(errJSON("Only workspace owners can remove members"))
	}

	result := database.DB.
		Where("workspace_id = ? AND user_id = ? AND role != 'owner'", wsID, tid).
		Delete(&models.WorkspaceMember{})

	if result.Error != nil {
		return serverError(c, "Failed to remove member")
	}
	if result.RowsAffected == 0 {
		return c.Status(fiber.StatusNotFound).JSON(errJSON("Member not found or cannot remove an owner"))
	}

	return c.JSON(fiber.Map{"message": "Member removed"})
}
