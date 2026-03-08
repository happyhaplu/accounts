<template>
  <div class="settings-page">

    <!-- Page header -->
    <div class="page-header">
      <div class="page-header-inner">
        <h1>Billing &amp; Subscriptions</h1>
        <p>Manage your Stripe billing and active subscriptions.</p>
      </div>
    </div>

    <div class="page-body">

      <!-- ── Workspace selector ───────────────────────────────── -->
      <section class="section">
        <div class="card">

          <!-- skeleton -->
          <div v-if="wsLoading" class="profile-skeleton">
            <div class="skeleton-row" style="width:55%;height:28px"></div>
            <div class="skeleton-row" style="height:36px"></div>
          </div>

          <!-- error -->
          <div v-else-if="wsListError" class="alert alert-error">{{ wsListError }}</div>

          <!-- workspace tabs -->
          <template v-else>
            <p class="ws-selector-label">Select workspace</p>
            <div class="ws-tabs">
              <button
                v-for="item in wsList" :key="item.id"
                class="ws-tab"
                :class="{ active: item.id === activeWsId }"
                @click="switchWorkspace(item.id)"
              >
                <span class="ws-tab-dot" :class="item.my_role"></span>
                {{ item.name }}
              </button>
            </div>
          </template>

        </div>
      </section>

      <!-- ── Products & Plans ─────────────────────────────────── -->
      <section v-if="ws" class="section">
        <div class="section-head">
          <div class="section-icon">
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none"
                 stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <polyline points="22 12 18 12 15 21 9 3 6 12 2 12"/>
            </svg>
          </div>
          <div>
            <h2>Products &amp; Plans</h2>
            <p class="section-sub">Subscribe to products for <strong>{{ ws.name }}</strong>.</p>
          </div>
        </div>

        <div class="card">
          <div v-if="prodLoading" class="profile-skeleton">
            <div class="skeleton-row" style="height:72px"></div>
            <div class="skeleton-row" style="height:72px"></div>
            <div class="skeleton-row" style="height:72px"></div>
          </div>
          <div v-else-if="prodError" class="alert alert-error">{{ prodError }}</div>
          <div v-else-if="products.length === 0" class="sub-empty">
            <p class="sub-empty-title">No products configured.</p>
          </div>
          <template v-else>
            <div class="prod-list">
              <div v-for="prod in products" :key="prod.id" class="prod-row">
                <!-- icon -->
                <div class="prod-icon" :data-name="prod.name">{{ prodInitial(prod.name) }}</div>

                <!-- info -->
                <div class="prod-info">
                  <div class="prod-name">{{ formatProductName(prod.name) }}</div>
                  <div class="prod-desc">{{ prod.description || 'No description.' }}</div>
                </div>

                <!-- status / action -->
                <div class="prod-action">
                  <template v-if="activeSubFor(prod.id)">
                    <span class="sub-badge active">✓ Active</span>
                    <div class="prod-renew">
                      Renews {{ formatDate(activeSubFor(prod.id).current_period_end) }}
                    </div>
                  </template>
                  <template v-else-if="ws.my_role === 'owner' && prod.stripe_price_id">
                    <button
                      class="btn-subscribe"
                      :disabled="checkoutLoading === prod.id"
                      @click="doCheckout(prod)"
                    >
                      <svg width="12" height="12" viewBox="0 0 24 24" fill="none"
                           stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round">
                        <line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/>
                      </svg>
                      {{ checkoutLoading === prod.id ? 'Redirecting…' : 'Subscribe' }}
                    </button>
                  </template>
                  <template v-else-if="ws.my_role !== 'owner'">
                    <span class="prod-no-sub">Not subscribed</span>
                  </template>
                  <template v-else>
                    <!-- owner, no price_id: product not yet priced -->
                    <span class="prod-no-price">Price not configured</span>
                  </template>
                </div>
              </div>
            </div>
            <div v-if="checkoutError" class="alert alert-error" style="margin-top:1rem;">{{ checkoutError }}</div>
          </template>
        </div>
      </section>

      <!-- ── Billing ──────────────────────────────────────────── -->
      <section v-if="ws" class="section">
        <div class="section-head">
          <div class="section-icon">
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none"
                 stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <rect x="1" y="4" width="22" height="16" rx="2" ry="2"/>
              <line x1="1" y1="10" x2="23" y2="10"/>
            </svg>
          </div>
          <div>
            <h2>Billing</h2>
            <p class="section-sub">Stripe payments for <strong>{{ ws.name }}</strong>.</p>
          </div>
        </div>

        <div class="card">

          <!-- Loading -->
          <template v-if="billingLoading">
            <div class="skel-line" style="width:40%;margin-bottom:.75rem;"></div>
            <div class="skel-line" style="width:60%;margin-bottom:.5rem;"></div>
            <div class="skel-line" style="width:50%;margin-bottom:.5rem;"></div>
          </template>

          <!-- Error -->
          <div v-else-if="billingError" class="alert alert-error">{{ billingError }}</div>

          <!-- No Stripe customer yet -->
          <div v-else-if="!billing?.has_stripe_customer" class="billing-empty">
            <svg width="28" height="28" viewBox="0 0 24 24" fill="none"
                 stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
              <rect x="1" y="4" width="22" height="16" rx="2" ry="2"/>
              <line x1="1" y1="10" x2="23" y2="10"/>
            </svg>
            <div>
              <p class="billing-empty-title">No active subscriptions</p>
              <p class="billing-empty-sub">
                This workspace has no billing history yet. Once a subscription is activated,
                your plan, renewal date, and payment details will appear here.
              </p>
            </div>
          </div>

          <!-- Has Stripe customer -->
          <template v-else>
            <div class="billing-status-row">
              <span class="sub-badge active">✓ Billing active</span>
              <code class="billing-cus-id">{{ billing?.stripe_customer_id }}</code>
            </div>

            <div class="billing-actions">
              <button class="billing-btn" @click="openPortal" :disabled="portalLoading">
                <svg width="13" height="13" viewBox="0 0 24 24" fill="none"
                     stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round">
                  <path d="M18 13v6a2 2 0 01-2 2H5a2 2 0 01-2-2V8a2 2 0 012-2h6"/>
                  <polyline points="15 3 21 3 21 9"/><line x1="10" y1="14" x2="21" y2="3"/>
                </svg>
                {{ portalLoading ? 'Opening…' : 'Manage billing' }}
              </button>
              <button class="billing-btn billing-btn-ghost" @click="doSyncBilling()" :disabled="syncLoading">
                <svg width="13" height="13" viewBox="0 0 24 24" fill="none"
                     stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round">
                  <polyline points="23 4 23 10 17 10"/>
                  <path d="M20.49 15a9 9 0 1 1-2.12-9.36L23 10"/>
                </svg>
                {{ syncLoading ? 'Syncing…' : 'Refresh from Stripe' }}
              </button>
            </div>

            <div v-if="portalError" class="alert alert-error" style="margin-top:.75rem;">{{ portalError }}</div>
            <div v-if="syncMsg" class="billing-sync-msg">{{ syncMsg }}</div>

            <!-- Active subscriptions from Stripe -->
            <template v-if="billing?.subscriptions?.length">
              <div class="billing-subs-label">Active subscriptions</div>
              <div class="billing-sub-list">
                <div v-for="sub in billing.subscriptions" :key="sub.id" class="billing-sub-row">
                  <div class="billing-sub-icon">{{ (sub.product?.name ?? 'P')[0].toUpperCase() }}</div>
                  <div class="billing-sub-info">
                    <div class="billing-sub-product">{{ formatProductName(sub.product?.name) }}</div>
                    <div class="billing-sub-plan">{{ sub.plan_name }}</div>
                  </div>
                  <div class="billing-sub-meta">
                    <span class="sub-badge" :class="sub.status">{{ subStatusLabel(sub.status) }}</span>
                    <div v-if="sub.status === 'active'" class="billing-sub-renew">
                      Renews {{ formatDate(sub.current_period_end) }}
                    </div>
                  </div>
                </div>
              </div>
            </template>
            <div v-else class="billing-no-subs">No active subscriptions on this workspace.</div>
          </template>
        </div>
      </section>

      <!-- ── Subscriptions ────────────────────────────────────── -->
      <section v-if="ws" class="section">
        <div class="section-head">
          <div class="section-icon">
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none"
                 stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <path d="M21 16V8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16z"/>
            </svg>
          </div>
          <div>
            <h2>Subscriptions</h2>
            <p class="section-sub">Active plans for <strong>{{ ws.name }}</strong>.</p>
          </div>
        </div>

        <div class="card">

          <!-- loading -->
          <div v-if="subLoading" class="profile-skeleton">
            <div class="skeleton-row" style="height:52px"></div>
            <div class="skeleton-row" style="height:52px"></div>
          </div>

          <!-- error -->
          <div v-else-if="subError" class="alert alert-error">{{ subError }}</div>

          <!-- cancel feedback -->
          <div v-if="cancelSuccess" class="alert alert-success">{{ cancelSuccess }}</div>
          <div v-if="cancelError"   class="alert alert-error">{{ cancelError }}</div>

          <!-- empty -->
          <div v-if="!subLoading && !subError && subscriptions.length === 0" class="sub-empty">
            <svg width="24" height="24" viewBox="0 0 24 24" fill="none"
                 stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
              <path d="M21 16V8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16z"/>
            </svg>
            <div>
              <p class="sub-empty-title">No subscriptions</p>
              <p class="sub-empty-sub">This workspace has no subscriptions yet.</p>
            </div>
          </div>

          <!-- list -->
          <div v-else-if="!subLoading" class="sub-list">
            <div v-for="sub in subscriptions" :key="sub.id" class="sub-row">
              <div class="sub-icon">{{ (sub.product?.name ?? 'P')[0].toUpperCase() }}</div>
              <div class="sub-info">
                <div class="sub-product">{{ formatProductName(sub.product?.name) }}</div>
                <div class="sub-plan">{{ sub.plan_name }}</div>
              </div>
              <div class="sub-meta">
                <span class="sub-badge" :class="sub.status">{{ subStatusLabel(sub.status) }}</span>
                <div v-if="sub.status === 'active'" class="sub-expiry">
                  <svg width="11" height="11" viewBox="0 0 24 24" fill="none"
                       stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                    <circle cx="12" cy="12" r="10"/><polyline points="12 6 12 12 16 14"/>
                  </svg>
                  Renews {{ formatDate(sub.current_period_end) }}
                </div>
              </div>
              <button
                v-if="sub.status === 'active' && ws.my_role === 'owner'"
                class="cancel-btn"
                :disabled="cancelLoading === sub.id"
                @click="doCancel(sub)"
              >
                {{ cancelLoading === sub.id ? 'Canceling…' : 'Cancel' }}
              </button>
            </div>
          </div>

        </div>
      </section>

    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { authAPI } from '../services/api'

