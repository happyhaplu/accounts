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
        <h1>Password updated</h1>
        <p>Your password has been changed successfully.</p>
      </div>
      <div class="form-actions" style="justify-content: flex-end; margin-top: 24px;">
        <router-link to="/login" class="btn-primary" style="text-decoration:none; display:inline-flex; align-items:center; gap:6px;">
          Sign in
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none"
               stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round">
            <line x1="5" y1="12" x2="19" y2="12"/><polyline points="12 5 19 12 12 19"/>
          </svg>
        </router-link>
      </div>
    </template>

    <!-- Form state -->
    <template v-else>
      <div class="form-header">
        <h1>Set new password</h1>
        <p>Enter and confirm your new password below.</p>
      </div>

      <div v-if="error" class="alert alert-error">
        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor"
             stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
          <circle cx="12" cy="12" r="10"/><line x1="12" y1="8" x2="12" y2="12"/>
          <line x1="12" y1="16" x2="12.01" y2="16"/>
        </svg>
        {{ error }}
      </div>

      <form v-if="token" @submit.prevent="submit" novalidate>
        <!-- New password -->
        <div class="field">
          <label for="password">New password</label>
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
        </div>

        <!-- Confirm password -->
        <div class="field">
          <label for="confirm">Confirm new password</label>
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
          <router-link to="/login" class="link-btn">Back to sign in</router-link>
          <button type="submit" class="btn-primary" :disabled="loading || !canSubmit">
            <span>{{ loading ? 'Resetting…' : 'Reset password' }}</span>
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
import { ref, computed, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { authAPI } from '../services/api'
import AuthLayout from '../layouts/AuthLayout.vue'

const route = useRoute()

const form    = ref({ password: '', confirm: '' })
const token   = ref('')
const showPw  = ref(false)
const loading = ref(false)
const error   = ref('')
const success = ref(false)

const canSubmit = computed(() =>
  form.value.password.length >= 8 &&
  form.value.password === form.value.confirm
)

onMounted(() => {
  token.value = route.query.token ?? ''
  if (!token.value) error.value = 'Invalid or missing reset token. Please request a new one.'
})

async function submit() {
  if (!canSubmit.value) return
  loading.value = true
  error.value   = ''
  try {
    await authAPI.resetPassword({
      token: token.value,
      new_password: form.value.password,
    })
    success.value = true
  } catch (err) {
    error.value = err.response?.data?.error ?? 'Failed to reset password. The link may have expired.'
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
input.is-error { border-color: #d93025 !important; }
.field-err { font-size: 12px; color: #d93025; margin-top: 4px; }
</style>
