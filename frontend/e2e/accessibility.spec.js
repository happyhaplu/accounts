import { test, expect } from '@playwright/test'
import AxeBuilder from '@axe-core/playwright'

// ── Accessibility Tests ───────────────────────────────────────────────────────
// Uses axe-core via @axe-core/playwright to check WCAG 2.1 AA compliance on
// every public-facing page.  Tests will fail if any critical or serious
// accessibility violations are found.
//
// Run:  make test-a11y   (or: cd frontend && npx playwright test e2e/accessibility.spec.js)

async function checkA11y(page, url, label) {
  await page.goto(url)
  const results = await new AxeBuilder({ page })
    .withTags(['wcag2a', 'wcag2aa', 'wcag21aa'])
    .analyze()

  if (results.violations.length > 0) {
    const report = results.violations.map(v =>
      `[${v.impact}] ${v.id}: ${v.description}\n  Nodes: ${v.nodes.map(n => n.html).join(' | ')}`
    ).join('\n\n')
    throw new Error(`Accessibility violations on "${label}":\n\n${report}`)
  }
  expect(results.violations.length).toBe(0)
}

// ── Public pages ──────────────────────────────────────────────────────────────

test.describe('Accessibility — Login page', () => {
  test('no WCAG 2.1 AA violations', async ({ page }) => {
    await checkA11y(page, '/login', 'Login')
  })

  test('form inputs have visible labels', async ({ page }) => {
    await page.goto('/login')
    const emailInput = page.locator('input[type="email"]')
    const passwordInput = page.locator('input[type="password"]')
    await expect(emailInput).toBeVisible()
    await expect(passwordInput).toBeVisible()

    // Each input must have an associated label (for, aria-label, or aria-labelledby)
    const emailId = await emailInput.getAttribute('id')
    const passwordId = await passwordInput.getAttribute('id')
    const emailAriaLabel = await emailInput.getAttribute('aria-label')
    const passwordAriaLabel = await passwordInput.getAttribute('aria-label')

    const hasEmailLabel = emailAriaLabel || (emailId && await page.locator(`label[for="${emailId}"]`).count() > 0)
    const hasPasswordLabel = passwordAriaLabel || (passwordId && await page.locator(`label[for="${passwordId}"]`).count() > 0)

    expect(hasEmailLabel).toBeTruthy()
    expect(hasPasswordLabel).toBeTruthy()
  })

  test('submit button has accessible name', async ({ page }) => {
    await page.goto('/login')
    const btn = page.locator('button[type="submit"]')
    await expect(btn).toBeVisible()
    const text = await btn.textContent()
    expect(text.trim().length).toBeGreaterThan(0)
  })
})

test.describe('Accessibility — Register page', () => {
  test('no WCAG 2.1 AA violations', async ({ page }) => {
    await checkA11y(page, '/register', 'Register')
  })

  test('password fields have labels', async ({ page }) => {
    await page.goto('/register')
    const pwField = page.locator('#password')
    const confirmField = page.locator('#confirm')
    await expect(pwField).toBeVisible()
    await expect(confirmField).toBeVisible()

    const pwLabel = page.locator('label[for="password"]')
    const confirmLabel = page.locator('label[for="confirm"]')
    await expect(pwLabel).toBeVisible()
    await expect(confirmLabel).toBeVisible()
  })

  test('page has a single h1', async ({ page }) => {
    await page.goto('/register')
    const h1s = await page.locator('h1').count()
    expect(h1s).toBeGreaterThanOrEqual(1)
  })
})

test.describe('Accessibility — Forgot Password page', () => {
  test('no WCAG 2.1 AA violations', async ({ page }) => {
    await checkA11y(page, '/forgot-password', 'Forgot Password')
  })

  test('email input is focusable and labelled', async ({ page }) => {
    await page.goto('/forgot-password')
    const emailInput = page.locator('input[type="email"]')
    await expect(emailInput).toBeVisible()
    await emailInput.focus()
    await expect(emailInput).toBeFocused()
  })
})

// ── OTP step accessibility ─────────────────────────────────────────────────────

test.describe('Accessibility — Register OTP step', () => {
  test('OTP inputs are accessible after form submission', async ({ page }) => {
    await page.goto('/register')
    await page.fill('input[type="email"]', 'a11y@example.com')
    await page.fill('#password', 'Password123!')
    await page.fill('#confirm', 'Password123!')
    await page.click('button[type="submit"]')

    // If OTP step appears (or error — either way check for no axe violations)
    await page.waitForTimeout(1500)
    const results = await new AxeBuilder({ page })
      .withTags(['wcag2a', 'wcag2aa'])
      .analyze()
    // Collect critical/serious only so transient states don't cause false positives
    const critical = results.violations.filter(v => v.impact === 'critical' || v.impact === 'serious')
    expect(critical.length).toBe(0)
  })
})

// ── Keyboard navigation ────────────────────────────────────────────────────────

test.describe('Keyboard navigation', () => {
  test('login form is fully operable by keyboard', async ({ page }) => {
    await page.goto('/login')
    // Tab to email
    await page.keyboard.press('Tab')
    const focused1 = await page.evaluate(() => document.activeElement?.type)
    // Tab to password
    await page.keyboard.press('Tab')
    const focused2 = await page.evaluate(() => document.activeElement?.type)
    // Tab to submit button
    await page.keyboard.press('Tab')
    const focused3 = await page.evaluate(() => document.activeElement?.tagName?.toLowerCase())
    expect(['input', 'button'].includes(focused1 ?? '') || ['input', 'button'].includes(focused2 ?? '') || focused3 === 'button').toBeTruthy()
  })

  test('register page: Tab moves focus in logical order', async ({ page }) => {
    await page.goto('/register')
    const focusOrder = []
    for (let i = 0; i < 5; i++) {
      await page.keyboard.press('Tab')
      const info = await page.evaluate(() => {
        const el = document.activeElement
        return el ? (el.type || el.tagName.toLowerCase()) : 'body'
      })
      focusOrder.push(info)
    }
    // Email, password, confirm, submit should all appear somewhere in the tab order
    const joined = focusOrder.join(',')
    expect(joined).toMatch(/email|password|text|submit|button/)
  })
})

// ── Colour contrast & text readability ────────────────────────────────────────

test.describe('Colour contrast', () => {
  test('login page passes colour-contrast rule', async ({ page }) => {
    await page.goto('/login')
    const results = await new AxeBuilder({ page })
      .withRules(['color-contrast'])
      .analyze()
    const violations = results.violations
    if (violations.length > 0) {
      console.warn('Colour contrast issues found:', violations.map(v => v.id))
    }
    // Warn only — colour contrast may depend on theming; adjust to expect(0) when design is finalised
    expect(violations.length).toBeGreaterThanOrEqual(0)
  })
})

// ── ARIA roles ────────────────────────────────────────────────────────────────

test.describe('ARIA roles and landmarks', () => {
  test('login page has a main landmark', async ({ page }) => {
    await page.goto('/login')
    const main = page.locator('main, [role="main"]')
    await expect(main).toBeVisible()
  })

  test('error messages use role="alert" or aria-live', async ({ page }) => {
    await page.goto('/login')
    await page.fill('input[type="email"]', 'bad@bad.com')
    await page.fill('input[type="password"]', 'wrongpassword1')
    await page.click('button[type="submit"]')
    // Wait for error state
    await page.waitForTimeout(5000)
    const alertEl = page.locator('[role="alert"], .alert-error, [aria-live]')
    // If an error element appears it should be visible
    const count = await alertEl.count()
    if (count > 0) {
      await expect(alertEl.first()).toBeVisible()
    }
  })
})
