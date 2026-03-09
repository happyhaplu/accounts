<template>
  <AuthLayout>

    <!-- ── Step: success ──────────────────────────────── -->
    <template v-if="step === 'success'">
      <div class="form-header">
        <div class="check-icon">
          <svg width="28" height="28" viewBox="0 0 24 24" fill="none"
               stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round">
            <polyline points="20 6 9 17 4 12"/>
          </svg>
        </div>
        <h1>Password updated!</h1>
        <p>Your password has been changed successfully. You can now sign in.</p>
      </div>
      <div class="form-actions" style="margin-top:24px; justify-content:flex-end">
        <router-link to="/login" class="btn-primary" style="text-decoration:none; display:inline-flex; align-items:center; gap:6px">
          Sign in
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none"
               stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round">
            <line x1="5" y1="12" x2="19" y2="12"/><polyline points="12 5 19 12 12 19"/>
          </svg>
        </router-link>
      </div>
    </template>

    <!-- ── Step: new password ─────────────────────────── -->
    <template v-else-if="step === 'password'">
      <div class="form-header">
        <h1>Set new password</h1>
        <p>Choose a strong password for <strong>{{ sentTo }}</strong>.</p>
      </div>

      <div v-if="pwError" class="alert alert-error">
        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor"
             stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
          <circle cx="12" cy="12" r="10"/><line x1="12" y1="8" x2="12" y2="12"/>
          <line x1="12" y1="16" x2="12.01" y2="16"/>
        </svg>
        {{ pwError }}
      </div>

      <form @submit.prevent="submitReset" novalidate>
        <div class="field">
          <label for="password">New password</label>
          <div class="input-wrap">
            <input id="password" v-model="pwForm.password"
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
        <div class="field">
          <label for="confirm">Confirm new password</label>
          <div class="input-wrap">
            <input id="confirm" v-model="pwForm.confirm"
                   :type="showPw ? 'text' : 'password'"
                   placeholder="••••••••" autocomplete="new-password" required
                   :class="{ 'is-error': pwForm.confirm && pwForm.password !== pwForm.confirm }" />
          </div>
          <div v-if="pwForm.confirm && pwForm.password !== pwForm.confirm" class="field-err">
            Passwords do not match
          </div>
        </div>

        <div class="form-actions">
          <router-link to="/login" class="link-btn">Back to sign in</router-link>
          <button type="submit" class="btn-primary"
                  :disabled="pwLoading || pwForm.password.length < 8 || pwForm.password !== pwForm.confirm">
            <svg v-if="pwLoading" width="15" height="15" viewBox="0 0 24 24" fill="none"
                 stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round" class="spin">
              <path d="M21 12a9 9 0 1 1-6.219-8.56"/>
            </svg>
            <span>{{ pwLoading ? 'Saving…' : 'Set password' }}</span>
          </button>
        </div>
      </form>
    </template>

    <!-- ── Step: OTP input ───────────────────────────── -->
    <template v-else-if="step === 'otp'">
      <div class="form-header">
        <div class="otp-icon">
          <svg width="28" height="28" viewBox="0 0 24 24" fill="none" stroke="currentColor"
               stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <rect x="2" y="7" width="20" height="14" rx="2" ry="2"/>
            <path d="M16 21V5a2 2 0 0 0-2-2h-4a2 2 0 0 0-2 2v16"/>
          </svg>
        </div>
        <h1>Enter reset code</h1>
        <p>We sent a 6-digit code to <strong>{{ sentTo }}</strong>.<br>It expires in 10 minutes.</p>
      </div>

      <div v-if="otpError" class="alert alert-error">
        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor"
             stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
          <circle cx="12" cy="12" r="10"/><line x1="12" y1="8" x2="12" y2="12"/>
          <line x1="12" y1="16" x2="12.01" y2="16"/>
        </svg>
        {{ otpError }}
      </div>

      <div class="otp-inputs">
        <input
          v-for="i in 6" :key="i"
          :ref="el => { otpRefs[i-1] = el }"
          type="text" inputmode="numeric" pattern="[0-9]*"
          maxlength="1"
          class="otp-box"
          :class="{ filled: otpDigits[i-1] }"
          :value="otpDigits[i-1]"
          @input="onOtpInput(i-1, $event)"
          @keydown="onOtpKeydown(i-1, $event)"
          @paste.prevent="onOtpPaste($event)"
        />
      </div>

      <div class="form-actions" style="flex-direction:column; align-items:stretch; gap:12px">
        <button class="btn-primary" :disabled="otpValue.length < 6 || otpLoading" @click="verifyOTP">
          <svg v-if="otpLoading" width="15" height="15" viewBox="0 0 24 24" fill="none"
               stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round" class="spin">
            <path d="M21 12a9 9 0 1 1-6.219-8.56"/>
          </svg>
          <span>{{ otpLoading ? 'Verifying…' : 'Verify code' }}</span>
        </button>
        <div style="display:flex; align-items:center; justify-content:space-between; flex-wrap:wrap; gap:8px">
          <button class="link-btn" @click="step = 'email'">← Change email</button>
          <button class="link-btn" :disabled="resendLoading" @click="resendOTP">
            {{ resendLoading ? 'Sending…' : "Didn't get it? Resend" }}
          </button>
        </div>
        <div v-if="resendMsg" class="alert" :class="resendMsg.type === 'ok' ? 'alert-success' : 'alert-error'">
          {{ resendMsg.text }}
        </div>
      </div>
    </template>

    <!-- ── Step: email form ───────────────────────────── -->
    <template v-else>
      <div class="form-header">
        <h1>Forgot password?</h1>
        <p>Enter your email and we'll send you a reset code.</p>
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
            <svg v-if="loading" width="15" height="15" viewBox="0 0 24 24" fill="none"
                 stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round" class="spin">
              <path d="M21 12a9 9 0 1 1-6.219-8.56"/>
            </svg>
            <span>{{ loading ? 'Sending…' : 'Send reset code' }}</span>
          </button>
        </div>
      </form>
    </template>

  </AuthLayout>
