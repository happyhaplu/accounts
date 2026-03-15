<template>
  <div id="app">
    <!-- Top navbar — only for authenticated pages -->
    <nav v-if="auth.isAuthenticated" class="top-nav">
      <div class="nav-left">
        <div class="nav-logo-mark">
          <svg viewBox="0 0 48 48" fill="none" xmlns="http://www.w3.org/2000/svg" width="18" height="18">
            <defs>
              <linearGradient id="planeGradNav" x1="4" y1="38" x2="44" y2="4" gradientUnits="userSpaceOnUse">
                <stop offset="0%" stop-color="#1d4ed8"/>
                <stop offset="100%" stop-color="#4f8ef7"/>
              </linearGradient>
            </defs>
            <path d="M4 36 L44 6 L32 44 L22 28 Z" fill="url(#planeGradNav)"/>
            <path d="M22 28 L32 44 L26 30 Z" fill="#1535a8"/>
            <path d="M22 28 L4 36 L44 6 Z" fill="rgba(255,255,255,0.18)"/>
          </svg>
        </div>
        <span class="nav-brand">Outcraftly</span>
        <span class="nav-section">Accounts</span>
      </div>

      <div class="nav-right">

        <!-- ── Waffle / App launcher ── -->
        <div class="waffle-wrap" v-click-outside="closeApps">
          <button class="waffle-btn" @click="appsOpen = !appsOpen" :class="{ active: appsOpen }" title="Outcraftly apps">
            <svg width="20" height="20" viewBox="0 0 24 24" fill="currentColor">
              <circle cx="5"  cy="5"  r="1.8"/>
              <circle cx="12" cy="5"  r="1.8"/>
              <circle cx="19" cy="5"  r="1.8"/>
              <circle cx="5"  cy="12" r="1.8"/>
              <circle cx="12" cy="12" r="1.8"/>
              <circle cx="19" cy="12" r="1.8"/>
              <circle cx="5"  cy="19" r="1.8"/>
              <circle cx="12" cy="19" r="1.8"/>
              <circle cx="19" cy="19" r="1.8"/>
            </svg>
          </button>
          <transition name="dropdown">
            <div v-if="appsOpen" class="apps-panel">

              <!-- Panel header -->
              <div class="apps-panel-header">
                <div class="apps-panel-brand">
                  <svg viewBox="0 0 48 48" fill="none" width="18" height="18">
                    <defs>
                      <linearGradient id="appsLauncherGrad" x1="4" y1="38" x2="44" y2="4" gradientUnits="userSpaceOnUse">
                        <stop offset="0%" stop-color="#1d4ed8"/>
                        <stop offset="100%" stop-color="#4f8ef7"/>
                      </linearGradient>
                    </defs>
                    <path d="M4 36 L44 6 L32 44 L22 28 Z" fill="url(#appsLauncherGrad)"/>
                    <path d="M22 28 L32 44 L26 30 Z" fill="#1535a8"/>
                    <path d="M22 28 L4 36 L44 6 Z" fill="rgba(255,255,255,0.18)"/>
                  </svg>
                  <span>Outcraftly</span>
                </div>
                <span class="apps-panel-label">All products</span>
              </div>

              <!-- App grid -->
              <div class="apps-grid">
                <a v-for="app in apps" :key="app.name" :href="app.url" class="app-tile"
                   rel="noopener" @click="appsOpen = false">
                  <div class="app-icon-wrap"
                       :style="{ background: app.gradient, boxShadow: '0 4px 14px ' + app.shadow }">
                    <span class="app-letter">{{ app.short }}</span>
                  </div>
                  <span class="app-tile-name">{{ app.name }}</span>
                  <span class="app-tile-badge">
                    <span class="badge-dot"></span>
                    {{ app.status }}
                  </span>
                </a>
              </div>

              <!-- Panel footer -->
              <div class="apps-panel-footer">
                <span>More apps coming soon</span>
              </div>

            </div>
          </transition>
        </div>

        <!-- ── Account pill ── -->
        <div class="nav-avatar-wrap" @click="menuOpen = !menuOpen" v-click-outside="closeMenu">
          <div class="nav-user-pill">
            <div class="nav-avatar">{{ userInitial }}</div>
            <span class="nav-user-name">{{ auth.user?.name || auth.user?.email }}</span>
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor"
                 stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round" class="nav-caret">
              <polyline points="6 9 12 15 18 9"/>
            </svg>
          </div>
          <transition name="dropdown">
            <div v-if="menuOpen" class="nav-menu">
              <!-- Header: avatar + name + email -->
              <div class="nav-menu-header">
                <div class="menu-avatar-lg">{{ userInitial }}</div>
                <div class="menu-info">
                  <div class="menu-name">{{ auth.user?.name || 'Account' }}</div>
                  <div class="menu-email">{{ auth.user?.email }}</div>
                </div>
              </div>
              <div class="nav-menu-divider"></div>
              <router-link class="nav-menu-item" to="/settings" @click="menuOpen = false">
                <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor"
                     stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                  <circle cx="12" cy="12" r="3"/>
                  <path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1-2.83 2.83l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-4 0v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83-2.83l.06-.06A1.65 1.65 0 0 0 4.68 15a1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1 0-4h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 2.83-2.83l.06.06A1.65 1.65 0 0 0 9 4.68a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 4 0v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 2.83l-.06.06A1.65 1.65 0 0 0 19.4 9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 0 4h-.09a1.65 1.65 0 0 0-1.51 1z"/>
                </svg>
                Settings
              </router-link>
              <router-link class="nav-menu-item" to="/workspaces" @click="menuOpen = false">
                <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor"
                     stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                  <path d="M3 9l9-7 9 7v11a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z"/>
                  <polyline points="9 22 9 12 15 12 15 22"/>
                </svg>
                Workspaces
              </router-link>
              <router-link class="nav-menu-item" to="/billing" @click="menuOpen = false">
                <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor"
                     stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                  <rect x="1" y="4" width="22" height="16" rx="2" ry="2"/>
                  <line x1="1" y1="10" x2="23" y2="10"/>
                </svg>
                Billing
              </router-link>
              <div class="nav-menu-divider"></div>
              <button class="nav-menu-item sign-out" @click="handleLogout">
                <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor"
                     stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                  <path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4"/>
                  <polyline points="16 17 21 12 16 7"/>
                  <line x1="21" y1="12" x2="9" y2="12"/>
                </svg>
                Sign out
              </button>
            </div>
          </transition>
        </div>

      </div>
    </nav>

    <router-view />
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import { useAuthStore } from './stores/auth'
import { useRouter } from 'vue-router'

