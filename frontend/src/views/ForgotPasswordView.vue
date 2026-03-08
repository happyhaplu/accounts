<template>
  <AuthLayout>
    <!-- Success state -->
    <template v-if="success">
      <div class="form-header">
        <div class="check-icon">
          <svg width="28" height="28" viewBox="0 0 24 24" fill="none"
               stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round">
            <polyline points="20 6 9 17 4 12"/>
          </svg>
        </div>
        <h1>Check your inbox</h1>
        <p>If an account exists for <strong>{{ sentTo }}</strong>, a reset link has been sent. Check your email.</p>
      </div>

      <div v-if="resendMsg" class="alert" :class="resendMsg.type === 'ok' ? 'alert-success' : 'alert-error'" style="margin-top:16px">
        <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor"
             stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
          <polyline v-if="resendMsg.type === 'ok'" points="20 6 9 17 4 12"/>
          <template v-else><circle cx="12" cy="12" r="10"/><line x1="12" y1="8" x2="12" y2="12"/><line x1="12" y1="16" x2="12.01" y2="16"/></template>
        </svg>
        {{ resendMsg.text }}
      </div>

      <div class="form-actions" style="justify-content:flex-start; margin-top:20px; flex-direction:column; align-items:flex-start; gap:12px">
        <router-link to="/login" class="link-btn">← Back to sign in</router-link>
        <p style="font-size:0.82rem; color:#5f6368; margin:0">Didn't receive it? Check spam or</p>
        <button class="btn-resend" :disabled="resendLoading" @click="resend">
          <svg v-if="resendLoading" width="13" height="13" viewBox="0 0 24 24" fill="none"
               stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"
               class="spin">
            <path d="M21 12a9 9 0 1 1-6.219-8.56"/>
          </svg>
          {{ resendLoading ? 'Sending…' : 'Resend reset email' }}
        </button>
      </div>
    </template>

    <!-- Form state -->
    <template v-else>
      <div class="form-header">
        <h1>Forgot password?</h1>
        <p>Enter your email and we'll send you a reset link.</p>
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
        <div class="field">
          <label for="email">Email address</label>
          <div class="input-wrap">
            <input id="email" v-model="email" type="email"
                   placeholder="you@example.com" autocomplete="email" required />
          </div>
        </div>

        <div class="form-actions">
          <router-link to="/login" class="link-btn">Back to sign in</router-link>
          <button type="submit" class="btn-primary" :disabled="loading || !email">
            <span>{{ loading ? 'Sending…' : 'Send reset link' }}</span>
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
import { ref } from 'vue'
import { authAPI } from '../services/api'
import AuthLayout from '../layouts/AuthLayout.vue'

const email       = ref('')
const loading     = ref(false)
const error       = ref('')
const success     = ref(false)
const sentTo      = ref('')
const resendLoading = ref(false)
const resendMsg     = ref(null)   // { type: 'ok'|'err', text: string }

async function resend() {
  resendLoading.value = true
  resendMsg.value     = null
  try {
    await authAPI.forgotPassword({ email: sentTo.value })
    resendMsg.value = { type: 'ok', text: 'A new reset link has been sent — check your inbox.' }
  } catch {
    resendMsg.value = { type: 'err', text: 'Failed to send. Please try again.' }
  } finally {
    resendLoading.value = false
    setTimeout(() => { resendMsg.value = null }, 8000)
  }
}

async function submit() {
  loading.value  = true
  error.value    = ''
  try {
    await authAPI.forgotPassword({ email: email.value })
    sentTo.value  = email.value
    success.value = true
  } catch (err) {
    error.value = err.response?.data?.error ?? 'Something went wrong. Please try again.'
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.check-icon {
  width: 56px; height: 56px;
  border-radius: 50%;
  background: #e8f5e9;
  color: #2e7d32;
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
