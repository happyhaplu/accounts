/**
 * k6 Load Test — Outcraftly Accounts API
 *
 * Install k6:  https://k6.io/docs/getting-started/installation/
 * Run:         k6 run scripts/load_test.js
 *              k6 run --vus 50 --duration 60s scripts/load_test.js
 *
 * Stages defined below ramp from 5 → 50 VUs over 2 minutes,
 * hold for 1 minute, then ramp back down.
 *
 * Thresholds (SLOs):
 *   - 95th percentile response time < 500 ms
 *   - Error rate < 1%
 */

import http from 'k6/http'
import { check, sleep, group } from 'k6'
import { Rate, Trend } from 'k6/metrics'

// ── Custom metrics ────────────────────────────────────────────────────────────
const errorRate     = new Rate('errors')
const registerTrend = new Trend('register_duration', true)
const loginTrend    = new Trend('login_duration', true)
const otpTrend      = new Trend('verify_otp_duration', true)
const profileTrend  = new Trend('profile_duration', true)

// ── Config ────────────────────────────────────────────────────────────────────
const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080/api/v1'

export const options = {
  stages: [
    { duration: '30s', target: 5  },   // warm-up
    { duration: '60s', target: 20 },   // ramp up
    { duration: '60s', target: 50 },   // peak load
    { duration: '30s', target: 50 },   // hold
    { duration: '30s', target: 0  },   // ramp down
  ],
  thresholds: {
    http_req_duration:    ['p(95)<500'],  // 95% of requests under 500 ms
    errors:               ['rate<0.01'],  // error rate < 1%
    register_duration:    ['p(95)<800'],  // registration can be a bit slower
    login_duration:       ['p(95)<400'],
    verify_otp_duration:  ['p(95)<400'],
    profile_duration:     ['p(95)<300'],
  },
}

const JSON_HEADERS = { 'Content-Type': 'application/json' }

// ── Helpers ───────────────────────────────────────────────────────────────────

function uniqueEmail() {
  return `loadtest+${Date.now()}_${Math.random().toString(36).slice(2, 8)}@example.com`
}

function post(path, payload, headers = JSON_HEADERS) {
  return http.post(`${BASE_URL}${path}`, JSON.stringify(payload), { headers })
}

function get(path, token) {
  return http.get(`${BASE_URL}${path}`, {
    headers: { Authorization: `Bearer ${token}` },
  })
}

// ── Scenarios ─────────────────────────────────────────────────────────────────

export default function () {
  // ── 1. Health check ────────────────────────────────────────────────────────
  group('health', () => {
    const res = http.get(`${BASE_URL}/health`)
    check(res, { 'health 200': r => r.status === 200 })
    errorRate.add(res.status !== 200)
  })

  sleep(0.2)

  // ── 2. Register (creates a new unique user each VU iteration) ──────────────
  const email    = uniqueEmail()
  const password = 'LoadTest1234!'
  let registerOK = false

  group('register', () => {
    const res = post('/auth/register', { email, password })
    registerTrend.add(res.timings.duration)
    registerOK = res.status === 201
    check(res, { 'register 201': r => r.status === 201 })
    errorRate.add(res.status !== 201)
  })

  if (!registerOK) return
  sleep(0.3)

  // ── 3. Forgot-password (OTP generation throughput) ─────────────────────────
  group('forgot-password', () => {
    const res = post('/auth/forgot-password', { email })
    check(res, { 'forgot-password 200': r => r.status === 200 })
    errorRate.add(res.status !== 200)
  })

  sleep(0.2)

  // ── 4. Login with wrong password (error-path latency) ─────────────────────
  group('login-wrong-password', () => {
    const res = post('/auth/login', { email, password: 'WrongPass!' })
    check(res, { 'login-wrong 401': r => r.status === 401 })
    // 401 is expected — do NOT count as error
  })

  sleep(0.2)

  // ── 5. Login with correct password (happy path) ────────────────────────────
  // Note: user is not email-verified yet in load test (no OTP step),
  // so login may return 403 if the backend enforces verified-only login.
  // Adjust expected status below to match your handler's behaviour.
  let jwtToken = null
  group('login', () => {
    const res = post('/auth/login', { email, password })
    loginTrend.add(res.timings.duration)
    if (res.status === 200) {
      try {
        jwtToken = res.json('token')
      } catch (_) {}
    }
    // 200 (verified) or 403 (unverified) are both "OK" for load testing purposes
    const ok = res.status === 200 || res.status === 403
    check(res, { 'login ok': () => ok })
    errorRate.add(!ok)
  })

  sleep(0.2)

  // ── 6. Authenticated: get profile (if we got a token) ─────────────────────
  if (jwtToken) {
    group('get-profile', () => {
      const res = get('/profile', jwtToken)
      profileTrend.add(res.timings.duration)
      check(res, { 'profile 200': r => r.status === 200 })
      errorRate.add(res.status !== 200)
    })
    sleep(0.2)

    group('list-workspaces', () => {
      const res = get('/workspaces', jwtToken)
      check(res, { 'workspaces 200': r => r.status === 200 })
      errorRate.add(res.status !== 200)
    })
  }

  // ── 7. Resend-verification (unauthenticated, anti-spam path) ───────────────
  group('resend-verification', () => {
    const res = post('/auth/resend-verification', { email })
    check(res, { 'resend 200': r => r.status === 200 })
    errorRate.add(res.status !== 200)
  })

  sleep(0.5)
}

// ── Smoke test (run with: k6 run --vus 1 --iterations 1 scripts/load_test.js) ─
export function smoke() {
  const res = http.get(`${BASE_URL}/health`)
  check(res, { 'smoke: health 200': r => r.status === 200 })
}
