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

  return { token, user, isAuthenticated, setAuth, updateUser, logout }
})
