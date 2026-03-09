<template>
  <div class="settings-page">

    <!-- Page header -->
    <div class="page-header">
      <div class="page-header-inner">
        <h1>Settings</h1>
        <p>Manage your profile, security, and account preferences.</p>
      </div>
    </div>

    <div class="page-body">

      <!-- ── Profile ──────────────────────────────────────────── -->
      <section class="section">
        <div class="section-head">
          <div class="section-icon">
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none"
                 stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"/>
              <circle cx="12" cy="7" r="4"/>
            </svg>
          </div>
          <div>
            <h2>Profile</h2>
            <p class="section-sub">Update your personal and work information.</p>
          </div>
        </div>

        <div class="card">
          <div v-if="profileSuccess" class="alert alert-success">
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor"
                 stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"/><polyline points="22 4 12 14.01 9 11.01"/>
            </svg>
            {{ profileSuccess }}
          </div>
          <div v-if="profileError" class="alert alert-error">
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor"
                 stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <circle cx="12" cy="12" r="10"/><line x1="12" y1="8" x2="12" y2="12"/>
              <line x1="12" y1="16" x2="12.01" y2="16"/>
            </svg>
            {{ profileError }}
          </div>

          <div v-if="profileLoading" class="profile-skeleton">
            <div class="skeleton-row"></div>
            <div class="skeleton-row"></div>
            <div class="skeleton-row short"></div>
          </div>

          <form v-else @submit.prevent="saveProfile" novalidate class="profile-form">

            <!-- Email — read only -->
            <div class="field">
              <label>Email address</label>
              <div class="input-wrap readonly-wrap">
                <svg class="field-icon" width="16" height="16" viewBox="0 0 24 24" fill="none"
                     stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                  <path d="M4 4h16c1.1 0 2 .9 2 2v12c0 1.1-.9 2-2 2H4c-1.1 0-2-.9-2-2V6c0-1.1.9-2 2-2z"/>
                  <polyline points="22,6 12,13 2,6"/>
                </svg>
                <input type="email" :value="auth.user?.email" disabled class="readonly" />
                <span class="readonly-badge">
                  <svg width="11" height="11" viewBox="0 0 24 24" fill="none" stroke="currentColor"
                       stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round">
                    <rect x="3" y="11" width="18" height="11" rx="2"/><path d="M7 11V7a5 5 0 0 1 10 0v4"/>
                  </svg>
                  Locked
                </span>
              </div>
              <p class="field-hint">Email cannot be changed. Contact support if needed.</p>
            </div>

            <div class="field-row">
              <div class="field">
                <label for="prof-name">Full name <span class="req">*</span></label>
                <div class="input-wrap">
                  <svg class="field-icon" width="16" height="16" viewBox="0 0 24 24" fill="none"
                       stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                    <path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"/><circle cx="12" cy="7" r="4"/>
                  </svg>
                  <input id="prof-name" v-model="profile.name" type="text"
                         placeholder="Your full name" autocomplete="name" required />
                </div>
              </div>
              <div class="field">
                <label for="prof-company">Company <span class="req">*</span></label>
                <div class="input-wrap">
                  <svg class="field-icon" width="16" height="16" viewBox="0 0 24 24" fill="none"
                       stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                    <rect x="2" y="7" width="20" height="14" rx="2"/>
                    <path d="M16 21V5a2 2 0 0 0-2-2h-4a2 2 0 0 0-2 2v16"/>
                  </svg>
                  <input id="prof-company" v-model="profile.companyName" type="text"
                         placeholder="Your company name" autocomplete="organization" required />
                </div>
              </div>
            </div>

            <div class="field">
              <label for="prof-job">Job title <span class="optional">optional</span></label>
              <div class="input-wrap">
                <svg class="field-icon" width="16" height="16" viewBox="0 0 24 24" fill="none"
                     stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                  <rect x="2" y="3" width="20" height="14" rx="2"/>
                  <path d="M8 21h8m-4-4v4"/>
                </svg>
                <input id="prof-job" v-model="profile.jobTitle" type="text"
                       placeholder="e.g. Founder, Engineer, Designer"
                       autocomplete="organization-title" />
              </div>
            </div>

            <div class="field">
              <label>Phone number <span class="optional">optional</span></label>
              <div class="phone-row">
                <select v-model="profile.phoneCountryCode" class="country-code-select">
                  <option value="">Code</option>
                  <option value="+1">🇺🇸 +1 (US/CA)</option>
                  <option value="+7">🇷🇺 +7 (RU)</option>
                  <option value="+20">🇪🇬 +20 (EG)</option>
                  <option value="+27">🇿🇦 +27 (ZA)</option>
                  <option value="+31">🇳🇱 +31 (NL)</option>
                  <option value="+32">🇧🇪 +32 (BE)</option>
                  <option value="+33">🇫🇷 +33 (FR)</option>
                  <option value="+34">🇪🇸 +34 (ES)</option>
                  <option value="+39">🇮🇹 +39 (IT)</option>
                  <option value="+41">🇨🇭 +41 (CH)</option>
                  <option value="+44">🇬🇧 +44 (GB)</option>
                  <option value="+45">🇩🇰 +45 (DK)</option>
                  <option value="+46">🇸🇪 +46 (SE)</option>
                  <option value="+47">🇳🇴 +47 (NO)</option>
                  <option value="+49">🇩🇪 +49 (DE)</option>
                  <option value="+52">🇲🇽 +52 (MX)</option>
                  <option value="+55">🇧🇷 +55 (BR)</option>
                  <option value="+60">🇲🇾 +60 (MY)</option>
                  <option value="+61">🇦🇺 +61 (AU)</option>
                  <option value="+62">🇮🇩 +62 (ID)</option>
                  <option value="+63">🇵🇭 +63 (PH)</option>
                  <option value="+65">🇸🇬 +65 (SG)</option>
                  <option value="+66">🇹🇭 +66 (TH)</option>
                  <option value="+81">🇯🇵 +81 (JP)</option>
                  <option value="+82">🇰🇷 +82 (KR)</option>
                  <option value="+86">🇨🇳 +86 (CN)</option>
                  <option value="+90">🇹🇷 +90 (TR)</option>
                  <option value="+91">🇮🇳 +91 (IN)</option>
                  <option value="+92">🇵🇰 +92 (PK)</option>
                  <option value="+94">🇱🇰 +94 (LK)</option>
                  <option value="+234">🇳🇬 +234 (NG)</option>
                  <option value="+254">🇰🇪 +254 (KE)</option>
                  <option value="+353">🇮🇪 +353 (IE)</option>
                  <option value="+358">🇫🇮 +358 (FI)</option>
                  <option value="+966">🇸🇦 +966 (SA)</option>
                  <option value="+971">🇦🇪 +971 (AE)</option>
                  <option value="+977">🇳🇵 +977 (NP)</option>
                  <option value="+998">🇺🇿 +998 (UZ)</option>
                </select>
                <div class="input-wrap phone-num-wrap">
                  <input v-model="profile.phoneNumber" type="tel"
                         placeholder="98765 43210" autocomplete="tel-national" />
                </div>
              </div>
            </div>

            <div class="form-actions right">
              <button type="button" class="btn-ghost" @click="resetProfile" :disabled="profileSaving">
                Discard changes
              </button>
              <button type="submit" class="btn-primary" :disabled="profileSaving || !canSaveProfile">
                {{ profileSaving ? 'Saving…' : 'Save profile' }}
              </button>
            </div>

          </form>
        </div>
      </section>

      <!-- ── Change Password ──────────────────────────────────── -->
      <section class="section">
        <div class="section-head">
          <div class="section-icon">
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none"
                 stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <rect x="3" y="11" width="18" height="11" rx="2" ry="2"/>
              <path d="M7 11V7a5 5 0 0 1 10 0v4"/>
            </svg>
          </div>
          <div>
            <h2>Change password</h2>
            <p class="section-sub">Update your password to keep your account secure.</p>
          </div>
        </div>

        <div class="card">
          <div v-if="pwSuccess" class="alert alert-success">
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor"
                 stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"/><polyline points="22 4 12 14.01 9 11.01"/>
            </svg>
            {{ pwSuccess }}
          </div>
          <div v-if="pwError" class="alert alert-error">
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor"
                 stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <circle cx="12" cy="12" r="10"/><line x1="12" y1="8" x2="12" y2="12"/>
              <line x1="12" y1="16" x2="12.01" y2="16"/>
            </svg>
            {{ pwError }}
          </div>

          <form @submit.prevent="changePassword" novalidate class="pw-form">

            <div class="field">
              <label for="current-pw">Current password</label>
              <div class="input-wrap">
                <input :type="showCurrent ? 'text' : 'password'"
                       id="current-pw" v-model="pw.current"
                       placeholder="Enter your current password"
                       autocomplete="current-password" required />
                <button type="button" class="pw-toggle" @click="showCurrent = !showCurrent" tabindex="-1">
                  <svg v-if="!showCurrent" width="17" height="17" viewBox="0 0 24 24" fill="none"
                       stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                    <path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z"/><circle cx="12" cy="12" r="3"/>
                  </svg>
                  <svg v-else width="17" height="17" viewBox="0 0 24 24" fill="none"
                       stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                    <path d="M17.94 17.94A10.07 10.07 0 0 1 12 20c-7 0-11-8-11-8a18.45 18.45 0 0 1 5.06-5.94"/>
                    <path d="M9.9 4.24A9.12 9.12 0 0 1 12 4c7 0 11 8 11 8a18.5 18.5 0 0 1-2.16 3.19"/>
                    <line x1="1" y1="1" x2="23" y2="23"/>
                  </svg>
                </button>
              </div>
            </div>

            <div class="field">
              <label for="new-pw">New password</label>
              <div class="input-wrap">
                <input :type="showNew ? 'text' : 'password'"
                       id="new-pw" v-model="pw.newPw"
                       placeholder="At least 8 characters"
                       autocomplete="new-password" required />
                <button type="button" class="pw-toggle" @click="showNew = !showNew" tabindex="-1">
                  <svg v-if="!showNew" width="17" height="17" viewBox="0 0 24 24" fill="none"
                       stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                    <path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z"/><circle cx="12" cy="12" r="3"/>
                  </svg>
                  <svg v-else width="17" height="17" viewBox="0 0 24 24" fill="none"
                       stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                    <path d="M17.94 17.94A10.07 10.07 0 0 1 12 20c-7 0-11-8-11-8a18.45 18.45 0 0 1 5.06-5.94"/>
                    <path d="M9.9 4.24A9.12 9.12 0 0 1 12 4c7 0 11 8 11 8a18.5 18.5 0 0 1-2.16 3.19"/>
                    <line x1="1" y1="1" x2="23" y2="23"/>
                  </svg>
                </button>
              </div>
              <div v-if="pw.newPw" class="pw-strength">
                <div class="pw-bars">
                  <div v-for="i in 4" :key="i" class="pw-bar"
                       :class="i <= pwStrength.score ? `lvl-${pwStrength.score}` : ''"></div>
                </div>
                <span class="pw-strength-lbl" :class="`s${pwStrength.score}`">{{ pwStrength.label }}</span>
              </div>
            </div>

            <div class="field">
              <label for="confirm-pw">Confirm new password</label>
              <div class="input-wrap">
                <input :type="showConfirm ? 'text' : 'password'"
                       id="confirm-pw" v-model="pw.confirm"
                       :class="{ 'is-error': pw.confirm && pw.newPw !== pw.confirm }"
                       placeholder="Re-enter new password"
                       autocomplete="new-password" required />
                <button type="button" class="pw-toggle" @click="showConfirm = !showConfirm" tabindex="-1">
                  <svg v-if="!showConfirm" width="17" height="17" viewBox="0 0 24 24" fill="none"
                       stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                    <path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z"/><circle cx="12" cy="12" r="3"/>
                  </svg>
                  <svg v-else width="17" height="17" viewBox="0 0 24 24" fill="none"
                       stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                    <path d="M17.94 17.94A10.07 10.07 0 0 1 12 20c-7 0-11-8-11-8a18.45 18.45 0 0 1 5.06-5.94"/>
                    <path d="M9.9 4.24A9.12 9.12 0 0 1 12 4c7 0 11 8 11 8a18.5 18.5 0 0 1-2.16 3.19"/>
                    <line x1="1" y1="1" x2="23" y2="23"/>
                  </svg>
                </button>
              </div>
              <p v-if="pw.confirm && pw.newPw !== pw.confirm" class="field-err">Passwords do not match</p>
            </div>

            <div class="form-actions right">
              <button type="submit" class="btn-primary" :disabled="pwLoading || !canChangePw">
                {{ pwLoading ? 'Updating…' : 'Update password' }}
              </button>
            </div>

          </form>
        </div>
      </section>

      <!-- ── Account Info ─────────────────────────────────────── -->
      <section class="section">
        <div class="section-head">
          <div class="section-icon">
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none"
                 stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <circle cx="12" cy="12" r="10"/>
              <line x1="12" y1="8" x2="12" y2="12"/>
              <line x1="12" y1="16" x2="12.01" y2="16"/>
            </svg>
          </div>
          <div>
            <h2>Account info</h2>
            <p class="section-sub">Read-only account metadata.</p>
          </div>
        </div>

        <div class="card info-rows">
          <div class="info-row">
            <span class="info-row-label">Email</span>
            <span class="info-row-value">{{ auth.user?.email }}</span>
          </div>
          <div class="info-row">
            <span class="info-row-label">Member since</span>
            <span class="info-row-value">{{ formatDate(auth.user?.created_at) }}</span>
          </div>
          <div class="info-row" :class="{ 'info-row-tall': !auth.user?.email_verified }">
            <span class="info-row-label">Email verified</span>
            <span v-if="auth.user?.email_verified" class="info-row-value text-green">
              <svg width="13" height="13" viewBox="0 0 24 24" fill="none"
                   stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round">
                <polyline points="20 6 9 17 4 12"/>
              </svg>
              Verified
            </span>
            <span v-else class="info-row-value-col">
              <span class="text-orange" style="display:flex;align-items:center;gap:5px;font-size:0.875rem;font-weight:600">
                <svg width="13" height="13" viewBox="0 0 24 24" fill="none"
                     stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round">
                  <circle cx="12" cy="12" r="10"/><line x1="12" y1="8" x2="12" y2="12"/>
                  <line x1="12" y1="16" x2="12.01" y2="16"/>
                </svg>
                Not verified
              </span>

              <!-- idle: send OTP button -->
              <button v-if="verifyStep === 'idle'" class="btn-resend" :disabled="sendingOTP" @click="sendVerifyOTP">
                <svg v-if="sendingOTP" width="12" height="12" viewBox="0 0 24 24" fill="none"
                     stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round" class="spin">
                  <path d="M21 12a9 9 0 1 1-6.219-8.56"/>
                </svg>
                {{ sendingOTP ? 'Sending…' : 'Send verification code' }}
              </button>

              <!-- otp: inline 6-box input -->
              <template v-if="verifyStep === 'otp'">
                <div style="display:flex;flex-direction:column;align-items:flex-end;gap:8px;width:100%">
                  <div v-if="verifyOtpError" class="resend-msg err" style="width:100%;text-align:right">{{ verifyOtpError }}</div>
                  <div class="otp-inputs-sm">
                    <input
                      v-for="i in 6" :key="i"
                      :ref="el => { verifyOtpRefs[i-1] = el }"
                      type="text" inputmode="numeric" pattern="[0-9]*"
                      maxlength="1"
                      class="otp-box-sm"
                      :class="{ filled: verifyOtpDigits[i-1] }"
                      :value="verifyOtpDigits[i-1]"
                      @input="onVerifyOtpInput(i-1, $event)"
                      @keydown="onVerifyOtpKeydown(i-1, $event)"
                      @paste.prevent="onVerifyOtpPaste($event)"
                    />
                  </div>
                  <div style="display:flex;gap:10px;align-items:center">
                    <button class="link-btn-sm" :disabled="resendLoading" @click="sendVerifyOTP">
                      {{ resendLoading ? 'Sending…' : 'Resend' }}
                    </button>
                    <button class="btn-resend" style="border-color:var(--blue);color:var(--blue);background:var(--blue-light)"
                            :disabled="verifyOtpValue.length < 6 || verifyOtpLoading" @click="submitVerifyOTP">
                      <svg v-if="verifyOtpLoading" width="12" height="12" viewBox="0 0 24 24" fill="none"
                           stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round" class="spin">
                        <path d="M21 12a9 9 0 1 1-6.219-8.56"/>
                      </svg>
                      {{ verifyOtpLoading ? 'Verifying…' : 'Verify' }}
                    </button>
                  </div>
                </div>
              </template>
            </span>
          </div>
          <div class="info-row">
            <span class="info-row-label">Last login</span>
            <span class="info-row-value">{{ formatDate(auth.user?.last_login_at) }}</span>
          </div>
        </div>
      </section>

    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted, nextTick } from 'vue'
