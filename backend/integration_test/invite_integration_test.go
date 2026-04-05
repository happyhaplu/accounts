package integration_test

// ── Invite integration tests ──────────────────────────────────────────────────

import (
"bytes"
"encoding/json"
"net/http/httptest"
"testing"

"outcraftly/accounts/database"
"outcraftly/accounts/models"
"outcraftly/accounts/testhelpers"

"github.com/stretchr/testify/assert"
"github.com/stretchr/testify/require"
)

// ownerAndWorkspace creates a fresh DB, registers a verified user, and returns
// that user's JWT plus their auto-created workspace ID.
func ownerAndWorkspace(t *testing.T, email string) (tok, wsID string) {
t.Helper()
db := testhelpers.SetupTestDB()
app := testhelpers.NewApp()
_, tok = testhelpers.CreateVerifiedUser(db, email, "password123")

listReq := httptest.NewRequest("GET", "/api/v1/workspaces", nil)
listReq.Header.Set("Authorization", testhelpers.AuthBearer(tok))
listResp, err := app.Test(listReq)
require.NoError(t, err)
var listBody map[string]interface{}
json.NewDecoder(listResp.Body).Decode(&listBody) //nolint:errcheck
wsID = listBody["workspaces"].([]interface{})[0].(map[string]interface{})["id"].(string)
return
}

// ── SendInvite ────────────────────────────────────────────────────────────────

func TestSendInvite_Success(t *testing.T) {
tok, wsID := ownerAndWorkspace(t, "invite-owner@example.com")
app := testhelpers.NewApp()

payload, _ := json.Marshal(map[string]string{"email": "invited@example.com", "role": "member"})
req := httptest.NewRequest("POST", "/api/v1/workspaces/"+wsID+"/invites", bytes.NewBuffer(payload))
req.Header.Set("Content-Type", "application/json")
req.Header.Set("Authorization", testhelpers.AuthBearer(tok))

resp, err := app.Test(req)
require.NoError(t, err)
assert.Equal(t, 201, resp.StatusCode)

var body map[string]interface{}
json.NewDecoder(resp.Body).Decode(&body) //nolint:errcheck
invite := body["invite"].(map[string]interface{})
assert.Equal(t, "invited@example.com", invite["email"])
assert.NotEmpty(t, invite["id"], "invite id must be set")

// Token is NOT in the JSON response — verify it was persisted in DB
var dbInvite models.WorkspaceInvite
require.NoError(t, database.DB.Where("email = ?", "invited@example.com").First(&dbInvite).Error)
assert.NotEmpty(t, dbInvite.Token)
}

func TestSendInvite_Unauthenticated_Returns401(t *testing.T) {
testhelpers.SetupTestDB()
app := testhelpers.NewApp()

payload, _ := json.Marshal(map[string]string{"email": "x@y.com", "role": "member"})
req := httptest.NewRequest("POST", "/api/v1/workspaces/some-ws/invites", bytes.NewBuffer(payload))
req.Header.Set("Content-Type", "application/json")
resp, _ := app.Test(req)
assert.Equal(t, 401, resp.StatusCode)
}

// TestSendInvite_ReinviteSameEmail_RevokesOldAndCreatesNew verifies that sending
// a second invite to the same address revokes the previous pending invite and
// issues a fresh one (HTTP 201). Only inviting an existing workspace *member*
// returns HTTP 409.
func TestSendInvite_ReinviteSameEmail_RevokesOldAndCreatesNew(t *testing.T) {
	tok, wsID := ownerAndWorkspace(t, "dup-invite-owner@example.com")
	app := testhelpers.NewApp()

	for i := 0; i < 2; i++ {
		payload, _ := json.Marshal(map[string]string{"email": "dup@example.com", "role": "member"})
		req := httptest.NewRequest("POST", "/api/v1/workspaces/"+wsID+"/invites", bytes.NewBuffer(payload))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", testhelpers.AuthBearer(tok))
		resp, _ := app.Test(req)
		assert.Equal(t, 201, resp.StatusCode, "re-inviting same email must return 201 (revokes old, creates new)")
	}

	// Only one pending invite should exist after two sends
	var count int64
	database.DB.Model(&models.WorkspaceInvite{}).Where("email = ? AND status = 'pending'", "dup@example.com").Count(&count)
	assert.Equal(t, int64(1), count, "only one pending invite should remain")
}

// ── ListInvites ───────────────────────────────────────────────────────────────

