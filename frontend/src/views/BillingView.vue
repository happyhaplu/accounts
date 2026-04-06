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

      <!-- \u2500\u2500 Your Subscriptions \u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500 -->
      <section v-if="ws" class="section">
        <div class="section-head">
          <div class="section-icon">
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none"
                 stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <path d="M21 16V8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16z"/>
            </svg>
          </div>
          <div>
            <h2>Your Subscriptions</h2>
            <p class="section-sub">Active plans for <strong>{{ ws.name }}</strong>.</p>
          </div>
        </div>

        <div class="card">
          <!-- loading -->
          <div v-if="prodLoading || subLoading" class="profile-skeleton">
            <div class="skeleton-row" style="height:56px"></div>
            <div class="skeleton-row" style="height:56px"></div>
            <div class="skeleton-row" style="height:56px"></div>
          </div>

          <!-- error -->
          <div v-else-if="prodError || subError" class="alert alert-error">{{ prodError || subError }}</div>

          <!-- product rows -->
          <template v-else>
            <!-- cancel feedback -->
            <div v-if="cancelSuccess" class="alert alert-success" style="margin-bottom:1rem">{{ cancelSuccess }}</div>
            <div v-if="cancelError"   class="alert alert-error"   style="margin-bottom:1rem">{{ cancelError }}</div>

            <div class="subs-grid">
              <div v-for="prod in products" :key="prod.id" class="subs-row">

                <!-- product identity -->
                <div class="subs-prod">
                  <div class="subs-icon" :data-name="prod.name">{{ prodInitial(prod.name) }}</div>
                  <div class="subs-prod-info">
                    <div class="subs-prod-name">{{ formatProductName(prod.name) }}</div>
                    <div class="subs-prod-desc">{{ prod.description || '' }}</div>
                  </div>
                </div>

                <!-- plan + status -->
                <template v-if="activeSubFor(prod.id)">
                  <div class="subs-plan">{{ activeSubFor(prod.id).plan_name }}</div>
                  <div class="subs-status">
                    <span class="sub-badge active">\u2713 Active</span>
                    <span class="subs-until">until {{ formatDate(activeSubFor(prod.id).current_period_end) }}</span>
                  </div>
                  <div class="subs-actions">
                    <button class="subs-btn-manage" @click="openPortal" :disabled="portalLoading">
                      {{ portalLoading ? 'Opening\u2026' : 'Manage billing' }}
                    </button>
                    <button
                      v-if="ws.my_role === 'owner'"
                      class="subs-btn-cancel"
                      :disabled="cancelLoading === activeSubFor(prod.id).id"
                      @click="doCancel(activeSubFor(prod.id))"
                    >
                      {{ cancelLoading === activeSubFor(prod.id).id ? 'Canceling\u2026' : 'Cancel' }}
                    </button>
                  </div>
                </template>

                <template v-else>
                  <div class="subs-plan subs-plan-none">\u2014</div>
                  <div class="subs-status">
                    <span class="sub-badge">Not subscribed</span>
                  </div>
                  <div class="subs-actions">
                    <a
                      v-if="productHomeURL(prod)"
                      :href="productHomeURL(prod)"
                      target="_blank"
                      rel="noopener"
                      class="subs-btn-goto"
                    >
                      Go to {{ formatProductName(prod.name) }}
                      <svg width="11" height="11" viewBox="0 0 24 24" fill="none" stroke="currentColor"
                           stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round" style="margin-left:4px">
                        <path d="M18 13v6a2 2 0 01-2 2H5a2 2 0 01-2-2V8a2 2 0 012-2h6"/>
                        <polyline points="15 3 21 3 21 9"/><line x1="10" y1="14" x2="21" y2="3"/>
                      </svg>
                    </a>
                  </div>
                </template>

              </div>
            </div>

            <div v-if="products.length === 0" class="sub-empty">
              <p class="sub-empty-title">No products configured.</p>
            </div>
          </template>
        </div>
      </section>

      <!-- \u2500\u2500 Billing (portal + sync) \u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500 -->
      <section v-if="ws && billing?.has_stripe_customer" class="section">
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
            <p class="section-sub">Stripe account for <strong>{{ ws.name }}</strong>.</p>
          </div>
        </div>

        <div class="card">
          <div class="billing-status-row">
            <span class="sub-badge active">\u2713 Billing active</span>
            <code class="billing-cus-id">{{ billing?.stripe_customer_id }}</code>
          </div>
          <div class="billing-actions">
            <button class="billing-btn" @click="openPortal" :disabled="portalLoading">
              <svg width="13" height="13" viewBox="0 0 24 24" fill="none"
                   stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round">
                <path d="M18 13v6a2 2 0 01-2 2H5a2 2 0 01-2-2V8a2 2 0 012-2h6"/>
                <polyline points="15 3 21 3 21 9"/><line x1="10" y1="14" x2="21" y2="3"/>
              </svg>
              {{ portalLoading ? 'Opening\u2026' : 'Manage billing' }}
            </button>
            <button class="billing-btn billing-btn-ghost" @click="doSyncBilling()" :disabled="syncLoading">
              <svg width="13" height="13" viewBox="0 0 24 24" fill="none"
                   stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round">
                <polyline points="23 4 23 10 17 10"/>
                <path d="M20.49 15a9 9 0 1 1-2.12-9.36L23 10"/>
              </svg>
              {{ syncLoading ? 'Syncing\u2026' : 'Refresh from Stripe' }}
            </button>
          </div>
          <div v-if="portalError" class="alert alert-error" style="margin-top:.75rem;">{{ portalError }}</div>
          <div v-if="syncMsg" class="billing-sync-msg">{{ syncMsg }}</div>
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
  activeWsId.value    = id
  billing.value       = null
  billingError.value  = ''
  subscriptions.value = []
  cancelError.value   = ''
  cancelSuccess.value = ''
  try {
    const { data } = await authAPI.getWorkspace(id)
    ws.value = data.workspace
    // Auto-sync when returning from a product's Stripe checkout success
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
const products    = ref([])
const prodLoading = ref(false)
const prodError   = ref('')

onMounted(async () => {
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

// Returns the product's home URL (first redirect_url) for the "Go to Product" link
function productHomeURL(prod) {
  return prod.redirect_urls?.[0] ?? null
}

// ── Billing ────────────────────────────────────────────────────────────────────
const billing        = ref(null)
const billingLoading = ref(false)
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
.prod-list  { display: flex; flex-direction: column; gap: 0; }
.prod-entry { border-bottom: 1px solid var(--border, #dadce0); }
.prod-entry:last-child { border-bottom: none; }

.prod-pricing-table-wrap {
  padding: 1rem 0 1.5rem;
  /* Reset Stripe's internal overflow so it expands naturally */
  overflow: visible;
}
stripe-pricing-table { display: block; width: 100%; }
.prod-row {
  display: flex; align-items: center; gap: 14px;
  padding: 14px 0;
}
.prod-row:last-child { border-bottom: none; }

.prod-icon {
  width: 42px; height: 42px; border-radius: 10px; flex-shrink: 0;
  font-size: 16px; font-weight: 800;
  display: flex; align-items: center; justify-content: center;
  background: var(--blue-light, #e8f0fe); color: var(--blue, #1a73e8);
}

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

/* ── Pricing Cards ───────────────────────────────────────────────────────── */
.prod-entry-plans { border-bottom: none; padding-bottom: 1.5rem; }

.pc-head {
  display: flex; flex-direction: column; align-items: center;
  gap: 1rem; padding: 1.75rem 0 1rem; text-align: center;
}
.pc-title { font-size: 1.1rem; font-weight: 700; color: var(--text, #202124); }

.pc-toggle {
  display: inline-flex; align-items: center;
  background: var(--bg, #f1f3f4); border-radius: 100px;
  padding: 3px; gap: 2px;
}
.pc-tgl {
  height: 30px; padding: 0 14px; border-radius: 100px;
  border: none; background: transparent;
  font-size: 0.8rem; font-weight: 600; color: var(--text-muted, #5f6368);
  cursor: pointer; transition: all .15s; font-family: inherit;
  display: inline-flex; align-items: center; gap: 5px;
}
.pc-tgl.active { background: #fff; color: var(--text, #202124); box-shadow: 0 1px 3px rgba(0,0,0,.12); }
.pc-save { font-size: 0.72rem; color: var(--blue, #1a73e8); font-weight: 700; }

.pc-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 16px;
  padding: 0.25rem 0 1.25rem;
}
@media (max-width: 720px) { .pc-grid { grid-template-columns: 1fr; } }
@media (max-width: 960px) and (min-width: 721px) { .pc-grid { grid-template-columns: 1fr 1fr; } }

.pc-card {
  position: relative;
  border: 1.5px solid var(--border, #dadce0);
  border-radius: 14px;
  padding: 1.5rem 1.25rem;
  background: #fff;
  display: flex; flex-direction: column; gap: 0;
  transition: box-shadow .18s;
}
.pc-card:hover { box-shadow: 0 4px 18px rgba(0,0,0,.08); }
.pc-card.popular {
  border-color: var(--text, #202124);
  box-shadow: 0 2px 12px rgba(0,0,0,.1);
}

.pc-badge {
  position: absolute; top: -11px; left: 50%; transform: translateX(-50%);
  background: var(--text, #202124); color: #fff;
  font-size: 0.68rem; font-weight: 700; letter-spacing: .04em;
  padding: 3px 12px; border-radius: 100px; white-space: nowrap;
}

.pc-icon {
  width: 44px; height: 44px; border-radius: 12px;
  background: var(--bg, #f1f3f4);
  display: flex; align-items: center; justify-content: center;
  margin-bottom: 12px; color: var(--text, #202124);
}
.pc-card.popular .pc-icon { background: #f0f4ff; color: var(--blue, #1a73e8); }

.pc-name    { font-size: 1rem; font-weight: 700; color: var(--text, #202124); margin-bottom: 3px; }
.pc-tagline { font-size: 0.78rem; color: var(--text-muted, #5f6368); margin-bottom: 1.1rem; line-height: 1.45; }

.pc-price {
  margin-bottom: 1.1rem;
}
.pc-amount {
  font-size: 2rem; font-weight: 800; color: var(--text, #202124);
  letter-spacing: -.02em;
}
.pc-unit-wrap { display: flex; flex-direction: column; margin-top: 1px; }
.pc-unit      { font-size: 0.72rem; color: var(--text-muted, #5f6368); font-weight: 600; }
.pc-note      { font-size: 0.7rem; color: var(--text-muted, #5f6368); margin-top: 1px; }
.pc-contact-price { font-size: 1.6rem; font-weight: 800; color: var(--text, #202124); }
.pc-contact-note  { font-size: 0.72rem; color: var(--text-muted, #5f6368); margin-top: 2px; }

.pc-cta {
  width: 100%; height: 40px; border-radius: 8px;
  border: 1.5px solid var(--border, #dadce0);
  background: #fff; color: var(--text, #202124);
  font-size: 0.85rem; font-weight: 700; cursor: pointer;
  transition: all .15s; font-family: inherit;
  margin-bottom: 1.25rem; display: flex; align-items: center; justify-content: center;
  text-decoration: none;
}
.pc-cta:hover:not(:disabled) { background: var(--bg, #f1f3f4); }
.pc-cta:disabled { opacity: 0.5; cursor: not-allowed; }
.pc-cta.primary {
  background: var(--text, #202124); color: #fff; border-color: var(--text, #202124);
}
.pc-cta.primary:hover:not(:disabled) { background: #333; border-color: #333; }
.pc-cta-outline { border-color: var(--border, #dadce0); color: var(--text, #202124); }

.pc-features {
  list-style: none; padding: 0; margin: 0;
  display: flex; flex-direction: column; gap: 8px;
}
.pc-features li {
  display: flex; align-items: flex-start; gap: 7px;
  font-size: 0.8rem; color: var(--text-muted, #5f6368); line-height: 1.45;
}
.pc-features li.pc-feat-heading {
  font-size: 0.8rem; font-weight: 700; color: var(--text, #202124);
  margin-top: 2px; display: block;
}
.pc-check { flex-shrink: 0; margin-top: 1px; color: var(--text-muted, #5f6368); }
.pc-card.popular .pc-check { color: var(--blue, #1a73e8); }

/* ── Your Subscriptions grid ──────────────────────────────────────────────── */
.subs-grid { display: flex; flex-direction: column; gap: 0; }
.subs-row {
  display: grid;
  grid-template-columns: 1fr 120px 160px auto;
  align-items: center;
  gap: 12px;
  padding: 14px 0;
  border-bottom: 1px solid var(--border, #dadce0);
}
.subs-row:last-child { border-bottom: none; }
@media (max-width: 640px) {
  .subs-row { grid-template-columns: 1fr; gap: 8px; }
}

.subs-prod { display: flex; align-items: center; gap: 12px; min-width: 0; }
.subs-icon {
  width: 40px; height: 40px; border-radius: 10px; flex-shrink: 0;
  background: var(--blue-light, #e8f0fe); color: var(--blue, #1a73e8);
  font-size: 15px; font-weight: 800;
  display: flex; align-items: center; justify-content: center;
}
.subs-prod-info { min-width: 0; }
.subs-prod-name {
  font-size: 0.875rem; font-weight: 700; color: var(--text, #202124);
  white-space: nowrap; overflow: hidden; text-overflow: ellipsis;
}
.subs-prod-desc {
  font-size: 0.75rem; color: var(--text-muted, #5f6368); margin-top: 1px;
  white-space: nowrap; overflow: hidden; text-overflow: ellipsis;
}

.subs-plan      { font-size: 0.78rem; font-weight: 600; color: var(--text, #202124); text-transform: capitalize; }
.subs-plan-none { color: var(--text-muted, #5f6368); font-weight: 400; }

.subs-status { display: flex; flex-direction: column; align-items: flex-start; gap: 3px; }
.subs-until  { font-size: 0.72rem; color: var(--text-muted, #5f6368); }

.subs-actions { display: flex; align-items: center; gap: 6px; flex-shrink: 0; }

.subs-btn-manage {
  display: inline-flex; align-items: center; gap: 5px;
  height: 30px; padding: 0 12px; border-radius: 7px;
  background: var(--blue, #1a73e8); color: #fff; border: none;
  font-size: 0.78rem; font-weight: 600; cursor: pointer;
  transition: background .15s; font-family: inherit;
}
.subs-btn-manage:hover:not(:disabled) { background: #1557b0; }
.subs-btn-manage:disabled { opacity: 0.5; cursor: not-allowed; }

.subs-btn-cancel {
  height: 30px; padding: 0 12px; border-radius: 7px;
  background: none; color: #c5221f; font-weight: 600;
  font-size: 0.78rem; border: 1.5px solid #f5c6c4; cursor: pointer;
  transition: background .15s; font-family: inherit;
}
.subs-btn-cancel:hover:not(:disabled) { background: #fce8e6; }
.subs-btn-cancel:disabled { opacity: 0.5; cursor: not-allowed; }

.subs-btn-goto {
  display: inline-flex; align-items: center;
  height: 30px; padding: 0 12px; border-radius: 7px;
  background: var(--bg, #f1f3f4); color: var(--text, #202124);
  border: 1.5px solid var(--border, #dadce0);
  font-size: 0.78rem; font-weight: 600; text-decoration: none;
  transition: all .15s; white-space: nowrap;
}
.subs-btn-goto:hover {
  background: #e8f0fe; border-color: var(--blue, #1a73e8); color: var(--blue, #1a73e8);
}
</style>
