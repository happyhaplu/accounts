<template>
  <div class="settings-page">

    <!-- Page header -->
    <div class="page-header">
      <div class="page-header-inner">
        <h1>Workspaces</h1>
        <p>Manage your workspaces and team members.</p>
      </div>
    </div>

    <div class="page-body">

      <section class="section">

        <div class="card">

          <!-- skeleton -->
          <div v-if="wsLoading" class="profile-skeleton">
            <div class="skeleton-row" style="width:55%;height:28px"></div>
            <div class="skeleton-row" style="height:60px"></div>
            <div class="skeleton-row" style="height:60px"></div>
          </div>

          <!-- error -->
          <div v-else-if="wsListError" class="alert alert-error">{{ wsListError }}</div>

          <!-- loaded -->
          <template v-else>

            <!-- workspace switcher tabs -->
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
              <!-- create new -->
              <button
                v-if="!showCreateWs"
                class="ws-tab ws-tab-new"
                @click="showCreateWs = true"
                title="Create a new workspace"
              >
                <svg width="13" height="13" viewBox="0 0 24 24" fill="none"
                     stroke="currentColor" stroke-width="2.5"
                     stroke-linecap="round" stroke-linejoin="round">
                  <line x1="12" y1="5" x2="12" y2="19"/>
                  <line x1="5" y1="12" x2="19" y2="12"/>
                </svg>
                New workspace
              </button>
            </div>

            <!-- inline create-workspace form -->
            <div v-if="showCreateWs" class="ws-create-form">
              <div v-if="createWsError" class="alert alert-error">{{ createWsError }}</div>
              <div class="ws-invite-row">
                <div class="input-wrap" style="flex:1">
                  <svg class="field-icon" width="16" height="16" viewBox="0 0 24 24" fill="none"
                       stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                    <rect x="2" y="7" width="20" height="14" rx="2"/>
                    <path d="M16 7V5a2 2 0 0 0-4 0v2M8 7V5a2 2 0 0 1 4 0"/>
                  </svg>
                  <input
                    v-model="newWsName"
                    type="text"
                    placeholder="e.g. Marketing Team"
                    maxlength="80"
                    @keyup.enter="createWorkspace"
                    @keyup.esc="showCreateWs = false"
                    autofocus
                  />
                </div>
                <button class="btn-primary" :disabled="createWsLoading || !newWsName.trim()" @click="createWorkspace">
                  {{ createWsLoading ? 'Creating…' : 'Create' }}
                </button>
                <button class="btn-ghost" @click="showCreateWs = false; newWsName = ''; createWsError = ''">Cancel</button>
              </div>
            </div>

            <!-- active workspace detail -->
            <template v-if="ws">
              <div v-if="wsError" class="alert alert-error" style="margin-top:1rem">{{ wsError }}</div>

              <!-- header: name + member count -->
              <div class="ws-header">
                <div>
                  <div class="ws-name">{{ ws.name }}</div>
                  <span class="ws-role-badge" :class="ws.my_role">
                    {{ ws.my_role === 'owner' ? '★ Owner' : 'Member' }}
                  </span>
                </div>
                <div class="ws-meta">
                  {{ ws.members?.length || 0 }} member{{ ws.members?.length !== 1 ? 's' : '' }}
                </div>
              </div>

              <!-- member rows -->
              <div v-if="wsDetailLoading" class="profile-skeleton" style="margin-bottom:1rem">
                <div class="skeleton-row" style="height:52px"></div>
                <div class="skeleton-row" style="height:52px"></div>
              </div>
              <div v-else class="ws-members">
                <div v-for="m in ws.members" :key="m.id" class="ws-member">
                  <div class="ws-avatar">{{ memberInitial(m) }}</div>
                  <div class="ws-member-info">
                    <div class="ws-member-name">{{ m.user.name || m.user.email }}</div>
                    <div class="ws-member-email">{{ m.user.email }}</div>
                  </div>
                  <span class="ws-role-chip" :class="m.role">
                    {{ m.role === 'owner' ? '★ Owner' : 'Member' }}
                  </span>
                  <button
                    v-if="ws.my_role === 'owner' && m.role !== 'owner'"
                    class="ws-remove-btn"
                    :disabled="removeLoading === m.user.id"
                    title="Remove member"
                    @click="removeMember(m)"
                  >
                    <svg width="14" height="14" viewBox="0 0 24 24" fill="none"
                         stroke="currentColor" stroke-width="2.5"
                         stroke-linecap="round" stroke-linejoin="round">
                      <line x1="18" y1="6" x2="6" y2="18"/>
                      <line x1="6" y1="6" x2="18" y2="18"/>
                    </svg>
                  </button>
                </div>
              </div>

              <!-- add-member / invite form (owners only) -->
              <div v-if="ws.my_role === 'owner'" class="ws-invite">
                <div class="ws-invite-title">Invite a member</div>
                <div v-if="inviteSuccess" class="alert alert-success">{{ inviteSuccess }}</div>
                <div v-if="inviteError"   class="alert alert-error">{{ inviteError }}</div>
                <div class="ws-invite-row">
                  <div class="input-wrap ws-invite-email">
                    <svg class="field-icon" width="16" height="16" viewBox="0 0 24 24" fill="none"
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
                  <select v-model="inviteRole" class="country-code-select ws-role-select">
                    <option value="member">Member</option>
                    <option value="owner">Owner</option>
                  </select>
                  <button
                    class="btn-primary"
                    :disabled="inviteLoading || !inviteEmail.trim()"
                    @click="addMember"
                  >
                    {{ inviteLoading ? 'Sending…' : 'Send invite' }}
                  </button>
                </div>
              </div>

              <!-- pending invites (owners only) -->
              <div v-if="ws.my_role === 'owner' && pendingInvites.length > 0" class="ws-invite">
                <div class="ws-invite-title">
                  Pending invites
                  <span class="ws-pending-count">{{ pendingInvites.length }}</span>
                </div>
                <div class="ws-members">
                  <div v-for="inv in pendingInvites" :key="inv.id" class="ws-member">
                    <div class="ws-avatar ws-avatar-invite">✉</div>
                    <div class="ws-member-info">
                      <div class="ws-member-name">{{ inv.email }}</div>
                      <div class="ws-member-email">Invited · expires {{ formatDate(inv.expires_at) }}</div>
                    </div>
                    <span class="ws-role-chip member">{{ inv.role === 'owner' ? '★ Owner' : 'Member' }}</span>
                    <span class="ws-badge-pending">Pending</span>
                    <button
                      class="ws-remove-btn"
                      :disabled="revokeInviteLoading === inv.id"
                      title="Revoke invite"
                      @click="doRevokeInvite(inv)"
                    >
                      <svg width="14" height="14" viewBox="0 0 24 24" fill="none"
                           stroke="currentColor" stroke-width="2.5"
                           stroke-linecap="round" stroke-linejoin="round">
                        <line x1="18" y1="6" x2="6" y2="18"/>
                        <line x1="6" y1="6" x2="18" y2="18"/>
                      </svg>
                    </button>
                  </div>
                </div>
              </div>
            </template>

          </template>

        </div>
      </section>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useAuthStore } from '../stores/auth'
