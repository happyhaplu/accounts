<template>
  <div class="dashboard">

    <!-- Blue header bar -->
    <div class="dash-hero">
      <div class="dash-hero-inner">
        <div class="dash-greeting">
          <p>Welcome back</p>
          <h1>{{ auth.user?.name || auth.user?.email }}</h1>
          <p v-if="auth.user?.name" class="dash-sub">{{ auth.user?.email }}</p>
        </div>
        <span class="status-chip">
          <span class="status-dot"></span>Active
        </span>
      </div>
    </div>

    <div class="page-body">

      <!-- Identity section -->
      <section class="section">
        <div class="section-head">
          <div class="section-icon">
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none"
                 stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"/>
              <circle cx="12" cy="7" r="4"/>
            </svg>
          </div>
          <h2>Account details</h2>
        </div>

        <div class="cards-row">
          <div class="info-card">
            <div class="info-card-icon blue-bg">
              <svg width="18" height="18" viewBox="0 0 24 24" fill="none"
                   stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <rect x="2" y="7" width="20" height="14" rx="2"/><path d="M16 21V5a2 2 0 0 0-2-2h-4a2 2 0 0 0-2 2v16"/>
              </svg>
            </div>
            <div>
              <div class="info-label">Account ID</div>
              <div class="info-value mono">{{ auth.user?.id }}</div>
            </div>
          </div>

          <div class="info-card">
            <div class="info-card-icon blue-bg">
              <svg width="18" height="18" viewBox="0 0 24 24" fill="none"
                   stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <path d="M4 4h16c1.1 0 2 .9 2 2v12c0 1.1-.9 2-2 2H4c-1.1 0-2-.9-2-2V6c0-1.1.9-2 2-2z"/>
                <polyline points="22,6 12,13 2,6"/>
              </svg>
            </div>
            <div>
              <div class="info-label">Email address</div>
              <div class="info-value">{{ auth.user?.email }}</div>
            </div>
          </div>

          <div class="info-card">
            <div class="info-card-icon blue-bg">
              <svg width="18" height="18" viewBox="0 0 24 24" fill="none"
                   stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"/>
                <circle cx="12" cy="7" r="4"/>
              </svg>
            </div>
            <div>
              <div class="info-label">Full name</div>
              <div class="info-value">{{ auth.user?.name || '—' }}</div>
            </div>
          </div>

          <div class="info-card">
            <div class="info-card-icon blue-bg">
              <svg width="18" height="18" viewBox="0 0 24 24" fill="none"
                   stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <path d="M3 9l9-7 9 7v11a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z"/>
                <polyline points="9 22 9 12 15 12 15 22"/>
              </svg>
            </div>
            <div>
              <div class="info-label">Company</div>
              <div class="info-value">{{ auth.user?.company_name || '—' }}</div>
            </div>
          </div>

          <div v-if="auth.user?.job_title" class="info-card">
            <div class="info-card-icon blue-bg">
              <svg width="18" height="18" viewBox="0 0 24 24" fill="none"
                   stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <rect x="2" y="7" width="20" height="14" rx="2"/>
                <path d="M16 21V5a2 2 0 0 0-2-2h-4a2 2 0 0 0-2 2v16"/>
              </svg>
            </div>
            <div>
              <div class="info-label">Job title</div>
              <div class="info-value">{{ auth.user?.job_title }}</div>
            </div>
          </div>

          <div v-if="auth.user?.phone_number" class="info-card">
            <div class="info-card-icon blue-bg">
              <svg width="18" height="18" viewBox="0 0 24 24" fill="none"
                   stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <path d="M22 16.92v3a2 2 0 0 1-2.18 2 19.79 19.79 0 0 1-8.63-3.07A19.5 19.5 0 0 1 4.69 12 19.79 19.79 0 0 1 1.61 3.4 2 2 0 0 1 3.6 1.22h3a2 2 0 0 1 2 1.72 12.84 12.84 0 0 0 .7 2.81 2 2 0 0 1-.45 2.11L7.91 8.8a16 16 0 0 0 6.29 6.29l.96-.96a2 2 0 0 1 2.11-.45 12.84 12.84 0 0 0 2.81.7A2 2 0 0 1 22 16.92z"/>
              </svg>
            </div>
            <div>
              <div class="info-label">Phone number</div>
              <div class="info-value">{{ auth.user?.phone_country_code }} {{ auth.user?.phone_number }}</div>
            </div>
          </div>

          <div class="info-card">
            <div class="info-card-icon blue-bg">
              <svg width="18" height="18" viewBox="0 0 24 24" fill="none"
                   stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <rect x="3" y="11" width="18" height="11" rx="2" ry="2"/>
                <path d="M7 11V7a5 5 0 0 1 10 0v4"/>
              </svg>
            </div>
            <div>
              <div class="info-label">Auth method</div>
              <div class="info-value">Email &amp; Password</div>
            </div>
          </div>

          <div class="info-card">
            <div class="info-card-icon" :class="auth.user?.email_verified ? 'green-bg' : 'yellow-bg'">
              <svg width="18" height="18" viewBox="0 0 24 24" fill="none"
                   stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <path v-if="auth.user?.email_verified" d="M22 11.08V12a10 10 0 1 1-5.93-9.14"/>
                <polyline v-if="auth.user?.email_verified" points="22 4 12 14.01 9 11.01"/>
                <circle v-if="!auth.user?.email_verified" cx="12" cy="12" r="10"/>
                <line v-if="!auth.user?.email_verified" x1="12" y1="8" x2="12" y2="12"/>
                <line v-if="!auth.user?.email_verified" x1="12" y1="16" x2="12.01" y2="16"/>
              </svg>
            </div>
            <div>
              <div class="info-label">Email status</div>
              <div class="info-value">
                {{ auth.user?.email_verified ? 'Verified' : 'Not verified' }}
              </div>
            </div>
          </div>

          <div v-if="auth.user?.last_login_at" class="info-card">
            <div class="info-card-icon blue-bg">
              <svg width="18" height="18" viewBox="0 0 24 24" fill="none"
                   stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <circle cx="12" cy="12" r="10"/>
                <polyline points="12 6 12 12 16 14"/>
              </svg>
            </div>
            <div>
              <div class="info-label">Last sign-in</div>
              <div class="info-value">{{ formatDate(auth.user?.last_login_at) }}</div>
            </div>
          </div>

          <div class="info-card">
            <div class="info-card-icon blue-bg">
              <svg width="18" height="18" viewBox="0 0 24 24" fill="none"
                   stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <rect x="3" y="4" width="18" height="18" rx="2" ry="2"/>
                <line x1="16" y1="2" x2="16" y2="6"/><line x1="8" y1="2" x2="8" y2="6"/>
                <line x1="3" y1="10" x2="21" y2="10"/>
              </svg>
            </div>
            <div>
              <div class="info-label">Member since</div>
              <div class="info-value">{{ formatDate(auth.user?.created_at) }}</div>
            </div>
          </div>
        </div>
      </section>

      <!-- Products section -->
      <section class="section">
        <div class="section-head">
          <div class="section-icon">
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none"
                 stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <rect x="2" y="3" width="6" height="6" rx="1"/><rect x="16" y="3" width="6" height="6" rx="1"/>
              <rect x="2" y="15" width="6" height="6" rx="1"/><rect x="16" y="15" width="6" height="6" rx="1"/>
              <path d="M9 6h6M9 18h6M5 9v6M19 9v6"/>
            </svg>
          </div>
          <h2>Products</h2>
        </div>
        <p class="section-sub">Outcraftly products available to your workspace.</p>

        <div v-if="productsLoading" class="products-loading">Loading products…</div>

        <div v-else class="products-grid">
          <div v-for="prod in enrichedProducts" :key="prod.id" class="product-card">
            <div class="product-icon-wrap">{{ prod.icon }}</div>
            <div class="product-name">{{ prod.displayName }}</div>

            <!-- Active subscription → Launch button -->
            <template v-if="prod.subscribed">
              <span class="product-badge badge-active">Active</span>
              <button
                v-if="prod.redirect_urls && prod.redirect_urls.length"
                class="launch-btn"
                :disabled="launching === prod.name"
                @click="launch(prod.name)"
              >
                {{ launching === prod.name ? 'Launching…' : 'Launch' }}
              </button>
            </template>

            <!-- No subscription -->
            <span v-else class="product-badge">No subscription</span>
          </div>

          <!-- Empty state -->
          <div v-if="!enrichedProducts.length" class="product-card empty-card">
            <div class="product-icon-wrap">📦</div>
            <div class="product-name">No products yet</div>
            <span class="product-badge">Coming soon</span>
          </div>
        </div>

        <p v-if="launchError" class="launch-error">{{ launchError }}</p>
      </section>

    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useAuthStore } from '../stores/auth'
