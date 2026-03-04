<template>
  <AuthLayout>

    <!-- Loading -->
    <template v-if="status === 'loading'">
      <div class="form-header centered">
        <div class="spinner"></div>
        <h1>Verifying your email…</h1>
        <p>Just a moment, please.</p>
      </div>
    </template>

    <!-- Success -->
    <template v-else-if="status === 'success'">
      <div class="form-header">
        <div class="check-icon">
          <svg width="28" height="28" viewBox="0 0 24 24" fill="none" stroke="currentColor"
               stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round">
            <polyline points="20 6 9 17 4 12"/>
          </svg>
        </div>
        <h1>Email verified!</h1>
        <p>Your account is active. Let's set up your profile.</p>
      </div>
      <div class="form-actions" style="margin-top:24px; justify-content: flex-end;">
        <button class="btn-primary" @click="proceed">
          <span>Continue</span>
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor"
               stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round">
            <line x1="5" y1="12" x2="19" y2="12"/><polyline points="12 5 19 12 12 19"/>
          </svg>
        </button>
      </div>
    </template>

    <!-- Error -->
    <template v-else>
      <div class="form-header">
        <div class="error-icon">
          <svg width="28" height="28" viewBox="0 0 24 24" fill="none" stroke="currentColor"
               stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round">
            <circle cx="12" cy="12" r="10"/>
            <line x1="12" y1="8" x2="12" y2="12"/>
            <line x1="12" y1="16" x2="12.01" y2="16"/>
          </svg>
        </div>
        <h1>Verification failed</h1>
        <p>{{ errorMsg }}</p>
      </div>
      <div class="form-actions" style="margin-top:24px; justify-content: flex-start;">
        <router-link to="/register" class="link-btn">Create a new account</router-link>
      </div>
    </template>

  </AuthLayout>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import { authAPI } from '../services/api'
import AuthLayout from '../layouts/AuthLayout.vue'

const router   = useRouter()
const route    = useRoute()
const auth     = useAuthStore()
const status   = ref('loading')  // 'loading' | 'success' | 'error'
const errorMsg = ref('')
const needsSetup = ref(false)

onMounted(async () => {
  const token = route.query.token
  if (!token) {
    errorMsg.value = 'No verification token found in the link.'
    status.value = 'error'
    return
  }
  try {
    const { data } = await authAPI.verifyEmail(token)
    auth.setAuth(data.token, data.user)
    needsSetup.value = data.needs_profile_setup
    status.value = 'success'
    // Auto-proceed after 1.5s
    setTimeout(proceed, 1500)
  } catch (err) {
    errorMsg.value = err.response?.data?.error ?? 'Invalid or expired verification link.'
    status.value = 'error'
  }
})

function proceed() {
  router.push(needsSetup.value ? '/profile-setup' : '/dashboard')
}
</script>

<style scoped>
.centered { text-align: center; }

.spinner {
  width: 44px; height: 44px; border-radius: 50%;
  border: 3px solid var(--blue-light, #e8f0fe);
  border-top-color: var(--blue, #1a73e8);
  animation: spin 0.7s linear infinite;
  margin: 0 auto 16px;
}
@keyframes spin { to { transform: rotate(360deg); } }

.check-icon {
  width: 56px; height: 56px; border-radius: 50%;
  background: #e8f5e9; color: #2e7d32;
  display: flex; align-items: center; justify-content: center;
  margin-bottom: 16px;
}
.error-icon {
  width: 56px; height: 56px; border-radius: 50%;
  background: #fce8e6; color: #d93025;
  display: flex; align-items: center; justify-content: center;
  margin-bottom: 16px;
}
</style>
