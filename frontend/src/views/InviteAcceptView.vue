<template>
  <div class="invite-page">
    <div class="invite-card">

      <!-- Logo / brand -->
      <div class="invite-brand">
        <img src="/logo.svg" alt="Gour" class="invite-brand-icon" />
      </div>

      <!-- Loading state -->
      <div v-if="loading" class="invite-state">
        <div class="invite-spinner"></div>
        <p>Loading invite…</p>
      </div>

      <!-- Error: expired / invalid / already used -->
      <div v-else-if="errorMsg" class="invite-state invite-state--error">
        <div class="invite-icon invite-icon--error">
          <svg width="28" height="28" viewBox="0 0 24 24" fill="none"
               stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <circle cx="12" cy="12" r="10"/>
            <line x1="12" y1="8" x2="12" y2="12"/>
            <line x1="12" y1="16" x2="12.01" y2="16"/>
          </svg>
        </div>
        <h2>{{ errorMsg }}</h2>
        <p class="invite-sub">This invite link may have expired or already been used.</p>
        <router-link to="/login" class="invite-btn">Go to login</router-link>
      </div>

      <!-- Accepted already -->
      <div v-else-if="accepted" class="invite-state invite-state--success">
        <div class="invite-icon invite-icon--success">
          <svg width="28" height="28" viewBox="0 0 24 24" fill="none"
               stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round">
            <polyline points="20 6 9 17 4 12"/>
          </svg>
        </div>
        <h2>You're in!</h2>
        <p class="invite-sub">
          You've joined <strong>{{ invite.workspace_name }}</strong> as a
          <strong>{{ invite.role }}</strong>.
        </p>
        <router-link to="/settings" class="invite-btn">Go to workspace settings</router-link>
      </div>

      <!-- Invite info: show details + action -->
      <div v-else-if="invite" class="invite-content">
        <div class="invite-icon invite-icon--info">
          <svg width="28" height="28" viewBox="0 0 24 24" fill="none"
               stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"/>
            <circle cx="9" cy="7" r="4"/>
            <path d="M23 21v-2a4 4 0 0 0-3-3.87"/>
            <path d="M16 3.13a4 4 0 0 1 0 7.75"/>
          </svg>
        </div>

        <h2>You're invited to join<br><em>{{ invite.workspace_name }}</em></h2>
        <p class="invite-sub">
          <strong>{{ invite.invited_by }}</strong> has invited you to collaborate
          as a <strong>{{ invite.role }}</strong>.
        </p>
        <p class="invite-email-chip">{{ invite.email }}</p>

        <!-- Logged in with correct email → show accept button -->
        <template v-if="auth.isAuthenticated">
          <div v-if="emailMismatch" class="invite-alert">
            This invite was sent to <strong>{{ invite.email }}</strong>. You're signed in as
            <strong>{{ auth.user?.email }}</strong>. Please sign in with the correct account to accept.
            <button class="invite-btn-ghost" @click="handleLogout">Switch account</button>
          </div>
          <template v-else>
            <div v-if="acceptError" class="invite-alert">{{ acceptError }}</div>
            <button class="invite-btn" :disabled="accepting" @click="handleAccept">
              {{ accepting ? 'Joining…' : 'Accept & join workspace' }}
            </button>
          </template>
        </template>

        <!-- Not logged in → prompt to sign in / register -->
        <template v-else>
          <p class="invite-prompt">To accept this invite, sign in to your account or create a new one.</p>
          <div class="invite-actions">
            <router-link :to="loginPath" class="invite-btn">Sign in</router-link>
            <router-link :to="registerPath" class="invite-btn-outline">Create account</router-link>
          </div>
        </template>
      </div>

    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter }      from 'vue-router'
import { useAuthStore }             from '../stores/auth'
import { authAPI }                  from '../services/api'

const route  = useRoute()
const router = useRouter()
const auth   = useAuthStore()

const token   = computed(() => route.query.token ?? '')

const loading     = ref(true)
const invite      = ref(null)
const errorMsg    = ref('')
const accepted    = ref(false)
const accepting   = ref(false)
const acceptError = ref('')

const emailMismatch = computed(() =>
  auth.isAuthenticated &&
  invite.value?.email &&
  auth.user?.email?.toLowerCase() !== invite.value.email.toLowerCase()
)

const loginPath    = computed(() => `/login?invite=${token.value}`)
const registerPath = computed(() => `/register?invite=${token.value}`)

onMounted(async () => {
  if (!token.value) {
    errorMsg.value = 'No invite token provided'
    loading.value  = false
    return
  }
  try {
    const { data } = await authAPI.previewInvite(token.value)
    invite.value = data.invite
  } catch (err) {
    errorMsg.value = err.response?.data?.error ?? 'Invite not found or has expired'
  } finally {
    loading.value = false
  }

  // If the user is already logged in with the matching email, auto-accept silently
  // (they might have just registered / logged in via the invite link)
  if (
    auth.isAuthenticated &&
    invite.value &&
    auth.user?.email?.toLowerCase() === invite.value.email.toLowerCase()
  ) {
    await handleAccept()
  }
})