import { authAPI } from '../services/api'

const auth = useAuthStore()

// ── Product display meta ──────────────────────────────────────────────────
const PRODUCT_META = {
  'email-warmup': { icon: '🔥', displayName: 'Email Warmup' },
  'cold_email':   { icon: '📧', displayName: 'Cold Email' },
  'linkedin':     { icon: '💼', displayName: 'LinkedIn Outreach' },
  'reach':        { icon: '📡', displayName: 'Reach' },
  'warmup':       { icon: '🌡️', displayName: 'Inbox Warmup' },
}

// ── State ─────────────────────────────────────────────────────────────────
const rawProducts      = ref([])
const activeSubNames   = ref(new Set())
const productsLoading  = ref(true)
const launching        = ref(null)
const launchError      = ref('')

// ── Computed: merge product list with subscription info ───────────────────
const enrichedProducts = computed(() =>
  rawProducts.value.map(p => ({
    ...p,
    icon:        PRODUCT_META[p.name]?.icon        ?? '🚀',
    displayName: PRODUCT_META[p.name]?.displayName ?? p.name,
    subscribed:  activeSubNames.value.has(p.name),
  }))
)

// ── Data loading ──────────────────────────────────────────────────────────
onMounted(async () => {
  try {
    // Load all active products
    const prodRes = await authAPI.listProducts()
    rawProducts.value = prodRes.data.products ?? []

    // Load subscriptions for the user's workspace
    const wsRes = await authAPI.listWorkspaces()
    const workspaces = wsRes.data.workspaces ?? []
    if (workspaces.length) {
      const wsId = workspaces[0].id
      const subRes = await authAPI.listSubscriptions(wsId)
      const subs = subRes.data.subscriptions ?? []
      // Build a set of product names with active subscriptions
      const productIds = new Set(subs.filter(s => s.status === 'active').map(s => s.product_id))
      // Map product UUIDs → names
      const nameSet = new Set(
        rawProducts.value
          .filter(p => productIds.has(p.id))
          .map(p => p.name)
      )
      activeSubNames.value = nameSet
    }
  } catch (e) {
    console.error('Failed to load products/subscriptions', e)
  } finally {
    productsLoading.value = false
  }
})

