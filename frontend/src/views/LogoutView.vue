<template>
  <div class="logout-page">
    <div class="logout-card">
      <div class="logout-icon">
        <svg viewBox="0 0 48 48" fill="none" xmlns="http://www.w3.org/2000/svg" width="48" height="48">
          <defs>
            <linearGradient id="planeGrad" x1="4" y1="38" x2="44" y2="4" gradientUnits="userSpaceOnUse">
              <stop offset="0%" stop-color="#1d4ed8"/>
              <stop offset="100%" stop-color="#4f8ef7"/>
            </linearGradient>
          </defs>
          <path d="M4 36 L44 6 L32 44 L22 28 Z" fill="url(#planeGrad)"/>
          <path d="M22 28 L32 44 L26 30 Z" fill="#1535a8"/>
          <path d="M22 28 L4 36 L44 6 Z" fill="rgba(255,255,255,0.18)"/>
        </svg>
      </div>
      <h2>Signing you out…</h2>
      <p class="logout-sub">You have been signed out of your Outcraftly account.</p>
    </div>
  </div>
</template>

<script setup>
import { onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'

const route  = useRoute()
const router = useRouter()
const auth   = useAuthStore()

onMounted(() => {
  // Clear auth state (token + user from localStorage).
  auth.logout()

  // If a product sent ?redirect_uri=, go back there after logout.
  const redirectUri = route.query.redirect_uri
  if (redirectUri) {
    // Small delay so the user sees the confirmation.
    setTimeout(() => {
      window.location.href = redirectUri
    }, 800)
  } else {
    // No redirect — go to the login page.
    setTimeout(() => {
      router.push('/login')
    }, 800)
  }
})
</script>

<style scoped>
.logout-page {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 100vh;
  background: #f8f9fa;
}
.logout-card {
  text-align: center;
  padding: 3rem 2.5rem;
  background: #fff;
  border-radius: 12px;
  box-shadow: 0 1px 3px rgba(0,0,0,.08);
  max-width: 400px;
}
.logout-icon {
  margin-bottom: 1.5rem;
}
.logout-card h2 {
  font-size: 1.25rem;
  font-weight: 600;
  color: #202124;
  margin: 0 0 .5rem;
}
.logout-sub {
  font-size: .9rem;
  color: #5f6368;
  margin: 0;
}
</style>