async function handleAccept() {
  accepting.value   = true
  acceptError.value = ''
  try {
    await authAPI.acceptInvite(token.value)
    accepted.value = true
  } catch (err) {
    acceptError.value = err.response?.data?.error ?? 'Failed to accept invite. Please try again.'
  } finally {
    accepting.value = false
  }
}

function handleLogout() {
  auth.logout()
  router.push(loginPath.value)
}
</script>

<style scoped>
.invite-page {
  min-height: 100vh;
  background: var(--bg, #f1f3f4);
  display: flex; align-items: center; justify-content: center;
  padding: 24px 16px;
}

.invite-card {
  background: #fff;
  border: 1px solid var(--border, #dadce0);
  border-radius: 16px;
  box-shadow: 0 4px 24px rgba(0,0,0,0.08);
  padding: 48px 40px;
  max-width: 480px; width: 100%;
  text-align: center;
}

.invite-brand {
  display: flex; align-items: center; justify-content: center; gap: 10px;
  margin-bottom: 32px;
  font-size: 18px; font-weight: 700; color: var(--text, #202124);
}
.invite-brand-icon {
  height: 26px;
  width: auto;
}

/* ── States ─────────────────────────────────────────────── */
.invite-state {
  display: flex; flex-direction: column; align-items: center; gap: 12px;
}
.invite-state p { font-size: 14px; color: var(--text-muted, #5f6368); }
.invite-state h2 { font-size: 1.25rem; font-weight: 700; color: var(--text, #202124); }
.invite-state--error h2 { color: var(--error, #d93025); }
.invite-state--success h2 { color: var(--success, #1e8e3e); }

/* ── Icon circle ───────────────────────────────────────── */
.invite-icon {
  width: 64px; height: 64px; border-radius: 50%;
  display: flex; align-items: center; justify-content: center;
  margin-bottom: 4px;
}
.invite-icon--info    { background: var(--blue-light, #e8f0fe);  color: var(--blue, #1a73e8); }
.invite-icon--success { background: var(--success-bg, #e6f4ea);  color: var(--success, #1e8e3e); }
.invite-icon--error   { background: var(--error-bg, #fce8e6);    color: var(--error, #d93025); }

/* ── Content ────────────────────────────────────────────── */
.invite-content {
  display: flex; flex-direction: column; align-items: center; gap: 12px;
}
.invite-content h2 {
  font-size: 1.25rem; font-weight: 700; color: var(--text, #202124); line-height: 1.35;
}
.invite-content h2 em { font-style: normal; color: var(--blue, #1a73e8); }
.invite-sub  { font-size: 14px; color: var(--text-muted, #5f6368); line-height: 1.6; }

.invite-email-chip {
  display: inline-flex; align-items: center; gap: 6px;
  padding: 5px 14px; border-radius: 20px;
  background: var(--bg, #f1f3f4); font-size: 13px;
  color: var(--text, #202124); font-weight: 500;
}

.invite-prompt {
  font-size: 13.5px; color: var(--text-muted, #5f6368); margin-top: 4px;
}

/* ── Buttons ────────────────────────────────────────────── */
.invite-btn {
  display: inline-flex; align-items: center; justify-content: center;
  height: 44px; padding: 0 28px; border-radius: 10px;
  background: var(--blue, #1a73e8); color: #fff;
  font-size: 14.5px; font-weight: 600; text-decoration: none;
  border: none; cursor: pointer; transition: background .15s;
  margin-top: 6px;
}
.invite-btn:hover:not(:disabled) { background: var(--blue-dark, #1557b0); text-decoration: none; }
.invite-btn:disabled { opacity: .55; cursor: not-allowed; }

.invite-btn-outline {
  display: inline-flex; align-items: center; justify-content: center;
  height: 44px; padding: 0 28px; border-radius: 10px;
  background: #fff; color: var(--blue, #1a73e8);
  border: 1.5px solid var(--blue, #1a73e8);
  font-size: 14.5px; font-weight: 600; text-decoration: none;
  cursor: pointer; transition: background .15s;
  margin-top: 6px;
}
.invite-btn-outline:hover { background: var(--blue-light, #e8f0fe); text-decoration: none; }

.invite-btn-ghost {
  background: none; border: none; cursor: pointer; color: var(--blue, #1a73e8);
  font-size: 13.5px; font-weight: 600; padding: 4px 8px;
  text-decoration: underline; margin-top: 4px;
}

.invite-actions {
  display: flex; gap: 12px; flex-wrap: wrap; justify-content: center; margin-top: 4px;
}

.invite-alert {
  width: 100%; padding: 12px 16px; border-radius: 8px;
  background: var(--error-bg, #fce8e6); color: var(--error, #d93025);
  font-size: 13.5px; line-height: 1.5; text-align: left;
  display: flex; flex-direction: column; gap: 8px;
}

/* ── Spinner ────────────────────────────────────────────── */
.invite-spinner {
  width: 40px; height: 40px; border-radius: 50%;
  border: 3px solid var(--border, #dadce0);
  border-top-color: var(--blue, #1a73e8);
  animation: spin .8s linear infinite;
}
@keyframes spin { to { transform: rotate(360deg); } }
</style>