const auth        = useAuthStore()
const router      = useRouter()
const menuOpen    = ref(false)
const appsOpen    = ref(false)

const apps = [
  { name: 'Email Warmup', short: 'EW', gradient: 'linear-gradient(135deg,#ea4335,#fb8c00)', shadow: 'rgba(234,67,53,0.35)', url: '/products/email-warmup/launch', status: 'Active' },
  { name: 'Reach',        short: 'RE', gradient: 'linear-gradient(135deg,#4285f4,#1a73e8)', shadow: 'rgba(66,133,244,0.35)', url: '/products/reach/launch',        status: 'Active' },
  { name: 'Cold Email',   short: 'CE', gradient: 'linear-gradient(135deg,#34a853,#137333)', shadow: 'rgba(52,168,83,0.35)',  url: '/products/cold_email/launch',   status: 'Coming soon' },
  { name: 'LinkedIn',     short: 'LI', gradient: 'linear-gradient(135deg,#0077b5,#005582)', shadow: 'rgba(0,119,181,0.35)',  url: '/products/linkedin/launch',     status: 'Coming soon' },
]

const userInitial = computed(() => {
  const name = auth.user?.name
  if (name && name.trim()) return name.trim()[0].toUpperCase()
  return auth.user?.email?.[0]?.toUpperCase() ?? 'U'
})

function closeMenu() { menuOpen.value = false }
function closeApps() { appsOpen.value = false }
function handleLogout() { menuOpen.value = false; auth.logout(); router.push('/login') }

const vClickOutside = {
  mounted(el, binding) {
    el._oco = (e) => { if (!el.contains(e.target)) binding.value() }
    document.addEventListener('click', el._oco, true)
  },
  unmounted(el) { document.removeEventListener('click', el._oco, true) },
}
</script>
<style>
/* ── Inter font ─────────────────────────────────────────────── */
@import url('https://fonts.googleapis.com/css2?family=Inter:wght@300;400;500;600;700&display=swap');

