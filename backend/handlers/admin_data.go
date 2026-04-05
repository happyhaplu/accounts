package handlers

import (
	"time"

	"outcraftly/accounts/database"
	"outcraftly/accounts/models"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ─────────────────────────────────────────────────────────────────────────────
// DELETE /api/v1/admin/users/purge-unverified  (admin only)
//
// Permanently deletes all users whose email has never been verified AND who
// have no subscription history.  Users with any subscription row (even
// expired/canceled) are left untouched to preserve billing history.
//
// Cascade order inside a single transaction:
//   workspace_invites → workspace_members → workspaces → billing_customers → users
//
// Returns: { "deleted": N }
// ─────────────────────────────────────────────────────────────────────────────

func AdminPurgeUnverifiedUsers(c *fiber.Ctx) error {
	// ── 1. Find user IDs that have any subscription (must never be deleted) ───
	type idRow struct{ UserID uuid.UUID `gorm:"column:user_id"` }
	var subUserRows []idRow
	database.DB.Table("workspace_members").
		Select("DISTINCT workspace_members.user_id").
		Joins("JOIN subscriptions ON subscriptions.workspace_id = workspace_members.workspace_id").
		Where("workspace_members.role = 'owner'").
		Scan(&subUserRows)

	protectedIDs := make([]uuid.UUID, 0, len(subUserRows))
	for _, r := range subUserRows {
		protectedIDs = append(protectedIDs, r.UserID)
	}

	// ── 2. Collect the user IDs we actually want to delete ────────────────────
	userQ := database.DB.Model(&models.User{}).Where("email_verified = false")
	if len(protectedIDs) > 0 {
		userQ = userQ.Where("id NOT IN ?", protectedIDs)
	}
	var targetIDs []uuid.UUID
	userQ.Pluck("id", &targetIDs)

	if len(targetIDs) == 0 {
		return c.JSON(fiber.Map{"deleted": 0})
	}

	// ── 3. Find workspace IDs owned by those users ────────────────────────────
	var wsIDs []uuid.UUID
	database.DB.Table("workspace_members").
		Where("user_id IN ? AND role = 'owner'", targetIDs).
		Pluck("workspace_id", &wsIDs)

	// ── 4. Cascade-delete inside a transaction ────────────────────────────────
	// FK order (must delete children before parents):
	//   workspace_invites  (FK → workspaces, FK → users via invited_by)
	//   workspace_members  (FK → workspaces, FK → users)
	//   subscriptions      (FK → workspaces)
	//   billing_customers  (FK → workspaces)
	//   workspaces         (FK → users via owner_id)
	//   users
	var deleted int64
	err := database.DB.Transaction(func(tx *gorm.DB) error {
		if len(wsIDs) > 0 {
			// 4a. workspace_invites
			if err := tx.Where("workspace_id IN ?", wsIDs).
				Delete(&models.WorkspaceInvite{}).Error; err != nil {
				return err
			}
			// 4b. workspace_members for these workspaces (member rows of other users)
			if err := tx.Where("workspace_id IN ?", wsIDs).
				Delete(&models.WorkspaceMember{}).Error; err != nil {
				return err
			}
			// 4c. subscriptions (FK → workspaces — must go before workspaces)
			if err := tx.Where("workspace_id IN ?", wsIDs).
				Delete(&models.Subscription{}).Error; err != nil {
				return err
			}
			// 4d. billing_customers (FK → workspaces — must go before workspaces)
			if err := tx.Where("workspace_id IN ?", wsIDs).
				Delete(&models.BillingCustomer{}).Error; err != nil {
				return err
			}
			// 4e. workspaces
			if err := tx.Where("id IN ?", wsIDs).
				Delete(&models.Workspace{}).Error; err != nil {
				return err
			}
		}

		// 4f. any remaining workspace_member rows for the target users
		//     (cases where they were a member of someone else's workspace)
		if err := tx.Where("user_id IN ?", targetIDs).
			Delete(&models.WorkspaceMember{}).Error; err != nil {
			return err
		}

		// 4g. workspace_invites sent by the target users (invited_by FK → users)
		if err := tx.Where("invited_by IN ?", targetIDs).
			Delete(&models.WorkspaceInvite{}).Error; err != nil {
			return err
		}

		// 4h. users — safe now that all FK children are gone
		result := tx.Where("id IN ?", targetIDs).Delete(&models.User{})
		if result.Error != nil {
			return result.Error
		}
		deleted = result.RowsAffected
		return nil
	})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Failed to purge unverified users",
			"details": err.Error(),
		})
	}

	return c.JSON(fiber.Map{"deleted": deleted})
}

