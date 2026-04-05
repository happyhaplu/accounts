import { describe, it, expect, vi, beforeEach } from 'vitest'
import axios from 'axios'

// ── Billing + Admin API surface tests ─────────────────────────────────────────
// All billing methods live on authAPI; admin CRUD lives on adminAPI.
// Axios is mocked at the module level — no real network traffic.

vi.mock('axios', async () => {
  const calls = []
  const mockInstance = {
    get:     (url, cfg)    => { calls.push({ method: 'GET',    url, cfg }); return Promise.resolve({ data: {} }) },
    post:    (url, d, cfg) => { calls.push({ method: 'POST',   url, data: d, cfg }); return Promise.resolve({ data: {} }) },
    patch:   (url, d, cfg) => { calls.push({ method: 'PATCH',  url, data: d, cfg }); return Promise.resolve({ data: {} }) },
    delete:  (url, cfg)    => { calls.push({ method: 'DELETE', url, cfg }); return Promise.resolve({ data: {} }) },
    interceptors: {
      request:  { use: vi.fn() },
      response: { use: vi.fn() },
    },
    _calls: calls,
  }
  return { default: { create: () => mockInstance } }
})

const { authAPI, adminAPI } = await import('../services/api.js')
function lastCall() { return axios.create()._calls.at(-1) }

beforeEach(() => {
  axios.create()._calls.length = 0
  vi.stubGlobal('localStorage',   { getItem: () => null, setItem: vi.fn(), removeItem: vi.fn() })
  vi.stubGlobal('sessionStorage', { getItem: () => null, setItem: vi.fn(), removeItem: vi.fn() })
})

// ── authAPI billing ───────────────────────────────────────────────────────────

describe('authAPI.getBillingStatus', () => {
  it('GETs /workspaces/:id/billing', async () => {
    await authAPI.getBillingStatus('ws-abc')
    expect(lastCall().method).toBe('GET')
    expect(lastCall().url).toBe('/workspaces/ws-abc/billing')
  })
})

describe('authAPI.createPortalSession', () => {
  it('POSTs to /workspaces/:id/billing/portal', async () => {
    await authAPI.createPortalSession('ws-portal-1')
    expect(lastCall().method).toBe('POST')
    expect(lastCall().url).toBe('/workspaces/ws-portal-1/billing/portal')
  })
})

describe('authAPI.syncBilling', () => {
  it('POSTs to /workspaces/:id/billing/sync', async () => {
    await authAPI.syncBilling('ws-sync-1')
    expect(lastCall().method).toBe('POST')
    expect(lastCall().url).toBe('/workspaces/ws-sync-1/billing/sync')
  })
})

describe('authAPI.listSubscriptions', () => {
  it('GETs /workspaces/:id/subscriptions', async () => {
    await authAPI.listSubscriptions('ws-sub-1')
    expect(lastCall().method).toBe('GET')
    expect(lastCall().url).toBe('/workspaces/ws-sub-1/subscriptions')
  })
})

describe('authAPI.cancelSubscription', () => {
  it('DELETEs /workspaces/:wsId/subscriptions/:subId', async () => {
    await authAPI.cancelSubscription('ws-1', 'sub-99')
    expect(lastCall().method).toBe('DELETE')
    expect(lastCall().url).toBe('/workspaces/ws-1/subscriptions/sub-99')
  })
})

describe('authAPI.listProducts', () => {
  it('GETs /products', async () => {
    await authAPI.listProducts()
    expect(lastCall().method).toBe('GET')
    expect(lastCall().url).toBe('/products')
  })
})

// ── adminAPI product CRUD ─────────────────────────────────────────────────────

describe('adminAPI.listProducts', () => {
  it('GETs /admin/products', async () => {
    await adminAPI.listProducts()
    expect(lastCall().method).toBe('GET')
    expect(lastCall().url).toBe('/admin/products')
  })
})

