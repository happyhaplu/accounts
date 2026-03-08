package handlers

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"outcraftly/accounts/database"
	"outcraftly/accounts/models"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	stripe "github.com/stripe/stripe-go/v76"
	checkoutsession "github.com/stripe/stripe-go/v76/checkout/session"
	portalsession "github.com/stripe/stripe-go/v76/billingportal/session"
	stripecustomer "github.com/stripe/stripe-go/v76/customer"
	stripesubscription "github.com/stripe/stripe-go/v76/subscription"
	"github.com/stripe/stripe-go/v76/webhook"
)

// stripeReady returns false and logs a warning if the Stripe key is not set.
func stripeReady(c *fiber.Ctx) bool {
	if os.Getenv("STRIPE_SECRET_KEY") == "" {
		_ = c.Status(fiber.StatusServiceUnavailable).JSON(errJSON("Stripe is not configured on this server"))
		return false
	}
	return true
}

// ─────────────────────────────────────────────────────────────────────────────
// GET /api/v1/workspaces/:id/billing  (any member)
// Returns billing customer status + active subscriptions for the workspace.
// ─────────────────────────────────────────────────────────────────────────────

func GetBillingStatus(c *fiber.Ctx) error {
	uid, _ := uuid.Parse(c.Locals("userID").(string))
	wsID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return badRequest(c, "Invalid workspace ID")
	}
	if !isWorkspaceMember(uid, wsID) {
		return c.Status(fiber.StatusForbidden).JSON(errJSON("Access denied"))
	}

	var bc models.BillingCustomer
	hasCustomer := database.DB.Where("workspace_id = ?", wsID).First(&bc).Error == nil

	var subs []models.Subscription
	database.DB.Preload("Product").
		Where("workspace_id = ?", wsID).
		Order("created_at DESC").
		Find(&subs)

	// Auto-mark expired subscriptions.
	now := time.Now()
	for i := range subs {
		if subs[i].Status == "active" && now.After(subs[i].CurrentPeriodEnd) {
			subs[i].Status = "expired"
			database.DB.Model(&subs[i]).Update("status", "expired")
		}
	}

	return c.JSON(fiber.Map{
		"has_stripe_customer": hasCustomer,
		"stripe_customer_id":  func() string { if hasCustomer { return bc.StripeCustomerID }; return "" }(),
		"subscriptions":       subs,
	})
}

// ─────────────────────────────────────────────────────────────────────────────
// POST /api/v1/workspaces/:id/billing/checkout  (owner only)
// Creates a Stripe Checkout session and returns the URL.
// Body: { "price_id", "product_id", "plan_name", "success_url"?, "cancel_url"? }
// ─────────────────────────────────────────────────────────────────────────────

func CreateCheckoutSession(c *fiber.Ctx) error {
	if !stripeReady(c) {
		return nil
	}
	uid, _ := uuid.Parse(c.Locals("userID").(string))
	wsID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return badRequest(c, "Invalid workspace ID")
	}
	if !isWorkspaceOwner(uid, wsID) {
		return c.Status(fiber.StatusForbidden).JSON(errJSON("Only workspace owners can manage billing"))
	}

	type body struct {
		PriceID    string `json:"price_id"`
		ProductID  string `json:"product_id"`
		PlanName   string `json:"plan_name"`
		SuccessURL string `json:"success_url"`
		CancelURL  string `json:"cancel_url"`
	}
	req := new(body)
	if err := c.BodyParser(req); err != nil {
		return badRequest(c, "Invalid request body")
	}
	if req.PriceID == "" {
		return badRequest(c, "price_id is required")
	}
	if req.PlanName == "" {
		return badRequest(c, "plan_name is required")
	}

	appURL := os.Getenv("APP_URL")
	if req.SuccessURL == "" {
		req.SuccessURL = appURL + "/settings?billing=success"
	}
	if req.CancelURL == "" {
		req.CancelURL = appURL + "/settings?billing=canceled"
	}

	// Get or create workspace billing customer.
	stripeCustomerID, err := getOrCreateStripeCustomer(wsID, uid)
	if err != nil {
		return serverError(c, "Failed to set up billing customer")
	}

	stripe.Key = os.Getenv("STRIPE_SECRET_KEY")

	params := &stripe.CheckoutSessionParams{
		Customer:   stripe.String(stripeCustomerID),
		Mode:       stripe.String(string(stripe.CheckoutSessionModeSubscription)),
		SuccessURL: stripe.String(req.SuccessURL),
		CancelURL:  stripe.String(req.CancelURL),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{Price: stripe.String(req.PriceID), Quantity: stripe.Int64(1)},
		},
		SubscriptionData: &stripe.CheckoutSessionSubscriptionDataParams{
			Metadata: map[string]string{
				"workspace_id": wsID.String(),
				"product_id":  req.ProductID,
				"plan_name":   req.PlanName,
			},
		},
	}
	sess, err := checkoutsession.New(params)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[billing] checkout session error: %v\n", err)
		return serverError(c, "Failed to create checkout session")
	}
	return c.JSON(fiber.Map{"url": sess.URL})
}

