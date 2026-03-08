<template>
  <AuthLayout>

    <!-- ── Success: check inbox ───────────────────────── -->
    <template v-if="registered">
      <div class="form-header">
        <div class="inbox-icon">
          <svg width="28" height="28" viewBox="0 0 24 24" fill="none" stroke="currentColor"
               stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <path d="M4 4h16c1.1 0 2 .9 2 2v12c0 1.1-.9 2-2 2H4c-1.1 0-2-.9-2-2V6c0-1.1.9-2 2-2z"/>
            <polyline points="22,6 12,13 2,6"/>
          </svg>
        </div>
        <h1>Check your inbox</h1>
        <p>We sent a verification link to <strong>{{ registeredEmail }}</strong>.<br>Click it to activate your account.</p>
      </div>

      <div v-if="resendMsg" class="alert" :class="resendMsg.type === 'ok' ? 'alert-success' : 'alert-error'" style="margin-top:20px">
        <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor"
             stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
          <polyline v-if="resendMsg.type === 'ok'" points="20 6 9 17 4 12"/>
          <template v-else><circle cx="12" cy="12" r="10"/><line x1="12" y1="8" x2="12" y2="12"/><line x1="12" y1="16" x2="12.01" y2="16"/></template>
        </svg>
        {{ resendMsg.text }}
      </div>

      <div class="form-actions" style="margin-top:20px; flex-direction:column; align-items:flex-start; gap:12px">
        <router-link to="/login" class="link-btn">← Back to sign in</router-link>
        <p style="font-size:0.82rem; color:#5f6368; margin:0">Didn't receive it? Check spam or</p>
        <button class="btn-resend" :disabled="resendLoading" @click="resendEmail">
          <svg v-if="resendLoading" width="13" height="13" viewBox="0 0 24 24" fill="none"
               stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"
               class="spin">
            <path d="M21 12a9 9 0 1 1-6.219-8.56"/>
          </svg>
          {{ resendLoading ? 'Sending…' : 'Resend verification email' }}
        </button>
      </div>
    </template>

    <!-- ── Registration form ──────────────────────────── -->
    <template v-else>
      <div class="form-header">
        <h1>Create your account</h1>
        <p>Start with Outcraftly — it's free</p>
      </div>

      <div v-if="error" class="alert alert-error">
        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor"
             stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
          <circle cx="12" cy="12" r="10"/><line x1="12" y1="8" x2="12" y2="12"/>
          <line x1="12" y1="16" x2="12.01" y2="16"/>
        </svg>
        {{ error }}
      </div>

      <form @submit.prevent="submit" novalidate>
        <!-- Email -->
        <div class="field">
          <label for="email">Email address</label>
          <div class="input-wrap">
            <input id="email" v-model="form.email" type="email"
                   placeholder="you@example.com" autocomplete="email" required />
          </div>
        </div>

        <!-- Password -->
        <div class="field">
          <label for="password">Password</label>
          <div class="input-wrap">
            <input id="password" v-model="form.password"
                   :type="showPw ? 'text' : 'password'"
                   placeholder="Min. 8 characters" autocomplete="new-password"
                   minlength="8" required />
            <button type="button" class="pw-toggle" @click="showPw = !showPw">
              <svg v-if="showPw" width="18" height="18" viewBox="0 0 24 24" fill="none"
                   stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <path d="M17.94 17.94A10.07 10.07 0 0 1 12 20c-7 0-11-8-11-8a18.45 18.45 0 0 1 5.06-5.94"/>
                <path d="M9.9 4.24A9.12 9.12 0 0 1 12 4c7 0 11 8 11 8a18.5 18.5 0 0 1-2.16 3.19"/>
                <path d="M14.12 14.12a3 3 0 1 1-4.24-4.24"/>
                <line x1="1" y1="1" x2="23" y2="23"/>
              </svg>
              <svg v-else width="18" height="18" viewBox="0 0 24 24" fill="none"
                   stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z"/>
                <circle cx="12" cy="12" r="3"/>
              </svg>
            </button>
          </div>
          <!-- Password strength -->
          <div v-if="form.password" class="pw-strength">
            <div class="pw-bars">
              <div v-for="i in 4" :key="i" class="pw-bar"
                   :class="i <= strength.level ? `lvl-${strength.level}` : ''"></div>
            </div>
            <span class="pw-strength-lbl" :class="`s${strength.level}`">{{ strength.label }}</span>
          </div>
        </div>

        <!-- Confirm password -->
        <div class="field">
          <label for="confirm">Confirm password</label>
          <div class="input-wrap">
            <input id="confirm" v-model="form.confirm"
                   :type="showPw ? 'text' : 'password'"
                   placeholder="••••••••" autocomplete="new-password" required
                   :class="{ 'is-error': form.confirm && form.password !== form.confirm }" />
          </div>
          <div v-if="form.confirm && form.password !== form.confirm" class="field-err">
            Passwords do not match
          </div>
        </div>

        <div class="form-actions">
          <router-link to="/login" class="link-btn">Sign in instead</router-link>
          <button type="submit" class="btn-primary" :disabled="loading || !canSubmit">
            <span>{{ loading ? 'Creating…' : 'Create account' }}</span>
            <svg v-if="!loading" width="16" height="16" viewBox="0 0 24 24" fill="none"
                 stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round">
              <line x1="5" y1="12" x2="19" y2="12"/><polyline points="12 5 19 12 12 19"/>
            </svg>
          </button>
        </div>
      </form>
    </template>

  </AuthLayout>
