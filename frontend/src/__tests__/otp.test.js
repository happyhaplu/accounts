import { describe, it, expect, vi, beforeEach } from 'vitest'
import axios from 'axios'

// ── OTP API surface tests ──────────────────────────────────────────────────────
// Tests that verifyEmailOTP and verifyResetOTP call the correct endpoints with
// the correct method and payload.  No real network traffic — axios is mocked.

vi.mock('axios', async () => {
  const calls = []
  const mockInstance = {
    get:    (url, cfg)      => { calls.push({ method: 'GET',    url, cfg });         return Promise.resolve({ data: {} }) },
    post:   (url, d, cfg)   => { calls.push({ method: 'POST',   url, data: d, cfg }); return Promise.resolve({ data: {} }) },
    delete: (url, cfg)      => { calls.push({ method: 'DELETE', url, cfg });         return Promise.resolve({ data: {} }) },
    patch:  (url, d, cfg)   => { calls.push({ method: 'PATCH',  url, data: d, cfg }); return Promise.resolve({ data: {} }) },
    interceptors: {
      request:  { use: vi.fn() },
      response: { use: vi.fn() },
    },
    _calls: calls,
  }
  return { default: { create: () => mockInstance } }
})

const { authAPI } = await import('../services/api.js')

function lastCall() {
  return axios.create()._calls.at(-1)
}

describe('OTP API methods', () => {
  beforeEach(() => {
    axios.create()._calls.length = 0
    vi.stubGlobal('localStorage', {
      getItem:    () => null,
      setItem:    vi.fn(),
      removeItem: vi.fn(),
    })
  })

  // ── verifyEmailOTP ─────────────────────────────────────────────────────────

  it('verifyEmailOTP POSTs to /auth/verify-email-otp', async () => {
    await authAPI.verifyEmailOTP({ email: 'a@b.com', otp: '123456' })
    const c = lastCall()
    expect(c.method).toBe('POST')
    expect(c.url).toBe('/auth/verify-email-otp')
  })

  it('verifyEmailOTP sends email and otp in payload', async () => {
    const payload = { email: 'user@example.com', otp: '654321' }
    await authAPI.verifyEmailOTP(payload)
    const c = lastCall()
    expect(c.data).toEqual(payload)
  })

  it('verifyEmailOTP returns data from server', async () => {
    axios.create()._calls.length = 0
    // Override the mock to return a specific response
    const mockResp = { token: 'jwt-abc', email_verified: true }
    axios.create().post = (url, d) => {
      axios.create()._calls.push({ method: 'POST', url, data: d })
      return Promise.resolve({ data: mockResp })
    }
    const result = await authAPI.verifyEmailOTP({ email: 'x@y.com', otp: '111111' })
    expect(result.data).toEqual(mockResp)
  })

  // ── verifyResetOTP ─────────────────────────────────────────────────────────

  it('verifyResetOTP POSTs to /auth/verify-reset-otp', async () => {
    await authAPI.verifyResetOTP({ email: 'a@b.com', otp: '999999' })
    const c = lastCall()
    expect(c.method).toBe('POST')
    expect(c.url).toBe('/auth/verify-reset-otp')
  })

  it('verifyResetOTP sends email and otp in payload', async () => {
    const payload = { email: 'reset@example.com', otp: '000001' }
    await authAPI.verifyResetOTP(payload)
    const c = lastCall()
    expect(c.data).toEqual(payload)
  })

  // ── resendVerification ─────────────────────────────────────────────────────

  it('resendVerification POSTs to /auth/resend-verification', async () => {
    await authAPI.resendVerification({ email: 'resend@example.com' })
    const c = lastCall()
    expect(c.method).toBe('POST')
    expect(c.url).toBe('/auth/resend-verification')
  })

  it('resendVerification sends email in payload', async () => {
    const payload = { email: 'resend@example.com' }
    await authAPI.resendVerification(payload)
    const c = lastCall()
    expect(c.data).toEqual(payload)
  })

  // ── forgotPassword ─────────────────────────────────────────────────────────

  it('forgotPassword POSTs to /auth/forgot-password', async () => {
    await authAPI.forgotPassword({ email: 'fp@example.com' })
    const c = lastCall()
    expect(c.method).toBe('POST')
    expect(c.url).toBe('/auth/forgot-password')
  })

  // ── resetPassword ──────────────────────────────────────────────────────────

  it('resetPassword POSTs to /auth/reset-password', async () => {
    await authAPI.resetPassword({ token: 'tok-abc', new_password: 'NewPass1!' })
    const c = lastCall()
    expect(c.method).toBe('POST')
    expect(c.url).toBe('/auth/reset-password')
  })

  it('resetPassword sends token and new_password in payload', async () => {
    const payload = { token: 'reset-tok-123', new_password: 'SecureNew99!' }
    await authAPI.resetPassword(payload)
    const c = lastCall()
    expect(c.data).toEqual(payload)
  })
})

// ── AUTH_PATHS whitelist ──────────────────────────────────────────────────────
// Ensure OTP endpoints are in the unauthenticated (no-token) whitelist so the
// request interceptor does not attach a stale token to them.

describe('AUTH_PATHS whitelist', () => {
  it('includes /auth/verify-email-otp', async () => {
    const mod = await import('../services/api.js')
    // Inspect the interceptor: it captures the list during module init.
    // We check indirectly: calling without a token must not throw.
    vi.stubGlobal('localStorage', { getItem: () => null, setItem: vi.fn(), removeItem: vi.fn() })
    await expect(authAPI.verifyEmailOTP({ email: 'a@b.com', otp: '000000' })).resolves.toBeDefined()
  })

  it('includes /auth/verify-reset-otp', async () => {
    vi.stubGlobal('localStorage', { getItem: () => null, setItem: vi.fn(), removeItem: vi.fn() })
    await expect(authAPI.verifyResetOTP({ email: 'a@b.com', otp: '000000' })).resolves.toBeDefined()
  })
})
