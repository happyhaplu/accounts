<template>
  <div class="ws-page">

    <!-- ── Hero header ─────────────────────────────────────────── -->
    <div class="ws-hero">
      <div class="ws-hero-inner">
        <div class="ws-hero-left">
          <div class="ws-hero-icon">
            <svg width="22" height="22" viewBox="0 0 24 24" fill="none"
                 stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <rect x="2" y="7" width="20" height="14" rx="2"/>
              <path d="M16 7V5a2 2 0 0 0-4 0v2M8 7V5a2 2 0 0 1 4 0"/>
            </svg>
          </div>
          <div>
            <h1>Workspaces</h1>
            <p>Manage your teams and collaborate with members</p>
          </div>
        </div>
        <button class="btn-new-ws" @click="openCreateModal">
          <svg width="15" height="15" viewBox="0 0 24 24" fill="none"
               stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round">
            <line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/>
          </svg>
          New workspace
        </button>
      </div>
    </div>

    <!-- ── Body ───────────────────────────────────────────────── -->
    <div class="ws-body">

      <!-- Global skeleton while list loads -->
      <div v-if="wsLoading" class="skeleton-page">
        <div class="skel-sidebar">
          <div v-for="i in 2" :key="i" class="skel-ws-item"></div>
          <div class="skel-ws-item short"></div>
        </div>
        <div class="skel-main">
          <div class="skel-header"></div>
          <div class="skel-row"></div>
          <div class="skel-row"></div>
          <div class="skel-row"></div>
        </div>
      </div>

      <!-- List error -->
      <div v-else-if="wsListError" class="full-alert alert-error">
        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor"
             stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
          <circle cx="12" cy="12" r="10"/><line x1="12" y1="8" x2="12" y2="12"/>
          <line x1="12" y1="16" x2="12.01" y2="16"/>
        </svg>
        {{ wsListError }}
      </div>

      <!-- Main layout: sidebar + detail -->
      <div v-else class="ws-layout">

        <!-- ── Sidebar: workspace list ─────────────────────────── -->
        <aside class="ws-sidebar">
          <div class="sidebar-label">Your workspaces</div>
          <button
            v-for="item in wsList" :key="item.id"
            class="ws-sidebar-item"
            :class="{ active: item.id === activeWsId }"
            @click="switchWorkspace(item.id)"
          >
            <div class="ws-sidebar-avatar" :class="item.my_role">
              {{ wsInitial(item.name) }}
            </div>
            <div class="ws-sidebar-info">
              <div class="ws-sidebar-name">{{ item.name }}</div>
              <div class="ws-sidebar-role">{{ item.my_role === 'owner' ? '★ Owner' : 'Member' }}</div>
            </div>
            <svg v-if="item.id === activeWsId" width="14" height="14" viewBox="0 0 24 24"
                 fill="none" stroke="currentColor" stroke-width="2.5"
                 stroke-linecap="round" stroke-linejoin="round" class="sidebar-check">
              <polyline points="20 6 9 17 4 12"/>
            </svg>
          </button>
        </aside>

        <!-- ── Detail panel ────────────────────────────────────── -->
        <main class="ws-detail">

          <!-- Detail skeleton while switching -->
          <div v-if="wsDetailLoading" class="skel-detail">
            <div class="skel-header"></div>
            <div class="skel-row"></div>
            <div class="skel-row"></div>
            <div class="skel-row short"></div>
          </div>

          <template v-else-if="ws">

            <!-- Workspace header card -->
            <div class="ws-detail-header card">
              <div class="ws-detail-meta">
                <div class="ws-detail-avatar">{{ wsInitial(ws.name) }}</div>
                <div>
                  <div class="ws-detail-name">{{ ws.name }}</div>
                  <div class="ws-detail-badges">
                    <span class="role-badge" :class="ws.my_role">
                      {{ ws.my_role === 'owner' ? '★ Owner' : 'Member' }}
                    </span>
                    <span class="member-count-badge">
                      {{ ws.members?.length || 0 }} member{{ ws.members?.length !== 1 ? 's' : '' }}
                    </span>
                    <span class="created-badge">
                      Created {{ formatDate(ws.created_at) }}
                    </span>
                  </div>
                </div>
              </div>
            </div>

            <!-- ── Members section ─────────────────────────────── -->
            <div class="section-card card">
              <div class="section-card-head">
                <div class="section-card-title">
                  <svg width="16" height="16" viewBox="0 0 24 24" fill="none"
                       stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                    <path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"/>
                    <circle cx="9" cy="7" r="4"/>
                    <path d="M23 21v-2a4 4 0 0 0-3-3.87"/>
                    <path d="M16 3.13a4 4 0 0 1 0 7.75"/>
                  </svg>
                  Team members
                </div>
              </div>

              <!-- Member list -->
              <div class="member-list">
                <div v-for="m in ws.members" :key="m.id" class="member-row">
                  <div class="member-avatar">{{ memberInitial(m) }}</div>
                  <div class="member-info">
                    <div class="member-name">{{ m.user.name || m.user.email }}</div>
                    <div class="member-email">{{ m.user.email }}</div>
                  </div>
                  <div class="member-role-wrap">
                    <span class="role-chip" :class="m.role">
                      {{ m.role === 'owner' ? '★ Owner' : 'Member' }}
                    </span>
                  </div>
                  <div class="member-joined">
                    Joined {{ formatDate(m.joined_at) }}
                  </div>
                  <button
                    v-if="ws.my_role === 'owner' && m.role !== 'owner'"
                    class="remove-btn"
                    :disabled="removeLoading === m.user.id"
                    title="Remove from workspace"
                    @click="confirmRemove(m)"
                  >
                    <svg v-if="removeLoading !== m.user.id" width="14" height="14" viewBox="0 0 24 24"
                         fill="none" stroke="currentColor" stroke-width="2.5"
                         stroke-linecap="round" stroke-linejoin="round">
                      <polyline points="3 6 5 6 21 6"/>
                      <path d="M19 6l-1 14H6L5 6"/>
                      <path d="M10 11v6M14 11v6"/>
                      <path d="M9 6V4h6v2"/>
                    </svg>
                    <svg v-else width="14" height="14" viewBox="0 0 24 24" fill="none"
                         stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
                         class="spin">
                      <path d="M21 12a9 9 0 1 1-6.219-8.56"/>
                    </svg>
                  </button>
                </div>

                <div v-if="!ws.members?.length" class="empty-members">
                  No members yet. Add someone below.
                </div>
              </div>

              <!-- ── Add member form (owner only) ─────────────── -->
              <div v-if="ws.my_role === 'owner'" class="invite-form">
                <div class="invite-form-title">
                  <svg width="14" height="14" viewBox="0 0 24 24" fill="none"
                       stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round">
                    <line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/>
                  </svg>
                  Add member
                </div>
                <div v-if="inviteSuccess" class="inline-alert alert-success">
                  <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor"
                       stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round">
                    <polyline points="20 6 9 17 4 12"/>
                  </svg>
                  {{ inviteSuccess }}
                </div>
                <div v-if="inviteError" class="inline-alert alert-error">
                  <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor"
                       stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                    <circle cx="12" cy="12" r="10"/>
                    <line x1="12" y1="8" x2="12" y2="12"/><line x1="12" y1="16" x2="12.01" y2="16"/>
                  </svg>
                  {{ inviteError }}
                </div>
                <div class="invite-row">
                  <div class="invite-email-wrap">
                    <svg class="invite-icon" width="15" height="15" viewBox="0 0 24 24" fill="none"
                         stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                      <rect x="2" y="4" width="20" height="16" rx="2"/>
                      <path d="m22 7-10 7L2 7"/>
                    </svg>
                    <input
                      v-model="inviteEmail"
                      type="email"
                      placeholder="colleague@company.com"
                      @keyup.enter="addMember"
                    />
                  </div>
                  <select v-model="inviteRole" class="role-select">
                    <option value="member">Member</option>
                    <option value="owner">Owner</option>
                  </select>
                  <button class="btn-invite" :disabled="inviteLoading || !inviteEmail.trim()" @click="addMember">
                    <svg v-if="!inviteLoading" width="14" height="14" viewBox="0 0 24 24" fill="none"
                         stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round">
                      <line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/>
                    </svg>
                    <svg v-else width="14" height="14" viewBox="0 0 24 24" fill="none"
                         stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
                         class="spin">
                      <path d="M21 12a9 9 0 1 1-6.219-8.56"/>
                    </svg>
                    {{ inviteLoading ? 'Adding…' : 'Add member' }}
                  </button>
                </div>
              </div>
            </div>

          </template>
        </main>
      </div>
    </div>

    <!-- ── Create workspace modal ──────────────────────────────── -->
    <transition name="fade">
      <div v-if="showCreate" class="modal-overlay" @click.self="closeCreateModal">
        <div class="modal">
          <div class="modal-header">
            <div class="modal-title">
              <svg width="18" height="18" viewBox="0 0 24 24" fill="none"
                   stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <rect x="2" y="7" width="20" height="14" rx="2"/>
                <path d="M16 7V5a2 2 0 0 0-4 0v2M8 7V5a2 2 0 0 1 4 0"/>
              </svg>
              Create new workspace
            </div>
            <button class="modal-close" @click="closeCreateModal">
              <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor"
                   stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round">
                <line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/>
              </svg>
            </button>
          </div>
          <div class="modal-body">
            <p class="modal-hint">Give your workspace a name. You can invite members after creating it.</p>
            <div v-if="createError" class="inline-alert alert-error" style="margin-bottom:14px">
              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor"
                   stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <circle cx="12" cy="12" r="10"/>
                <line x1="12" y1="8" x2="12" y2="12"/><line x1="12" y1="16" x2="12.01" y2="16"/>
              </svg>
              {{ createError }}
            </div>
            <div class="modal-field">
              <label>Workspace name</label>
              <input
                ref="createInput"
                v-model="newWsName"
                type="text"
                placeholder="e.g. Marketing Team"
                maxlength="80"
                @keyup.enter="createWorkspace"
                @keyup.esc="closeCreateModal"
              />
              <div class="char-count">{{ newWsName.length }}/80</div>
            </div>
          </div>
          <div class="modal-footer">
            <button class="btn-ghost" @click="closeCreateModal">Cancel</button>
            <button class="btn-primary" :disabled="createLoading || !newWsName.trim()" @click="createWorkspace">
              <svg v-if="!createLoading" width="14" height="14" viewBox="0 0 24 24" fill="none"
                   stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round">
                <line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/>
              </svg>
              <svg v-else width="14" height="14" viewBox="0 0 24 24" fill="none"
                   stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
                   class="spin">
                <path d="M21 12a9 9 0 1 1-6.219-8.56"/>
              </svg>
              {{ createLoading ? 'Creating…' : 'Create workspace' }}
            </button>
          </div>
        </div>
      </div>
    </transition>

    <!-- ── Confirm remove modal ────────────────────────────────── -->
    <transition name="fade">
      <div v-if="confirmMember" class="modal-overlay" @click.self="confirmMember = null">
        <div class="modal modal-sm">
          <div class="modal-header">
            <div class="modal-title danger">Remove member</div>
            <button class="modal-close" @click="confirmMember = null">
              <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor"
                   stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round">
                <line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/>
              </svg>
            </button>
          </div>
          <div class="modal-body">
            <p class="modal-hint">
              Remove <strong>{{ confirmMember.user.name || confirmMember.user.email }}</strong> from
              <strong>{{ ws?.name }}</strong>? They will lose access immediately.
            </p>
          </div>
          <div class="modal-footer">
            <button class="btn-ghost" @click="confirmMember = null">Cancel</button>
            <button class="btn-danger" :disabled="removeLoading" @click="doRemove">
              {{ removeLoading ? 'Removing…' : 'Remove member' }}
            </button>
          </div>
        </div>
      </div>
    </transition>

  </div>
