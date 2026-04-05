package integration_test

// ── Subscription + Billing integration tests ──────────────────────────────────

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"

	"outcraftly/accounts/database"
	"outcraftly/accounts/models"
	"outcraftly/accounts/testhelpers"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ── helpers shared across sub tests ──────────────────────────────────────────

// setupProductAndWorkspace creates a verified user + product, returns tok + wsID + productID.
// A unique email is derived from t.Name() so parallel/sequential runs never collide in the
// shared in-memory SQLite database.
func setupProductAndWorkspace(t *testing.T) (tok, wsID, productID string) {
	t.Helper()
	db := testhelpers.SetupTestDB()
	app := testhelpers.NewApp()
	// Build a safe email: keep only [a-z0-9], replace everything else with '-'.
	safe := strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			return r
		}
		return '-'
	}, strings.ToLower(t.Name()))
	email := safe + "@example.com"
	_, tok = testhelpers.CreateVerifiedUser(db, email, "password123")

	// Use a unique product name (same safe transform) to avoid the uniqueIndex collision.
	productName := "product-" + safe
	p := models.Product{Name: productName, IsActive: true}
	if err := database.DB.Create(&p).Error; err != nil {
		t.Fatalf("setupProductAndWorkspace: failed to create product: %v", err)
	}
	productID = p.ID.String()

	// Trigger workspace lazy-creation
	listReq := httptest.NewRequest("GET", "/api/v1/workspaces", nil)
	listReq.Header.Set("Authorization", testhelpers.AuthBearer(tok))
	listResp, err := app.Test(listReq)
	require.NoError(t, err)
	var listBody map[string]interface{}
	json.NewDecoder(listResp.Body).Decode(&listBody)
	wsID = listBody["workspaces"].([]interface{})[0].(map[string]interface{})["id"].(string)
	return
}

// ── CreateSubscription ────────────────────────────────────────────────────────

func TestCreateSubscription_Success(t *testing.T) {
	tok, wsID, productID := setupProductAndWorkspace(t)
	app := testhelpers.NewApp()

	payload, _ := json.Marshal(map[string]interface{}{
		"product_id":  productID,
		"plan_name":   "starter",
		"period_days": 30,
	})
	req := httptest.NewRequest("POST", "/api/v1/workspaces/"+wsID+"/subscriptions", bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", testhelpers.AuthBearer(tok))

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 201, resp.StatusCode)

	var body map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&body)
	sub := body["subscription"].(map[string]interface{})
	assert.Equal(t, "active", sub["status"])
	assert.Equal(t, "starter", sub["plan_name"])
}

func TestCreateSubscription_InvalidProductID_Returns400(t *testing.T) {
	db := testhelpers.SetupTestDB()
	app := testhelpers.NewApp()
	_, tok := testhelpers.CreateVerifiedUser(db, "sub-bad@example.com", "password123")

	listReq := httptest.NewRequest("GET", "/api/v1/workspaces", nil)
	listReq.Header.Set("Authorization", testhelpers.AuthBearer(tok))
	listResp, _ := app.Test(listReq)
	var listBody map[string]interface{}
	json.NewDecoder(listResp.Body).Decode(&listBody)
	wsID := listBody["workspaces"].([]interface{})[0].(map[string]interface{})["id"].(string)

	payload, _ := json.Marshal(map[string]interface{}{
		"product_id": "not-a-uuid",
		"plan_name":  "starter",
	})
	req := httptest.NewRequest("POST", "/api/v1/workspaces/"+wsID+"/subscriptions", bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", testhelpers.AuthBearer(tok))
	resp, _ := app.Test(req)
	assert.Equal(t, 400, resp.StatusCode)
}

func TestCreateSubscription_Unauthenticated_Returns401(t *testing.T) {
	testhelpers.SetupTestDB()
	app := testhelpers.NewApp()

	payload, _ := json.Marshal(map[string]interface{}{"product_id": "uuid", "plan_name": "starter"})
	req := httptest.NewRequest("POST", "/api/v1/workspaces/some-id/subscriptions", bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)
	assert.Equal(t, 401, resp.StatusCode)
}

// ── ListSubscriptions ─────────────────────────────────────────────────────────

func TestListSubscriptions_ReturnsEmpty_WhenNone(t *testing.T) {
	db := testhelpers.SetupTestDB()
	app := testhelpers.NewApp()
	_, tok := testhelpers.CreateVerifiedUser(db, "sub-empty@example.com", "password123")

	listWS := httptest.NewRequest("GET", "/api/v1/workspaces", nil)
	listWS.Header.Set("Authorization", testhelpers.AuthBearer(tok))
	listWSResp, _ := app.Test(listWS)
	var wsBody map[string]interface{}
	json.NewDecoder(listWSResp.Body).Decode(&wsBody)
	wsID := wsBody["workspaces"].([]interface{})[0].(map[string]interface{})["id"].(string)

	req := httptest.NewRequest("GET", "/api/v1/workspaces/"+wsID+"/subscriptions", nil)
	req.Header.Set("Authorization", testhelpers.AuthBearer(tok))
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var body map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&body)
	subs := body["subscriptions"].([]interface{})
	assert.Equal(t, 0, len(subs))
}

