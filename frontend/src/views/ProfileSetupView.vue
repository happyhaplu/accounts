<template>
  <AuthLayout>
    <div class="form-header">
      <h1>Set up your profile</h1>
      <p>Tell us about yourself to personalise your workspace.</p>
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

      <!-- Full name (required) -->
      <div class="field">
        <label for="name">
          Full name
          <span class="req">*</span>
        </label>
        <div class="input-wrap">
          <input id="name" v-model="form.name" type="text"
                 placeholder="Alex Johnson" autocomplete="name" required />
        </div>
      </div>

      <!-- Company name (required) -->
      <div class="field">
        <label for="company">
          Company name
          <span class="req">*</span>
        </label>
        <div class="input-wrap">
          <input id="company" v-model="form.companyName" type="text"
                 placeholder="Acme Inc." autocomplete="organization" required />
        </div>
      </div>

      <!-- Job title (optional) -->
      <div class="field">
        <label for="jobtitle">
          Job title
          <span class="optional">optional</span>
        </label>
        <div class="input-wrap">
          <input id="jobtitle" v-model="form.jobTitle" type="text"
                 placeholder="e.g. Founder, Engineer, Marketing" autocomplete="organization-title" />
        </div>
      </div>

      <!-- Phone number (optional) -->
      <div class="field">
        <label>
          Phone number
          <span class="optional">optional</span>
        </label>
        <div class="phone-row">
          <select v-model="form.phoneCountryCode" class="country-code-select">
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
          <div class="input-wrap phone-input-wrap">
            <input v-model="form.phoneNumber" type="tel"
                   placeholder="98765 43210" autocomplete="tel-national" />
          </div>
        </div>
      </div>

      <div class="form-actions">
        <span class="req-note"><span class="req">*</span> required</span>
        <button type="submit" class="btn-primary" :disabled="loading || !canSubmit">
          <span>{{ loading ? 'Saving…' : 'Continue to dashboard' }}</span>
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
import { ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import { authAPI } from '../services/api'
import AuthLayout from '../layouts/AuthLayout.vue'

const router  = useRouter()
const auth    = useAuthStore()
const loading = ref(false)
const error   = ref('')
const form    = ref({ name: '', companyName: '', jobTitle: '', phoneCountryCode: '+91', phoneNumber: '' })

const canSubmit = computed(() =>
  form.value.name.trim().length > 0 && form.value.companyName.trim().length > 0
)

async function submit() {
  if (!canSubmit.value) return
  loading.value = true
  error.value   = ''
  try {
    const { data } = await authAPI.setupProfile({
      name:               form.value.name.trim(),
      company_name:       form.value.companyName.trim(),
      job_title:          form.value.jobTitle.trim(),
      phone_country_code: form.value.phoneCountryCode,
      phone_number:       form.value.phoneNumber.trim(),
    })
    auth.updateUser(data.user)
    router.push('/dashboard')
  } catch (err) {
    error.value = err.response?.data?.error ?? 'Failed to save profile. Please try again.'
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.req        { color: #d93025; font-weight: 600; margin-left: 2px; }
.optional   { font-size: 11px; color: var(--text-muted, #5f6368); font-weight: 400; margin-left: 4px; }
.req-note   { font-size: 12px; color: var(--text-muted, #5f6368); }

.phone-row {
  display: flex;
  gap: 8px;
  align-items: stretch;
}
.country-code-select {
  flex: 0 0 auto;
  height: 44px;
  padding: 0 10px;
  border: 1.5px solid #dadce0;
  border-radius: 8px;
  font-size: 14px;
  color: #202124;
  background: #fff;
  cursor: pointer;
  outline: none;
  transition: border-color .15s;
  min-width: 130px;
}
.country-code-select:focus {
  border-color: #1a73e8;
  box-shadow: 0 0 0 3px rgba(26,115,232,.12);
}
.phone-input-wrap {
  flex: 1;
}
</style>