// ─────────────────────────────────────────────────────────────────────────────
// GET /api/v1/admin/users  (admin only)
//
// Returns all registered users, each enriched with their subscription status
// across all owner workspaces. Does NOT return sensitive fields (password hash,
// OTP codes, reset tokens).
//
// Query params:
//   q  — optional substring filter (matches email or name, case-insensitive)
// ─────────────────────────────────────────────────────────────────────────────

func AdminListUsers(c *fiber.Ctx) error {
	// ── 1. Fetch users ────────────────────────────────────────────────────────
	var users []models.User
	q := database.DB.Order("created_at DESC")
	if search := c.Query("q"); search != "" {
		like := "%" + search + "%"
		q = q.Where("email ILIKE ? OR name ILIKE ?", like, like)
	}
	q.Find(&users)

	// ── 2. Fetch subscriptions for every owner workspace in one query ─────────
	type subRow struct {
		UserID           uuid.UUID `gorm:"column:user_id"`
		ProductName      string    `gorm:"column:product_name"`
		Status           string    `gorm:"column:status"`
		CurrentPeriodEnd time.Time `gorm:"column:current_period_end"`
	}
	var subRows []subRow
	database.DB.Table("workspace_members").
		Select("workspace_members.user_id, products.name AS product_name, subscriptions.status, subscriptions.current_period_end").
		Joins("JOIN subscriptions ON subscriptions.workspace_id = workspace_members.workspace_id").
		Joins("JOIN products ON products.id = subscriptions.product_id").
		Where("workspace_members.role = 'owner'").
		Scan(&subRows)

	type subSummary struct {
		ProductName      string    `json:"product_name"`
		Status           string    `json:"status"`
		CurrentPeriodEnd time.Time `json:"current_period_end"`
	}
	subsByUser := map[uuid.UUID][]subSummary{}
	for _, r := range subRows {
		subsByUser[r.UserID] = append(subsByUser[r.UserID], subSummary{
			ProductName:      r.ProductName,
			Status:           r.Status,
			CurrentPeriodEnd: r.CurrentPeriodEnd,
		})
	}

	// ── 3. Build response (no sensitive fields) ───────────────────────────────
	type userResp struct {
		ID              uuid.UUID    `json:"id"`
		Email           string       `json:"email"`
		Name            string       `json:"name"`
		CompanyName     string       `json:"company_name"`
		JobTitle        string       `json:"job_title"`
		EmailVerified   bool         `json:"email_verified"`
		ProfileComplete bool         `json:"profile_complete"`
		Role            string       `json:"role"`
		LastLoginAt     *time.Time   `json:"last_login_at"`
		CreatedAt       time.Time    `json:"created_at"`
		Subscriptions   []subSummary `json:"subscriptions"`
	}
	result := make([]userResp, 0, len(users))
	for _, u := range users {
		subs := subsByUser[u.ID]
		if subs == nil {
			subs = []subSummary{}
		}
		result = append(result, userResp{
			ID:              u.ID,
			Email:           u.Email,
			Name:            u.Name,
			CompanyName:     u.CompanyName,
			JobTitle:        u.JobTitle,
			EmailVerified:   u.EmailVerified,
			ProfileComplete: u.ProfileComplete,
			Role:            u.Role,
			LastLoginAt:     u.LastLoginAt,
			CreatedAt:       u.CreatedAt,
			Subscriptions:   subs,
		})
	}

	return c.JSON(fiber.Map{"total": len(result), "users": result})
}

// ─────────────────────────────────────────────────────────────────────────────
// GET /api/v1/admin/workspaces  (admin only)
//
// Returns all workspaces enriched with owner info, member count, and
// per-product subscription status.
//
// Query params:
//   q  — optional substring filter (matches workspace name or owner email)
// ─────────────────────────────────────────────────────────────────────────────