// ── Actions ───────────────────────────────────────────────────────────────
async function launch(productName) {
  if (launching.value) return
  launchError.value = ''
  launching.value = productName
  try {
    const { data } = await authAPI.launchProduct(productName)
    if (data.redirect_url) {
      window.location.href = data.redirect_url
    } else {
      launching.value = null
      launchError.value = 'No redirect URL returned. Please try again.'
    }
  } catch (err) {
    launching.value = null
    launchError.value = err.response?.data?.error ?? 'Could not launch product. Please try again.'
  }
}

function formatDate(iso) {
  if (!iso) return '—'
  return new Date(iso).toLocaleDateString('en-US', {
    year: 'numeric', month: 'short', day: 'numeric',
  })
}
</script>

<style scoped>
.dashboard { min-height: 100vh; background: var(--bg, #f1f3f4); }

/* Hero bar */
.dash-hero {
  background: linear-gradient(135deg, var(--blue, #1a73e8) 0%, var(--blue-darker, #0d47a1) 100%);
  padding: 2.25rem 2rem;
  color: #fff;
}
.dash-hero-inner {
  max-width: 1100px; margin: 0 auto;
  display: flex; align-items: center; justify-content: space-between; flex-wrap: wrap; gap: 1rem;
}
.dash-greeting p  { font-size: 0.85rem; opacity: 0.8; margin-bottom: 4px; }
.dash-greeting h1 { font-size: 1.4rem; font-weight: 600; letter-spacing: -0.3px; }
.dash-sub { font-size: 0.8rem; opacity: 0.65; margin-top: 3px; margin-bottom: 0; }

.status-chip {
  display: inline-flex; align-items: center; gap: 6px;
  background: rgba(255,255,255,0.18);
  border: 1px solid rgba(255,255,255,0.3);
  padding: 6px 14px; border-radius: 20px;
  font-size: 0.8rem; font-weight: 600; letter-spacing: 0.02em;
  color: #c8e6ff;
}
.status-dot {
  width: 7px; height: 7px; border-radius: 50%;
  background: #69f0ae; box-shadow: 0 0 0 2px rgba(105,240,174,0.3);
}

/* Page body */
.page-body { max-width: 1100px; margin: 0 auto; padding: 2rem; }

/* Section */
.section { margin-bottom: 2.5rem; }
.section-head {
  display: flex; align-items: center; gap: 10px; margin-bottom: 1.25rem;
}
.section-icon {
  width: 34px; height: 34px; border-radius: 8px;
  background: var(--blue-light, #e8f0fe); color: var(--blue, #1a73e8);
  display: flex; align-items: center; justify-content: center; flex-shrink: 0;
}
.section-head h2 { font-size: 1rem; font-weight: 700; color: var(--text, #202124); }
.section-sub { color: var(--text-muted, #5f6368); font-size: 0.875rem; margin: -0.75rem 0 1.25rem; }

/* Info cards */
.cards-row {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(260px, 1fr));
  gap: 1rem;
}
.info-card {
  background: #fff;
  border: 1px solid var(--border, #dadce0);
  border-radius: 12px;
  padding: 1.1rem 1.25rem;
  display: flex; align-items: flex-start; gap: 12px;
  transition: box-shadow 0.18s;
}
.info-card:hover { box-shadow: 0 2px 10px rgba(0,0,0,0.08); }

.info-card-icon {
  width: 36px; height: 36px; border-radius: 9px; flex-shrink: 0;
  display: flex; align-items: center; justify-content: center;
}
.blue-bg   { background: var(--blue-light, #e8f0fe); color: var(--blue, #1a73e8); }
.green-bg  { background: #e8f5e9; color: #2e7d32; }
.yellow-bg { background: #fff8e1; color: #f57f17; }

.info-label {
  font-size: 11px; text-transform: uppercase; letter-spacing: 0.06em;
  color: var(--text-muted, #5f6368); margin-bottom: 3px; font-weight: 500;
}
.info-value { font-size: 0.875rem; font-weight: 600; color: var(--text, #202124); word-break: break-all; }
.info-value.mono { font-family: 'Courier New', monospace; font-size: 0.75rem; font-weight: 400; }

/* Products grid */
.products-loading { color: var(--text-muted, #5f6368); font-size: 0.875rem; padding: 1rem 0; }
.products-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
  gap: 1rem;
}
.product-card {
  background: #fff;
  border: 1.5px dashed var(--border, #dadce0);
  border-radius: 12px;
  padding: 1.75rem 1rem;
  text-align: center;
  transition: border-color 0.18s, box-shadow 0.18s;
  display: flex; flex-direction: column; align-items: center; gap: 0.5rem;
}
.product-card:hover {
  border-color: var(--blue, #1a73e8);
  box-shadow: 0 2px 10px rgba(26,115,232,0.08);
}
.empty-card { opacity: 0.5; }
.product-icon-wrap { font-size: 2rem; margin-bottom: 0.25rem; }
.product-name  { font-weight: 600; font-size: 0.9rem; color: var(--text, #202124); }
.product-badge {
  display: inline-block;
  background: var(--blue-light, #e8f0fe);
  color: var(--blue, #1a73e8);
  font-size: 11px; font-weight: 600;
  padding: 3px 10px; border-radius: 20px;
}
.badge-active {
  background: #e8f5e9;
  color: #2e7d32;
}
.launch-btn {
  margin-top: 0.25rem;
  background: var(--blue, #1a73e8);
  color: #fff;
  border: none;
  border-radius: 8px;
  padding: 7px 20px;
  font-size: 0.82rem;
  font-weight: 600;
  cursor: pointer;
  transition: background 0.18s, opacity 0.18s;
  width: 100%;
}
.launch-btn:hover:not(:disabled) { background: var(--blue-darker, #0d47a1); }
.launch-btn:disabled { opacity: 0.6; cursor: not-allowed; }
.launch-error {
  margin-top: 0.75rem;
  color: #c62828;
  font-size: 0.85rem;
}

@media (max-width: 600px) {
  .dash-hero { padding: 1.5rem 1rem; }
  .page-body { padding: 1.25rem 1rem; }
}
</style>
