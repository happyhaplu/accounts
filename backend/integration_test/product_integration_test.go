package integration_test

// ── Admin product CRUD + product access integration tests ─────────────────────

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"outcraftly/accounts/database"
	"outcraftly/accounts/models"
	"outcraftly/accounts/testhelpers"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createProductViaAdmin(t *testing.T, name, description string) map[string]interface{} {
	t.Helper()
	testhelpers.SetupTestDB()
	app := testhelpers.NewApp()

	body, _ := json.Marshal(map[string]string{"name": name, "description": description})
	req := httptest.NewRequest("POST", "/api/v1/admin/products", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Admin-Secret", testhelpers.AdminSecret)
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, 201, resp.StatusCode)

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	return result["product"].(map[string]interface{})
}

// ── Admin: list products ──────────────────────────────────────────────────────

func TestAdminListProducts_NoSecret_Returns403(t *testing.T) {
	testhelpers.SetupTestDB()
	app := testhelpers.NewApp()

	req := httptest.NewRequest("GET", "/api/v1/admin/products", nil)
	resp, _ := app.Test(req)
	assert.Equal(t, 403, resp.StatusCode)
}

func TestAdminListProducts_WithSecret_Returns200(t *testing.T) {
	testhelpers.SetupTestDB()
	app := testhelpers.NewApp()

	req := httptest.NewRequest("GET", "/api/v1/admin/products", nil)
	req.Header.Set("X-Admin-Secret", testhelpers.AdminSecret)
	resp, _ := app.Test(req)
	assert.Equal(t, 200, resp.StatusCode)

	var body map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&body)
	_, ok := body["products"]
	assert.True(t, ok, "response must contain 'products' key")
}

// ── Admin: create product ─────────────────────────────────────────────────────

func TestAdminCreateProduct_Success(t *testing.T) {
	testhelpers.SetupTestDB()
	app := testhelpers.NewApp()

	payload, _ := json.Marshal(map[string]string{
		"name":        "test-product",
		"description": "A test product for integration tests",
	})
	req := httptest.NewRequest("POST", "/api/v1/admin/products", bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Admin-Secret", testhelpers.AdminSecret)

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 201, resp.StatusCode)

	var body map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&body)
	product := body["product"].(map[string]interface{})
	assert.Equal(t, "test-product", product["name"])
	assert.NotEmpty(t, product["id"])
	assert.NotEmpty(t, product["api_key"], "api_key must be auto-generated")
	assert.Equal(t, true, product["is_active"])
}

func TestAdminCreateProduct_DuplicateName_Returns409(t *testing.T) {
	testhelpers.SetupTestDB()
	app := testhelpers.NewApp()

	for i := 0; i < 2; i++ {
		payload, _ := json.Marshal(map[string]string{"name": "dup-product"})
		req := httptest.NewRequest("POST", "/api/v1/admin/products", bytes.NewBuffer(payload))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Admin-Secret", testhelpers.AdminSecret)
		resp, _ := app.Test(req)
		if i == 1 {
			assert.Equal(t, 409, resp.StatusCode)
		}
	}
}