import { authAPI } from '../services/api'

const auth = useAuthStore()

// ── Workspace state ────────────────────────────────────────────────────────────
const wsList          = ref([])
const activeWsId      = ref(null)
const ws              = ref(null)
const wsLoading       = ref(true)
const wsDetailLoading = ref(false)
const wsListError     = ref('')
const wsError         = ref('')

// create-new form
const showCreateWs    = ref(false)
const newWsName       = ref('')
const createWsLoading = ref(false)
const createWsError   = ref('')

// invite form
const inviteEmail         = ref('')
const inviteRole          = ref('member')
const inviteLoading       = ref(false)
const inviteError         = ref('')
const inviteSuccess       = ref('')
const removeLoading       = ref('')

// pending invites
const pendingInvites      = ref([])
const revokeInviteLoading = ref('')

onMounted(() => loadWorkspaceList())

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
    wsListError.value = 'Failed to load workspaces.'
  } finally {
    wsLoading.value = false
  }
}

async function switchWorkspace(id) {
  activeWsId.value      = id
  wsDetailLoading.value = true
  wsError.value         = ''
  inviteError.value     = ''
  inviteSuccess.value   = ''
  inviteEmail.value     = ''
  pendingInvites.value  = []
  try {
    const { data } = await authAPI.getWorkspace(id)
    ws.value = data.workspace
    if (ws.value?.my_role === 'owner') loadInvites()
  } catch {
    wsError.value = 'Failed to load workspace details.'
  } finally {
    wsDetailLoading.value = false
  }
}

async function createWorkspace() {
  if (!newWsName.value.trim()) return
  createWsLoading.value = true
  createWsError.value   = ''
  try {
    const { data } = await authAPI.createWorkspace({ name: newWsName.value.trim() })
    wsList.value.push(data.workspace)
    showCreateWs.value = false
    newWsName.value    = ''
    await switchWorkspace(data.workspace.id)
  } catch (err) {
    createWsError.value = err.response?.data?.error ?? 'Failed to create workspace.'
  } finally {
    createWsLoading.value = false
  }
}

