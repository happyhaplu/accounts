<template>
  <div class="au-page">

    <!-- Header -->
    <div class="au-header">
      <div>
        <div class="au-title">Users</div>
        <div class="au-sub">All registered accounts and their subscription status</div>
      </div>
      <div class="au-header-actions">
        <button class="btn-purge" @click="purgeConfirm = true" :disabled="loading || purging"
                v-if="unverifiedCount > 0">
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor"
               stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <polyline points="3 6 5 6 21 6"/>
            <path d="M19 6l-1 14a2 2 0 0 1-2 2H8a2 2 0 0 1-2-2L5 6"/>
            <path d="M10 11v6M14 11v6"/>
            <path d="M9 6V4a1 1 0 0 1 1-1h4a1 1 0 0 1 1 1v2"/>
          </svg>
          Purge unverified ({{ unverifiedCount }})
        </button>
        <button class="btn-refresh" @click="fetchUsers" :disabled="loading">
          <svg :class="['refresh-icon', { spinning: loading }]" width="15" height="15" viewBox="0 0 24 24"
               fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round">
            <polyline points="23 4 23 10 17 10"/>
            <path d="M20.49 15a9 9 0 1 1-2.12-9.36L23 10"/>
          </svg>
          Refresh
        </button>
      </div>
    </div>

    <!-- Purge confirmation dialog -->
    <Teleport to="body">
      <transition name="modal-fade">
        <div v-if="purgeConfirm" class="modal-overlay" @click.self="purgeConfirm = false">
          <div class="modal-card" role="dialog" aria-modal="true">
            <div class="purge-modal-icon">
              <svg width="32" height="32" viewBox="0 0 24 24" fill="none" stroke="#c5221f"
                   stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round">
                <path d="M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z"/>
                <line x1="12" y1="9" x2="12" y2="13"/>
                <line x1="12" y1="17" x2="12.01" y2="17"/>
              </svg>
            </div>
            <h2 class="purge-modal-title">Purge {{ unverifiedCount }} unverified users?</h2>
            <p class="purge-modal-body">
              This will <strong>permanently delete</strong> all {{ unverifiedCount }} users who never verified their email address.
              Users with any subscription history are <strong>excluded</strong> and kept safe.
              <br/><br/>
              <strong>This cannot be undone.</strong>
            </p>
            <div v-if="purgeError" class="purge-error">{{ purgeError }}</div>
            <div class="purge-modal-footer">
              <button class="btn-ghost" @click="purgeConfirm = false" :disabled="purging">Cancel</button>
              <button class="btn-danger" @click="purgeUnverified" :disabled="purging">
                <svg v-if="purging" class="spin-icon" width="14" height="14" viewBox="0 0 24 24"
                     fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round">
                  <path d="M12 2v4M12 18v4M4.93 4.93l2.83 2.83M16.24 16.24l2.83 2.83M2 12h4M18 12h4M4.93 19.07l2.83-2.83M16.24 7.76l2.83-2.83"/>
                </svg>
                {{ purging ? 'Deleting…' : `Yes, delete ${unverifiedCount} users` }}
              </button>
            </div>
          </div>
        </div>
      </transition>
    </Teleport>

    <!-- Stats row -->
    <div class="au-stats" v-if="!loading && !loadError">
      <div class="stat-card">
        <div class="stat-value">{{ users.length }}</div>
        <div class="stat-label">Total Users</div>
      </div>
      <div class="stat-card">
        <div class="stat-value">{{ verifiedCount }}</div>
        <div class="stat-label">Verified</div>
      </div>
      <div class="stat-card">
        <div class="stat-value">{{ unverifiedCount }}</div>
        <div class="stat-label">Unverified</div>
      </div>
      <div class="stat-card">
        <div class="stat-value stat-green">{{ subscribedCount }}</div>
        <div class="stat-label">With Active Sub</div>
      </div>
    </div>

    <!-- Search -->
    <div class="au-toolbar">
      <div class="search-wrap">
        <svg class="search-icon" width="14" height="14" viewBox="0 0 24 24" fill="none"
             stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
          <circle cx="11" cy="11" r="8"/><line x1="21" y1="21" x2="16.65" y2="16.65"/>
        </svg>
        <input
          v-model="search"
          type="text"
          class="search-input"
          placeholder="Search by email or name…"
        />
        <button v-if="search" class="search-clear" @click="search = ''" title="Clear">
          <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor"
               stroke-width="2.5" stroke-linecap="round">
            <line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/>
          </svg>
        </button>
      </div>
      <div class="filter-wrap">
        <select v-model="filterStatus" class="filter-select">
          <option value="">All subscriptions</option>
          <option value="active">Has active</option>
          <option value="none">No subscriptions</option>
        </select>
        <select v-model="filterVerified" class="filter-select">
          <option value="">All users</option>
          <option value="verified">Verified</option>
          <option value="unverified">Unverified</option>
        </select>
      </div>
    </div>

    <!-- Error -->
    <div v-if="loadError" class="au-error">
      <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor"
           stroke-width="2" stroke-linecap="round" stroke-linejoin="round" style="flex-shrink:0">
        <circle cx="12" cy="12" r="10"/><line x1="12" y1="8" x2="12" y2="12"/>
        <line x1="12" y1="16" x2="12.01" y2="16"/>
      </svg>
      {{ loadError }}
    </div>

    <!-- Loading skeleton -->
    <div v-else-if="loading" class="au-table-wrap skeleton-wrap">
      <div v-for="i in 5" :key="i" class="skeleton-row"></div>
    </div>

    <!-- Table -->
    <div v-else class="au-table-wrap">
      <table class="au-table">
        <thead>
          <tr>
            <th>User</th>
            <th>Company</th>
            <th>Status</th>
            <th>Subscriptions</th>
            <th>Joined</th>
          </tr>
        </thead>
        <tbody>
          <tr v-if="filtered.length === 0">
            <td colspan="5" class="au-empty">
              {{ search || filterStatus || filterVerified ? 'No users match this filter.' : 'No users yet.' }}
            </td>
          </tr>
          <tr v-for="u in filtered" :key="u.id" :class="{ 'row-unverified': !u.email_verified }">

            <!-- User -->
            <td>
              <div class="user-email">{{ u.email }}</div>
              <div class="user-name" v-if="u.name">{{ u.name }}</div>
              <div class="user-tags">
                <span v-if="u.role === 'admin'" class="tag tag-admin">admin</span>
                <span v-if="!u.email_verified" class="tag tag-unverified">unverified</span>
                <span v-if="!u.profile_complete" class="tag tag-incomplete">profile incomplete</span>
              </div>
            </td>

            <!-- Company -->
            <td>
              <div v-if="u.company_name" class="company-name">{{ u.company_name }}</div>
              <div v-if="u.job_title" class="job-title">{{ u.job_title }}</div>
              <span v-if="!u.company_name && !u.job_title" class="text-muted">—</span>
            </td>

            <!-- Verified status -->
            <td>
              <span :class="['status-badge', u.email_verified ? 'badge-verified' : 'badge-pending']">
                <span class="badge-dot"></span>
                {{ u.email_verified ? 'Verified' : 'Pending' }}
              </span>
            </td>

            <!-- Subscriptions -->
            <td class="subs-cell">
              <div v-if="u.subscriptions.length" class="sub-pills">
                <span
                  v-for="s in u.subscriptions"
                  :key="s.product_name"
                  :class="['sub-pill', subPillClass(s)]"
                  :title="`${s.product_name} — expires ${fmtDate(s.current_period_end)}`"
                >
                  {{ s.product_name }}
                </span>
              </div>
              <span v-else class="text-muted">—</span>
            </td>

            <!-- Joined -->
            <td class="date-cell">
              <div class="date-main">{{ fmtDate(u.created_at) }}</div>
              <div v-if="u.last_login_at" class="date-sub">Last login {{ fmtRelative(u.last_login_at) }}</div>
            </td>

          </tr>
        </tbody>
      </table>
    </div>

    <!-- Footer count -->
    <div v-if="!loading && !loadError && filtered.length > 0" class="au-footer">
      Showing {{ filtered.length }} of {{ users.length }} users
    </div>

  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { adminAPI } from '../../services/api'