</template>

<script setup>
import { ref, nextTick, onMounted } from 'vue'
import { authAPI } from '../services/api'

// ── State ─────────────────────────────────────────────────────────────────────
const wsList          = ref([])
const activeWsId      = ref(null)
const ws              = ref(null)
const wsLoading       = ref(true)
const wsDetailLoading = ref(false)
const wsListError     = ref('')

const showCreate  = ref(false)
const newWsName   = ref('')
const createLoading = ref(false)
const createError   = ref('')
const createInput   = ref(null)

const inviteEmail   = ref('')
const inviteRole    = ref('member')
const inviteLoading = ref(false)
const inviteError   = ref('')
const inviteSuccess = ref('')

const removeLoading  = ref('')
const confirmMember  = ref(null)

// ── Lifecycle ─────────────────────────────────────────────────────────────────
onMounted(loadWorkspaceList)

// ── Functions ─────────────────────────────────────────────────────────────────
async function loadWorkspaceList() {
  wsLoading.value   = true
  wsListError.value = ''
  try {
    const { data } = await authAPI.listWorkspaces()
    wsList.value = data.workspaces ?? []
    if (wsList.value.length > 0) {
      await switchWorkspace(wsList.value[0].id)
    }
  } catch {
    wsListError.value = 'Failed to load workspaces. Please refresh.'
  } finally {
    wsLoading.value = false
  }
}