async function addMember() {
  if (!inviteEmail.value.trim() || !ws.value) return
  inviteLoading.value = true
  inviteError.value   = ''
  inviteSuccess.value = ''
  try {
    const { data } = await authAPI.sendInvite(ws.value.id, {
      email: inviteEmail.value.trim(),
      role:  inviteRole.value,
    })
    inviteSuccess.value = data.message
    inviteEmail.value   = ''
    if (data.invite) pendingInvites.value.push(data.invite)
    setTimeout(() => { inviteSuccess.value = '' }, 5000)
  } catch (err) {
    inviteError.value = err.response?.data?.error ?? 'Failed to send invite.'
  } finally {
    inviteLoading.value = false
  }
}

async function removeMember(m) {
  if (!ws.value) return
  removeLoading.value = m.user.id
  try {
    await authAPI.removeMember(ws.value.id, m.user.id)
    ws.value.members = ws.value.members.filter(x => x.id !== m.id)
  } catch (err) {
    wsError.value = err.response?.data?.error ?? 'Failed to remove member.'
  } finally {
    removeLoading.value = ''
  }
}

async function loadInvites() {
  if (!ws.value) return
  try {
    const { data } = await authAPI.listInvites(ws.value.id)
    pendingInvites.value = data.invites ?? []
  } catch {
    // non-critical — silently ignore
  }
}

async function doRevokeInvite(inv) {
  revokeInviteLoading.value = inv.id
  try {
    await authAPI.revokeInvite(ws.value.id, inv.id)
    pendingInvites.value = pendingInvites.value.filter(i => i.id !== inv.id)
  } catch (err) {
    wsError.value = err.response?.data?.error ?? 'Failed to revoke invite.'
  } finally {
    revokeInviteLoading.value = ''
  }
}

function memberInitial(m) {
  const name = m.user?.name?.trim()
  if (name) return name[0].toUpperCase()
  return (m.user?.email?.[0] ?? '?').toUpperCase()
}