const route = useRoute()

// ── Workspace list ─────────────────────────────────────────────────────────────
const wsList      = ref([])
const activeWsId  = ref(null)
const ws          = ref(null)
const wsLoading   = ref(true)
const wsListError = ref('')

onMounted(async () => {
  wsLoading.value   = true
  wsListError.value = ''
  try {
    const { data } = await authAPI.listWorkspaces()
    wsList.value = data.workspaces ?? []
    if (wsList.value.length > 0) {
      await switchWorkspace(wsList.value[0].id)
    }
  } catch {
    wsListError.value = 'Failed to load workspaces.'
  } finally {
    wsLoading.value = false
  }
})

async function switchWorkspace(id) {
  activeWsId.value = id
  billing.value    = null
  billingError.value = ''
  subscriptions.value = []
  cancelError.value   = ''
  cancelSuccess.value = ''
  checkoutError.value = ''
  try {
    const { data } = await authAPI.getWorkspace(id)
    ws.value = data.workspace
    // Auto-sync when returning from Stripe Checkout success
    if (route.query.billing === 'success' && ws.value?.my_role === 'owner') {
      await doSyncBilling(id)
    }
    loadBilling(id)
    loadSubscriptions(id)
  } catch {
    wsListError.value = 'Failed to load workspace details.'
  }
}