func TestAdminCreateProduct_EmptyName_Returns400(t *testing.T) {
	testhelpers.SetupTestDB()
	app := testhelpers.NewApp()

	payload, _ := json.Marshal(map[string]string{"name": "", "description": "x"})
	req := httptest.NewRequest("POST", "/api/v1/admin/products", bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Admin-Secret", testhelpers.AdminSecret)
	resp, _ := app.Test(req)
	assert.Equal(t, 400, resp.StatusCode)
}

func TestAdminCreateProduct_NoSecret_Returns403(t *testing.T) {
	testhelpers.SetupTestDB()
	app := testhelpers.NewApp()

	payload, _ := json.Marshal(map[string]string{"name": "secret-test"})
	req := httptest.NewRequest("POST", "/api/v1/admin/products", bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)
	assert.Equal(t, 403, resp.StatusCode)
}

// ── Admin: update product ─────────────────────────────────────────────────────

func TestAdminUpdateProduct_Success(t *testing.T) {
	testhelpers.SetupTestDB()
	app := testhelpers.NewApp()

	// Create
	createPayload, _ := json.Marshal(map[string]string{"name": "update-me", "description": "old"})
	createReq := httptest.NewRequest("POST", "/api/v1/admin/products", bytes.NewBuffer(createPayload))
	createReq.Header.Set("Content-Type", "application/json")
	createReq.Header.Set("X-Admin-Secret", testhelpers.AdminSecret)
	createResp, _ := app.Test(createReq)
	var createBody map[string]interface{}
	json.NewDecoder(createResp.Body).Decode(&createBody)
	id := createBody["product"].(map[string]interface{})["id"].(string)

	// Update description
	patchPayload, _ := json.Marshal(map[string]string{"description": "updated description"})
	patchReq := httptest.NewRequest("PATCH", "/api/v1/admin/products/"+id, bytes.NewBuffer(patchPayload))
	patchReq.Header.Set("Content-Type", "application/json")
	patchReq.Header.Set("X-Admin-Secret", testhelpers.AdminSecret)
	patchResp, err := app.Test(patchReq)
	require.NoError(t, err)
	assert.Equal(t, 200, patchResp.StatusCode)

	var patchBody map[string]interface{}
	json.NewDecoder(patchResp.Body).Decode(&patchBody)
	updated := patchBody["product"].(map[string]interface{})
	assert.Equal(t, "updated description", updated["description"])
}

// ── Admin: deactivate product ─────────────────────────────────────────────────

func TestAdminDeactivateProduct_Success(t *testing.T) {
	testhelpers.SetupTestDB()
	app := testhelpers.NewApp()

	// Create
	createPayload, _ := json.Marshal(map[string]string{"name": "deactivate-me"})
	createReq := httptest.NewRequest("POST", "/api/v1/admin/products", bytes.NewBuffer(createPayload))
	createReq.Header.Set("Content-Type", "application/json")
	createReq.Header.Set("X-Admin-Secret", testhelpers.AdminSecret)
	createResp, _ := app.Test(createReq)
	var createBody map[string]interface{}
	json.NewDecoder(createResp.Body).Decode(&createBody)
	id := createBody["product"].(map[string]interface{})["id"].(string)

	// Deactivate
	delReq := httptest.NewRequest("DELETE", "/api/v1/admin/products/"+id, nil)
	delReq.Header.Set("X-Admin-Secret", testhelpers.AdminSecret)
	delResp, err := app.Test(delReq)
	require.NoError(t, err)
	assert.Equal(t, 200, delResp.StatusCode)

	// Confirm is_active = false in DB
	var product models.Product
	database.DB.First(&product, "id = ?", id)
	assert.False(t, product.IsActive)
}

// ── Admin: regenerate API key ─────────────────────────────────────────────────

func TestAdminRegenerateAPIKey_ChangesKey(t *testing.T) {
	testhelpers.SetupTestDB()
	app := testhelpers.NewApp()

	// Create
	createPayload, _ := json.Marshal(map[string]string{"name": "regen-key-prod"})
	createReq := httptest.NewRequest("POST", "/api/v1/admin/products", bytes.NewBuffer(createPayload))
	createReq.Header.Set("Content-Type", "application/json")
	createReq.Header.Set("X-Admin-Secret", testhelpers.AdminSecret)
	createResp, _ := app.Test(createReq)
	var createBody map[string]interface{}
	json.NewDecoder(createResp.Body).Decode(&createBody)
	product := createBody["product"].(map[string]interface{})
	id := product["id"].(string)
	oldKey := product["api_key"].(string)

	// Regenerate
	regenReq := httptest.NewRequest("POST", "/api/v1/admin/products/"+id+"/regenerate-key", nil)
	regenReq.Header.Set("X-Admin-Secret", testhelpers.AdminSecret)
	regenResp, err := app.Test(regenReq)
	require.NoError(t, err)
	assert.Equal(t, 200, regenResp.StatusCode)

	var regenBody map[string]interface{}
	json.NewDecoder(regenResp.Body).Decode(&regenBody)
	newKey := regenBody["product"].(map[string]interface{})["api_key"].(string)
	assert.NotEqual(t, oldKey, newKey, "api_key must change after regeneration")
	assert.NotEmpty(t, newKey)
}

// ── GET /api/v1/products (authenticated user) ─────────────────────────────────

func TestListProducts_Authenticated_ReturnsActiveOnly(t *testing.T) {
	db := testhelpers.SetupTestDB()
	app := testhelpers.NewApp()
	_, tok := testhelpers.CreateVerifiedUser(db, "prod-list@example.com", "password123")

	// Create two products: one active, one deactivated
	var active models.Product
	database.DB.Create(&models.Product{Name: "active-prod", IsActive: true})
	database.DB.Create(&models.Product{Name: "inactive-prod", IsActive: false})
	database.DB.Where("name = ?", "active-prod").First(&active)

	req := httptest.NewRequest("GET", "/api/v1/products", nil)
	req.Header.Set("Authorization", testhelpers.AuthBearer(tok))
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var body map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&body)
	products := body["products"].([]interface{})
	for _, p := range products {
		prod := p.(map[string]interface{})
		assert.Equal(t, true, prod["is_active"],
			"inactive products must not appear for regular users")
	}
}