func AdminListWorkspaces(c *fiber.Ctx) error {
	// ── 1. Fetch workspaces ───────────────────────────────────────────────────
	var workspaces []models.Workspace
	database.DB.Order("created_at DESC").Find(&workspaces)

	// ── 2. Fetch owner email + name per workspace ─────────────────────────────
	type ownerRow struct {
		WorkspaceID uuid.UUID `gorm:"column:workspace_id"`
		UserID      uuid.UUID `gorm:"column:user_id"`
		Email       string    `gorm:"column:email"`
		Name        string    `gorm:"column:name"`
	}
	var owners []ownerRow
	database.DB.Table("workspace_members").
		Select("workspace_members.workspace_id, workspace_members.user_id, users.email, users.name").
		Joins("JOIN users ON users.id = workspace_members.user_id").
		Where("workspace_members.role = 'owner'").
		Scan(&owners)

	ownerByWS := map[uuid.UUID]ownerRow{}
	for _, o := range owners {
		ownerByWS[o.WorkspaceID] = o
	}

	// ── 3. Fetch member counts per workspace ──────────────────────────────────
	type memberCount struct {
		WorkspaceID uuid.UUID `gorm:"column:workspace_id"`
		Count       int       `gorm:"column:count"`
	}
	var counts []memberCount
	database.DB.Table("workspace_members").
		Select("workspace_id, COUNT(*) AS count").
		Group("workspace_id").
		Scan(&counts)

	countByWS := map[uuid.UUID]int{}
	for _, mc := range counts {
		countByWS[mc.WorkspaceID] = mc.Count
	}

	// ── 4. Fetch subscriptions with products ──────────────────────────────────
	type wsSubRow struct {
		WorkspaceID      uuid.UUID `gorm:"column:workspace_id"`
		ProductName      string    `gorm:"column:product_name"`
		Status           string    `gorm:"column:status"`
		CurrentPeriodEnd time.Time `gorm:"column:current_period_end"`
	}
	var wsSubRows []wsSubRow
	database.DB.Table("subscriptions").
		Select("subscriptions.workspace_id, products.name AS product_name, subscriptions.status, subscriptions.current_period_end").
		Joins("JOIN products ON products.id = subscriptions.product_id").
		Scan(&wsSubRows)

	type subSummary struct {
		ProductName      string    `json:"product_name"`
		Status           string    `json:"status"`
		CurrentPeriodEnd time.Time `json:"current_period_end"`
	}
	subsByWS := map[uuid.UUID][]subSummary{}
	for _, r := range wsSubRows {
		subsByWS[r.WorkspaceID] = append(subsByWS[r.WorkspaceID], subSummary{
			ProductName:      r.ProductName,
			Status:           r.Status,
			CurrentPeriodEnd: r.CurrentPeriodEnd,
		})
	}

	// ── 5. Build response ─────────────────────────────────────────────────────
	type wsResp struct {
		ID            uuid.UUID    `json:"id"`
		Name          string       `json:"name"`
		OwnerID       uuid.UUID    `json:"owner_id"`
		OwnerEmail    string       `json:"owner_email"`
		OwnerName     string       `json:"owner_name"`
		MemberCount   int          `json:"member_count"`
		CreatedAt     time.Time    `json:"created_at"`
		Subscriptions []subSummary `json:"subscriptions"`
	}
	result := make([]wsResp, 0, len(workspaces))
	for _, ws := range workspaces {
		owner := ownerByWS[ws.ID]
		subs := subsByWS[ws.ID]
		if subs == nil {
			subs = []subSummary{}
		}

		// Apply optional search filter post-fetch (simpler than SQL ILIKE join)
		if search := c.Query("q"); search != "" {
			like := search
			matchName  := containsCI(ws.Name, like)
			matchEmail := containsCI(owner.Email, like)
			if !matchName && !matchEmail {
				continue
			}
		}

		result = append(result, wsResp{
			ID:            ws.ID,
			Name:          ws.Name,
			OwnerID:       owner.UserID,
			OwnerEmail:    owner.Email,
			OwnerName:     owner.Name,
			MemberCount:   countByWS[ws.ID],
			CreatedAt:     ws.CreatedAt,
			Subscriptions: subs,
		})
	}

	return c.JSON(fiber.Map{"total": len(result), "workspaces": result})
}

// containsCI is a simple case-insensitive substring check.
func containsCI(s, sub string) bool {
	if sub == "" {
		return true
	}
	sLow  := toLower(s)
	subLow := toLower(sub)
	return len(sLow) >= len(subLow) && containsStr(sLow, subLow)
}

func toLower(s string) string {
	b := make([]byte, len(s))
	for i := range s {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c += 'a' - 'A'
		}
		b[i] = c
	}
	return string(b)
}

func containsStr(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
