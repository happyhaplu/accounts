<template>
  <div class="admin-login-page">
    <div class="admin-login-card">

      <!-- Brand header -->
      <div class="alc-header">
        <img src="/logo.svg" alt="Gour" class="alc-logo" />
        <span class="alc-badge">Admin Panel</span>
      </div>

      <h1 class="alc-title">Sign in to Admin</h1>
      <p class="alc-sub">Restricted access — authorised personnel only.</p>

      <!-- Error -->
      <div v-if="error" class="alert alert-error">
        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor"
             stroke-width="2" stroke-linecap="round" stroke-linejoin="round" style="flex-shrink:0;margin-top:1px">
          <circle cx="12" cy="12" r="10"/><line x1="12" y1="8" x2="12" y2="12"/>
          <line x1="12" y1="16" x2="12.01" y2="16"/>
        </svg>
        {{ error }}
      </div>

      <!-- Form -->
      <form @submit.prevent="handleLogin" novalidate>

        <div class="field">
          <label for="email">Email address</label>
          <input
            id="email"
            v-model="form.email"
            type="email"
            autocomplete="email"
            placeholder="admin@example.com"
            :class="{ 'is-error': errors.email }"
            :disabled="loading"
          />
          <p v-if="errors.email" class="field-err">{{ errors.email }}</p>
        </div>

        <div class="field">
          <label for="password">Password</label>
          <div class="input-wrap">
            <input
              id="password"
              v-model="form.password"
              :type="showPw ? 'text' : 'password'"
              autocomplete="current-password"
              placeholder="••••••••"
              :class="{ 'is-error': errors.password }"
              :disabled="loading"
            />
            <button type="button" class="pw-toggle" @click="showPw = !showPw" tabindex="-1">
              <svg v-if="!showPw" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z"/><circle cx="12" cy="12" r="3"/>
              </svg>
              <svg v-else width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <path d="M17.94 17.94A10.07 10.07 0 0 1 12 20c-7 0-11-8-11-8a18.45 18.45 0 0 1 5.06-5.94M9.9 4.24A9.12 9.12 0 0 1 12 4c7 0 11 8 11 8a18.5 18.5 0 0 1-2.16 3.19m-6.72-1.07a3 3 0 1 1-4.24-4.24"/>
                <line x1="1" y1="1" x2="23" y2="23"/>
              </svg>
            </button>
          </div>
          <p v-if="errors.password" class="field-err">{{ errors.password }}</p>
        </div>

        <button type="submit" class="btn-primary btn-full" :disabled="loading">
          <svg v-if="loading" class="spin-icon" width="16" height="16" viewBox="0 0 24 24" fill="none"
               stroke="currentColor" stroke-width="2.5" stroke-linecap="round">
            <path d="M12 2v4M12 18v4M4.93 4.93l2.83 2.83M16.24 16.24l2.83 2.83M2 12h4M18 12h4M4.93 19.07l2.83-2.83M16.24 7.76l2.83-2.83"/>
          </svg>
          {{ loading ? 'Signing in…' : 'Sign in to Admin' }}
        </button>

      </form>

    </div>
  </div>
</template>

<script setup>
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { useAdminStore } from '../../stores/admin'
import { adminAPI } from '../../services/api'

const router = useRouter()
const admin  = useAdminStore()

const loading = ref(false)
const error   = ref('')
const showPw  = ref(false)

const form = reactive({ email: '', password: '' })
const errors = reactive({ email: '', password: '' })

function validate() {
  errors.email    = ''
  errors.password = ''
  let ok = true
  if (!form.email.trim() || !form.email.includes('@')) {
    errors.email = 'Valid email address is required'
    ok = false
  }
  if (!form.password) {
    errors.password = 'Password is required'
    ok = false
  }
  return ok
}

