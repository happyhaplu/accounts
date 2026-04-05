<template>
  <div class="aw-page">

    <!-- Header -->
    <div class="aw-header">
      <div>
        <div class="aw-title">Workspaces</div>
        <div class="aw-sub">All workspaces, their owners, members, and product subscriptions</div>
      </div>
      <button class="btn-refresh" @click="fetchWorkspaces" :disabled="loading">
        <svg :class="['refresh-icon', { spinning: loading }]" width="15" height="15" viewBox="0 0 24 24"
             fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round">
          <polyline points="23 4 23 10 17 10"/>
          <path d="M20.49 15a9 9 0 1 1-2.12-9.36L23 10"/>
        </svg>
        Refresh
      </button>
    </div>

    <!-- Stats row -->
    <div class="aw-stats" v-if="!loading && !loadError">
      <div class="stat-card">
        <div class="stat-value">{{ workspaces.length }}</div>
        <div class="stat-label">Total Workspaces</div>
      </div>
      <div class="stat-card">
        <div class="stat-value stat-green">{{ activeSubsCount }}</div>
        <div class="stat-label">Active Subscriptions</div>
      </div>
      <div class="stat-card">
        <div class="stat-value">{{ withSubsCount }}</div>
        <div class="stat-label">Paying Workspaces</div>
      </div>
      <div class="stat-card">
        <div class="stat-value">{{ noSubsCount }}</div>
        <div class="stat-label">Free / No Sub</div>
      </div>
    </div>

    <!-- Search & filters -->
    <div class="aw-toolbar">
      <div class="search-wrap">
        <svg class="search-icon" width="14" height="14" viewBox="0 0 24 24" fill="none"
             stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
          <circle cx="11" cy="11" r="8"/><line x1="21" y1="21" x2="16.65" y2="16.65"/>
        </svg>
        <input
          v-model="search"
          type="text"
          class="search-input"
          placeholder="Search by workspace name or owner email…"
        />
        <button v-if="search" class="search-clear" @click="search = ''" title="Clear">
          <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor"
               stroke-width="2.5" stroke-linecap="round">
            <line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/>
          </svg>
        </button>
      </div>
      <div class="filter-wrap">
        <select v-model="filterSubs" class="filter-select">
          <option value="">All workspaces</option>
          <option value="active">Has active sub</option>
          <option value="none">No subscriptions</option>
        </select>
      </div>
    </div>

    <!-- Error -->
    <div v-if="loadError" class="aw-error">
      <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor"
           stroke-width="2" stroke-linecap="round" stroke-linejoin="round" style="flex-shrink:0">
        <circle cx="12" cy="12" r="10"/><line x1="12" y1="8" x2="12" y2="12"/>
        <line x1="12" y1="16" x2="12.01" y2="16"/>
      </svg>
      {{ loadError }}
    </div>

    <!-- Loading skeleton -->
    <div v-else-if="loading" class="aw-table-wrap skeleton-wrap">
      <div v-for="i in 5" :key="i" class="skeleton-row"></div>
    </div>

    <!-- Table -->
    <div v-else class="aw-table-wrap">
      <table class="aw-table">
        <thead>
          <tr>
            <th>Workspace</th>
            <th>Owner</th>
            <th class="center-th">Members</th>
            <th>Products / Subscriptions</th>
            <th>Created</th>
          </tr>
        </thead>
        <tbody>
          <tr v-if="filtered.length === 0">
            <td colspan="5" class="aw-empty">
              {{ search || filterSubs ? 'No workspaces match this filter.' : 'No workspaces yet.' }}
            </td>
          </tr>
          <tr v-for="ws in filtered" :key="ws.id" :class="{ 'row-no-subs': ws.subscriptions.length === 0 }">

            <!-- Workspace name -->
            <td>
              <div class="ws-name">{{ ws.name }}</div>
              <div class="ws-id">{{ ws.id.slice(0, 8) }}…</div>
            </td>

            <!-- Owner -->
            <td>
              <div class="owner-email">{{ ws.owner_email || '—' }}</div>
              <div class="owner-name" v-if="ws.owner_name">{{ ws.owner_name }}</div>
            </td>

            <!-- Members count -->
            <td class="center-td">
              <span class="member-badge">
                <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor"
                     stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                  <path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"/>
                  <circle cx="9" cy="7" r="4"/>
                  <path d="M23 21v-2a4 4 0 0 0-3-3.87"/>
                  <path d="M16 3.13a4 4 0 0 1 0 7.75"/>
                </svg>
                {{ ws.member_count }}
              </span>
            </td>

            <!-- Subscriptions -->
            <td class="subs-cell">
              <div v-if="ws.subscriptions.length" class="sub-pills">
                <div
                  v-for="s in ws.subscriptions"
                  :key="s.product_name"
                  :class="['sub-pill', subPillClass(s)]"
                >
                  <span class="pill-dot"></span>
                  <span class="pill-name">{{ s.product_name }}</span>
                  <span class="pill-status">{{ s.status }}</span>
                  <span class="pill-exp" v-if="s.status === 'active'">· {{ fmtDate(s.current_period_end) }}</span>
                </div>
              </div>
              <span v-else class="text-muted no-subs">No subscriptions</span>
            </td>

            <!-- Created -->
            <td class="date-cell">
              {{ fmtDate(ws.created_at) }}
            </td>

          </tr>
        </tbody>
      </table>
    </div>

    <!-- Footer count -->
    <div v-if="!loading && !loadError && filtered.length > 0" class="aw-footer">
      Showing {{ filtered.length }} of {{ workspaces.length }} workspaces
    </div>

  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { adminAPI } from '../../services/api'