const users     = ref([])
const loading   = ref(false)
const loadError = ref('')
const search        = ref('')
const filterStatus  = ref('')
const filterVerified = ref('')
const purgeConfirm  = ref(false)
const purging       = ref(false)
const purgeError    = ref('')

// ── Computed stats ────────────────────────────────────────────────────────────
const verifiedCount   = computed(() => users.value.filter(u => u.email_verified).length)
const unverifiedCount = computed(() => users.value.filter(u => !u.email_verified).length)
const subscribedCount = computed(() =>
  users.value.filter(u => u.subscriptions.some(s => s.status === 'active')).length
)

// ── Filtered list ─────────────────────────────────────────────────────────────
const filtered = computed(() => {
  let list = users.value
  const q = search.value.toLowerCase().trim()
  if (q) {
    list = list.filter(u =>
      u.email.toLowerCase().includes(q) ||
      (u.name || '').toLowerCase().includes(q)
    )
  }
  if (filterStatus.value === 'active') {
    list = list.filter(u => u.subscriptions.some(s => s.status === 'active'))
  } else if (filterStatus.value === 'none') {
    list = list.filter(u => u.subscriptions.length === 0)
  }
  if (filterVerified.value === 'verified') {
    list = list.filter(u => u.email_verified)
  } else if (filterVerified.value === 'unverified') {
    list = list.filter(u => !u.email_verified)
  }
  return list
})