import { useAuthStore } from '../stores/auth'
import { authAPI } from '../services/api'

const auth = useAuthStore()

// ── Profile ───────────────────────────────────────────────────────────────────
const profileLoading = ref(true)
const profileSaving  = ref(false)
const profileError   = ref('')
const profileSuccess = ref('')

const profile = ref({
  name:             '',
  companyName:      '',
  jobTitle:         '',
  phoneCountryCode: '+91',
  phoneNumber:      '',
})

let profileSnapshot = { ...profile.value }

async function fetchProfile() {
  try {
    const { data } = await authAPI.getProfile()
    const u = data.user
    // Always sync the auth store with fresh server data so email_verified
    // (and any other field changed in another tab/session) is up to date.
    auth.updateUser(u)
    profile.value = {
      name:             u.name              ?? '',
      companyName:      u.company_name      ?? '',
      jobTitle:         u.job_title         ?? '',
      phoneCountryCode: u.phone_country_code ?? '+91',
      phoneNumber:      u.phone_number      ?? '',
    }
    profileSnapshot = { ...profile.value }
  } catch {
    profileError.value = 'Failed to load profile. Please refresh the page.'
  } finally {
    profileLoading.value = false
  }
}

// Re-fetch when the user switches back to this tab (e.g. verified email in
// another tab — their Pinia store here is stale until we re-sync from the API).
function onVisibilityChange() {
  if (document.visibilityState === 'visible') fetchProfile()
}

