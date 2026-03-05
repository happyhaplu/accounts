import { describe, it, expect, beforeEach, vi } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import { useAuthStore } from '../stores/auth'

// Mock localStorage
const localStorageMock = (() => {
  let store = {}
  return {
    getItem:    (k) => store[k] ?? null,
    setItem:    (k, v) => { store[k] = String(v) },
    removeItem: (k) => { delete store[k] },
    clear:      () => { store = {} },
  }
})()
Object.defineProperty(global, 'localStorage', { value: localStorageMock })

describe('useAuthStore', () => {
  beforeEach(() => {
    localStorage.clear()
    setActivePinia(createPinia())
  })

  // ── Initial state ──────────────────────────────────────────────────────────

  it('starts unauthenticated when localStorage is empty', () => {
    const auth = useAuthStore()
    expect(auth.isAuthenticated).toBe(false)
    expect(auth.token).toBeNull()
    expect(auth.user).toBeNull()
  })

  it('reads existing token from localStorage on creation', () => {
    localStorage.setItem('oc_token', 'existing-token')
    localStorage.setItem('oc_user', JSON.stringify({ id: '1', email: 'x@y.com' }))
    setActivePinia(createPinia())
    const auth = useAuthStore()
    expect(auth.isAuthenticated).toBe(true)
    expect(auth.token).toBe('existing-token')
    expect(auth.user.email).toBe('x@y.com')
  })

  // ── setAuth ────────────────────────────────────────────────────────────────

  it('setAuth stores token and user', () => {
    const auth = useAuthStore()
    const user = { id: 'uid-1', email: 'test@example.com', profile_complete: true }
    auth.setAuth('jwt-token-123', user)

    expect(auth.token).toBe('jwt-token-123')
    expect(auth.user).toEqual(user)
    expect(auth.isAuthenticated).toBe(true)
    expect(localStorage.getItem('oc_token')).toBe('jwt-token-123')
    expect(JSON.parse(localStorage.getItem('oc_user')).email).toBe('test@example.com')
  })

  // ── updateUser ────────────────────────────────────────────────────────────

  it('updateUser updates user without changing token', () => {
    const auth = useAuthStore()
    auth.setAuth('tok-abc', { id: '1', email: 'a@b.com', profile_complete: false })
    auth.updateUser({ id: '1', email: 'a@b.com', profile_complete: true, name: 'Alice' })

    expect(auth.token).toBe('tok-abc')
    expect(auth.user.profile_complete).toBe(true)
    expect(auth.user.name).toBe('Alice')
    const stored = JSON.parse(localStorage.getItem('oc_user'))
    expect(stored.name).toBe('Alice')
  })

  // ── logout ────────────────────────────────────────────────────────────────

  it('logout clears token, user, and localStorage', () => {
    const auth = useAuthStore()
    auth.setAuth('tok-xyz', { id: '2', email: 'b@c.com' })
    auth.logout()

    expect(auth.token).toBeNull()
    expect(auth.user).toBeNull()
    expect(auth.isAuthenticated).toBe(false)
    expect(localStorage.getItem('oc_token')).toBeNull()
    expect(localStorage.getItem('oc_user')).toBeNull()
  })

  // ── isAuthenticated reactive behaviour ────────────────────────────────────

  it('isAuthenticated is false with empty string token', () => {
    const auth = useAuthStore()
    auth.setAuth('', { id: '3' })
    expect(auth.isAuthenticated).toBe(false)
  })

  it('isAuthenticated becomes true after setAuth and false after logout', () => {
    const auth = useAuthStore()
    expect(auth.isAuthenticated).toBe(false)
    auth.setAuth('tok', { id: '4', email: 'c@d.com' })
    expect(auth.isAuthenticated).toBe(true)
    auth.logout()
    expect(auth.isAuthenticated).toBe(false)
  })
})
