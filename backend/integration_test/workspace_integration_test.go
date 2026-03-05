package integration_test

import (
"bytes"
"encoding/json"
"fmt"
"net/http/httptest"
"testing"

"outcraftly/accounts/testhelpers"

"github.com/stretchr/testify/assert"
"github.com/stretchr/testify/require"
)

// ── ListWorkspaces ────────────────────────────────────────────────────────────

func TestListWorkspaces_Unauthorised(t *testing.T) {
testhelpers.SetupTestDB()
app := testhelpers.NewApp()

req := httptest.NewRequest("GET", "/api/v1/workspaces", nil)
resp, _ := app.Test(req)
assert.Equal(t, 401, resp.StatusCode)
}

func TestListWorkspaces_CreatesDefaultWorkspace(t *testing.T) {
db := testhelpers.SetupTestDB()
app := testhelpers.NewApp()
_, tok := testhelpers.CreateVerifiedUser(db, "ws-list@example.com", "password123")

req := httptest.NewRequest("GET", "/api/v1/workspaces", nil)
req.Header.Set("Authorization", testhelpers.AuthBearer(tok))
resp, err := app.Test(req)
require.NoError(t, err)
assert.Equal(t, 200, resp.StatusCode)

var body map[string]interface{}
json.NewDecoder(resp.Body).Decode(&body)
wsList := body["workspaces"].([]interface{})
// CreateVerifiedUser does not auto-create workspace; ListWorkspaces lazy-creates it
assert.GreaterOrEqual(t, len(wsList), 1)
}

// ── CreateWorkspace ───────────────────────────────────────────────────────────

func TestCreateWorkspace_Success(t *testing.T) {
db := testhelpers.SetupTestDB()
app := testhelpers.NewApp()
_, tok := testhelpers.CreateVerifiedUser(db, "ws-create@example.com", "password123")

payload := `{"name":"My New Workspace"}`
req := httptest.NewRequest("POST", "/api/v1/workspaces", bytes.NewBufferString(payload))
req.Header.Set("Content-Type", "application/json")
req.Header.Set("Authorization", testhelpers.AuthBearer(tok))
resp, err := app.Test(req)
require.NoError(t, err)
assert.Equal(t, 201, resp.StatusCode)

var body map[string]interface{}
json.NewDecoder(resp.Body).Decode(&body)
ws := body["workspace"].(map[string]interface{})
assert.Equal(t, "My New Workspace", ws["name"])
assert.Equal(t, "owner", ws["my_role"])
}

func TestCreateWorkspace_EmptyName(t *testing.T) {
db := testhelpers.SetupTestDB()
app := testhelpers.NewApp()
_, tok := testhelpers.CreateVerifiedUser(db, "ws-empty@example.com", "password123")

payload := `{"name":""}`
req := httptest.NewRequest("POST", "/api/v1/workspaces", bytes.NewBufferString(payload))
req.Header.Set("Content-Type", "application/json")
req.Header.Set("Authorization", testhelpers.AuthBearer(tok))
resp, _ := app.Test(req)
assert.Equal(t, 400, resp.StatusCode)
}

func TestCreateWorkspace_NameTooLong(t *testing.T) {
db := testhelpers.SetupTestDB()
app := testhelpers.NewApp()
_, tok := testhelpers.CreateVerifiedUser(db, "ws-long@example.com", "password123")

longName := fmt.Sprintf(`{"name":"%s"}`, string(make([]byte, 81)))
req := httptest.NewRequest("POST", "/api/v1/workspaces", bytes.NewBufferString(longName))
req.Header.Set("Content-Type", "application/json")
req.Header.Set("Authorization", testhelpers.AuthBearer(tok))
resp, _ := app.Test(req)
assert.Equal(t, 400, resp.StatusCode)
}

func TestCreateWorkspace_Unauthorised(t *testing.T) {
testhelpers.SetupTestDB()
app := testhelpers.NewApp()

payload := `{"name":"Test WS"}`
req := httptest.NewRequest("POST", "/api/v1/workspaces", bytes.NewBufferString(payload))
req.Header.Set("Content-Type", "application/json")
resp, _ := app.Test(req)
assert.Equal(t, 401, resp.StatusCode)
}

// ── GetWorkspace ──────────────────────────────────────────────────────────────

func TestGetWorkspace_Success(t *testing.T) {
db := testhelpers.SetupTestDB()
app := testhelpers.NewApp()
_, tok := testhelpers.CreateVerifiedUser(db, "ws-get@example.com", "password123")

// First list to trigger lazy-create
listReq := httptest.NewRequest("GET", "/api/v1/workspaces", nil)
listReq.Header.Set("Authorization", testhelpers.AuthBearer(tok))
listResp, _ := app.Test(listReq)
var listBody map[string]interface{}
json.NewDecoder(listResp.Body).Decode(&listBody)
ws0 := listBody["workspaces"].([]interface{})[0].(map[string]interface{})
wsID := ws0["id"].(string)

// Now get by ID
req := httptest.NewRequest("GET", "/api/v1/workspaces/"+wsID, nil)
req.Header.Set("Authorization", testhelpers.AuthBearer(tok))
resp, err := app.Test(req)
require.NoError(t, err)
assert.Equal(t, 200, resp.StatusCode)

var body map[string]interface{}
json.NewDecoder(resp.Body).Decode(&body)
ws := body["workspace"].(map[string]interface{})
assert.Equal(t, wsID, ws["id"])
assert.NotNil(t, ws["members"])
}

