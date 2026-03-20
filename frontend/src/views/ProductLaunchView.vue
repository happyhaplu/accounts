<template>
  <div class="launch-shell">
    <div class="launch-card">
      <!-- Logo -->
      <div class="launch-logo">
        <img src="/logo.svg" alt="Gour" class="launch-logo-img" />
      </div>

      <!-- Loading state -->
      <template v-if="state === 'loading'">
        <div class="launch-spinner"></div>
        <p class="launch-msg">Preparing to launch <strong>{{ displayName }}</strong>…</p>
        <p class="launch-sub">Checking your subscription and signing you in.</p>
      </template>

      <!-- Redirecting state -->
      <template v-else-if="state === 'redirecting'">
        <div class="launch-spinner"></div>
        <p class="launch-msg">Redirecting to <strong>{{ redirectHost }}</strong>…</p>
        <p class="launch-sub">You'll arrive in a moment.</p>
      </template>

      <!-- Error state -->
      <template v-else-if="state === 'error'">
        <div class="launch-icon-wrap error-icon">
          <svg width="28" height="28" viewBox="0 0 24 24" fill="none"
               stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <circle cx="12" cy="12" r="10"/>
            <line x1="12" y1="8" x2="12" y2="12"/>
            <line x1="12" y1="16" x2="12.01" y2="16"/>
          </svg>
        </div>
        <p class="launch-msg">{{ errorTitle }}</p>
        <p class="launch-sub">{{ errorDetail }}</p>
        <div class="launch-actions">
          <button v-if="showBillingBtn" class="btn-primary" @click="goToBilling">
            View billing &amp; plans
          </button>
          <button class="btn-secondary" @click="goToDashboard">
            Go to dashboard
          </button>
        </div>
      </template>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import { authAPI } from '../services/api'

const route  = useRoute()
const router = useRouter()
const auth   = useAuthStore()

// ── Route params & query ──────────────────────────────────────────────────
const slug        = route.params.slug
const redirectUri = route.query.redirect_uri ?? ''

// ── Display helpers ───────────────────────────────────────────────────────
const PRODUCT_META = {
  'email-warmup': 'Email Warmup',
  'cold_email':   'Cold Email',
  'reach':        'Reach — LinkedIn Automation',
  'warmup':       'Inbox Warmup',
}
const displayName = PRODUCT_META[slug] ?? slug

const redirectHost = (() => {
  try { return new URL(redirectUri).hostname } catch { return '' }
})()

// ── State machine: loading → redirecting | error ──────────────────────────
const state        = ref('loading')
const errorTitle   = ref('')
const errorDetail  = ref('')
const showBillingBtn = ref(false)

// ── Navigation helpers ────────────────────────────────────────────────────
function goToBilling() {
  router.push('/billing')
}
function goToDashboard() {
  router.push('/dashboard')
}

// ── Main orchestration ────────────────────────────────────────────────────
onMounted(async () => {
  // 1. If user is not logged in → redirect to /login with redirect_uri
  //    The LoginView already handles sending redirect_uri to the backend,
  //    which signs a launch JWT and returns redirect_url in the response.
  if (!auth.isAuthenticated) {
    const loginQuery = { redirect_uri: redirectUri }
    router.replace({ path: '/login', query: loginQuery })
    return
  }

  // 2. User is logged in → call the launch API
  try {
    const { data } = await authAPI.launchProduct(slug, redirectUri || undefined)

    if (data.redirect_url) {
      state.value = 'redirecting'
      // Small delay so the user sees "Redirecting…" before navigation
      setTimeout(() => {
        window.location.href = data.redirect_url
      }, 300)
    } else {
      state.value     = 'error'
      errorTitle.value  = 'No redirect URL'
      errorDetail.value = 'The server did not return a redirect URL. Please try launching from the dashboard.'
    }
  } catch (err) {
    state.value = 'error'
    const status  = err.response?.status
    const message = err.response?.data?.error ?? ''

    if (status === 403) {
      // No active subscription
      errorTitle.value    = 'Subscription required'
      errorDetail.value   = `You don't have an active subscription for ${displayName}. Subscribe from the billing page to get started.`
      showBillingBtn.value = true
    } else if (status === 404) {
      errorTitle.value  = 'Product not found'
      errorDetail.value = `"${slug}" is not a recognised product. Please check the URL or go to the dashboard.`
    } else if (status === 400) {
      errorTitle.value  = 'Invalid redirect'
      errorDetail.value = message || 'The redirect URL is not allowed for this product.'
    } else {
      errorTitle.value  = 'Something went wrong'
      errorDetail.value = message || 'An unexpected error occurred. Please try again.'
    }
  }
})
</script>

<style scoped>
.launch-shell {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #f1f3f4;
  padding: 2rem;
}

.launch-card {
  background: #fff;
  border-radius: 12px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.08);
  padding: 2.5rem 2.5rem 2rem;
  max-width: 420px;
  width: 100%;
  text-align: center;
}

/* Logo */
.launch-logo {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 10px;
  margin-bottom: 2rem;
}
.launch-logo-img {
  height: 32px;
  width: auto;
}

/* Spinner */
.launch-spinner {
  width: 36px;
  height: 36px;
  border: 3px solid #e8eaed;
  border-top-color: #1a73e8;
  border-radius: 50%;
  margin: 0 auto 1.25rem;
  animation: spin 0.8s linear infinite;
}
@keyframes spin {
  to { transform: rotate(360deg); }
}

/* Messages */
.launch-msg {
  font-size: 1rem;
  font-weight: 600;
  color: #202124;
  margin-bottom: 0.35rem;
}
.launch-sub {
  font-size: 0.85rem;
  color: #5f6368;
  margin-bottom: 0;
}

/* Error icon */
.launch-icon-wrap {
  width: 52px;
  height: 52px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  margin: 0 auto 1.25rem;
}
.error-icon {
  background: #fce8e6;
  color: #d93025;
}

/* Action buttons */
.launch-actions {
  margin-top: 1.5rem;
  display: flex;
  flex-direction: column;
  gap: 0.6rem;
}
.btn-primary {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  background: #1a73e8;
  color: #fff;
  border: none;
  border-radius: 6px;
  padding: 10px 20px;
  font-size: 0.875rem;
  font-weight: 600;
  cursor: pointer;
  transition: background 0.15s;
}
.btn-primary:hover { background: #1765cc; }
.btn-secondary {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  background: transparent;
  color: #1a73e8;
  border: 1px solid #dadce0;
  border-radius: 6px;
  padding: 10px 20px;
  font-size: 0.875rem;
  font-weight: 600;
  cursor: pointer;
  transition: background 0.15s, border-color 0.15s;
}
.btn-secondary:hover {
  background: #f1f3f4;
  border-color: #c6c9cc;
}
</style>