const workspaces = ref([])
const loading    = ref(false)
const loadError  = ref('')
const search     = ref('')
const filterSubs = ref('')

// ── Computed stats ─────────────────────────────────────────────────────────────
const activeSubsCount = computed(() =>
  workspaces.value.reduce((n, ws) =>
    n + ws.subscriptions.filter(s => s.status === 'active' && new Date(s.current_period_end) > new Date()).length
  , 0)
)
const withSubsCount = computed(() =>
  workspaces.value.filter(ws => ws.subscriptions.some(s => s.status === 'active')).length
)
const noSubsCount = computed(() =>
  workspaces.value.filter(ws => ws.subscriptions.length === 0).length
)

// ── Filtered list ──────────────────────────────────────────────────────────────
const filtered = computed(() => {
  let list = workspaces.value
  const q = search.value.toLowerCase().trim()
  if (q) {
    list = list.filter(ws =>
      ws.name.toLowerCase().includes(q) ||
      (ws.owner_email || '').toLowerCase().includes(q) ||
      (ws.owner_name || '').toLowerCase().includes(q)
    )
  }
  if (filterSubs.value === 'active') {
    list = list.filter(ws => ws.subscriptions.some(s => s.status === 'active'))
  } else if (filterSubs.value === 'none') {
    list = list.filter(ws => ws.subscriptions.length === 0)
  }
  return list
})

// ── Fetch ──────────────────────────────────────────────────────────────────────
async function fetchWorkspaces() {
  loading.value   = true
  loadError.value = ''
  try {
    const { data } = await adminAPI.listWorkspaces()
    workspaces.value = data.workspaces ?? []
  } catch (err) {
    loadError.value = err.response?.data?.error ?? 'Failed to load workspaces.'
  } finally {
    loading.value = false
  }
}

onMounted(fetchWorkspaces)

// ── Helpers ────────────────────────────────────────────────────────────────────
function subPillClass(s) {
  if (s.status === 'active') {
    return new Date(s.current_period_end) > new Date() ? 'pill-active' : 'pill-expired'
  }
  return s.status === 'canceled' ? 'pill-canceled' : 'pill-expired'
}

function fmtDate(iso) {
  if (!iso) return '—'
  return new Date(iso).toLocaleDateString('en-GB', { day: 'numeric', month: 'short', year: 'numeric' })
}
</script>

<style scoped>
.aw-page { padding: 0; }

