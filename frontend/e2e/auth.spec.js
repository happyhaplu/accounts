import { test, expect } from '@playwright/test'

// ── Helpers ───────────────────────────────────────────────────────────────────

// Generate a unique email for each test run so tests are independent
function uniqueEmail(prefix = 'user') {
  return `${prefix}+${Date.now()}@example.com`
}

async function register(page, email, password) {
  await page.goto('/register')
  await page.fill('input[type="email"]', email)
  await page.fill('#password', password)
  await page.fill('#confirm', password)
  await page.click('button[type="submit"]')
}

async function loginDirect(page, email, password) {
  await page.goto('/login')
  await page.fill('input[type="email"]', email)
  await page.fill('input[type="password"]', password)
  await page.click('button[type="submit"]')
}

// ── Auth flows ────────────────────────────────────────────────────────────────

test.describe('Register page', () => {
  test('shows register form with email + password fields', async ({ page }) => {
    await page.goto('/register')
    await expect(page.locator('input[type="email"]')).toBeVisible()
    await expect(page.locator('#password')).toBeVisible()
    await expect(page.locator('#confirm')).toBeVisible()
    await expect(page.locator('button[type="submit"]')).toBeVisible()
  })

  test('submit button disabled for short password (< 8 chars)', async ({ page }) => {
    await page.goto('/register')
    await page.fill('input[type="email"]', 'test@example.com')
    await page.fill('#password', 'short')
    // canSubmit requires password >= 8 chars AND password === confirm, so button stays disabled
    await expect(page.locator('button[type="submit"]')).toBeDisabled()
  })

  test('submit button disabled when email field is empty', async ({ page }) => {
    await page.goto('/register')
    // canSubmit requires email to be non-empty — button is disabled on fresh load
    await expect(page.locator('button[type="submit"]')).toBeDisabled()
  })
})

test.describe('Login page', () => {
  test('shows login form', async ({ page }) => {
    await page.goto('/login')
    await expect(page.locator('input[type="email"]')).toBeVisible()
    await expect(page.locator('input[type="password"]')).toBeVisible()
    await expect(page.locator('button[type="submit"]')).toBeVisible()
  })

  test('shows error for wrong credentials', async ({ page }) => {
    await page.goto('/login')
    await page.fill('input[type="email"]', 'nobody@example.com')
    await page.fill('input[type="password"]', 'wrongpass123')
    await page.click('button[type="submit"]')
    // alert alert-error div appears when login fails; allow up to 10s for backend round-trip
    await expect(page.locator('.alert-error')).toBeVisible({ timeout: 10000 })
  })

  test('redirects unauthenticated user from /dashboard to /login', async ({ page }) => {
    await page.goto('/dashboard')
    await expect(page).toHaveURL(/\/login/)
  })

  test('redirects unauthenticated user from /settings to /login', async ({ page }) => {
    await page.goto('/settings')
    await expect(page).toHaveURL(/\/login/)
  })

  test('redirects unauthenticated user from /workspaces to /login', async ({ page }) => {
    await page.goto('/workspaces')
    await expect(page).toHaveURL(/\/login/)
  })

  test('redirects unauthenticated user from /billing to /login', async ({ page }) => {
    await page.goto('/billing')
    await expect(page).toHaveURL(/\/login/)
  })
})

test.describe('Forgot password page', () => {
  test('shows forgot password form', async ({ page }) => {
    await page.goto('/forgot-password')
    await expect(page.locator('input[type="email"]')).toBeVisible()
    await expect(page.locator('button[type="submit"]')).toBeVisible()
  })

  test('shows success message for any email (anti-enumeration)', async ({ page }) => {
    await page.goto('/forgot-password')
    await page.fill('input[type="email"]', 'nobody@example.com')
    await page.click('button[type="submit"]')
    // Success state renders .check-icon + "Check your inbox" heading (no success/alert class)
    await expect(page.locator('.check-icon')).toBeVisible({ timeout: 10000 })
    await expect(page.locator('h1')).toHaveText('Check your inbox')
  })
})

test.describe('Navigation', () => {
  test('login page has link to register', async ({ page }) => {
    await page.goto('/login')
    const link = page.locator('a[href*="register"]')
    await expect(link).toBeVisible()
  })

  test('register page has link to login', async ({ page }) => {
    await page.goto('/register')
    const link = page.locator('a[href*="login"]')
    await expect(link).toBeVisible()
  })

  test('/ redirects to /dashboard (which redirects to /login when unauthenticated)', async ({ page }) => {
    await page.goto('/')
    await expect(page).toHaveURL(/\/login/)
  })
})