// ── Fetch ─────────────────────────────────────────────────────────────────────
async function fetchUsers() {
  loading.value   = true
  loadError.value = ''
  try {
    const { data } = await adminAPI.listUsers()
    users.value = data.users ?? []
  } catch (err) {
    loadError.value = err.response?.data?.error ?? 'Failed to load users.'
  } finally {
    loading.value = false
  }
}

onMounted(fetchUsers)

// ── Purge unverified ──────────────────────────────────────────────────────────
async function purgeUnverified() {
  purging.value   = true
  purgeError.value = ''
  try {
    const { data } = await adminAPI.purgeUnverifiedUsers()
    purgeConfirm.value = false
    await fetchUsers()
    // brief highlight via a toast-style message could be added here
    console.info(`[admin] purged ${data.deleted} unverified users`)
  } catch (err) {
    purgeError.value = err.response?.data?.error ?? 'Failed to purge users. Try again.'
  } finally {
    purging.value = false
  }
}

// ── Helpers ───────────────────────────────────────────────────────────────────
function subPillClass(s) {
  if (s.status === 'active') {
    const expires = new Date(s.current_period_end)
    return expires > new Date() ? 'pill-active' : 'pill-expired'
  }
  return s.status === 'canceled' ? 'pill-canceled' : 'pill-expired'
}

function fmtDate(iso) {
  if (!iso) return '—'
  return new Date(iso).toLocaleDateString('en-GB', { day: 'numeric', month: 'short', year: 'numeric' })
}

function fmtRelative(iso) {
  if (!iso) return ''
  const diff = Date.now() - new Date(iso).getTime()
  const mins  = Math.floor(diff / 60000)
  const hours = Math.floor(mins / 60)
  const days  = Math.floor(hours / 24)
  if (days > 0)  return `${days}d ago`
  if (hours > 0) return `${hours}h ago`
  if (mins > 0)  return `${mins}m ago`
  return 'just now'
}
</script>

<style scoped>
.au-page { padding: 0; }