onMounted(() => {
  fetchProfile()
  document.addEventListener('visibilitychange', onVisibilityChange)
})

onUnmounted(() => {
  document.removeEventListener('visibilitychange', onVisibilityChange)
})

const canSaveProfile = computed(() =>
  profile.value.name.trim().length > 0 &&
  profile.value.companyName.trim().length > 0
)

function resetProfile() {
  profile.value        = { ...profileSnapshot }
  profileError.value   = ''
  profileSuccess.value = ''
}

async function saveProfile() {
  if (!canSaveProfile.value) return
  profileSaving.value  = true
  profileError.value   = ''
  profileSuccess.value = ''
  try {
    const { data } = await authAPI.setupProfile({
      name:               profile.value.name.trim(),
      company_name:       profile.value.companyName.trim(),
      job_title:          profile.value.jobTitle.trim(),
      phone_country_code: profile.value.phoneCountryCode,
      phone_number:       profile.value.phoneNumber.trim(),
    })
    auth.updateUser(data.user)
    profileSnapshot      = { ...profile.value }
    profileSuccess.value = 'Profile updated successfully.'
    setTimeout(() => { profileSuccess.value = '' }, 4000)
  } catch (err) {
    profileError.value = err.response?.data?.error ?? 'Failed to save profile. Please try again.'
  } finally {
    profileSaving.value = false
  }
}