// ── Products ───────────────────────────────────────────────────────────────────
const products       = ref([])
const prodLoading    = ref(false)
const prodError      = ref('')
const checkoutLoading = ref('')
const checkoutError  = ref('')

onMounted(async () => {
  // Load products list once (same for all workspaces)
  prodLoading.value = true
  try {
    const { data } = await authAPI.listProducts()
    products.value = data.products ?? []
  } catch {
    prodError.value = 'Failed to load products.'
  } finally {
    prodLoading.value = false
  }
})

// Returns the active Subscription for a given product ID, or null
function activeSubFor(productId) {
  return subscriptions.value.find(
    s => s.product_id === productId && s.status === 'active'
  ) ?? null
}

async function doCheckout(prod) {
  checkoutLoading.value = prod.id
  checkoutError.value   = ''
  try {
    const { data } = await authAPI.createCheckoutSession(ws.value.id, {
      price_id:    prod.stripe_price_id,
      product_id:  prod.id,
      plan_name:   'standard',
      success_url: window.location.origin + '/billing?billing=success',
      cancel_url:  window.location.origin + '/billing?billing=canceled',
    })
    // Redirect to Stripe Checkout
    window.location.href = data.url
  } catch (err) {
    checkoutError.value   = err.response?.data?.error ?? 'Failed to start checkout. Try again.'
    checkoutLoading.value = ''
  }
}