</template>

<script setup>
import { ref, computed } from 'vue'
import { authAPI } from '../services/api'
import AuthLayout from '../layouts/AuthLayout.vue'

const form            = ref({ email: '', password: '', confirm: '' })
const showPw          = ref(false)
const loading         = ref(false)
const error           = ref('')
const registered      = ref(false)
const registeredEmail = ref('')
const resendLoading   = ref(false)
const resendMsg       = ref(null)   // { type: 'ok'|'err', text: string }

async function resendEmail() {
  resendLoading.value = true
  resendMsg.value     = null
  try {
    await authAPI.resendVerification({ email: registeredEmail.value })
    resendMsg.value = { type: 'ok', text: 'A new link has been sent — check your inbox.' }
  } catch {
    resendMsg.value = { type: 'err', text: 'Failed to send. Please try again.' }
  } finally {
    resendLoading.value = false
    setTimeout(() => { resendMsg.value = null }, 8000)
  }
}

const strength = computed(() => {
  const p = form.value.password
  if (!p) return { level: 0, label: '' }
  let s = 0
  if (p.length >= 8)  s++
  if (p.length >= 12) s++
  if (/[A-Z]/.test(p) && /[0-9]/.test(p)) s++
  if (/[^A-Za-z0-9]/.test(p)) s++
  const labels = ['', 'Weak', 'Fair', 'Good', 'Strong']
  return { level: Math.min(s, 4), label: labels[Math.min(s, 4)] }
})

const canSubmit = computed(() =>
  form.value.email && form.value.password.length >= 8 &&
  form.value.password === form.value.confirm
)

async function submit() {
  if (!canSubmit.value) return
  loading.value = true
  error.value   = ''
  try {
    await authAPI.register({
      email: form.value.email, password: form.value.password,
    })
    registeredEmail.value = form.value.email
    registered.value = true
  } catch (err) {
    error.value = err.response?.data?.error ?? 'Registration failed. Please try again.'
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.inbox-icon {
  width: 56px; height: 56px; border-radius: 50%;
  background: var(--blue-light, #e8f0fe); color: var(--blue, #1a73e8);
  display: flex; align-items: center; justify-content: center;
  margin-bottom: 16px;
}
.btn-resend {
  display: inline-flex; align-items: center; gap: 6px;
  height: 34px; padding: 0 16px; border-radius: 8px;
  font-size: 0.82rem; font-weight: 600; cursor: pointer;
  border: 1.5px solid var(--blue, #1a73e8);
  color: var(--blue, #1a73e8); background: var(--blue-light, #e8f0fe);
  transition: background .15s;
}
.btn-resend:hover:not(:disabled) { background: #d2e3fc; }
.btn-resend:disabled { opacity: 0.6; cursor: not-allowed; }
.alert {
  display: flex; align-items: center; gap: 8px;
  padding: 10px 14px; border-radius: 8px;
  font-size: 0.84rem; font-weight: 500;
}
.alert-success { background: #e6f4ea; color: #1e8e3e; }
.alert-error   { background: #fce8e6; color: #c5221f; }
@keyframes spin { to { transform: rotate(360deg); } }
.spin { animation: spin .8s linear infinite; }
</style>