async function handleLogin() {
  error.value = ''
  if (!validate()) return

  loading.value = true
  try {
    const { data } = await adminAPI.login({
      email:    form.email.trim().toLowerCase(),
      password: form.password,
    })
    admin.setAuth(data.admin_secret, form.email.trim().toLowerCase())
    router.push('/admin/products')
  } catch (err) {
    const status = err.response?.status
    if (status === 401) {
      error.value = 'Invalid email or password. Please try again.'
    } else if (status === 503) {
      error.value = 'Admin access is not configured on this server.'
    } else {
      error.value = err.response?.data?.error ?? 'Something went wrong. Please try again.'
    }
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.admin-login-page {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #f1f3f4;
  padding: 24px;
}

.admin-login-card {
  background: #fff;
  border-radius: 12px;
  box-shadow: 0 2px 20px rgba(0,0,0,0.09);
  padding: 40px 44px;
  width: 100%;
  max-width: 420px;
}

/* Brand header */
.alc-header {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 28px;
}
.alc-logo {
  height: 28px;
  width: auto;
}
.alc-badge {
  font-size: 11px;
  font-weight: 700;
  color: #1a2332;
  background: #e8f0fe;
  border: 1px solid #c5d8fb;
  padding: 3px 9px;
  border-radius: 20px;
  letter-spacing: 0.05em;
  text-transform: uppercase;
}

.alc-title {
  font-size: 24px;
  font-weight: 600;
  color: #202124;
  letter-spacing: -0.3px;
  margin-bottom: 6px;
}
.alc-sub {
  font-size: 14px;
  color: #5f6368;
  margin-bottom: 28px;
}

/* Field */
.field { margin-bottom: 20px; }
.field label {
  display: block;
  font-size: 13px;
  font-weight: 500;
  color: #3c4043;
  margin-bottom: 6px;
}
.field input {
  width: 100%;
  height: 46px;
  padding: 0 44px 0 13px;
  border: 1.5px solid #dadce0;
  border-radius: 6px;
  font-size: 14.5px;
  color: #202124;
  outline: none;
  font-family: inherit;
  transition: border-color 0.15s, box-shadow 0.15s;
  background: #fff;
}
.field input:focus {
  border-color: #1a73e8;
  box-shadow: 0 0 0 3px rgba(26,115,232,0.1);
}
.field input.is-error {
  border-color: #d93025;
}
.field input:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}
.field-err { font-size: 12px; color: #d93025; margin-top: 5px; }

.input-wrap { position: relative; }
.pw-toggle {
  position: absolute;
  right: 12px; top: 50%;
  transform: translateY(-50%);
  background: none; border: none;
  cursor: pointer; color: #5f6368;
  padding: 4px;
  display: flex; align-items: center;
  border-radius: 4px;
  transition: color 0.15s;
}
.pw-toggle:hover { color: #1a73e8; }

/* Alert */
.alert {
  display: flex;
  align-items: flex-start;
  gap: 10px;
  padding: 11px 14px;
  border-radius: 6px;
  font-size: 13.5px;
  line-height: 1.5;
  margin-bottom: 20px;
}
.alert-error {
  background: #fce8e6;
  color: #c5221f;
  border-left: 3px solid #d93025;
}

/* Buttons */
.btn-primary {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  background: #1a73e8;
  color: #fff;
  border: none;
  border-radius: 6px;
  height: 44px;
  padding: 0 22px;
  font-size: 14.5px;
  font-weight: 500;
  cursor: pointer;
  font-family: inherit;
  transition: background 0.18s, box-shadow 0.18s;
  white-space: nowrap;
}
.btn-primary:hover:not(:disabled) {
  background: #1557b0;
  box-shadow: 0 2px 8px rgba(26,115,232,0.35);
}
.btn-primary:disabled { opacity: 0.6; cursor: not-allowed; }
.btn-full { width: 100%; margin-top: 8px; }

/* Spinner */
.spin-icon { animation: spin 1s linear infinite; }
@keyframes spin { to { transform: rotate(360deg); } }
</style>