</template>

<script setup>
import { ref, computed, nextTick } from 'vue'
import { authAPI } from '../services/api'
import AuthLayout from '../layouts/AuthLayout.vue'

// 'email' | 'otp' | 'password' | 'success'
const step  = ref('email')
const email = ref('')
const sentTo = ref('')
const error  = ref('')
const loading = ref(false)

// OTP
const otpDigits    = ref(['', '', '', '', '', ''])
const otpRefs      = ref([])
const otpLoading   = ref(false)
const otpError     = ref('')
const resendLoading = ref(false)
const resendMsg     = ref(null)
const otpValue = computed(() => otpDigits.value.join(''))

// Password reset
const resetToken = ref('')
const pwForm  = ref({ password: '', confirm: '' })
const showPw  = ref(false)
const pwLoading = ref(false)
const pwError   = ref('')

// ── Email step ────────────────────────────────────────────────────────────────
async function submit() {
  loading.value = true
  error.value   = ''
  try {
    await authAPI.forgotPassword({ email: email.value })
    sentTo.value = email.value
    otpDigits.value = ['', '', '', '', '', '']
    step.value = 'otp'
    nextTick(() => otpRefs.value[0]?.focus())
  } catch (err) {
    error.value = err.response?.data?.error ?? 'Something went wrong. Please try again.'
  } finally {
    loading.value = false
  }
}

