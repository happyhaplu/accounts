package routes

import (
	"outcraftly/accounts/handlers"
	"outcraftly/accounts/middleware"

	"github.com/gofiber/fiber/v2"
)

// Setup registers all API routes on the Fiber app.
//
// Middleware strategy
// ───────────────────
// Fiber's Group(prefix, handlers...) registers those handlers as USE
// middleware for "<prefix>/*", meaning they fire for every request whose
// path starts with that prefix — regardless of which group the concrete
// route was added to later.
//
// Therefore we NEVER use an empty-string prefix group for JWT protection:
//   api.Group("", Protected())   ← would match /api/v1/admin/* too!
//
// Instead each logical section owns its own path prefix and carries only
// the middleware it needs.
func Setup(app *fiber.App) {
	api := app.Group("/api/v1")

	// ── Health check ─────────────────────────────────────────────
	api.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "ok",
			"service": "Gour Accounts",
		})
	})

	// ── Public config (Stripe publishable key etc.) ───────────────
	api.Get("/config", handlers.GetConfig)

	// ── Public auth routes ────────────────────────────────────────
	auth := api.Group("/auth")
	auth.Post("/register",              handlers.Register)
        auth.Post("/verify-email-otp",      handlers.VerifyEmailOTP)
        auth.Post("/resend-verification",   handlers.ResendVerification)
        auth.Post("/login",                 handlers.Login)
        auth.Post("/logout",                handlers.Logout)
        auth.Post("/forgot-password",       handlers.ForgotPassword)
        auth.Post("/verify-reset-otp",      handlers.VerifyResetOTP)
        auth.Post("/reset-password",        handlers.ResetPassword)
	// stripe listen --forward-to localhost:8080/api/v1/billing/webhook
	api.Post("/billing/webhook", handlers.HandleStripeWebhook)

	// ── Public invite preview ─────────────────────────────────────
	api.Get("/invites/preview", handlers.PreviewInvite)

	// ── Admin login (public — no secret needed to call this) ──────
	api.Post("/admin/auth/login", handlers.AdminLogin)

	// ── Admin routes (X-Admin-Secret only, no JWT) ────────────────
	// Uses its own distinct prefix "/admin" so its USE middleware does
	// NOT overlap with any user-facing prefix below.
	admin := api.Group("/admin", middleware.AdminOnly())
	admin.Get("/products",                    handlers.AdminListProducts)
	admin.Post("/products",                   handlers.CreateProduct)
	admin.Patch("/products/:id",              handlers.UpdateProduct)
	admin.Delete("/products/:id",             handlers.DeactivateProduct)
	admin.Delete("/products/:id/permanent",   handlers.PermanentDeleteProduct)
	admin.Post("/products/:id/regenerate-key", handlers.RegenerateProductAPIKey)
	admin.Get("/subscriptions",               handlers.AdminListSubscriptions)
	admin.Get("/billing",                     handlers.AdminBillingOverview)
	admin.Get("/users",                       handlers.AdminListUsers)
	admin.Delete("/users/purge-unverified",   handlers.AdminPurgeUnverifiedUsers)
	admin.Get("/workspaces",                  handlers.AdminListWorkspaces)

	// ── Protected: profile ────────────────────────────────────────
	// Inline middleware so we avoid a catch-all USE handler.
	p := middleware.Protected()
	api.Get("/profile",               p, handlers.GetProfile)
	api.Post("/profile",              p, handlers.UpdateProfile)
	api.Post("/auth/change-password", p, handlers.ChangePassword)
	// Legacy single-workspace lookup
	api.Get("/workspace", p, handlers.GetWorkspace)
	// Products list (authenticated users)
	api.Get("/products", p, handlers.ListProducts)
	// Product launch & subscription check (for external Gour apps)
	api.Get("/products/:name/launch", p, handlers.LaunchProduct)
	api.Get("/products/:name/check",     handlers.CheckProductSubscription)
	// Server-to-server token verification (X-API-Key, no JWT needed)
	api.Post("/products/verify",         handlers.VerifyToken)
	// Accept invite (user must be logged in)
	api.Post("/invites/accept", p, handlers.AcceptInvite)

	// ── Protected: workspace routes ───────────────────────────────
	ws := api.Group("/workspaces", middleware.Protected())
	ws.Get("",                           handlers.ListWorkspaces)
	ws.Post("",                          handlers.CreateWorkspace)
	ws.Get("/:id",                       handlers.GetWorkspace)
	ws.Post("/:id/members",              handlers.AddMember)
	ws.Delete("/:id/members/:userID",    handlers.RemoveMember)
	// Invites (workspace-scoped)
	ws.Post("/:id/invites",              handlers.SendInvite)
	ws.Get("/:id/invites",               handlers.ListInvites)
	ws.Delete("/:id/invites/:inviteID",  handlers.RevokeInvite)
	// Subscriptions
	ws.Post("/:id/subscriptions",        handlers.CreateSubscription)
	ws.Get("/:id/subscriptions",         handlers.ListSubscriptions)
	ws.Get("/:id/subscriptions/access",  handlers.CheckAccess)
	ws.Delete("/:id/subscriptions/:subID", handlers.CancelSubscription)
	// Billing
	ws.Get("/:id/billing",               handlers.GetBillingStatus)
	ws.Post("/:id/billing/portal",       handlers.CreatePortalSession)
	ws.Post("/:id/billing/sync",         handlers.SyncBilling)
}