func TestListInvites_AfterSend_ReturnsList(t *testing.T) {
tok, wsID := ownerAndWorkspace(t, "listinvite-owner@example.com")
app := testhelpers.NewApp()

payload, _ := json.Marshal(map[string]string{"email": "li@example.com", "role": "member"})
sendReq := httptest.NewRequest("POST", "/api/v1/workspaces/"+wsID+"/invites", bytes.NewBuffer(payload))
sendReq.Header.Set("Content-Type", "application/json")
sendReq.Header.Set("Authorization", testhelpers.AuthBearer(tok))
app.Test(sendReq) //nolint:errcheck

listReq := httptest.NewRequest("GET", "/api/v1/workspaces/"+wsID+"/invites", nil)
listReq.Header.Set("Authorization", testhelpers.AuthBearer(tok))
listResp, err := app.Test(listReq)
require.NoError(t, err)
assert.Equal(t, 200, listResp.StatusCode)

var body map[string]interface{}
json.NewDecoder(listResp.Body).Decode(&body) //nolint:errcheck
invites := body["invites"].([]interface{})
assert.Equal(t, 1, len(invites))
}

// ── RevokeInvite ──────────────────────────────────────────────────────────────

func TestRevokeInvite_Success(t *testing.T) {
tok, wsID := ownerAndWorkspace(t, "revoke-owner@example.com")
app := testhelpers.NewApp()

payload, _ := json.Marshal(map[string]string{"email": "revoke-me@example.com", "role": "member"})
sendReq := httptest.NewRequest("POST", "/api/v1/workspaces/"+wsID+"/invites", bytes.NewBuffer(payload))
sendReq.Header.Set("Content-Type", "application/json")
sendReq.Header.Set("Authorization", testhelpers.AuthBearer(tok))
sendResp, _ := app.Test(sendReq)
var sendBody map[string]interface{}
json.NewDecoder(sendResp.Body).Decode(&sendBody) //nolint:errcheck
inviteID := sendBody["invite"].(map[string]interface{})["id"].(string)

delReq := httptest.NewRequest("DELETE", "/api/v1/workspaces/"+wsID+"/invites/"+inviteID, nil)
delReq.Header.Set("Authorization", testhelpers.AuthBearer(tok))
delResp, err := app.Test(delReq)
require.NoError(t, err)
assert.Equal(t, 200, delResp.StatusCode)

var inv models.WorkspaceInvite
	require.NoError(t, database.DB.First(&inv, "id = ?", inviteID).Error, "invite record must still exist")
	assert.Equal(t, "revoked", inv.Status, "invite status must be 'revoked'")
}

// ── PreviewInvite ─────────────────────────────────────────────────────────────

func TestPreviewInvite_MissingToken_Returns400(t *testing.T) {
testhelpers.SetupTestDB()
app := testhelpers.NewApp()

req := httptest.NewRequest("GET", "/api/v1/invites/preview", nil)
resp, _ := app.Test(req)
assert.Equal(t, 400, resp.StatusCode)
}

func TestPreviewInvite_InvalidToken_Returns404(t *testing.T) {
testhelpers.SetupTestDB()
app := testhelpers.NewApp()

req := httptest.NewRequest("GET", "/api/v1/invites/preview?token=invalid-token", nil)
resp, _ := app.Test(req)
assert.Equal(t, 404, resp.StatusCode)
}

// ── AcceptInvite ──────────────────────────────────────────────────────────────

func TestAcceptInvite_ValidToken_JoinsWorkspace(t *testing.T) {
ownerTok, wsID := ownerAndWorkspace(t, "accept-owner@example.com")
app := testhelpers.NewApp()

// ownerAndWorkspace called SetupTestDB, so database.DB is the current in-memory DB.
// Create the member in the SAME DB by passing database.DB directly.
_, memberTok := testhelpers.CreateVerifiedUser(database.DB, "accept-member@example.com", "password123")

// Send invite as owner
payload, _ := json.Marshal(map[string]string{"email": "accept-member@example.com", "role": "member"})
sendReq := httptest.NewRequest("POST", "/api/v1/workspaces/"+wsID+"/invites", bytes.NewBuffer(payload))
sendReq.Header.Set("Content-Type", "application/json")
sendReq.Header.Set("Authorization", testhelpers.AuthBearer(ownerTok))
sendResp, _ := app.Test(sendReq)
require.Equal(t, 201, sendResp.StatusCode)

// The token is NOT in the response body — look it up in the DB
var dbInvite models.WorkspaceInvite
require.NoError(t, database.DB.Where("email = ?", "accept-member@example.com").First(&dbInvite).Error)

// Accept as member
acceptPayload, _ := json.Marshal(map[string]string{"token": dbInvite.Token})
acceptReq := httptest.NewRequest("POST", "/api/v1/invites/accept", bytes.NewBuffer(acceptPayload))
acceptReq.Header.Set("Content-Type", "application/json")
acceptReq.Header.Set("Authorization", testhelpers.AuthBearer(memberTok))
acceptResp, err := app.Test(acceptReq)
require.NoError(t, err)
assert.Equal(t, 200, acceptResp.StatusCode)
}