/* ── Design tokens ──────────────────────────────────────────── */
:root {
  --blue:        #1a73e8;
  --blue-dark:   #1557b0;
  --blue-darker: #0d47a1;
  --blue-light:  #e8f0fe;
  --text:        #202124;
  --text-muted:  #5f6368;
  --text-hint:   #9aa0a6;
  --border:      #dadce0;
  --bg:          #f1f3f4;
  --surface:     #ffffff;
  --error:       #d93025;
  --error-bg:    #fce8e6;
  --success:     #1e8e3e;
  --success-bg:  #e6f4ea;
  --warning:     #f9ab00;
  --warning-bg:  #fef7e0;
  --radius:      6px;
}

/* ── Reset ──────────────────────────────────────────────────── */
*, *::before, *::after { box-sizing: border-box; margin: 0; padding: 0; }
html { font-size: 16px; }
body {
  font-family: 'Inter', -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif;
  background: var(--bg);
  color: var(--text);
  -webkit-font-smoothing: antialiased;
  line-height: 1.5;
}
a { color: var(--blue); text-decoration: none; }
a:hover { text-decoration: underline; }
button { font-family: inherit; }

/* ── Top Navbar ─────────────────────────────────────────────── */
.top-nav {
  position: sticky; top: 0; z-index: 200;
  height: 64px; background: var(--surface);
  border-bottom: 1px solid var(--border);
  display: flex; align-items: center;
  justify-content: space-between; padding: 0 24px; gap: 16px;
}
.nav-left { display: flex; align-items: center; gap: 10px; }
.nav-logo-mark {
  width: 30px; height: 30px;
  display: flex; align-items: center; justify-content: center; flex-shrink: 0;
  filter: drop-shadow(0 1px 3px rgba(26,115,232,0.25));
}
.nav-brand  { font-size: 15px; font-weight: 700; color: var(--text); letter-spacing: -0.2px; }
.nav-section {
  font-size: 13px; color: var(--text-muted);
  padding-left: 8px; border-left: 1.5px solid var(--border); margin-left: 2px;
}