/* Header */
.au-header {
  display: flex; align-items: flex-start; justify-content: space-between;
  gap: 16px; margin-bottom: 24px;
}
.au-title { font-size: 22px; font-weight: 600; color: #202124; letter-spacing: -0.3px; margin-bottom: 4px; }
.au-sub   { font-size: 13.5px; color: #5f6368; }

.au-header-actions { display: flex; align-items: center; gap: 8px; flex-shrink: 0; }

.btn-purge {
  display: inline-flex; align-items: center; gap: 7px;
  background: #fce8e6; border: 1px solid #f5c6c4; color: #c5221f;
  border-radius: 7px; padding: 8px 16px; font-size: 13.5px; font-weight: 500;
  cursor: pointer; font-family: inherit; transition: background 0.15s, border-color 0.15s;
  white-space: nowrap;
}
.btn-purge:hover:not(:disabled) { background: #f5c6c4; border-color: #c5221f; }
.btn-purge:disabled { opacity: 0.5; cursor: not-allowed; }

.btn-refresh {
  display: inline-flex; align-items: center; gap: 7px;
  background: #fff; border: 1px solid #dadce0; color: #3c4043;
  border-radius: 7px; padding: 8px 16px; font-size: 13.5px; font-weight: 500;
  cursor: pointer; font-family: inherit; transition: background 0.15s;
  white-space: nowrap; flex-shrink: 0;
}
.btn-refresh:hover:not(:disabled) { background: #f1f3f4; }
.btn-refresh:disabled { opacity: 0.5; cursor: not-allowed; }
.refresh-icon { transition: transform 0.3s; }
.refresh-icon.spinning { animation: spin 0.7s linear infinite; }
@keyframes spin { to { transform: rotate(360deg); } }

/* Stats */
.au-stats { display: flex; gap: 12px; margin-bottom: 24px; flex-wrap: wrap; }
.stat-card {
  background: #fff; border: 1px solid #dadce0; border-radius: 8px;
  padding: 14px 22px; min-width: 120px;
}
.stat-value { font-size: 26px; font-weight: 700; color: #202124; line-height: 1; margin-bottom: 5px; }
.stat-value.stat-green { color: #137333; }
.stat-label { font-size: 12px; color: #5f6368; font-weight: 500; }

/* Toolbar */
.au-toolbar {
  display: flex; align-items: center; gap: 12px; margin-bottom: 16px; flex-wrap: wrap;
}
.search-wrap {
  position: relative; flex: 1; min-width: 220px;
}
.search-icon { position: absolute; left: 11px; top: 50%; transform: translateY(-50%); color: #80868b; pointer-events: none; }
.search-input {
  width: 100%; box-sizing: border-box;
  padding: 8px 32px 8px 34px; border: 1px solid #dadce0; border-radius: 7px;
  font-size: 13.5px; font-family: inherit; color: #202124;
  background: #fff; outline: none; transition: border-color 0.15s, box-shadow 0.15s;
}
.search-input:focus { border-color: #1a73e8; box-shadow: 0 0 0 3px rgba(26,115,232,0.12); }
.search-clear {
  position: absolute; right: 8px; top: 50%; transform: translateY(-50%);
  background: none; border: none; cursor: pointer; color: #80868b; padding: 3px;
  border-radius: 4px; line-height: 0;
}
.search-clear:hover { color: #202124; background: #f1f3f4; }
.filter-wrap { display: flex; gap: 8px; }
.filter-select {
  border: 1px solid #dadce0; border-radius: 7px; background: #fff;
  font-family: inherit; font-size: 13px; color: #3c4043;
  padding: 7px 10px; cursor: pointer; outline: none;
}
.filter-select:focus { border-color: #1a73e8; }

/* Error */
.au-error {
  display: flex; align-items: center; gap: 10px;
  background: #fce8e6; border: 1px solid #f5c6c4; border-radius: 8px;
  color: #c5221f; padding: 14px 16px; font-size: 13.5px; margin-bottom: 16px;
}

/* Skeleton */
.skeleton-wrap { background: #fff; border: 1px solid #dadce0; border-radius: 10px; overflow: hidden; }
.skeleton-row { height: 52px; background: linear-gradient(90deg, #f1f3f4 25%, #e8eaed 50%, #f1f3f4 75%); background-size: 400% 100%; animation: shimmer 1.4s ease infinite; border-bottom: 1px solid #e8eaed; }
@keyframes shimmer { 0% { background-position: 100% 0 } 100% { background-position: -100% 0 } }

/* Table */
.au-table-wrap { background: #fff; border: 1px solid #dadce0; border-radius: 10px; overflow: hidden; }
.au-table { width: 100%; border-collapse: collapse; font-size: 13.5px; }
.au-table thead th {
  background: #f8f9fa; font-weight: 600; font-size: 11.5px;
  text-transform: uppercase; letter-spacing: 0.05em; color: #5f6368;
  padding: 11px 16px; text-align: left; border-bottom: 1px solid #e8eaed;
  white-space: nowrap;
}
.au-table tbody tr { border-bottom: 1px solid #f1f3f4; transition: background 0.1s; }
.au-table tbody tr:last-child { border-bottom: none; }
.au-table tbody tr:hover { background: #f8f9fa; }
.au-table tbody tr.row-unverified { opacity: 0.75; }
.au-table td { padding: 13px 16px; vertical-align: top; }
.au-empty { text-align: center; color: #9aa0a6; padding: 48px 16px !important; font-size: 14px; }

/* User cell */
.user-email { font-weight: 500; color: #202124; margin-bottom: 2px; }
.user-name  { font-size: 12.5px; color: #5f6368; margin-bottom: 4px; }
.user-tags  { display: flex; flex-wrap: wrap; gap: 4px; }
.tag { font-size: 10.5px; font-weight: 600; padding: 2px 7px; border-radius: 4px; }
.tag-admin      { background: #e8f0fe; color: #1a73e8; }
.tag-unverified { background: #fef7e0; color: #b06000; }
.tag-incomplete { background: #f1f3f4; color: #5f6368; }

/* Company */
.company-name { color: #202124; font-size: 13px; }
.job-title    { font-size: 12px; color: #5f6368; margin-top: 2px; }

/* Status badge */
.status-badge { display: inline-flex; align-items: center; gap: 5px; font-size: 12px; font-weight: 600; padding: 3px 9px; border-radius: 20px; white-space: nowrap; }
.badge-dot { width: 6px; height: 6px; border-radius: 50%; flex-shrink: 0; }
.badge-verified { background: #e6f4ea; color: #137333; }
.badge-verified .badge-dot { background: #34a853; }
.badge-pending  { background: #fef7e0; color: #b06000; }
.badge-pending .badge-dot  { background: #f9ab00; }

/* Subscription pills */
.subs-cell { min-width: 180px; }
.sub-pills { display: flex; flex-wrap: wrap; gap: 4px; }
.sub-pill {
  display: inline-block; font-size: 11px; font-weight: 600;
  padding: 3px 9px; border-radius: 20px; white-space: nowrap;
  cursor: default;
}
.pill-active   { background: #e6f4ea; color: #137333; }
.pill-canceled { background: #f1f3f4; color: #5f6368; }
.pill-expired  { background: #fce8e6; color: #c5221f; }

/* Dates */
.date-cell { white-space: nowrap; }
.date-main { font-size: 13px; color: #3c4043; }
.date-sub  { font-size: 11.5px; color: #9aa0a6; margin-top: 2px; }
.text-muted { color: #9aa0a6; }

/* Footer */
.au-footer { margin-top: 12px; font-size: 12.5px; color: #5f6368; text-align: right; }

/* Purge confirmation modal */
.modal-overlay {
  position: fixed; inset: 0; background: rgba(0,0,0,0.45);
  display: flex; align-items: center; justify-content: center;
  z-index: 1000; padding: 16px;
}
.modal-card {
  background: #fff; border-radius: 12px; padding: 32px 28px;
  width: 100%; max-width: 440px; box-shadow: 0 8px 40px rgba(0,0,0,0.22);
}
.purge-modal-icon { text-align: center; margin-bottom: 16px; }
.purge-modal-title {
  font-size: 18px; font-weight: 700; color: #202124;
  text-align: center; margin: 0 0 12px;
}
.purge-modal-body {
  font-size: 14px; color: #3c4043; line-height: 1.65;
  text-align: center; margin: 0 0 20px;
}
.purge-error {
  background: #fce8e6; border: 1px solid #f5c6c4; border-radius: 7px;
  color: #c5221f; padding: 10px 14px; font-size: 13px; margin-bottom: 16px;
}
.purge-modal-footer {
  display: flex; justify-content: flex-end; gap: 10px;
}
.btn-ghost {
  background: #fff; border: 1px solid #dadce0; color: #3c4043;
  border-radius: 7px; padding: 9px 18px; font-size: 13.5px; font-weight: 500;
  cursor: pointer; font-family: inherit; transition: background 0.15s;
}
.btn-ghost:hover:not(:disabled) { background: #f1f3f4; }
.btn-ghost:disabled { opacity: 0.5; cursor: not-allowed; }
.btn-danger {
  display: inline-flex; align-items: center; gap: 7px;
  background: #c5221f; border: 1px solid #c5221f; color: #fff;
  border-radius: 7px; padding: 9px 18px; font-size: 13.5px; font-weight: 600;
  cursor: pointer; font-family: inherit; transition: background 0.15s;
}
.btn-danger:hover:not(:disabled) { background: #a50e0e; }
.btn-danger:disabled { opacity: 0.5; cursor: not-allowed; }
.spin-icon { animation: spin 0.7s linear infinite; }

/* Modal transitions */
.modal-fade-enter-active, .modal-fade-leave-active { transition: opacity 0.18s ease; }
.modal-fade-enter-from, .modal-fade-leave-to { opacity: 0; }
</style>