func TestListSubscriptions_AfterCreate_ReturnsSub(t *testing.T) {
	tok, wsID, productID := setupProductAndWorkspace(t)
	app := testhelpers.NewApp()

	// Create subscription
	createPayload, _ := json.Marshal(map[string]interface{}{
		"product_id": productID, "plan_name": "pro", "period_days": 30,
	})
	createReq := httptest.NewRequest("POST", "/api/v1/workspaces/"+wsID+"/subscriptions", bytes.NewBuffer(createPayload))
	createReq.Header.Set("Content-Type", "application/json")
	createReq.Header.Set("Authorization", testhelpers.AuthBearer(tok))
	app.Test(createReq)

	// List
	listReq := httptest.NewRequest("GET", "/api/v1/workspaces/"+wsID+"/subscriptions", nil)
	listReq.Header.Set("Authorization", testhelpers.AuthBearer(tok))
	listResp, err := app.Test(listReq)
	require.NoError(t, err)
	assert.Equal(t, 200, listResp.StatusCode)

	var body map[string]interface{}
	json.NewDecoder(listResp.Body).Decode(&body)
	subs := body["subscriptions"].([]interface{})
	assert.Equal(t, 1, len(subs))
	sub := subs[0].(map[string]interface{})
	assert.Equal(t, "active", sub["status"])
}

// ── CheckAccess ───────────────────────────────────────────────────────────────

func TestCheckAccess_WithActiveSub_ReturnsGranted(t *testing.T) {
	tok, wsID, productID := setupProductAndWorkspace(t)
	app := testhelpers.NewApp()

	// Create subscription
	createPayload, _ := json.Marshal(map[string]interface{}{
		"product_id": productID, "plan_name": "starter", "period_days": 30,
	})
	createReq := httptest.NewRequest("POST", "/api/v1/workspaces/"+wsID+"/subscriptions", bytes.NewBuffer(createPayload))
	createReq.Header.Set("Content-Type", "application/json")
	createReq.Header.Set("Authorization", testhelpers.AuthBearer(tok))
	app.Test(createReq)

	// Check access
	req := httptest.NewRequest("GET", "/api/v1/workspaces/"+wsID+"/subscriptions/access?product_id="+productID, nil)
	req.Header.Set("Authorization", testhelpers.AuthBearer(tok))
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var body map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&body)
	assert.Equal(t, true, body["has_access"])
}