// ─────────────────────────────────────────────────────────────────────────────
// POST /api/v1/workspaces/:id/billing/portal  (owner only)
// Creates a Stripe customer portal session and returns the URL.
// Body: { "return_url"? }
// ─────────────────────────────────────────────────────────────────────────────

func CreatePortalSession(c *fiber.Ctx) error {
	if !stripeReady(c) {
		return nil
	}
	uid, _ := uuid.Parse(c.Locals("userID").(string))
	wsID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return badRequest(c, "Invalid workspace ID")
	}
	if !isWorkspaceOwner(uid, wsID) {
		return c.Status(fiber.StatusForbidden).JSON(errJSON("Only workspace owners can access billing"))
	}

	var bc models.BillingCustomer
	if database.DB.Where("workspace_id = ?", wsID).First(&bc).Error != nil {
		return c.Status(fiber.StatusNotFound).JSON(errJSON("No billing account found for this workspace. Set up billing first."))
	}

	type body struct {
		ReturnURL string `json:"return_url"`
	}
	req := new(body)
	_ = c.BodyParser(req)
	if req.ReturnURL == "" {
		req.ReturnURL = os.Getenv("APP_URL") + "/settings"
	}

	stripe.Key = os.Getenv("STRIPE_SECRET_KEY")

	params := &stripe.BillingPortalSessionParams{
		Customer:  stripe.String(bc.StripeCustomerID),
		ReturnURL: stripe.String(req.ReturnURL),
	}
	sess, err := portalsession.New(params)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[billing] portal session error: %v\n", err)
		return serverError(c, "Failed to create billing portal session")
	}
	return c.JSON(fiber.Map{"url": sess.URL})
}

// ─────────────────────────────────────────────────────────────────────────────
// POST /api/v1/workspaces/:id/billing/sync  (owner only)
// Pulls the latest subscription state from Stripe and upserts into the DB.
// Call this after the user returns from a Stripe Checkout success redirect,
// or via the "Refresh" button — no webhook or CLI needed.
// ─────────────────────────────────────────────────────────────────────────────