// ── Billing ────────────────────────────────────────────────────────────────────
const billing        = ref(null)
const billingLoading = ref(true)
const billingError   = ref('')
const portalLoading  = ref(false)
const portalError    = ref('')
const syncLoading    = ref(false)
const syncMsg        = ref('')

async function loadBilling(wsId) {
  billingLoading.value = true
  billingError.value   = ''
  try {
    const { data } = await authAPI.getBillingStatus(wsId)
    billing.value = data
  } catch {
    billingError.value = 'Failed to load billing info.'
  } finally {
    billingLoading.value = false
  }
}

async function openPortal() {
  portalLoading.value = true
  portalError.value   = ''
  try {
    const { data } = await authAPI.createPortalSession(ws.value.id, {
      return_url: window.location.href,
    })
    window.location.href = data.url
  } catch (err) {
    portalError.value   = err.response?.data?.error ?? 'Failed to open billing portal.'
    portalLoading.value = false
  }
}

async function doSyncBilling(wsId) {
  syncLoading.value = true
  syncMsg.value     = ''
  try {
    const { data } = await authAPI.syncBilling(wsId ?? ws.value.id)
    billing.value = data
    syncMsg.value = data.synced > 0
      ? `Synced ${data.synced} subscription(s).`
      : 'No active subscriptions found on Stripe.'
    setTimeout(() => { syncMsg.value = '' }, 5000)
  } catch (err) {
    syncMsg.value = err.response?.data?.error ?? 'Sync failed.'
  } finally {
    syncLoading.value = false
  }
}

// ── Subscriptions ──────────────────────────────────────────────────────────────
const subscriptions = ref([])
const subLoading    = ref(false)
const subError      = ref('')
const cancelLoading = ref('')
const cancelSuccess = ref('')
const cancelError   = ref('')

async function loadSubscriptions(wsId) {
  subLoading.value = true
  subError.value   = ''
  try {
    const { data } = await authAPI.listSubscriptions(wsId)
    subscriptions.value = data.subscriptions ?? []
  } catch {
    subError.value = 'Failed to load subscriptions.'
  } finally {
    subLoading.value = false
  }
}

async function doCancel(sub) {
  cancelLoading.value = sub.id
  cancelError.value   = ''
  cancelSuccess.value = ''
  try {
    await authAPI.cancelSubscription(ws.value.id, sub.id)
    sub.status          = 'canceled'
    cancelSuccess.value = `${formatProductName(sub.product?.name)} subscription canceled.`
    setTimeout(() => { cancelSuccess.value = '' }, 5000)
  } catch (err) {
    cancelError.value = err.response?.data?.error ?? 'Failed to cancel subscription.'
  } finally {
    cancelLoading.value = ''
  }
}

// ── Helpers ────────────────────────────────────────────────────────────────────
function formatDate(iso) {
  if (!iso) return '—'
  return new Date(iso).toLocaleDateString('en-US', { year: 'numeric', month: 'long', day: 'numeric' })
}