async function switchWorkspace(id) {
  if (activeWsId.value === id && ws.value) return
  activeWsId.value      = id
  wsDetailLoading.value = true
  ws.value              = null
  inviteEmail.value     = ''
  inviteError.value     = ''
  inviteSuccess.value   = ''
  try {
    const { data } = await authAPI.getWorkspace(id)
    ws.value = data.workspace
  } catch {
    wsListError.value = 'Failed to load workspace details.'
  } finally {
    wsDetailLoading.value = false
  }
}

// Create
function openCreateModal() {
  showCreate.value  = true
  newWsName.value   = ''
  createError.value = ''
  nextTick(() => createInput.value?.focus())
}
function closeCreateModal() {
  showCreate.value  = false
  newWsName.value   = ''
  createError.value = ''
}
async function createWorkspace() {
  if (!newWsName.value.trim()) return
  createLoading.value = true
  createError.value   = ''
  try {
    const { data } = await authAPI.createWorkspace({ name: newWsName.value.trim() })
    wsList.value.push(data.workspace)
    closeCreateModal()
    await switchWorkspace(data.workspace.id)
  } catch (err) {
    createError.value = err.response?.data?.error ?? 'Failed to create workspace.'
  } finally {
    createLoading.value = false
  }
}

// Add member
async function addMember() {
  if (!inviteEmail.value.trim() || !ws.value) return
  inviteLoading.value = true
  inviteError.value   = ''
  inviteSuccess.value = ''
  try {
    const { data } = await authAPI.addMember(ws.value.id, {
      email: inviteEmail.value.trim(),
      role:  inviteRole.value,
    })
    ws.value.members.push(data.member)
    // Update sidebar count
    const sidebar = wsList.value.find(w => w.id === ws.value.id)
    if (sidebar) sidebar.member_count = (sidebar.member_count || 0) + 1
    inviteSuccess.value = `${inviteEmail.value.trim()} added successfully.`
    inviteEmail.value   = ''
    inviteRole.value    = 'member'
    setTimeout(() => { inviteSuccess.value = '' }, 4000)
  } catch (err) {
    inviteError.value = err.response?.data?.error ?? 'Failed to add member.'
  } finally {
    inviteLoading.value = false
  }
}