func SyncBilling(c *fiber.Ctx) error {
	if !stripeReady(c) {
		return nil
	}
	uid, _ := uuid.Parse(c.Locals("userID").(string))
	wsID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return badRequest(c, "Invalid workspace ID")
	}
	if !isWorkspaceOwner(uid, wsID) {
		return c.Status(fiber.StatusForbidden).JSON(errJSON("Only workspace owners can sync billing"))
	}

	var bc models.BillingCustomer
	if database.DB.Where("workspace_id = ?", wsID).First(&bc).Error != nil {
		return c.JSON(fiber.Map{"synced": 0, "message": "No billing account found — complete a checkout first"})
	}

	stripe.Key = os.Getenv("STRIPE_SECRET_KEY")

	params := &stripe.SubscriptionListParams{}
	params.Customer = stripe.String(bc.StripeCustomerID)
	params.Filters.AddFilter("limit", "", "20")

	iter := stripesubscription.List(params)
	synced := 0
	for iter.Next() {
		s := iter.Subscription()
		ssID      := s.ID
		status    := stripeStatusToOurs(string(s.Status))
		periodEnd := time.Unix(s.CurrentPeriodEnd, 0)
		planName  := s.Metadata["plan_name"]
		if planName == "" {
			planName = "standard"
		}
		prodIDStr := s.Metadata["product_id"]
		wsIDStr   := s.Metadata["workspace_id"]

		// Skip subscriptions that belong to a different workspace.
		if wsIDStr != "" && wsIDStr != wsID.String() {
			continue
		}

		var sub models.Subscription
		if database.DB.Where("stripe_subscription_id = ?", ssID).First(&sub).Error == nil {
			// Already exists — just sync status and period.
			database.DB.Model(&sub).Updates(map[string]interface{}{
				"status":             status,
				"current_period_end": periodEnd,
				"plan_name":          planName,
			})
		} else if prodIDStr != "" {
			// New subscription — create a row.
			prodID, pErr := uuid.Parse(prodIDStr)
			if pErr == nil {
				// Cancel old active subs for same workspace+product.
				database.DB.Model(&models.Subscription{}).
					Where("workspace_id = ? AND product_id = ? AND status = 'active'", wsID, prodID).
					Update("status", "canceled")

				newSub := models.Subscription{
					WorkspaceID:          wsID,
					ProductID:            prodID,
					PlanName:             planName,
					Status:               status,
					CurrentPeriodEnd:     periodEnd,
					StripeSubscriptionID: &ssID,
				}
				database.DB.Create(&newSub)
			}
		}
		synced++
	}

	if err := iter.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "[billing] sync error for customer %s: %v\n", bc.StripeCustomerID, err)
		return serverError(c, "Failed to sync with Stripe")
	}

	fmt.Printf("[billing] synced %d subscriptions for workspace %s\n", synced, wsID)

	// Return fresh billing status.
	var subs []models.Subscription
	database.DB.Preload("Product").
		Where("workspace_id = ?", wsID).
		Order("created_at DESC").
		Find(&subs)

	return c.JSON(fiber.Map{
		"synced":              synced,
		"has_stripe_customer": true,
		"stripe_customer_id":  bc.StripeCustomerID,
		"subscriptions":       subs,
	})
}

// ─────────────────────────────────────────────────────────────────────────────
// GET /api/v1/admin/billing  (admin only)
// Overview of all billing customers and Stripe-linked subscriptions.
// ─────────────────────────────────────────────────────────────────────────────

func AdminBillingOverview(c *fiber.Ctx) error {
	var customers []models.BillingCustomer
	database.DB.Preload("Workspace").Order("created_at DESC").Find(&customers)

	var subs []models.Subscription
	database.DB.Preload("Workspace").Preload("Product").
		Where("stripe_subscription_id IS NOT NULL").
		Order("created_at DESC").
		Limit(200).
		Find(&subs)

	return c.JSON(fiber.Map{
		"billing_customers":    customers,
		"stripe_subscriptions": subs,
	})
}

// ─────────────────────────────────────────────────────────────────────────────
// Helpers
// ─────────────────────────────────────────────────────────────────────────────

// getOrCreateStripeCustomer returns the Stripe customer ID for the workspace,
// creating one on Stripe (and a BillingCustomer row) if it doesn't exist yet.
func getOrCreateStripeCustomer(wsID, userID uuid.UUID) (string, error) {
	var bc models.BillingCustomer
	if database.DB.Where("workspace_id = ?", wsID).First(&bc).Error == nil {
		return bc.StripeCustomerID, nil
	}

	// Load workspace + owner info for the Stripe customer record.
	var ws models.Workspace
	database.DB.First(&ws, "id = ?", wsID)
	var owner models.User
	database.DB.First(&owner, "id = ?", userID)

	stripe.Key = os.Getenv("STRIPE_SECRET_KEY")

	custParams := &stripe.CustomerParams{
		Name:  stripe.String(ws.Name),
		Email: stripe.String(owner.Email),
		Metadata: map[string]string{
			"workspace_id": wsID.String(),
		},
	}
	cust, err := stripecustomer.New(custParams)
	if err != nil {
		return "", fmt.Errorf("stripe customer create: %w", err)
	}

	bc = models.BillingCustomer{
		WorkspaceID:      wsID,
		StripeCustomerID: cust.ID,
	}
	database.DB.Create(&bc)
	return cust.ID, nil
}

