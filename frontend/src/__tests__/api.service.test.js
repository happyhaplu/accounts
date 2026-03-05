import { describe, it, expect, vi, beforeEach } from 'vitest'
import axios from 'axios'

// ── api.js endpoint surface tests ─────────────────────────────────────────────
// We don't hit the network; we verify each helper calls the correct
// HTTP method + URL. Axios is mocked at the module level.

vi.mock('axios', async () => {
  const calls = []
  const mockInstance = {
    get:     (url, cfg) => { calls.push({ method: 'GET',    url, cfg }); return Promise.resolve({ data: {} }) },
    post:    (url, d, cfg) => { calls.push({ method: 'POST',   url, data: d, cfg }); return Promise.resolve({ data: {} }) },
    delete:  (url, cfg) => { calls.push({ method: 'DELETE', url, cfg }); return Promise.resolve({ data: {} }) },
    patch:   (url, d, cfg) => { calls.push({ method: 'PATCH',  url, data: d, cfg }); return Promise.resolve({ data: {} }) },
    interceptors: {
      request:  { use: vi.fn() },
      response: { use: vi.fn() },
    },
    _calls: calls,
  }
  return {
    default: { create: () => mockInstance },
  }
})

// Re-import authAPI AFTER the mock is in place
const { authAPI } = await import('../services/api.js')

// Helper: peek at the last recorded call from the mocked axios instance
function lastCall() {
  // We need access to the mock instance — easiest is to grab it from the module
  return axios.create()._calls.at(-1)
}

describe('authAPI surface', () => {
  beforeEach(() => {
    // Reset call log before each test
    axios.create()._calls.length = 0
    // Mock localStorage so the request interceptor doesn't throw
    vi.stubGlobal('localStorage', {
      getItem:    () => null,
      setItem:    vi.fn(),
      removeItem: vi.fn(),
    })
  })

  it('register POSTs to /auth/register', async () => {
    await authAPI.register({ email: 'a@b.com', password: 'pw123456' })
    const c = lastCall()
    expect(c.method).toBe('POST')
    expect(c.url).toBe('/auth/register')
  })

  it('login POSTs to /auth/login', async () => {
    await authAPI.login({ email: 'a@b.com', password: 'pw' })
    const c = lastCall()
    expect(c.method).toBe('POST')
    expect(c.url).toBe('/auth/login')
  })

  it('logout POSTs to /auth/logout', async () => {
    await authAPI.logout()
    const c = lastCall()
    expect(c.method).toBe('POST')
    expect(c.url).toBe('/auth/logout')
  })

  it('forgotPassword POSTs to /auth/forgot-password', async () => {
    await authAPI.forgotPassword({ email: 'a@b.com' })
    const c = lastCall()
    expect(c.method).toBe('POST')
    expect(c.url).toBe('/auth/forgot-password')
  })

  it('resetPassword POSTs to /auth/reset-password', async () => {
    await authAPI.resetPassword({ token: 'tok', new_password: 'newpass1' })
    const c = lastCall()
    expect(c.method).toBe('POST')
    expect(c.url).toBe('/auth/reset-password')
  })

  it('getProfile GETs /profile', async () => {
    await authAPI.getProfile()
    const c = lastCall()
    expect(c.method).toBe('GET')
    expect(c.url).toBe('/profile')
  })

  it('setupProfile POSTs to /profile', async () => {
    await authAPI.setupProfile({ name: 'Alice', company_name: 'Acme' })
    const c = lastCall()
    expect(c.method).toBe('POST')
    expect(c.url).toBe('/profile')
  })

  it('listWorkspaces GETs /workspaces', async () => {
    await authAPI.listWorkspaces()
    const c = lastCall()
    expect(c.method).toBe('GET')
    expect(c.url).toBe('/workspaces')
  })

  it('createWorkspace POSTs to /workspaces', async () => {
    await authAPI.createWorkspace({ name: 'My WS' })
    const c = lastCall()
    expect(c.method).toBe('POST')
    expect(c.url).toBe('/workspaces')
  })

  it('getWorkspace GETs /workspaces/:id when id is provided', async () => {
    await authAPI.getWorkspace('ws-uuid-123')
    const c = lastCall()
    expect(c.method).toBe('GET')
    expect(c.url).toBe('/workspaces/ws-uuid-123')
  })

  it('getWorkspace GETs /workspace (legacy) when id is falsy', async () => {
    await authAPI.getWorkspace(null)
    const c = lastCall()
    expect(c.method).toBe('GET')
    expect(c.url).toBe('/workspace')
  })

  it('addMember POSTs to /workspaces/:id/members', async () => {
    await authAPI.addMember('ws-1', { email: 'b@c.com', role: 'member' })
    const c = lastCall()
    expect(c.method).toBe('POST')
    expect(c.url).toBe('/workspaces/ws-1/members')
  })

  it('removeMember DELETEs /workspaces/:id/members/:uid', async () => {
    await authAPI.removeMember('ws-1', 'user-uuid-1')
    const c = lastCall()
    expect(c.method).toBe('DELETE')
    expect(c.url).toBe('/workspaces/ws-1/members/user-uuid-1')
  })

  it('listSubscriptions GETs /workspaces/:id/subscriptions', async () => {
    await authAPI.listSubscriptions('ws-abc')
    const c = lastCall()
    expect(c.method).toBe('GET')
    expect(c.url).toBe('/workspaces/ws-abc/subscriptions')
  })

  it('cancelSubscription DELETEs /workspaces/:wsId/subscriptions/:subId', async () => {
    await authAPI.cancelSubscription('ws-abc', 'sub-xyz')
    const c = lastCall()
    expect(c.method).toBe('DELETE')
    expect(c.url).toBe('/workspaces/ws-abc/subscriptions/sub-xyz')
  })

  it('getBillingStatus GETs /workspaces/:id/billing', async () => {
    await authAPI.getBillingStatus('ws-abc')
    const c = lastCall()
    expect(c.method).toBe('GET')
    expect(c.url).toBe('/workspaces/ws-abc/billing')
  })
})