function formatProductName(name) {
  if (!name) return 'Unknown product'
  return name.replace(/_/g, ' ').replace(/\b\w/g, c => c.toUpperCase())
}

function prodInitial(name) {
  if (!name) return '?'
  return name[0].toUpperCase()
}

function subStatusLabel(status) {
  return { active: 'Active', canceled: 'Canceled', expired: 'Expired' }[status] ?? status
}
</script>

<style scoped>
.settings-page { min-height: 100vh; background: var(--bg, #f1f3f4); }

.page-header {
  background: #fff;
  border-bottom: 1px solid var(--border, #dadce0);
  padding: 2rem 2rem 1.5rem;
}
.page-header-inner { max-width: 860px; margin: 0 auto; }
.page-header h1 { font-size: 1.4rem; font-weight: 700; color: var(--text, #202124); margin-bottom: 4px; }
.page-header p  { font-size: 0.875rem; color: var(--text-muted, #5f6368); }

.page-body { max-width: 860px; margin: 0 auto; padding: 2rem; }

.section { margin-bottom: 2.5rem; }
.section-head {
  display: flex; align-items: flex-start; gap: 12px; margin-bottom: 1.25rem;
}
.section-icon {
  width: 36px; height: 36px; border-radius: 9px; flex-shrink: 0;
  background: var(--blue-light, #e8f0fe); color: var(--blue, #1a73e8);
  display: flex; align-items: center; justify-content: center; margin-top: 2px;
}
.section-head h2   { font-size: 1rem; font-weight: 700; color: var(--text, #202124); }
.section-sub { font-size: 0.8rem; color: var(--text-muted, #5f6368); margin-top: 2px; }

.card {
  background: #fff;
  border: 1px solid var(--border, #dadce0);
  border-radius: 12px;
  padding: 1.75rem;
}

/* Skeleton */
.profile-skeleton { display: flex; flex-direction: column; gap: 14px; }
.skeleton-row {
  height: 44px; border-radius: 8px;
  background: linear-gradient(90deg, #f0f0f0 25%, #e4e4e4 50%, #f0f0f0 75%);
  background-size: 200% 100%;
  animation: shimmer 1.4s infinite;
}
.skel-line {
  height: 14px; border-radius: 6px;
  background: linear-gradient(90deg, #f0f0f0 25%, #e4e4e4 50%, #f0f0f0 75%);
  background-size: 200% 100%;
  animation: shimmer 1.4s infinite;
}
@keyframes shimmer { 0% { background-position: 200% 0 } 100% { background-position: -200% 0 } }

/* Alerts */
.alert {
  display: flex; align-items: center; gap: 8px;
  padding: 10px 14px; border-radius: 8px;
  font-size: 0.85rem; font-weight: 500; margin-bottom: 1.25rem;
}
.alert-success { background: #e6f4ea; color: #1e8e3e; }
.alert-error   { background: #fce8e6; color: #c5221f; }

/* Workspace selector */
.ws-selector-label {
  font-size: 0.78rem; font-weight: 600; color: var(--text-muted, #5f6368);
  margin-bottom: 0.75rem;
}
.ws-tabs {
  display: flex; flex-wrap: wrap; gap: 6px;
}
.ws-tab {
  display: inline-flex; align-items: center; gap: 7px;
  height: 34px; padding: 0 14px; border-radius: 8px;
  background: var(--bg, #f1f3f4); border: 1.5px solid transparent;
  font-size: 0.82rem; font-weight: 600; color: var(--text-muted, #5f6368);
  cursor: pointer; transition: all .15s;
  font-family: inherit;
}
.ws-tab:hover { border-color: var(--border, #dadce0); color: var(--text, #202124); }
.ws-tab.active {
  background: var(--blue-light, #e8f0fe); border-color: var(--blue, #1a73e8);
  color: var(--blue, #1a73e8);
}
.ws-tab-dot {
  width: 7px; height: 7px; border-radius: 50%; flex-shrink: 0;
}
.ws-tab-dot.owner  { background: var(--blue, #1a73e8); }
.ws-tab-dot.member { background: var(--text-muted, #5f6368); }

/* Billing */
.billing-empty {
  display: flex; align-items: flex-start; gap: 14px;
  padding: 1.25rem; border-radius: 10px;
  background: var(--bg, #f1f3f4); color: var(--text-muted, #5f6368);
}
.billing-empty svg          { flex-shrink: 0; opacity: .4; margin-top: 2px; }
.billing-empty-title        { font-size: 0.875rem; font-weight: 600; color: var(--text, #202124); margin-bottom: 3px; }
.billing-empty-sub          { font-size: 0.78rem; line-height: 1.55; }

.billing-status-row         { display: flex; align-items: center; gap: 8px; margin-bottom: .75rem; flex-wrap: wrap; }
.billing-cus-id             { font-size: 0.72rem; color: var(--text-muted, #5f6368); background: var(--bg, #f1f3f4); padding: 2px 8px; border-radius: 6px; }

.billing-actions            { display: flex; gap: .5rem; margin-bottom: .75rem; flex-wrap: wrap; }

.billing-btn {
  display: inline-flex; align-items: center; gap: 6px;
  padding: 7px 14px; border-radius: 8px; font-size: 0.82rem; font-weight: 600;
  cursor: pointer; border: 1.5px solid var(--blue, #1a73e8);
  background: var(--blue, #1a73e8); color: #fff; transition: background .15s;
  font-family: inherit;
}
.billing-btn:hover:not(:disabled) { background: #1557b0; border-color: #1557b0; }
.billing-btn:disabled             { opacity: 0.5; cursor: not-allowed; }
.billing-btn-ghost {
  background: transparent; color: var(--text, #202124);
  border-color: var(--border, #dadce0);
}
.billing-btn-ghost:hover:not(:disabled) { background: var(--bg, #f1f3f4); border-color: #aaa; }

.billing-sync-msg {
  font-size: 0.8rem; color: var(--text-muted, #5f6368);
  background: var(--bg, #f1f3f4); border-radius: 6px;
  padding: 6px 12px; margin-top: .75rem;
}

.billing-subs-label {
  font-size: 0.72rem; font-weight: 700; text-transform: uppercase;
  letter-spacing: .06em; color: var(--text-muted, #5f6368);
  margin-top: .75rem; margin-bottom: .5rem;
}

.billing-sub-list           { display: flex; flex-direction: column; gap: 0; }
.billing-sub-row {
  display: flex; align-items: center; gap: 12px;
  padding: 10px 0; border-bottom: 1px solid var(--border, #dadce0);
}
.billing-sub-row:last-child { border-bottom: none; }
.billing-sub-icon {
  width: 34px; height: 34px; border-radius: 8px; flex-shrink: 0;
  background: var(--blue-light, #e8f0fe); color: var(--blue, #1a73e8);
  font-size: 14px; font-weight: 700;
  display: flex; align-items: center; justify-content: center;
}
.billing-sub-info           { flex: 1; min-width: 0; }
.billing-sub-product        { font-size: 0.875rem; font-weight: 600; color: var(--text, #202124); }
.billing-sub-plan           { font-size: 0.75rem; color: var(--text-muted, #5f6368); margin-top: 1px; text-transform: capitalize; }
.billing-sub-meta           { flex-shrink: 0; text-align: right; }
.billing-sub-renew          { font-size: 0.72rem; color: var(--text-muted, #5f6368); margin-top: 3px; }
.billing-no-subs            { font-size: 0.82rem; color: var(--text-muted, #5f6368); margin-top: .5rem; }

/* Subscriptions */
.sub-empty {
  display: flex; align-items: flex-start; gap: 14px;
  padding: 1.25rem; border-radius: 10px;
  background: var(--bg, #f1f3f4); color: var(--text-muted, #5f6368);
}
.sub-empty svg { flex-shrink: 0; opacity: .45; margin-top: 2px; }
.sub-empty-title { font-size: 0.875rem; font-weight: 600; color: var(--text, #202124); margin-bottom: 3px; }
.sub-empty-sub   { font-size: 0.78rem; line-height: 1.55; }

.sub-list { display: flex; flex-direction: column; gap: 0; }
.sub-row {
  display: flex; align-items: center; gap: 12px;
  padding: 12px 0; border-bottom: 1px solid var(--border, #dadce0);
}
.sub-row:last-child { border-bottom: none; }
.sub-icon {
  width: 38px; height: 38px; border-radius: 10px; flex-shrink: 0;
  background: var(--blue-light, #e8f0fe); color: var(--blue, #1a73e8);
  font-size: 15px; font-weight: 700;
  display: flex; align-items: center; justify-content: center;
}
.sub-info { flex: 1; min-width: 0; }
.sub-product { font-size: 0.875rem; font-weight: 600; color: var(--text, #202124); }
.sub-plan    { font-size: 0.75rem; color: var(--text-muted, #5f6368); margin-top: 1px; text-transform: capitalize; }
.sub-meta    { flex-shrink: 0; text-align: right; }
.sub-expiry  {
  display: flex; align-items: center; gap: 4px;
  font-size: 0.75rem; color: var(--text-muted, #5f6368);
}

.sub-badge {
  font-size: 0.7rem; font-weight: 700; padding: 3px 10px;
  border-radius: 12px; flex-shrink: 0;
}
.sub-badge.active   { background: var(--success-bg, #e6f4ea); color: var(--success, #1e8e3e); }
.sub-badge.canceled { background: var(--bg, #f1f3f4);         color: var(--text-muted, #5f6368); }
.sub-badge.expired  { background: var(--error-bg, #fce8e6);   color: var(--error, #d93025); }

.cancel-btn {
  height: 30px; padding: 0 14px; border-radius: 8px;
  background: none; color: #c5221f; font-weight: 600;
  font-size: 0.78rem; border: 1.5px solid #f5c6c4; cursor: pointer;
  transition: background .15s;
  font-family: inherit;
}
.cancel-btn:hover:not(:disabled) { background: #fce8e6; }
.cancel-btn:disabled { opacity: 0.5; cursor: not-allowed; }

/* ── Products & Plans ───────────────────────────────────────────────────── */
.prod-list { display: flex; flex-direction: column; gap: 0; }
.prod-row {
  display: flex; align-items: center; gap: 14px;
  padding: 14px 0; border-bottom: 1px solid var(--border, #dadce0);
}
.prod-row:last-child { border-bottom: none; }

.prod-icon {
  width: 42px; height: 42px; border-radius: 10px; flex-shrink: 0;
  font-size: 16px; font-weight: 800;
  display: flex; align-items: center; justify-content: center;
  background: var(--blue-light, #e8f0fe); color: var(--blue, #1a73e8);
}
.prod-icon[data-name="cold_email"] { background: #e8f0fe; color: #1a73e8; }
.prod-icon[data-name="linkedin"]   { background: #e8f5e9; color: #1b7e34; }
.prod-icon[data-name="warmup"]     { background: #fff3e0; color: #e65100; }

.prod-info  { flex: 1; min-width: 0; }
.prod-name  { font-size: 0.9rem; font-weight: 700; color: var(--text, #202124); }
.prod-desc  { font-size: 0.78rem; color: var(--text-muted, #5f6368); margin-top: 2px; line-height: 1.45; }

.prod-action { flex-shrink: 0; text-align: right; min-width: 110px; }
.prod-renew  { font-size: 0.72rem; color: var(--text-muted, #5f6368); margin-top: 3px; }
.prod-no-sub   { font-size: 0.78rem; color: var(--text-muted, #5f6368); }
.prod-no-price { font-size: 0.72rem; color: #e65100; background: #fff3e0; padding: 3px 10px; border-radius: 8px; }

.btn-subscribe {
  display: inline-flex; align-items: center; gap: 5px;
  height: 32px; padding: 0 14px; border-radius: 8px;
  background: var(--blue, #1a73e8); color: #fff;
  border: none; font-size: 0.8rem; font-weight: 700;
  cursor: pointer; transition: background .15s;
  font-family: inherit;
}
.btn-subscribe:hover:not(:disabled) { background: #1557b0; }
.btn-subscribe:disabled { opacity: 0.55; cursor: not-allowed; }
</style>