func TestGetWorkspace_ForbiddenOtherUser(t *testing.T) {
db := testhelpers.SetupTestDB()
app := testhelpers.NewApp()
_, tok1 := testhelpers.CreateVerifiedUser(db, "owner@example.com", "password123")
_, tok2 := testhelpers.CreateVerifiedUser(db, "intruder@example.com", "password123")

// owner creates workspace
createReq := httptest.NewRequest("POST", "/api/v1/workspaces", bytes.NewBufferString(`{"name":"Private WS"}`))
createReq.Header.Set("Content-Type", "application/json")
createReq.Header.Set("Authorization", testhelpers.AuthBearer(tok1))
createResp, _ := app.Test(createReq)
var createBody map[string]interface{}
json.NewDecoder(createResp.Body).Decode(&createBody)
wsID := createBody["workspace"].(map[string]interface{})["id"].(string)

// intruder tries to access it
req := httptest.NewRequest("GET", "/api/v1/workspaces/"+wsID, nil)
req.Header.Set("Authorization", testhelpers.AuthBearer(tok2))
resp, _ := app.Test(req)
assert.Equal(t, 403, resp.StatusCode)
}

// ── AddMember / RemoveMember ──────────────────────────────────────────────────

func TestAddMember_Success(t *testing.T) {
db := testhelpers.SetupTestDB()
app := testhelpers.NewApp()
_, ownerTok := testhelpers.CreateVerifiedUser(db, "addmem-owner@example.com", "password123")
testhelpers.CreateVerifiedUser(db, "addmem-member@example.com", "password123")

// owner creates workspace
createReq := httptest.NewRequest("POST", "/api/v1/workspaces", bytes.NewBufferString(`{"name":"Team WS"}`))
createReq.Header.Set("Content-Type", "application/json")
createReq.Header.Set("Authorization", testhelpers.AuthBearer(ownerTok))
createResp, _ := app.Test(createReq)
var createBody map[string]interface{}
json.NewDecoder(createResp.Body).Decode(&createBody)
wsID := createBody["workspace"].(map[string]interface{})["id"].(string)

// add member
addPayload := `{"email":"addmem-member@example.com","role":"member"}`
addReq := httptest.NewRequest("POST", "/api/v1/workspaces/"+wsID+"/members", bytes.NewBufferString(addPayload))
addReq.Header.Set("Content-Type", "application/json")
addReq.Header.Set("Authorization", testhelpers.AuthBearer(ownerTok))
addResp, err := app.Test(addReq)
require.NoError(t, err)
assert.Equal(t, 201, addResp.StatusCode)
}

func TestAddMember_NonOwnerForbidden(t *testing.T) {
db := testhelpers.SetupTestDB()
app := testhelpers.NewApp()
_, ownerTok := testhelpers.CreateVerifiedUser(db, "addmem2-owner@example.com", "password123")
_, memberTok := testhelpers.CreateVerifiedUser(db, "addmem2-member@example.com", "password123")
testhelpers.CreateVerifiedUser(db, "addmem2-target@example.com", "password123")

// create workspace as owner
createReq := httptest.NewRequest("POST", "/api/v1/workspaces", bytes.NewBufferString(`{"name":"WS2"}`))
createReq.Header.Set("Content-Type", "application/json")
createReq.Header.Set("Authorization", testhelpers.AuthBearer(ownerTok))
createResp, _ := app.Test(createReq)
var createBody map[string]interface{}
json.NewDecoder(createResp.Body).Decode(&createBody)
wsID := createBody["workspace"].(map[string]interface{})["id"].(string)

// add member as owner
addPayload := `{"email":"addmem2-member@example.com","role":"member"}`
addReq := httptest.NewRequest("POST", "/api/v1/workspaces/"+wsID+"/members", bytes.NewBufferString(addPayload))
addReq.Header.Set("Content-Type", "application/json")
addReq.Header.Set("Authorization", testhelpers.AuthBearer(ownerTok))
app.Test(addReq)

// member tries to add someone else — forbidden
addPayload2 := `{"email":"addmem2-target@example.com","role":"member"}`
req := httptest.NewRequest("POST", "/api/v1/workspaces/"+wsID+"/members", bytes.NewBufferString(addPayload2))
req.Header.Set("Content-Type", "application/json")
req.Header.Set("Authorization", testhelpers.AuthBearer(memberTok))
resp, _ := app.Test(req)
assert.Equal(t, 403, resp.StatusCode)
}

func TestAddMember_AlreadyMember(t *testing.T) {
db := testhelpers.SetupTestDB()
app := testhelpers.NewApp()
_, ownerTok := testhelpers.CreateVerifiedUser(db, "dup-owner@example.com", "password123")
testhelpers.CreateVerifiedUser(db, "dup-mem@example.com", "password123")

createReq := httptest.NewRequest("POST", "/api/v1/workspaces", bytes.NewBufferString(`{"name":"DupWS"}`))
createReq.Header.Set("Content-Type", "application/json")
createReq.Header.Set("Authorization", testhelpers.AuthBearer(ownerTok))
createResp, _ := app.Test(createReq)
var createBody map[string]interface{}
json.NewDecoder(createResp.Body).Decode(&createBody)
wsID := createBody["workspace"].(map[string]interface{})["id"].(string)

addPayload := `{"email":"dup-mem@example.com","role":"member"}`
// add once
r1 := httptest.NewRequest("POST", "/api/v1/workspaces/"+wsID+"/members", bytes.NewBufferString(addPayload))
r1.Header.Set("Content-Type", "application/json")
r1.Header.Set("Authorization", testhelpers.AuthBearer(ownerTok))
app.Test(r1)
// add again
r2 := httptest.NewRequest("POST", "/api/v1/workspaces/"+wsID+"/members", bytes.NewBufferString(addPayload))
r2.Header.Set("Content-Type", "application/json")
r2.Header.Set("Authorization", testhelpers.AuthBearer(ownerTok))
resp, _ := app.Test(r2)
assert.Equal(t, 409, resp.StatusCode)
}