// ── Change password ───────────────────────────────────────────────────────────
const pw          = ref({ current: '', newPw: '', confirm: '' })
const pwLoading   = ref(false)
const pwError     = ref('')
const pwSuccess   = ref('')
const showCurrent = ref(false)
const showNew     = ref(false)
const showConfirm = ref(false)

const pwStrength = computed(() => {
  const v = pw.value.newPw
  let score = 0
  if (v.length >= 8)                       score++
  if (/[A-Z]/.test(v) && /[a-z]/.test(v)) score++
  if (/\d/.test(v))                        score++
  if (/[^A-Za-z0-9]/.test(v))             score++
  const labels = ['', 'Weak', 'Fair', 'Good', 'Strong']
  return { score: Math.max(score, v.length ? 1 : 0), label: labels[Math.max(score, v.length ? 1 : 0)] }
})

const canChangePw = computed(() =>
  pw.value.current.length > 0 &&
  pw.value.newPw.length >= 8 &&
  pw.value.newPw === pw.value.confirm
)

async function changePassword() {
  if (!canChangePw.value) return
  pwLoading.value = true
  pwError.value   = ''
  pwSuccess.value = ''
  try {
    const { data } = await authAPI.changePassword({
      current_password: pw.value.current,
      new_password:     pw.value.newPw,
    })
    pwSuccess.value = data.message
    pw.value = { current: '', newPw: '', confirm: '' }
  } catch (err) {
    pwError.value = err.response?.data?.error ?? 'Failed to update password. Please try again.'
  } finally {
    pwLoading.value = false
  }
}