// ── OTP step ──────────────────────────────────────────────────────────────────
function onOtpInput(idx, e) {
  const val = e.target.value.replace(/\D/g, '')
  otpDigits.value[idx] = val.slice(-1)
  if (val && idx < 5) nextTick(() => otpRefs.value[idx + 1]?.focus())
}
function onOtpKeydown(idx, e) {
  if (e.key === 'Backspace' && !otpDigits.value[idx] && idx > 0) {
    nextTick(() => otpRefs.value[idx - 1]?.focus())
  }
}
function onOtpPaste(e) {
  const text = (e.clipboardData || window.clipboardData).getData('text')
  const digits = text.replace(/\D/g, '').slice(0, 6).split('')
  digits.forEach((d, i) => { otpDigits.value[i] = d })
  nextTick(() => otpRefs.value[Math.min(digits.length, 5)]?.focus())
}
function resetOtpBoxes() {
  otpDigits.value = ['', '', '', '', '', '']
  nextTick(() => otpRefs.value[0]?.focus())
}

async function verifyOTP() {
  otpError.value   = ''
  otpLoading.value = true
  try {
    const { data } = await authAPI.verifyResetOTP({
      email: sentTo.value,
      otp:   otpValue.value,
    })
    resetToken.value = data.reset_token
    step.value = 'password'
  } catch (err) {
    otpError.value = err.response?.data?.error ?? 'Invalid code. Please try again.'
    resetOtpBoxes()
  } finally {
    otpLoading.value = false
  }
}

async function resendOTP() {
  resendLoading.value = true
  resendMsg.value     = null
  try {
    await authAPI.forgotPassword({ email: sentTo.value })
    resendMsg.value = { type: 'ok', text: 'New code sent — check your inbox.' }
    resetOtpBoxes()
  } catch {
    resendMsg.value = { type: 'err', text: 'Failed to send. Please try again.' }
  } finally {
    resendLoading.value = false
    setTimeout(() => { resendMsg.value = null }, 6000)
  }
}

// ── Password step ─────────────────────────────────────────────────────────────
async function submitReset() {
  pwError.value   = ''
  pwLoading.value = true
  try {
    await authAPI.resetPassword({
      token:        resetToken.value,
      new_password: pwForm.value.password,
    })
    step.value = 'success'
  } catch (err) {
    pwError.value = err.response?.data?.error ?? 'Failed to reset password. The code may have expired.'
  } finally {
    pwLoading.value = false
  }
}
</script>

<style scoped>
.check-icon {
  width: 56px; height: 56px; border-radius: 50%;
  background: #e8f5e9; color: #2e7d32;
  display: flex; align-items: center; justify-content: center;
  margin-bottom: 16px;
}
.otp-icon {
  width: 56px; height: 56px; border-radius: 50%;
  background: var(--blue-light, #e8f0fe); color: var(--blue, #1a73e8);
  display: flex; align-items: center; justify-content: center;
  margin-bottom: 16px;
}
.otp-inputs { display: flex; gap: 10px; justify-content: center; margin: 28px 0; }
.otp-box {
  width: 48px; height: 58px;
  border: 2px solid var(--border, #dadce0); border-radius: 10px;
  font-size: 24px; font-weight: 700; text-align: center;
  outline: none; background: #fff; color: var(--text, #202124);
  transition: border-color .15s, box-shadow .15s; caret-color: transparent;
}
.otp-box:focus { border-color: var(--blue, #1a73e8); box-shadow: 0 0 0 3px var(--blue-light, #e8f0fe); }
.otp-box.filled { border-color: var(--blue, #1a73e8); background: var(--blue-light, #e8f0fe); }
.alert {
  display: flex; align-items: center; gap: 8px;
  padding: 10px 14px; border-radius: 8px; font-size: 0.84rem; font-weight: 500;
}
.alert-error   { background: #fce8e6; color: #c5221f; }
.alert-success { background: #e6f4ea; color: #1e8e3e; }
input.is-error { border-color: #d93025 !important; }
.field-err { font-size: 12px; color: #d93025; margin-top: 4px; }
@keyframes spin { to { transform: rotate(360deg); } }
.spin { animation: spin .8s linear infinite; flex-shrink: 0; }
</style>