function formatDate(iso) {
  if (!iso) return '—'
  return new Date(iso).toLocaleDateString('en-US', { year: 'numeric', month: 'long', day: 'numeric' })
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
@keyframes shimmer { 0% { background-position: 200% 0 } 100% { background-position: -200% 0 } }

/* Alerts */
.alert {
  display: flex; align-items: center; gap: 8px;
  padding: 10px 14px; border-radius: 8px;
  font-size: 0.85rem; font-weight: 500; margin-bottom: 1.25rem;
}
.alert-success { background: #e6f4ea; color: #1e8e3e; }
.alert-error   { background: #fce8e6; color: #c5221f; }

/* Buttons */
.btn-primary {
  height: 38px; padding: 0 20px; border-radius: 8px;
  background: var(--blue, #1a73e8); color: #fff; font-weight: 600;
  font-size: 0.875rem; border: none; cursor: pointer;
  transition: background .15s, opacity .15s;
}
.btn-primary:hover:not(:disabled) { background: #1557b0; }
.btn-primary:disabled { opacity: 0.55; cursor: not-allowed; }

.btn-ghost {
  height: 38px; padding: 0 18px; border-radius: 8px;
  background: none; color: var(--text-muted, #5f6368); font-weight: 600;
  font-size: 0.875rem; border: 1.5px solid var(--border, #dadce0); cursor: pointer;
  transition: border-color .15s, color .15s;
}
.btn-ghost:hover:not(:disabled) { border-color: #aaa; color: var(--text, #202124); }
.btn-ghost:disabled { opacity: 0.45; cursor: not-allowed; }

/* Workspace tabs */
.ws-tabs {
  display: flex; flex-wrap: wrap; gap: 6px; margin-bottom: 1.25rem;
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
.ws-tab-new {
  border-style: dashed; border-color: var(--border, #dadce0);
  color: var(--text-muted, #5f6368);
}
.ws-tab-new:hover { border-color: var(--blue, #1a73e8); color: var(--blue, #1a73e8); }

.ws-create-form {
  background: var(--bg, #f1f3f4); border-radius: 10px;
  padding: 14px 16px; margin-bottom: 1.25rem;
}

/* Workspace header / members */
.ws-header {
  display: flex; align-items: flex-start; justify-content: space-between;
  margin-bottom: 1.25rem;
}
.ws-name { font-size: 1.05rem; font-weight: 700; color: var(--text, #202124); margin-bottom: 6px; }
.ws-meta { font-size: 0.8rem; color: var(--text-muted, #5f6368); padding-top: 4px; }

.ws-role-badge {
  display: inline-flex; align-items: center; gap: 4px;
  font-size: 0.73rem; font-weight: 700;
  padding: 3px 10px; border-radius: 20px; letter-spacing: .01em;
}
.ws-role-badge.owner  { background: var(--blue-light, #e8f0fe); color: var(--blue, #1a73e8); }
.ws-role-badge.member { background: var(--bg, #f1f3f4);         color: var(--text-muted, #5f6368); }

.ws-members {
  border: 1.5px solid var(--border, #dadce0); border-radius: 10px;
  overflow: hidden; margin-bottom: 1.5rem;
}
.ws-member {
  display: flex; align-items: center; gap: 12px;
  padding: 11px 16px; border-bottom: 1px solid var(--border, #dadce0);
}
.ws-member:last-child { border-bottom: none; }

.ws-avatar {
  width: 36px; height: 36px; border-radius: 50%; flex-shrink: 0;
  background: var(--blue, #1a73e8); color: #fff;
  display: flex; align-items: center; justify-content: center;
  font-size: 14px; font-weight: 700;
}
.ws-member-info { flex: 1; min-width: 0; }
.ws-member-name  { font-size: 0.875rem; font-weight: 600; color: var(--text, #202124); }
.ws-member-email { font-size: 0.75rem;  color: var(--text-muted, #5f6368); }

.ws-role-chip {
  font-size: 0.72rem; font-weight: 700; padding: 3px 9px; border-radius: 12px; flex-shrink: 0;
}
.ws-role-chip.owner  { background: var(--blue-light, #e8f0fe); color: var(--blue, #1a73e8); }
.ws-role-chip.member { background: var(--bg, #f1f3f4);         color: var(--text-muted, #5f6368); }

.ws-remove-btn {
  background: none; border: none; cursor: pointer;
  width: 28px; height: 28px; border-radius: 6px; flex-shrink: 0;
  color: var(--text-muted, #5f6368);
  display: flex; align-items: center; justify-content: center;
  transition: background .15s, color .15s;
}
.ws-remove-btn:hover:not(:disabled) { background: #fce8e6; color: #c5221f; }
.ws-remove-btn:disabled { opacity: 0.4; cursor: not-allowed; }

.ws-invite { border-top: 1px solid var(--border, #dadce0); padding-top: 1.25rem; }
.ws-invite-title {
  font-size: 0.875rem; font-weight: 700; color: var(--text, #202124);
  margin-bottom: 0.75rem; display: flex; align-items: center; gap: 8px;
}
.ws-invite-row   { display: flex; gap: 8px; align-items: center; flex-wrap: wrap; }
.ws-invite-email { flex: 1; min-width: 200px; }
.ws-role-select  { height: 42px; width: 115px; flex-shrink: 0; }

.ws-pending-count {
  display: inline-flex; align-items: center; justify-content: center;
  min-width: 20px; height: 20px; padding: 0 6px;
  border-radius: 10px; font-size: 0.7rem; font-weight: 700;
  background: var(--warning-bg, #fef7e0); color: var(--warning, #f9ab00);
}
.ws-avatar-invite {
  background: var(--warning-bg, #fef7e0); color: var(--warning, #f9ab00);
  font-size: 13px;
}
.ws-badge-pending {
  font-size: 0.7rem; font-weight: 700; padding: 3px 9px; border-radius: 12px; flex-shrink: 0;
  background: var(--warning-bg, #fef7e0); color: #b06000;
}

/* Input / field helpers shared with invite form */
.input-wrap {
  position: relative; display: flex; align-items: center;
}
.input-wrap input {
  width: 100%; height: 42px; border: 1.5px solid var(--border, #dadce0);
  border-radius: 8px; padding: 0 12px 0 38px;
  font-size: 0.875rem; color: var(--text, #202124);
  background: #fff; outline: none;
  transition: border-color .15s, box-shadow .15s;
  font-family: inherit;
}
.input-wrap input:focus {
  border-color: var(--blue, #1a73e8);
  box-shadow: 0 0 0 3px rgba(26, 115, 232, 0.12);
}
.field-icon {
  position: absolute; left: 11px; color: var(--text-muted, #5f6368); pointer-events: none; flex-shrink: 0;
}

.country-code-select {
  height: 42px; border: 1.5px solid var(--border, #dadce0);
  border-radius: 8px; padding: 0 10px;
  font-size: 0.82rem; color: var(--text, #202124);
  background: #fff; outline: none; cursor: pointer;
  transition: border-color .15s, box-shadow .15s;
  font-family: inherit;
}
.country-code-select:focus {
  border-color: var(--blue, #1a73e8);
  box-shadow: 0 0 0 3px rgba(26, 115, 232, 0.12);
}
</style>