// ── Email verify OTP (inline in Settings) ────────────────────────────────────
const verifyStep      = ref('idle')   // 'idle' | 'otp'
const sendingOTP      = ref(false)
const resendLoading   = ref(false)
const verifyOtpDigits = ref(['', '', '', '', '', ''])
const verifyOtpRefs   = ref([])
const verifyOtpLoading = ref(false)
const verifyOtpError   = ref('')
const verifyOtpValue   = computed(() => verifyOtpDigits.value.join(''))

function onVerifyOtpInput(idx, e) {
  const val = e.target.value.replace(/\D/g, '')
  verifyOtpDigits.value[idx] = val.slice(-1)
  if (val && idx < 5) nextTick(() => verifyOtpRefs.value[idx + 1]?.focus())
}
function onVerifyOtpKeydown(idx, e) {
  if (e.key === 'Backspace' && !verifyOtpDigits.value[idx] && idx > 0) {
    nextTick(() => verifyOtpRefs.value[idx - 1]?.focus())
  }
}
function onVerifyOtpPaste(e) {
  const text = (e.clipboardData || window.clipboardData).getData('text')
  const digits = text.replace(/\D/g, '').slice(0, 6).split('')
  digits.forEach((d, i) => { verifyOtpDigits.value[i] = d })
  nextTick(() => verifyOtpRefs.value[Math.min(digits.length, 5)]?.focus())
}