func TestListProducts_Unauthenticated_Returns401(t *testing.T) {
	testhelpers.SetupTestDB()
	app := testhelpers.NewApp()

	req := httptest.NewRequest("GET", "/api/v1/products", nil)
	resp, _ := app.Test(req)
	assert.Equal(t, 401, resp.StatusCode)
}

// ── POST /api/v1/products/verify (server-to-server) ──────────────────────────

func TestVerifyToken_ValidAPIKeyAndJWT(t *testing.T) {
	db := testhelpers.SetupTestDB()
	app := testhelpers.NewApp()
	user, tok := testhelpers.CreateVerifiedUser(db, "verify-user@example.com", "password123")

	// Create product with known API key
	product := models.Product{Name: "verify-product", IsActive: true, APIKey: "gour_test_apikey_abc123"}
	database.DB.Create(&product)

	// Create active subscription for user's workspace
	// First get workspace (lazy-created via ListWorkspaces)
	listReq := httptest.NewRequest("GET", "/api/v1/workspaces", nil)
	listReq.Header.Set("Authorization", testhelpers.AuthBearer(tok))
	listResp, _ := app.Test(listReq)
	var listBody map[string]interface{}
	json.NewDecoder(listResp.Body).Decode(&listBody)
	ws := listBody["workspaces"].([]interface{})[0].(map[string]interface{})
	wsID := ws["id"].(string)

	wsUUID, _ := uuid.Parse(wsID)
	database.DB.Create(&models.Subscription{
		WorkspaceID: wsUUID,
		ProductID:   product.ID,
		PlanName:    "starter",
		Status:      "active",
	})

	_ = user // confirm user created

	payload, _ := json.Marshal(map[string]string{"token": tok})
	req := httptest.NewRequest("POST", "/api/v1/products/verify", bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", "gour_test_apikey_abc123")
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var body map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&body)
	assert.Equal(t, true, body["valid"])
}

func TestVerifyToken_InvalidAPIKey_Returns401(t *testing.T) {
	testhelpers.SetupTestDB()
	app := testhelpers.NewApp()

	payload, _ := json.Marshal(map[string]string{"token": "any-token"})
	req := httptest.NewRequest("POST", "/api/v1/products/verify", bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", "invalid-key")
	resp, _ := app.Test(req)
	assert.Equal(t, 401, resp.StatusCode)
}

func TestVerifyToken_MissingAPIKey_Returns401(t *testing.T) {
	testhelpers.SetupTestDB()
	app := testhelpers.NewApp()

	payload, _ := json.Marshal(map[string]string{"token": "any-token"})
	req := httptest.NewRequest("POST", "/api/v1/products/verify", bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)
	assert.Equal(t, 401, resp.StatusCode)
}
