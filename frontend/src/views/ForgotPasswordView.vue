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

      <div class="form-actions" style="justify-content: flex-start; margin-top: 24px;">
        <router-link to="/login" class="link-btn">← Back to sign in</router-link>
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

const email    = ref('')
const loading  = ref(false)
const error    = ref('')
const success  = ref(false)
const sentTo   = ref('')

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
</style>