async function sendVerifyOTP() {
  sendingOTP.value  = true
  resendLoading.value = true
  verifyOtpError.value = ''
  verifyOtpDigits.value = ['', '', '', '', '', '']
  try {
    await authAPI.resendVerification({ email: auth.user?.email })
    verifyStep.value = 'otp'
    nextTick(() => verifyOtpRefs.value[0]?.focus())
  } catch {
    verifyOtpError.value = 'Failed to send code. Please try again.'
  } finally {
    sendingOTP.value    = false
    resendLoading.value = false
  }
}

async function submitVerifyOTP() {
  verifyOtpError.value   = ''
  verifyOtpLoading.value = true
  try {
    const { data } = await authAPI.verifyEmailOTP({
      email: auth.user?.email,
      otp:   verifyOtpValue.value,
    })
    auth.setAuth(data.token, data.user)
    verifyStep.value = 'idle'
    verifyOtpDigits.value = ['', '', '', '', '', '']
  } catch (err) {
    verifyOtpError.value = err.response?.data?.error ?? 'Invalid code. Please try again.'
    verifyOtpDigits.value = ['', '', '', '', '', '']
    nextTick(() => verifyOtpRefs.value[0]?.focus())
  } finally {
    verifyOtpLoading.value = false
  }
}

// ── Helpers ───────────────────────────────────────────────────────────────────
function formatDate(iso) {
  if (!iso) return '—'
  return new Date(iso).toLocaleDateString('en-US', { year: 'numeric', month: 'long', day: 'numeric' })
}
</script>