describe('adminAPI.createProduct', () => {
  it('POSTs to /admin/products', async () => {
    await adminAPI.createProduct({ name: 'email-warmup' })
    expect(lastCall().method).toBe('POST')
    expect(lastCall().url).toBe('/admin/products')
  })
  it('sends payload as body', async () => {
    const payload = { name: 'reach', description: 'Reach out' }
    await adminAPI.createProduct(payload)
    expect(lastCall().data).toEqual(payload)
  })
})

describe('adminAPI.updateProduct', () => {
  it('PATCHes /admin/products/:id', async () => {
    await adminAPI.updateProduct('prod-1', { description: 'Updated' })
    expect(lastCall().method).toBe('PATCH')
    expect(lastCall().url).toBe('/admin/products/prod-1')
  })
})

describe('adminAPI.deactivateProduct', () => {
  it('DELETEs /admin/products/:id', async () => {
    await adminAPI.deactivateProduct('prod-2')
    expect(lastCall().method).toBe('DELETE')
    expect(lastCall().url).toBe('/admin/products/prod-2')
  })
})

describe('adminAPI.regenerateProductKey', () => {
  it('POSTs to /admin/products/:id/regenerate-key', async () => {
    await adminAPI.regenerateProductKey('prod-3')
    expect(lastCall().method).toBe('POST')
    expect(lastCall().url).toBe('/admin/products/prod-3/regenerate-key')
  })
})

// ── adminAPI users + workspaces + overview ────────────────────────────────────

describe('adminAPI.listUsers', () => {
  it('GETs /admin/users', async () => {
    await adminAPI.listUsers()
    expect(lastCall().method).toBe('GET')
    expect(lastCall().url).toBe('/admin/users')
  })
})

describe('adminAPI.listWorkspaces', () => {
  it('GETs /admin/workspaces', async () => {
    await adminAPI.listWorkspaces()
    expect(lastCall().method).toBe('GET')
    expect(lastCall().url).toBe('/admin/workspaces')
  })
})

describe('adminAPI.listSubscriptions', () => {
  it('GETs /admin/subscriptions', async () => {
    await adminAPI.listSubscriptions()
    expect(lastCall().method).toBe('GET')
    expect(lastCall().url).toBe('/admin/subscriptions')
  })
})

describe('adminAPI.billingOverview', () => {
  it('GETs /admin/billing', async () => {
    await adminAPI.billingOverview()
    expect(lastCall().method).toBe('GET')
    expect(lastCall().url).toBe('/admin/billing')
  })
})

// ── API contract regression: all methods exist ────────────────────────────────

describe('authAPI contract: required methods exist', () => {
  const required = [
    'register', 'login', 'logout', 'forgotPassword', 'resetPassword',
    'verifyEmailOTP', 'verifyResetOTP', 'resendVerification',
    'getProfile', 'setupProfile', 'changePassword',
    'listWorkspaces', 'createWorkspace', 'getWorkspace', 'addMember', 'removeMember',
    'sendInvite', 'listInvites', 'revokeInvite', 'previewInvite', 'acceptInvite',
    'listProducts', 'launchProduct',
    'listSubscriptions', 'cancelSubscription',
    'getBillingStatus', 'createPortalSession', 'syncBilling',
  ]
  required.forEach(m => {
    it(`authAPI.${m} is a function`, () => {
      expect(typeof authAPI[m]).toBe('function')
    })
  })
})

describe('adminAPI contract: required methods exist', () => {
  const required = [
    'login',
    'listProducts', 'createProduct', 'updateProduct', 'deactivateProduct',
    'permanentDeleteProduct', 'regenerateProductKey',
    'listUsers', 'purgeUnverifiedUsers',
    'listWorkspaces', 'listSubscriptions', 'billingOverview',
  ]
  required.forEach(m => {
    it(`adminAPI.${m} is a function`, () => {
      expect(typeof adminAPI[m]).toBe('function')
    })
  })
})
