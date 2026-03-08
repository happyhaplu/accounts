import { defineStore } from 'pinia'
import { ref, computed } from 'vue'

export const useAuthStore = defineStore('auth', () => {
  const token = ref(localStorage.getItem('oc_token') ?? null)
  const user  = ref(JSON.parse(localStorage.getItem('oc_user') ?? 'null'))

  const isAuthenticated = computed(() => !!token.value)

  /** Called after a successful login or register response. */
  function setAuth(newToken, newUser) {
    token.value = newToken
    user.value  = newUser
    localStorage.setItem('oc_token', newToken)
    localStorage.setItem('oc_user',  JSON.stringify(newUser))
  }

  /** Update stored user data (e.g. after profile setup). */
  function updateUser(newUser) {
    user.value = newUser
    localStorage.setItem('oc_user', JSON.stringify(newUser))
  }

  function logout() {
    token.value = null
    user.value  = null
    localStorage.removeItem('oc_token')
    localStorage.removeItem('oc_user')
  }

  // Cross-tab sync: when another tab writes oc_user / oc_token (e.g. after
  // email verification), update this tab's reactive state immediately so any
  // component bound to auth.user re-renders without a page refresh.
  window.addEventListener('storage', (e) => {
    if (e.key === 'oc_user') {
      user.value = e.newValue ? JSON.parse(e.newValue) : null
    }
    if (e.key === 'oc_token') {
      token.value = e.newValue ?? null
    }
  })

  return { token, user, isAuthenticated, setAuth, updateUser, logout }
})