<style scoped>
.settings-page { min-height: 100vh; background: var(--bg, #f1f3f4); }

.page-header {
  background: #fff;
  border-bottom: 1px solid var(--border, #dadce0);
  padding: 2rem 2rem 1.5rem;
}
.page-header-inner { max-width: 860px; margin: 0 auto; }
.page-header h1 { font-size: 1.4rem; font-weight: 700; color: var(--text, #202124); margin-bottom: 4px; }
.page-header p  { font-size: 0.875rem; color: var(--text-muted, #5f6368); }

.page-body { max-width: 860px; margin: 0 auto; padding: 2rem; }

.section { margin-bottom: 2.5rem; }
.section-head {
  display: flex; align-items: flex-start; gap: 12px; margin-bottom: 1.25rem;
}
.section-icon {
  width: 36px; height: 36px; border-radius: 9px; flex-shrink: 0;
  background: var(--blue-light, #e8f0fe); color: var(--blue, #1a73e8);
  display: flex; align-items: center; justify-content: center; margin-top: 2px;
}
.section-head h2   { font-size: 1rem; font-weight: 700; color: var(--text, #202124); }
.section-sub { font-size: 0.8rem; color: var(--text-muted, #5f6368); margin-top: 2px; }

.card {
  background: #fff;
  border: 1px solid var(--border, #dadce0);
  border-radius: 12px;
  padding: 1.75rem;
}

/* Skeleton */
.profile-skeleton { display: flex; flex-direction: column; gap: 14px; }
.skeleton-row {
  height: 44px; border-radius: 8px;
  background: linear-gradient(90deg, #f0f0f0 25%, #e4e4e4 50%, #f0f0f0 75%);
  background-size: 200% 100%;
  animation: shimmer 1.4s infinite;
}
.skeleton-row.short { width: 55%; }
@keyframes shimmer { 0% { background-position: 200% 0 } 100% { background-position: -200% 0 } }

/* Profile form */
.profile-form { max-width: 100%; }

.field-row {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 0 1.25rem;
}
@media (max-width: 600px) { .field-row { grid-template-columns: 1fr; } }

.field { margin-bottom: 1.25rem; }
.field label {
  display: block; font-size: 0.82rem; font-weight: 600;
  color: var(--text, #202124); margin-bottom: 6px;
}
.req      { color: var(--blue, #1a73e8); font-weight: 700; }
.optional { font-size: 0.75rem; font-weight: 400; color: var(--text-muted, #5f6368); margin-left: 4px; }

.input-wrap {
  position: relative; display: flex; align-items: center;
}
.input-wrap input {
  width: 100%; height: 42px; border: 1.5px solid var(--border, #dadce0);
  border-radius: 8px; padding: 0 42px 0 38px;
  font-size: 0.875rem; color: var(--text, #202124);
  background: #fff; outline: none;
  transition: border-color .15s, box-shadow .15s;
  font-family: inherit;
}
.input-wrap input:focus {
  border-color: var(--blue, #1a73e8);
  box-shadow: 0 0 0 3px rgba(26, 115, 232, 0.12);
}
.input-wrap input.is-error {
  border-color: #c5221f;
  box-shadow: 0 0 0 3px rgba(197, 34, 31, 0.1);
}
.field-icon {
  position: absolute; left: 11px; color: var(--text-muted, #5f6368); pointer-events: none; flex-shrink: 0;
}

/* Read-only */
.readonly-wrap input.readonly {
  background: var(--bg, #f1f3f4); color: var(--text-muted, #5f6368);
  cursor: not-allowed; user-select: none;
}
.readonly-badge {
  position: absolute; right: 10px;
  display: flex; align-items: center; gap: 4px;
  font-size: 0.7rem; font-weight: 600;
  color: var(--text-muted, #5f6368);
  background: var(--border, #dadce0);
  padding: 3px 8px; border-radius: 6px;
  pointer-events: none;
}
.field-hint { font-size: 0.75rem; color: var(--text-muted, #5f6368); margin-top: 5px; }

/* Phone */
.phone-row { display: flex; gap: 10px; }
.country-code-select {
  height: 42px; border: 1.5px solid var(--border, #dadce0);
  border-radius: 8px; padding: 0 10px;
  font-size: 0.82rem; color: var(--text, #202124);
  background: #fff; outline: none; cursor: pointer;
  flex-shrink: 0; width: 155px;
  transition: border-color .15s, box-shadow .15s;
  font-family: inherit;
}
.country-code-select:focus {
  border-color: var(--blue, #1a73e8);
  box-shadow: 0 0 0 3px rgba(26, 115, 232, 0.12);
}
.phone-num-wrap { flex: 1; }
.phone-num-wrap input { padding-left: 14px; }

/* Password toggle */
.pw-toggle {
  position: absolute; right: 10px; background: none; border: none;
  cursor: pointer; color: var(--text-muted, #5f6368); padding: 4px;
  display: flex; align-items: center; border-radius: 4px;
  transition: color .15s;
}
.pw-toggle:hover { color: var(--blue, #1a73e8); }

/* Password strength */
.pw-strength { display: flex; align-items: center; gap: 8px; margin-top: 7px; }
.pw-bars     { display: flex; gap: 4px; }
.pw-bar      { height: 3px; width: 44px; border-radius: 3px; background: #e0e0e0; transition: background .2s; }
.pw-bar.lvl-1 { background: #f44336; }
.pw-bar.lvl-2 { background: #ff9800; }
.pw-bar.lvl-3 { background: #2196f3; }
.pw-bar.lvl-4 { background: #4caf50; }
.pw-strength-lbl { font-size: 0.72rem; font-weight: 600; }
.s1 { color: #f44336; } .s2 { color: #ff9800; } .s3 { color: #2196f3; } .s4 { color: #4caf50; }

/* Alerts */
.alert {
  display: flex; align-items: center; gap: 8px;
  padding: 10px 14px; border-radius: 8px;
  font-size: 0.85rem; font-weight: 500; margin-bottom: 1.25rem;
}
.alert-success { background: #e6f4ea; color: #1e8e3e; }
.alert-error   { background: #fce8e6; color: #c5221f; }

.field-err { font-size: 0.75rem; color: #c5221f; margin-top: 4px; }

/* Form actions */
.form-actions { display: flex; gap: 10px; margin-top: 1.5rem; }
.form-actions.right { justify-content: flex-end; }

/* Buttons */
.btn-primary {
  height: 38px; padding: 0 20px; border-radius: 8px;
  background: var(--blue, #1a73e8); color: #fff; font-weight: 600;
  font-size: 0.875rem; border: none; cursor: pointer;
  transition: background .15s, opacity .15s;
}
.btn-primary:hover:not(:disabled) { background: #1557b0; }
.btn-primary:disabled { opacity: 0.55; cursor: not-allowed; }

.btn-ghost {
  height: 38px; padding: 0 18px; border-radius: 8px;
  background: none; color: var(--text-muted, #5f6368); font-weight: 600;
  font-size: 0.875rem; border: 1.5px solid var(--border, #dadce0); cursor: pointer;
  transition: border-color .15s, color .15s;
}
.btn-ghost:hover:not(:disabled) { border-color: #aaa; color: var(--text, #202124); }
.btn-ghost:disabled { opacity: 0.45; cursor: not-allowed; }

/* Password form */
.pw-form { max-width: 440px; }

/* Info rows */
.info-rows { padding: 0; }
.info-row {
  display: flex; align-items: center; justify-content: space-between;
  padding: 1rem 1.5rem;
  border-bottom: 1px solid var(--border, #dadce0);
}
.info-row:last-child { border-bottom: none; }
.info-row-label { font-size: 0.875rem; color: var(--text-muted, #5f6368); font-weight: 500; }
.info-row-value {
  font-size: 0.875rem; color: var(--text, #202124); font-weight: 600;
  display: flex; align-items: center; gap: 5px;
}
.text-green  { color: #1e8e3e; }
.text-orange { color: #e37400; }

/* Email verify inline OTP */
.info-row-tall { align-items: flex-start; padding-top: 1.1rem; padding-bottom: 1.1rem; }
.info-row-value-col {
  display: flex; flex-direction: column; align-items: flex-end; gap: 8px;
}
.btn-resend {
  display: inline-flex; align-items: center; gap: 6px;
  height: 30px; padding: 0 12px; border-radius: 7px;
  font-size: 0.78rem; font-weight: 600; cursor: pointer;
  border: 1.5px solid #e37400; color: #e37400; background: #fff9f0;
  transition: background .15s, color .15s;
}
.btn-resend:hover:not(:disabled) { background: #fff0d6; }
.btn-resend:disabled { opacity: 0.6; cursor: not-allowed; }
.link-btn-sm {
  font-size: 0.75rem; font-weight: 500; color: var(--blue); cursor: pointer;
  background: none; border: none; padding: 0; text-decoration: underline;
}
.link-btn-sm:disabled { opacity: 0.5; cursor: not-allowed; }
.resend-msg {
  font-size: 0.78rem; font-weight: 600; padding: 4px 10px; border-radius: 6px;
}
.resend-msg.ok  { background: #e6f4ea; color: #1e8e3e; }
.resend-msg.err { background: #fce8e6; color: #c5221f; }
/* Compact OTP boxes for Settings inline */
.otp-inputs-sm { display: flex; gap: 6px; }
.otp-box-sm {
  width: 34px; height: 40px;
  border: 1.5px solid var(--border, #dadce0); border-radius: 7px;
  font-size: 17px; font-weight: 700; text-align: center;
  outline: none; background: #fff; color: var(--text, #202124);
  transition: border-color .15s, box-shadow .15s; caret-color: transparent;
}
.otp-box-sm:focus { border-color: var(--blue, #1a73e8); box-shadow: 0 0 0 2px var(--blue-light, #e8f0fe); }
.otp-box-sm.filled { border-color: var(--blue, #1a73e8); background: var(--blue-light, #e8f0fe); }
@keyframes spin { to { transform: rotate(360deg); } }
.spin { animation: spin .8s linear infinite; }
</style>