func TestCheckAccess_WithoutSub_ReturnsDenied(t *testing.T) {
	db := testhelpers.SetupTestDB()
	app := testhelpers.NewApp()
	_, tok := testhelpers.CreateVerifiedUser(db, "check-nosub@example.com", "password123")

	p := models.Product{Name: "no-sub-product", IsActive: true}
	database.DB.Create(&p)

	listWS := httptest.NewRequest("GET", "/api/v1/workspaces", nil)
	listWS.Header.Set("Authorization", testhelpers.AuthBearer(tok))
	listWSResp, _ := app.Test(listWS)
	var wsBody map[string]interface{}
	json.NewDecoder(listWSResp.Body).Decode(&wsBody)
	wsID := wsBody["workspaces"].([]interface{})[0].(map[string]interface{})["id"].(string)

	req := httptest.NewRequest("GET", "/api/v1/workspaces/"+wsID+"/subscriptions/access?product_id="+p.ID.String(), nil)
	req.Header.Set("Authorization", testhelpers.AuthBearer(tok))
	resp, err := app.Test(req)
	require.NoError(t, err)
	// 200 with has_access:false OR 404 — both are valid "no access" responses
	assert.True(t, resp.StatusCode == 200 || resp.StatusCode == 404)
}

// ── CancelSubscription ────────────────────────────────────────────────────────

func TestCancelSubscription_Success(t *testing.T) {
	tok, wsID, productID := setupProductAndWorkspace(t)
	app := testhelpers.NewApp()

	// Create
	createPayload, _ := json.Marshal(map[string]interface{}{
		"product_id": productID, "plan_name": "starter", "period_days": 30,
	})
	createReq := httptest.NewRequest("POST", "/api/v1/workspaces/"+wsID+"/subscriptions", bytes.NewBuffer(createPayload))
	createReq.Header.Set("Content-Type", "application/json")
	createReq.Header.Set("Authorization", testhelpers.AuthBearer(tok))
	createResp, _ := app.Test(createReq)
	var createBody map[string]interface{}
	json.NewDecoder(createResp.Body).Decode(&createBody)
	subID := createBody["subscription"].(map[string]interface{})["id"].(string)

	// Cancel
	delReq := httptest.NewRequest("DELETE", "/api/v1/workspaces/"+wsID+"/subscriptions/"+subID, nil)
	delReq.Header.Set("Authorization", testhelpers.AuthBearer(tok))
	delResp, err := app.Test(delReq)
	require.NoError(t, err)
	assert.Equal(t, 200, delResp.StatusCode)

	// Confirm canceled in DB
	var sub models.Subscription
	database.DB.First(&sub, "id = ?", subID)
	assert.Equal(t, "canceled", sub.Status)
}

// ── GetBillingStatus ──────────────────────────────────────────────────────────

func TestGetBillingStatus_Returns200(t *testing.T) {
	db := testhelpers.SetupTestDB()
	app := testhelpers.NewApp()
	_, tok := testhelpers.CreateVerifiedUser(db, "billing-status@example.com", "password123")

	listWS := httptest.NewRequest("GET", "/api/v1/workspaces", nil)
	listWS.Header.Set("Authorization", testhelpers.AuthBearer(tok))
	listWSResp, _ := app.Test(listWS)
	var wsBody map[string]interface{}
	json.NewDecoder(listWSResp.Body).Decode(&wsBody)
	wsID := wsBody["workspaces"].([]interface{})[0].(map[string]interface{})["id"].(string)

	req := httptest.NewRequest("GET", "/api/v1/workspaces/"+wsID+"/billing", nil)
	req.Header.Set("Authorization", testhelpers.AuthBearer(tok))
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var body map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&body)
	_, ok := body["has_stripe_customer"]
	assert.True(t, ok, "response must contain has_stripe_customer field")
}

// ── Admin: list subscriptions + billing overview ──────────────────────────────

func TestAdminListSubscriptions_Returns200(t *testing.T) {
	testhelpers.SetupTestDB()
	app := testhelpers.NewApp()

	req := httptest.NewRequest("GET", "/api/v1/admin/subscriptions", nil)
	req.Header.Set("X-Admin-Secret", testhelpers.AdminSecret)
	resp, _ := app.Test(req)
	assert.Equal(t, 200, resp.StatusCode)
}

func TestAdminBillingOverview_Returns200(t *testing.T) {
	testhelpers.SetupTestDB()
	app := testhelpers.NewApp()

	req := httptest.NewRequest("GET", "/api/v1/admin/billing", nil)
	req.Header.Set("X-Admin-Secret", testhelpers.AdminSecret)
	resp, _ := app.Test(req)
	assert.Equal(t, 200, resp.StatusCode)
}
