import { defineStore } from 'pinia'
import { ref, computed } from 'vue'

export const useAdminStore = defineStore('admin', () => {
  const secret = ref(sessionStorage.getItem('admin_secret') ?? null)
  const email  = ref(sessionStorage.getItem('admin_email')  ?? null)

  const isAuthenticated = computed(() => !!secret.value)

  function setAuth(adminSecret, adminEmail) {
    secret.value = adminSecret
    email.value  = adminEmail
    sessionStorage.setItem('admin_secret', adminSecret)
    sessionStorage.setItem('admin_email',  adminEmail)
  }

  function logout() {
    secret.value = null
    email.value  = null
    sessionStorage.removeItem('admin_secret')
    sessionStorage.removeItem('admin_email')
  }

  return { secret, email, isAuthenticated, setAuth, logout }
})
