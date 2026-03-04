package routes

import (
	"outcraftly/accounts/handlers"
	"outcraftly/accounts/middleware"

	"github.com/gofiber/fiber/v2"
)

// Setup registers all API routes on the Fiber app.
func Setup(app *fiber.App) {
	api := app.Group("/api/v1")

	// Health check
	api.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "ok",
			"service": "Outcraftly Accounts",
		})
	})

	// ── Public auth routes ──────────────────────────────────────
	auth := api.Group("/auth")
	auth.Post("/register",        handlers.Register)
	auth.Get("/verify-email",     handlers.VerifyEmail)
	auth.Post("/login",           handlers.Login)
	auth.Post("/logout",          handlers.Logout)
	auth.Post("/forgot-password", handlers.ForgotPassword)
	auth.Post("/reset-password",  handlers.ResetPassword)

	// ── Protected routes (require valid JWT) ────────────────────
	protected := api.Group("", middleware.Protected())
	protected.Get("/profile",               handlers.GetProfile)
	protected.Post("/profile",              handlers.UpdateProfile)
	protected.Post("/auth/change-password", handlers.ChangePassword)

	// ── Workspace routes ────────────────────────────────────────
	protected.Get("/workspaces",                              handlers.ListWorkspaces)
	protected.Post("/workspaces",                             handlers.CreateWorkspace)
	protected.Get("/workspaces/:id",                          handlers.GetWorkspace)
	protected.Post("/workspaces/:id/members",                 handlers.AddMember)
	protected.Delete("/workspaces/:id/members/:userID",       handlers.RemoveMember)
	// Legacy single-workspace route (kept for backwards compat)
	protected.Get("/workspace",                               handlers.GetWorkspace)

	// ── Invite routes ────────────────────────────────────────────
	// Protected (owner only — send/list/revoke)
	protected.Post("/workspaces/:id/invites",                 handlers.SendInvite)
	protected.Get("/workspaces/:id/invites",                  handlers.ListInvites)
	protected.Delete("/workspaces/:id/invites/:inviteID",     handlers.RevokeInvite)
	// Protected — accept (requires logged-in user whose email matches invite)
	protected.Post("/invites/accept",                         handlers.AcceptInvite)
	// Public — preview invite info (no auth, used before login/register)
	api.Get("/invites/preview",                               handlers.PreviewInvite)

	// ── Product registry ─────────────────────────────────────────
	// Any authenticated user — list active products
	protected.Get("/products", handlers.ListProducts)
	// Hidden admin panel — requires X-Admin-Secret header
	admin := api.Group("/admin", middleware.AdminOnly())
	admin.Get("/products",        handlers.AdminListProducts)
	admin.Post("/products",       handlers.CreateProduct)
	admin.Patch("/products/:id",  handlers.UpdateProduct)
	admin.Delete("/products/:id", handlers.DeactivateProduct)

	// ── Subscription routes ───────────────────────────────────────
	// Workspace-scoped (owner or member)
	protected.Post("/workspaces/:id/subscriptions",             handlers.CreateSubscription)
	protected.Get("/workspaces/:id/subscriptions",              handlers.ListSubscriptions)
	protected.Get("/workspaces/:id/subscriptions/access",       handlers.CheckAccess)
	protected.Delete("/workspaces/:id/subscriptions/:subID",    handlers.CancelSubscription)
	// Admin — filterable list of all subscriptions
	admin.Get("/subscriptions", handlers.AdminListSubscriptions)

        // ── Billing routes ────────────────────────────────────────────
	// Protected — workspace billing (owner only inside handlers)
	protected.Get("/workspaces/:id/billing",          handlers.GetBillingStatus)
	protected.Post("/workspaces/:id/billing/checkout", handlers.CreateCheckoutSession)
	protected.Post("/workspaces/:id/billing/portal",   handlers.CreatePortalSession)
	protected.Post("/workspaces/:id/billing/sync",     handlers.SyncBilling)
        // Admin — billing overview
        admin.Get("/billing", handlers.AdminBillingOverview)

}