// Remove member
function confirmRemove(m) { confirmMember.value = m }
async function doRemove() {
  if (!confirmMember.value || !ws.value) return
  removeLoading.value = confirmMember.value.user.id
  try {
    await authAPI.removeMember(ws.value.id, confirmMember.value.user.id)
    ws.value.members = ws.value.members.filter(x => x.id !== confirmMember.value.id)
    confirmMember.value = null
  } catch (err) {
    wsListError.value = err.response?.data?.error ?? 'Failed to remove member.'
    confirmMember.value = null
  } finally {
    removeLoading.value = ''
  }
}

// ── Helpers ───────────────────────────────────────────────────────────────────
function wsInitial(name) {
  return (name?.trim()?.[0] ?? 'W').toUpperCase()
}
function memberInitial(m) {
  const name = m.user?.name?.trim()
  if (name) return name[0].toUpperCase()
  return (m.user?.email?.[0] ?? '?').toUpperCase()
}
function formatDate(iso) {
  if (!iso) return '—'
  return new Date(iso).toLocaleDateString('en-US', { year: 'numeric', month: 'short', day: 'numeric' })
}
</script>

<style scoped>
/* ── Page shell ──────────────────────────────────────────────────────────────── */
.ws-page { min-height: 100vh; background: var(--bg, #f1f3f4); }

/* ── Hero header ─────────────────────────────────────────────────────────────── */
.ws-hero {
  background: linear-gradient(135deg, #1557b0 0%, #1a73e8 60%, #4f8ef7 100%);
  padding: 2.25rem 2rem 2rem;
  border-bottom: 1px solid rgba(255,255,255,0.1);
}
.ws-hero-inner {
  max-width: 1060px; margin: 0 auto;
  display: flex; align-items: center; justify-content: space-between; gap: 1rem;
  flex-wrap: wrap;
}
.ws-hero-left { display: flex; align-items: center; gap: 16px; }
.ws-hero-icon {
  width: 48px; height: 48px; border-radius: 14px;
  background: rgba(255,255,255,0.2);
  display: flex; align-items: center; justify-content: center;
  color: #fff; flex-shrink: 0;
}
.ws-hero-left h1 { font-size: 1.35rem; font-weight: 700; color: #fff; margin-bottom: 4px; }
.ws-hero-left p  { font-size: 0.85rem; color: rgba(255,255,255,0.75); }

.btn-new-ws {
  display: inline-flex; align-items: center; gap: 7px;
  height: 40px; padding: 0 18px;
  background: #fff; color: var(--blue, #1a73e8);
  border: none; border-radius: 10px; cursor: pointer;
  font-size: 0.875rem; font-weight: 700;
  font-family: inherit;
  transition: box-shadow .15s, transform .1s;
  box-shadow: 0 2px 10px rgba(0,0,0,0.12);
}
.btn-new-ws:hover { box-shadow: 0 4px 18px rgba(0,0,0,0.18); transform: translateY(-1px); }

/* ── Body ────────────────────────────────────────────────────────────────────── */
.ws-body { max-width: 1060px; margin: 0 auto; padding: 2rem; }

/* ── Skeletons ───────────────────────────────────────────────────────────────── */
.skeleton-page { display: flex; gap: 1.5rem; }
.skel-sidebar { width: 240px; flex-shrink: 0; display: flex; flex-direction: column; gap: 10px; }
.skel-main    { flex: 1; display: flex; flex-direction: column; gap: 12px; }
.skel-ws-item, .skel-header, .skel-row {
  border-radius: 10px;
  background: linear-gradient(90deg, #e8e8e8 25%, #d4d4d4 50%, #e8e8e8 75%);
  background-size: 200% 100%;
  animation: shimmer 1.4s infinite;
}
.skel-ws-item { height: 60px; }
.skel-ws-item.short { width: 70%; height: 38px; }
.skel-header  { height: 90px; }
.skel-row     { height: 60px; }
.skel-detail  { display: flex; flex-direction: column; gap: 12px; }
@keyframes shimmer { 0% { background-position: 200% 0 } 100% { background-position: -200% 0 } }

/* ── Full-width alert ────────────────────────────────────────────────────────── */
.full-alert {
  display: flex; align-items: center; gap: 10px;
  padding: 14px 18px; border-radius: 10px; font-size: 0.875rem; font-weight: 500;
}
.alert-error   { background: #fce8e6; color: #c5221f; border: 1px solid #f5c6c6; }
.alert-success { background: #e6f4ea; color: #1e8e3e; border: 1px solid #b7dfba; }

/* ── Layout: sidebar + main ──────────────────────────────────────────────────── */
.ws-layout { display: flex; gap: 1.5rem; align-items: flex-start; }

/* ── Sidebar ─────────────────────────────────────────────────────────────────── */
.ws-sidebar { width: 240px; flex-shrink: 0; }
.sidebar-label {
  font-size: 0.72rem; font-weight: 700; color: var(--text-muted, #5f6368);
  letter-spacing: .05em; text-transform: uppercase;
  padding: 0 4px; margin-bottom: 10px;
}
.ws-sidebar-item {
  display: flex; align-items: center; gap: 10px;
  width: 100%; padding: 10px 12px; border-radius: 10px;
  background: none; border: 1.5px solid transparent;
  cursor: pointer; text-align: left; transition: all .15s;
  font-family: inherit; margin-bottom: 4px;
}
.ws-sidebar-item:hover { background: #fff; border-color: var(--border, #dadce0); }
.ws-sidebar-item.active {
  background: var(--blue-light, #e8f0fe);
  border-color: var(--blue, #1a73e8);
}
.ws-sidebar-avatar {
  width: 36px; height: 36px; border-radius: 10px; flex-shrink: 0;
  display: flex; align-items: center; justify-content: center;
  font-size: 15px; font-weight: 800; color: #fff;
}
.ws-sidebar-avatar.owner  { background: linear-gradient(135deg, #1a73e8, #1557b0); }
.ws-sidebar-avatar.member { background: linear-gradient(135deg, #5f6368, #3c4043); }
.ws-sidebar-info { flex: 1; min-width: 0; }
.ws-sidebar-name { font-size: 0.875rem; font-weight: 600; color: var(--text, #202124); white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.ws-sidebar-role { font-size: 0.72rem; color: var(--text-muted, #5f6368); margin-top: 1px; }
.ws-sidebar-item.active .ws-sidebar-name { color: var(--blue, #1a73e8); }
.sidebar-check { color: var(--blue, #1a73e8); flex-shrink: 0; }

/* ── Detail panel ────────────────────────────────────────────────────────────── */
.ws-detail { flex: 1; min-width: 0; display: flex; flex-direction: column; gap: 1.25rem; }

.card {
  background: #fff; border: 1px solid var(--border, #dadce0);
  border-radius: 14px; overflow: hidden;
}

/* Workspace header card */
.ws-detail-header { padding: 1.5rem; }
.ws-detail-meta { display: flex; align-items: center; gap: 16px; }
.ws-detail-avatar {
  width: 52px; height: 52px; border-radius: 14px; flex-shrink: 0;
  background: linear-gradient(135deg, #1a73e8, #1557b0);
  display: flex; align-items: center; justify-content: center;
  font-size: 22px; font-weight: 800; color: #fff;
}
.ws-detail-name { font-size: 1.15rem; font-weight: 700; color: var(--text, #202124); margin-bottom: 8px; }
.ws-detail-badges { display: flex; align-items: center; gap: 8px; flex-wrap: wrap; }

.role-badge {
  display: inline-flex; align-items: center;
  font-size: 0.72rem; font-weight: 700;
  padding: 3px 10px; border-radius: 20px;
}
.role-badge.owner  { background: var(--blue-light, #e8f0fe); color: var(--blue, #1a73e8); }
.role-badge.member { background: var(--bg, #f1f3f4);         color: var(--text-muted, #5f6368); }

.member-count-badge, .created-badge {
  font-size: 0.75rem; color: var(--text-muted, #5f6368);
  background: var(--bg, #f1f3f4); padding: 3px 10px; border-radius: 20px;
  font-weight: 500;
}

/* Members section card */
.section-card-head {
  display: flex; align-items: center; justify-content: space-between;
  padding: 1.1rem 1.5rem; border-bottom: 1px solid var(--border, #dadce0);
  background: #fafbff;
}
.section-card-title {
  display: flex; align-items: center; gap: 8px;
  font-size: 0.875rem; font-weight: 700; color: var(--text, #202124);
}

/* Member list */
.member-list { }
.member-row {
  display: flex; align-items: center; gap: 14px;
  padding: 13px 1.5rem;
  border-bottom: 1px solid var(--border, #dadce0);
  transition: background .12s;
}
.member-row:last-child { border-bottom: none; }
.member-row:hover { background: #fafbff; }

.member-avatar {
  width: 38px; height: 38px; border-radius: 50%; flex-shrink: 0;
  background: var(--blue, #1a73e8); color: #fff;
  display: flex; align-items: center; justify-content: center;
  font-size: 14px; font-weight: 700;
}
.member-info { flex: 1; min-width: 0; }
.member-name  { font-size: 0.875rem; font-weight: 600; color: var(--text, #202124); white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.member-email { font-size: 0.75rem;  color: var(--text-muted, #5f6368); margin-top: 1px; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }

.member-role-wrap { flex-shrink: 0; }
.role-chip {
  display: inline-flex; align-items: center;
  font-size: 0.72rem; font-weight: 700;
  padding: 3px 10px; border-radius: 12px;
}
.role-chip.owner  { background: var(--blue-light, #e8f0fe); color: var(--blue, #1a73e8); }
.role-chip.member { background: var(--bg, #f1f3f4);         color: var(--text-muted, #5f6368); }

.member-joined { font-size: 0.75rem; color: var(--text-muted, #5f6368); flex-shrink: 0; }

.remove-btn {
  width: 34px; height: 34px; border-radius: 8px; flex-shrink: 0;
  background: none; border: 1.5px solid transparent; cursor: pointer;
  color: var(--text-muted, #5f6368);
  display: flex; align-items: center; justify-content: center;
  transition: all .15s;
}
.remove-btn:hover:not(:disabled) {
  background: #fce8e6; color: #c5221f; border-color: #f5c6c6;
}
.remove-btn:disabled { opacity: 0.4; cursor: not-allowed; }

.empty-members {
  padding: 2rem 1.5rem;
  font-size: 0.875rem; color: var(--text-muted, #5f6368);
  text-align: center; font-style: italic;
}

/* ── Invite form ─────────────────────────────────────────────────────────────── */
.invite-form {
  padding: 1.25rem 1.5rem;
  border-top: 1px solid var(--border, #dadce0);
  background: #fafbff;
}
.invite-form-title {
  display: flex; align-items: center; gap: 7px;
  font-size: 0.82rem; font-weight: 700; color: var(--text, #202124);
  margin-bottom: 0.9rem;
}
.invite-row { display: flex; align-items: center; gap: 8px; flex-wrap: wrap; }
.invite-email-wrap {
  position: relative; flex: 1; min-width: 200px;
}
.invite-email-wrap input {
  width: 100%; height: 40px;
  border: 1.5px solid var(--border, #dadce0); border-radius: 8px;
  padding: 0 14px 0 36px;
  font-size: 0.875rem; color: var(--text, #202124);
  background: #fff; outline: none; font-family: inherit;
  transition: border-color .15s, box-shadow .15s;
}
.invite-email-wrap input:focus {
  border-color: var(--blue, #1a73e8);
  box-shadow: 0 0 0 3px rgba(26,115,232,.1);
}
.invite-icon {
  position: absolute; left: 11px; top: 50%; transform: translateY(-50%);
  color: var(--text-muted, #5f6368); pointer-events: none;
}
.role-select {
  height: 40px; padding: 0 10px;
  border: 1.5px solid var(--border, #dadce0); border-radius: 8px;
  font-size: 0.82rem; color: var(--text, #202124);
  background: #fff; cursor: pointer; outline: none; font-family: inherit;
  transition: border-color .15s;
}
.role-select:focus { border-color: var(--blue, #1a73e8); }

.btn-invite {
  display: inline-flex; align-items: center; gap: 7px;
  height: 40px; padding: 0 18px;
  background: var(--blue, #1a73e8); color: #fff;
  border: none; border-radius: 8px; cursor: pointer;
  font-size: 0.875rem; font-weight: 600; font-family: inherit;
  transition: background .15s, opacity .15s; white-space: nowrap;
}
.btn-invite:hover:not(:disabled) { background: #1557b0; }
.btn-invite:disabled { opacity: 0.55; cursor: not-allowed; }

.inline-alert {
  display: flex; align-items: center; gap: 8px;
  padding: 9px 13px; border-radius: 8px; font-size: 0.82rem; font-weight: 500;
  margin-bottom: 0.85rem;
}

/* ── Modal ───────────────────────────────────────────────────────────────────── */
.modal-overlay {
  position: fixed; inset: 0; z-index: 500;
  background: rgba(0,0,0,0.45);
  display: flex; align-items: center; justify-content: center;
  padding: 1rem;
}
.modal {
  background: #fff; border-radius: 16px;
  box-shadow: 0 24px 60px rgba(0,0,0,0.2);
  width: 100%; max-width: 460px;
  overflow: hidden;
}
.modal-sm { max-width: 400px; }
.modal-header {
  display: flex; align-items: center; justify-content: space-between;
  padding: 1.25rem 1.5rem;
  border-bottom: 1px solid var(--border, #dadce0);
}
.modal-title {
  display: flex; align-items: center; gap: 9px;
  font-size: 1rem; font-weight: 700; color: var(--text, #202124);
}
.modal-title.danger { color: #c5221f; }
.modal-close {
  background: none; border: none; cursor: pointer;
  color: var(--text-muted, #5f6368); padding: 4px; border-radius: 6px;
  display: flex; align-items: center; transition: background .15s, color .15s;
}
.modal-close:hover { background: var(--bg, #f1f3f4); color: var(--text, #202124); }

.modal-body { padding: 1.5rem; }
.modal-hint { font-size: 0.875rem; color: var(--text-muted, #5f6368); margin-bottom: 1.25rem; line-height: 1.5; }
.modal-field label {
  display: block; font-size: 0.82rem; font-weight: 600;
  color: var(--text, #202124); margin-bottom: 7px;
}
.modal-field input {
  width: 100%; height: 44px;
  border: 1.5px solid var(--border, #dadce0); border-radius: 8px;
  padding: 0 14px; font-size: 0.9rem; color: var(--text, #202124);
  background: #fff; outline: none; font-family: inherit;
  transition: border-color .15s, box-shadow .15s;
}
.modal-field input:focus {
  border-color: var(--blue, #1a73e8);
  box-shadow: 0 0 0 3px rgba(26,115,232,.1);
}
.char-count {
  font-size: 0.72rem; color: var(--text-muted, #5f6368);
  text-align: right; margin-top: 5px;
}

.modal-footer {
  display: flex; align-items: center; justify-content: flex-end; gap: 10px;
  padding: 1rem 1.5rem;
  border-top: 1px solid var(--border, #dadce0);
  background: #fafbff;
}

.btn-primary {
  display: inline-flex; align-items: center; gap: 7px;
  height: 40px; padding: 0 20px;
  background: var(--blue, #1a73e8); color: #fff;
  border: none; border-radius: 8px; cursor: pointer;
  font-size: 0.875rem; font-weight: 600; font-family: inherit;
  transition: background .15s, opacity .15s;
}
.btn-primary:hover:not(:disabled) { background: #1557b0; }
.btn-primary:disabled { opacity: 0.55; cursor: not-allowed; }

.btn-ghost {
  height: 40px; padding: 0 18px;
  background: none; border: 1.5px solid var(--border, #dadce0);
  border-radius: 8px; cursor: pointer;
  font-size: 0.875rem; font-weight: 600; color: var(--text-muted, #5f6368);
  font-family: inherit; transition: border-color .15s, color .15s;
}
.btn-ghost:hover { border-color: #aaa; color: var(--text, #202124); }

.btn-danger {
  display: inline-flex; align-items: center; gap: 7px;
  height: 40px; padding: 0 20px;
  background: #c5221f; color: #fff;
  border: none; border-radius: 8px; cursor: pointer;
  font-size: 0.875rem; font-weight: 600; font-family: inherit;
  transition: background .15s, opacity .15s;
}
.btn-danger:hover:not(:disabled) { background: #a50e0e; }
.btn-danger:disabled { opacity: 0.55; cursor: not-allowed; }

/* Spinner */
.spin { animation: spin 0.8s linear infinite; }
@keyframes spin { to { transform: rotate(360deg); } }

/* Modal transition */
.fade-enter-active, .fade-leave-active { transition: opacity .18s; }
.fade-enter-from, .fade-leave-to { opacity: 0; }
.fade-enter-active .modal, .fade-leave-active .modal { transition: transform .18s; }
.fade-enter-from .modal, .fade-leave-to .modal { transform: scale(0.96) translateY(8px); }

/* ── Responsive ──────────────────────────────────────────────────────────────── */
@media (max-width: 680px) {
  .ws-layout { flex-direction: column; }
  .ws-sidebar { width: 100%; }
  .member-joined { display: none; }
  .ws-hero-inner { flex-direction: column; align-items: flex-start; }
}
</style>