.nav-right { display: flex; align-items: center; gap: 12px; }
.nav-email {
  font-size: 13px; color: var(--text-muted);
  max-width: 200px; white-space: nowrap; overflow: hidden; text-overflow: ellipsis;
}
/* User pill button */
.nav-user-pill {
  display: flex; align-items: center; gap: 8px;
  padding: 4px 10px 4px 4px;
  border-radius: 24px; cursor: pointer;
  border: 1.5px solid var(--border, #dadce0);
  background: #fff;
  transition: border-color .15s, box-shadow .15s;
  user-select: none;
}
.nav-user-pill:hover { border-color: var(--blue); box-shadow: 0 0 0 3px var(--blue-light); }
.nav-avatar {
  width: 32px; height: 32px; border-radius: 50%;
  background: var(--blue); color: #fff; border: none;
  font-size: 13px; font-weight: 700;
  display: flex; align-items: center; justify-content: center;
  flex-shrink: 0; pointer-events: none;
}
.nav-user-name {
  font-size: 13.5px; font-weight: 500; color: var(--text);
  max-width: 140px; white-space: nowrap; overflow: hidden; text-overflow: ellipsis;
}
.nav-caret { color: var(--text-muted); flex-shrink: 0; }

.nav-avatar-wrap { position: relative; }
.nav-menu {
  position: absolute; top: calc(100% + 10px); right: 0;
  background: var(--surface); border: 1px solid var(--border);
  border-radius: 12px; box-shadow: 0 8px 28px rgba(0,0,0,0.14);
  min-width: 260px; z-index: 300; overflow: hidden;
}
.nav-menu-header { display: flex; align-items: center; gap: 12px; padding: 18px 16px; }
.menu-avatar-lg {
  width: 48px; height: 48px; border-radius: 50%;
  background: var(--blue); color: #fff;
  display: flex; align-items: center; justify-content: center;
  font-size: 20px; font-weight: 700; flex-shrink: 0;
}
.menu-info { min-width: 0; }
.menu-name  { font-size: 14px; font-weight: 600; color: var(--text); margin-bottom: 2px; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.menu-email { font-size: 12px; color: var(--text-muted); white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.nav-menu-divider { height: 1px; background: var(--border); }
.nav-menu-item {
  display: flex; align-items: center; gap: 10px;
  width: 100%; padding: 11px 16px; background: none; border: none;
  font-size: 13.5px; color: var(--text); cursor: pointer;
  text-align: left; transition: background 0.15s; text-decoration: none;
}
.nav-menu-item:hover { background: var(--bg); text-decoration: none; }
.nav-menu-item.sign-out { color: #c5221f; }
.nav-menu-item.sign-out:hover { background: #fce8e6; }
.dropdown-enter-active, .dropdown-leave-active { transition: opacity 0.15s, transform 0.15s; }
.dropdown-enter-from, .dropdown-leave-to { opacity: 0; transform: translateY(-6px); }

/* ── Waffle / App launcher ──────────────────────────────────── */
.waffle-wrap { position: relative; }
.waffle-btn {
  width: 38px; height: 38px; border-radius: 50%;
  background: none; border: none; cursor: pointer;
  color: var(--text-muted); display: flex; align-items: center; justify-content: center;
  transition: background .15s, color .15s;
}
.waffle-btn:hover { background: var(--bg); color: var(--text); }
.waffle-btn.active {
  background: var(--blue-light); color: var(--blue);
}

/* Panel shell */
.apps-panel {
  position: absolute; top: calc(100% + 10px); right: 0;
  background: #fff; border: 1px solid var(--border);
  border-radius: 18px;
  box-shadow: 0 16px 48px rgba(0,0,0,0.14), 0 2px 8px rgba(0,0,0,0.06);
  width: 316px; z-index: 300; overflow: hidden;
}

/* Panel header */
.apps-panel-header {
  display: flex; align-items: center; justify-content: space-between;
  padding: 15px 18px 13px;
  border-bottom: 1px solid var(--border);
  background: #fafbff;
}
.apps-panel-brand {
  display: flex; align-items: center; gap: 8px;
  font-size: 14px; font-weight: 700; color: var(--text);
  letter-spacing: -0.2px;
}
.apps-panel-label {
  font-size: 11px; font-weight: 600;
  color: var(--text-muted); letter-spacing: 0.02em;
  background: var(--bg); padding: 3px 9px; border-radius: 20px;
}

/* App grid */
.apps-grid {
  display: grid; grid-template-columns: repeat(3, 1fr);
  gap: 2px; padding: 16px 12px 12px;
}

/* Individual tile */
.app-tile {
  display: flex; flex-direction: column; align-items: center; gap: 9px;
  padding: 16px 8px 14px; border-radius: 14px; text-decoration: none;
  transition: background .14s, transform .14s, box-shadow .14s;
  cursor: pointer; position: relative;
}
.app-tile:hover {
  background: var(--bg);
  transform: translateY(-2px);
  box-shadow: 0 4px 16px rgba(0,0,0,0.07);
  text-decoration: none;
}

/* Icon block */
.app-icon-wrap {
  width: 58px; height: 58px; border-radius: 16px;
  display: flex; align-items: center; justify-content: center;
  transition: transform .14s;
}
.app-tile:hover .app-icon-wrap { transform: scale(1.06); }
.app-letter {
  font-size: 17px; font-weight: 800; color: #fff;
  letter-spacing: -0.5px; user-select: none;
}

/* Name */
.app-tile-name {
  font-size: 11.5px; font-weight: 600; color: var(--text);
  text-align: center; line-height: 1.25;
  white-space: nowrap; overflow: hidden; text-overflow: ellipsis;
  max-width: 80px;
}

/* Badge */
.app-tile-badge {
  display: flex; align-items: center; gap: 4px;
  font-size: 9.5px; font-weight: 600; color: #7d5000;
  background: var(--warning-bg); padding: 2px 7px 2px 5px;
  border-radius: 20px; white-space: nowrap;
}
.badge-dot {
  width: 5px; height: 5px; border-radius: 50%;
  background: var(--warning); flex-shrink: 0;
  animation: pulse-dot 2s ease-in-out infinite;
}
@keyframes pulse-dot {
  0%, 100% { opacity: 1; transform: scale(1); }
  50%       { opacity: 0.5; transform: scale(0.75); }
}

/* Footer */
.apps-panel-footer {
  border-top: 1px solid var(--border);
  padding: 10px 18px;
  font-size: 11.5px; color: var(--text-muted);
  text-align: center; background: #fafbff;
}

/* ── Auth form globals ──────────────────────────────────────── */
.form-header { margin-bottom: 32px; }
.form-header h1 { font-size: 26px; font-weight: 500; color: var(--text); letter-spacing: -0.3px; margin-bottom: 6px; }
.form-header p  { font-size: 15px; color: var(--text-muted); }

.field { margin-bottom: 20px; }
.field label { display: block; font-size: 13px; font-weight: 500; color: #3c4043; margin-bottom: 7px; }
.field .input-wrap { position: relative; }
.field input {
  width: 100%; height: 52px; padding: 0 48px 0 14px;
  border: 1.5px solid var(--border); border-radius: var(--radius);
  font-size: 15px; color: var(--text); background: var(--surface);
  outline: none; transition: border-color .15s, box-shadow .15s; font-family: inherit;
}
.field input:focus { border-color: var(--blue); box-shadow: 0 0 0 3px rgba(26,115,232,.10); }
.field input.is-error { border-color: var(--error); box-shadow: 0 0 0 3px rgba(217,48,37,.08); }
.field input::placeholder { color: var(--text-hint); }

.pw-toggle {
  position: absolute; right: 13px; top: 50%; transform: translateY(-50%);
  background: none; border: none; cursor: pointer; color: var(--text-muted);
  padding: 4px; display: flex; align-items: center; border-radius: 4px;
  transition: color .15s, background .15s;
}
.pw-toggle:hover { color: var(--blue); background: var(--blue-light); }

.pw-strength { margin-top: 8px; display: flex; align-items: center; gap: 8px; }
.pw-bars { display: flex; gap: 3px; flex: 1; }
.pw-bar { height: 3px; flex: 1; border-radius: 2px; background: var(--border); transition: background .3s; }
.pw-bar.lvl-1 { background: var(--error); }
.pw-bar.lvl-2 { background: var(--warning); }
.pw-bar.lvl-3 { background: #fbbc04; }
.pw-bar.lvl-4 { background: var(--success); }
.pw-strength-lbl { font-size: 11px; font-weight: 500; min-width: 44px; text-align: right; }
.pw-strength-lbl.s1 { color: var(--error); }
.pw-strength-lbl.s2 { color: var(--warning); }
.pw-strength-lbl.s3 { color: #e37400; }
.pw-strength-lbl.s4 { color: var(--success); }

.field-err { font-size: 12px; color: var(--error); margin-top: 5px; }

.alert {
  display: flex; align-items: flex-start; gap: 10px;
  padding: 12px 14px; border-radius: var(--radius);
  font-size: 13.5px; line-height: 1.5; margin-bottom: 20px;
}
.alert-error   { background: var(--error-bg);   color: #c5221f; border-left: 3px solid var(--error); }
.alert-success { background: var(--success-bg); color: #137333; border-left: 3px solid var(--success); }
.alert-warning { background: var(--warning-bg); color: #7d5000; border-left: 3px solid var(--warning); }

.form-actions {
  display: flex; align-items: center;
  justify-content: space-between; margin-top: 28px; gap: 12px;
}
.form-actions.right { justify-content: flex-end; }

.link-btn {
  font-size: 14px; font-weight: 500; color: var(--blue);
  text-decoration: none; cursor: pointer; background: none;
  border: none; padding: 0; flex-shrink: 0;
}
.link-btn:hover { text-decoration: underline; }

.btn-primary {
  display: inline-flex; align-items: center; justify-content: center; gap: 6px;
  background: var(--blue); color: #fff; border: none;
  border-radius: var(--radius); height: 40px; padding: 0 22px;
  font-size: 14px; font-weight: 500; cursor: pointer; letter-spacing: 0.2px;
  transition: background .18s, box-shadow .18s; flex-shrink: 0; white-space: nowrap;
}
.btn-primary:hover:not(:disabled) { background: var(--blue-dark); box-shadow: 0 2px 8px rgba(26,115,232,.35); }
.btn-primary:active:not(:disabled) { background: var(--blue-darker); }
.btn-primary:disabled { opacity: 0.6; cursor: not-allowed; }

.link-subtle { font-size: 13px; color: var(--blue); text-decoration: none; display: inline-block; margin-top: 7px; }
.link-subtle:hover { text-decoration: underline; }

.dev-box {
  background: var(--warning-bg); border: 1px solid var(--warning);
  border-radius: var(--radius); padding: 12px 14px; margin-bottom: 20px; font-size: 13px;
}
.dev-box strong { color: #7d5000; display: block; margin-bottom: 4px; }
.dev-box code { font-size: 11px; word-break: break-all; color: #5f4000; display: block; margin: 4px 0 8px; }
</style>