func stripeStatusToOurs(stripeStatus string) string {
	switch strings.ToLower(stripeStatus) {
	case "active", "trialing":
		return "active"
	case "canceled":
		return "canceled"
	default: // past_due, unpaid, incomplete, incomplete_expired
		return "expired"
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// POST /api/v1/billing/webhook  (PUBLIC — Stripe signs the payload)
// Handles Stripe event notifications so subscriptions are provisioned
// automatically — no manual "Refresh" button needed.
//
// Events handled:
//   checkout.session.completed       → provision subscription after payment
//   customer.subscription.updated    → sync status + renewal date
//   customer.subscription.deleted    → mark canceled
//   invoice.payment_failed           → mark subscription expired
//
// Set STRIPE_WEBHOOK_SECRET to the whsec_... value from:
//   stripe listen --forward-to localhost:8080/api/v1/billing/webhook
// or from your Stripe Dashboard > Webhooks.
// ─────────────────────────────────────────────────────────────────────────────

func HandleStripeWebhook(c *fiber.Ctx) error {
	secret := os.Getenv("STRIPE_WEBHOOK_SECRET")
	payload := c.Body()

	var event stripe.Event

	if secret != "" {
		// Verify the Stripe-Signature header — prevents spoofed webhooks.
		// IgnoreAPIVersionMismatch: account may be on an older Stripe API version
		// than the SDK expects; signature verification is what matters for security.
		sig := c.Get("Stripe-Signature")
		var err error
		event, err = webhook.ConstructEventWithOptions(payload, sig, secret,
			webhook.ConstructEventOptions{IgnoreAPIVersionMismatch: true},
		)
		if err != nil {
			fmt.Fprintf(os.Stderr, "[webhook] invalid signature: %v\n", err)
			return c.Status(fiber.StatusBadRequest).JSON(errJSON("Invalid webhook signature"))
		}
	} else {
		// STRIPE_WEBHOOK_SECRET not set — accept without verification (dev/testing only).
		fmt.Println("[webhook] ⚠  STRIPE_WEBHOOK_SECRET not set — signature verification skipped")
		if err := json.Unmarshal(payload, &event); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(errJSON("Invalid webhook payload"))
		}
	}

	stripe.Key = os.Getenv("STRIPE_SECRET_KEY")

	switch event.Type {

	case "checkout.session.completed":
		var session stripe.CheckoutSession
		if err := json.Unmarshal(event.Data.Raw, &session); err != nil {
			fmt.Fprintf(os.Stderr, "[webhook] parse checkout.session.completed: %v\n", err)
			break
		}
		webhookCheckoutCompleted(&session)

	case "customer.subscription.updated":
		var sub stripe.Subscription
		if err := json.Unmarshal(event.Data.Raw, &sub); err != nil {
			fmt.Fprintf(os.Stderr, "[webhook] parse subscription.updated: %v\n", err)
			break
		}
		webhookSyncSubscription(&sub)

	case "customer.subscription.deleted":
		var sub stripe.Subscription
		if err := json.Unmarshal(event.Data.Raw, &sub); err != nil {
			fmt.Fprintf(os.Stderr, "[webhook] parse subscription.deleted: %v\n", err)
			break
		}
		webhookMarkCanceled(sub.ID)

	case "invoice.payment_failed":
		var inv stripe.Invoice
		if err := json.Unmarshal(event.Data.Raw, &inv); err != nil {
			fmt.Fprintf(os.Stderr, "[webhook] parse invoice.payment_failed: %v\n", err)
			break
		}
		if inv.Subscription != nil {
			webhookMarkCanceled(inv.Subscription.ID)
		}

	default:
		fmt.Printf("[webhook] unhandled event type: %s\n", event.Type)
	}

	return c.SendStatus(fiber.StatusOK)
}

// webhookCheckoutCompleted is called after a successful Stripe Checkout.
// It fetches the subscription from Stripe and provisions it in our DB.
func webhookCheckoutCompleted(session *stripe.CheckoutSession) {
	if session.Subscription == nil {
		// Payment-mode (one-time charge) — no subscription object created.
		fmt.Printf("[webhook] checkout.session.completed for payment mode session %s — no sub to provision\n", session.ID)
		return
	}

	// Fetch the full subscription to get metadata + period details.
	sub, err := stripesubscription.Get(session.Subscription.ID, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[webhook] fetch subscription %s: %v\n", session.Subscription.ID, err)
		return
	}
	webhookSyncSubscription(sub)
}

// webhookSyncSubscription upserts a Subscription row from a Stripe Subscription object.
func webhookSyncSubscription(s *stripe.Subscription) {
	ssID      := s.ID
	status    := stripeStatusToOurs(string(s.Status))
	periodEnd := time.Unix(s.CurrentPeriodEnd, 0)
	planName  := s.Metadata["plan_name"]
	if planName == "" {
		planName = "standard"
	}
	prodIDStr := s.Metadata["product_id"]
	wsIDStr   := s.Metadata["workspace_id"]

	// Update an existing row if we already have it.
	var existing models.Subscription
	if database.DB.Where("stripe_subscription_id = ?", ssID).First(&existing).Error == nil {
		database.DB.Model(&existing).Updates(map[string]interface{}{
			"status":             status,
			"current_period_end": periodEnd,
			"plan_name":          planName,
		})
		fmt.Printf("[webhook] updated subscription %s → %s\n", ssID, status)
		return
	}

	// New subscription row — metadata must carry workspace_id + product_id.
	if wsIDStr == "" || prodIDStr == "" {
		fmt.Printf("[webhook] sub %s missing metadata (ws=%q prod=%q) — cannot provision\n", ssID, wsIDStr, prodIDStr)
		return
	}

	wsID, wErr := uuid.Parse(wsIDStr)
	prodID, pErr := uuid.Parse(prodIDStr)
	if wErr != nil || pErr != nil {
		fmt.Printf("[webhook] sub %s invalid UUID metadata — skip\n", ssID)
		return
	}

	// Ensure BillingCustomer row exists for this workspace.
	var bc models.BillingCustomer
	if database.DB.Where("workspace_id = ?", wsID).First(&bc).Error != nil {
		if s.Customer != nil {
			bc = models.BillingCustomer{
				WorkspaceID:      wsID,
				StripeCustomerID: s.Customer.ID,
			}
			database.DB.Create(&bc)
		}
	}

	// Deactivate any previous active subscription for the same product.
	database.DB.Model(&models.Subscription{}).
		Where("workspace_id = ? AND product_id = ? AND status = 'active'", wsID, prodID).
		Update("status", "canceled")

	newSub := models.Subscription{
		WorkspaceID:          wsID,
		ProductID:            prodID,
		PlanName:             planName,
		Status:               status,
		CurrentPeriodEnd:     periodEnd,
		StripeSubscriptionID: &ssID,
	}
	database.DB.Create(&newSub)
	fmt.Printf("[webhook] provisioned subscription %s for workspace=%s product=%s\n", ssID, wsIDStr, prodIDStr)
}

// webhookMarkCanceled sets the matching subscription's status to "canceled".
func webhookMarkCanceled(stripeSubID string) {
	result := database.DB.
		Model(&models.Subscription{}).
		Where("stripe_subscription_id = ?", stripeSubID).
		Update("status", "canceled")
	fmt.Printf("[webhook] canceled subscription %s (%d rows affected)\n", stripeSubID, result.RowsAffected)
}
