<template>
  <div class="logout-page">
    <div class="logout-card">
      <div class="logout-icon">
        <img src="/icon.svg" alt="Gour" class="logout-logo" />
      </div>
      <h2>Signing you out…</h2>
      <p class="logout-sub">You have been signed out of your Gour account.</p>
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
  display: flex; align-items: center; justify-content: center;
}
.logout-logo {
  height: 48px;
  width: auto;
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
