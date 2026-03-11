<template>
  <AuthLayout>
    <div class="form-header">
      <h1>Sign in</h1>
      <p>to continue to <strong>{{ redirectHost || 'Outcraftly Accounts' }}</strong></p>
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
          <input id="email" v-model="form.email" type="email"
                 placeholder="you@example.com" autocomplete="email" required />
        </div>
      </div>

      <div class="field">
        <label for="password">Password</label>
        <div class="input-wrap">
          <input id="password" v-model="form.password"
                 :type="showPw ? 'text' : 'password'"
                 placeholder="••••••••" autocomplete="current-password" required />
          <button type="button" class="pw-toggle" @click="showPw = !showPw"
                  :aria-label="showPw ? 'Hide password' : 'Show password'">
            <!-- eye-off -->
            <svg v-if="showPw" width="18" height="18" viewBox="0 0 24 24" fill="none"
                 stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <path d="M17.94 17.94A10.07 10.07 0 0 1 12 20c-7 0-11-8-11-8a18.45 18.45 0 0 1 5.06-5.94"/>
              <path d="M9.9 4.24A9.12 9.12 0 0 1 12 4c7 0 11 8 11 8a18.5 18.5 0 0 1-2.16 3.19"/>
              <path d="M14.12 14.12a3 3 0 1 1-4.24-4.24"/>
              <line x1="1" y1="1" x2="23" y2="23"/>
            </svg>
            <!-- eye-open -->
            <svg v-else width="18" height="18" viewBox="0 0 24 24" fill="none"
                 stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z"/>
              <circle cx="12" cy="12" r="3"/>
            </svg>
          </button>
        </div>
        <router-link to="/forgot-password" class="link-subtle">Forgot password?</router-link>
      </div>

      <div class="form-actions">
        <router-link to="/register" class="link-btn">Create account</router-link>
        <button type="submit" class="btn-primary" :disabled="loading">
          <span>{{ loading ? 'Signing in…' : 'Sign in' }}</span>
          <svg v-if="!loading" width="16" height="16" viewBox="0 0 24 24" fill="none"
               stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round">
            <line x1="5" y1="12" x2="19" y2="12"/><polyline points="12 5 19 12 12 19"/>
          </svg>
        </button>
      </div>
    </form>
  </AuthLayout>
</template>

<script setup>
import { ref } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import { authAPI } from '../services/api'
import AuthLayout from '../layouts/AuthLayout.vue'

const router = useRouter()
const route  = useRoute()
const auth   = useAuthStore()
const form   = ref({ email: '', password: '' })
const showPw  = ref(false)
const loading = ref(false)
const error   = ref('')

// Preserve redirect_uri from URL query (e.g. /login?redirect_uri=https://warmup.outcraftly.com/callback)
const redirectUri = route.query.redirect_uri ?? ''

// Extract a human-readable hostname for the "Sign in to continue to X" subtitle.
// e.g. "https://warmup.outcraftly.com/callback" → "warmup.outcraftly.com"
const redirectHost = redirectUri
  ? (() => { try { return new URL(redirectUri).hostname } catch { return '' } })()
  : ''

async function submit() {
  loading.value = true
  error.value   = ''
  try {
    const { data } = await authAPI.login({ ...form.value, redirect_uri: redirectUri })
    auth.setAuth(data.token, data.user)
    if (data.redirect_url) {
      // Backend signed a launch JWT and built the full callback URL for us.
      window.location.href = data.redirect_url
    } else if (data.needs_profile_setup) {
      router.push('/profile-setup')
    } else {
      router.push('/dashboard')
    }
  } catch (err) {
    error.value = err.response?.data?.error ?? 'Login failed. Please try again.'
  } finally {
    loading.value = false
  }
}
</script>