/* Header */
.aw-header {
  display: flex; align-items: flex-start; justify-content: space-between;
  gap: 16px; margin-bottom: 24px;
}
.aw-title { font-size: 22px; font-weight: 600; color: #202124; letter-spacing: -0.3px; margin-bottom: 4px; }
.aw-sub   { font-size: 13.5px; color: #5f6368; }

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
.aw-stats { display: flex; gap: 12px; margin-bottom: 24px; flex-wrap: wrap; }
.stat-card {
  background: #fff; border: 1px solid #dadce0; border-radius: 8px;
  padding: 14px 22px; min-width: 130px;
}
.stat-value { font-size: 26px; font-weight: 700; color: #202124; line-height: 1; margin-bottom: 5px; }
.stat-value.stat-green { color: #137333; }
.stat-label { font-size: 12px; color: #5f6368; font-weight: 500; }

/* Toolbar */
.aw-toolbar {
  display: flex; align-items: center; gap: 12px; margin-bottom: 16px; flex-wrap: wrap;
}
.search-wrap { position: relative; flex: 1; min-width: 220px; }
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
  background: none; border: none; cursor: pointer; color: #80868b; padding: 3px; border-radius: 4px; line-height: 0;
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
.aw-error {
  display: flex; align-items: center; gap: 10px;
  background: #fce8e6; border: 1px solid #f5c6c4; border-radius: 8px;
  color: #c5221f; padding: 14px 16px; font-size: 13.5px; margin-bottom: 16px;
}

/* Skeleton */
.skeleton-wrap { background: #fff; border: 1px solid #dadce0; border-radius: 10px; overflow: hidden; }
.skeleton-row { height: 56px; background: linear-gradient(90deg, #f1f3f4 25%, #e8eaed 50%, #f1f3f4 75%); background-size: 400% 100%; animation: shimmer 1.4s ease infinite; border-bottom: 1px solid #e8eaed; }
@keyframes shimmer { 0% { background-position: 100% 0 } 100% { background-position: -100% 0 } }

/* Table */
.aw-table-wrap { background: #fff; border: 1px solid #dadce0; border-radius: 10px; overflow: hidden; }
.aw-table { width: 100%; border-collapse: collapse; font-size: 13.5px; }
.aw-table thead th {
  background: #f8f9fa; font-weight: 600; font-size: 11.5px;
  text-transform: uppercase; letter-spacing: 0.05em; color: #5f6368;
  padding: 11px 16px; text-align: left; border-bottom: 1px solid #e8eaed; white-space: nowrap;
}
.center-th { text-align: center; }
.aw-table tbody tr { border-bottom: 1px solid #f1f3f4; transition: background 0.1s; }
.aw-table tbody tr:last-child { border-bottom: none; }
.aw-table tbody tr:hover { background: #f8f9fa; }
.aw-table tbody tr.row-no-subs { opacity: 0.8; }
.aw-table td { padding: 14px 16px; vertical-align: top; }
.aw-empty { text-align: center; color: #9aa0a6; padding: 48px 16px !important; font-size: 14px; }

/* Workspace */
.ws-name { font-weight: 600; color: #202124; font-size: 14px; margin-bottom: 3px; }
.ws-id   { font-size: 11px; color: #9aa0a6; font-family: 'SFMono-Regular', Consolas, monospace; }

/* Owner */
.owner-email { font-size: 13px; color: #202124; }
.owner-name  { font-size: 12px; color: #5f6368; margin-top: 2px; }

/* Members */
.center-td { text-align: center; }
.member-badge {
  display: inline-flex; align-items: center; gap: 5px;
  background: #e8f0fe; color: #1a73e8; font-size: 12.5px; font-weight: 600;
  padding: 4px 10px; border-radius: 20px;
}

/* Subscription pills */
.subs-cell { min-width: 220px; }
.sub-pills { display: flex; flex-direction: column; gap: 5px; }
.sub-pill {
  display: inline-flex; align-items: center; gap: 6px;
  padding: 4px 10px; border-radius: 6px; font-size: 12px; font-weight: 500;
  width: fit-content;
}
.pill-dot { width: 7px; height: 7px; border-radius: 50%; flex-shrink: 0; }
.pill-active   { background: #e6f4ea; color: #137333; }
.pill-active .pill-dot   { background: #34a853; }
.pill-canceled { background: #f1f3f4; color: #5f6368; }
.pill-canceled .pill-dot { background: #9aa0a6; }
.pill-expired  { background: #fce8e6; color: #c5221f; }
.pill-expired .pill-dot  { background: #ea4335; }
.pill-name   { font-weight: 600; }
.pill-status { opacity: 0.75; font-size: 11px; text-transform: capitalize; }
.pill-exp    { opacity: 0.65; font-size: 11px; }
.no-subs { font-size: 12.5px; }

/* Date */
.date-cell { font-size: 13px; color: #3c4043; white-space: nowrap; }
.text-muted { color: #9aa0a6; }

/* Footer */
.aw-footer { margin-top: 12px; font-size: 12.5px; color: #5f6368; text-align: right; }
</style>
