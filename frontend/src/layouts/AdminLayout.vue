<template>
  <div class="admin-shell">

    <!-- Admin Top Navigation -->
    <header class="admin-nav">
      <div class="admin-nav-left">
        <img src="/logo.svg" alt="Gour" class="admin-nav-logo" />
        <span class="admin-nav-sep">/</span>
        <span class="admin-nav-label">Admin</span>
      </div>

      <nav class="admin-nav-links">
        <RouterLink to="/admin/products" class="admin-nav-link">
          <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor"
               stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <rect x="1" y="3" width="15" height="13" rx="2"/>
            <path d="M16 8h6l-3 5-3-5z"/>
            <line x1="1" y1="21" x2="6" y2="21"/>
            <line x1="1" y1="17" x2="11" y2="17"/>
          </svg>
          Products
        </RouterLink>
        <RouterLink to="/admin/users" class="admin-nav-link">
          <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor"
               stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"/>
            <circle cx="9" cy="7" r="4"/>
            <path d="M23 21v-2a4 4 0 0 0-3-3.87"/>
            <path d="M16 3.13a4 4 0 0 1 0 7.75"/>
          </svg>
          Users
        </RouterLink>
        <RouterLink to="/admin/workspaces" class="admin-nav-link">
          <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor"
               stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <rect x="2" y="3" width="20" height="14" rx="2" ry="2"/>
            <line x1="8" y1="21" x2="16" y2="21"/>
            <line x1="12" y1="17" x2="12" y2="21"/>
          </svg>
          Workspaces
        </RouterLink>
      </nav>

      <div class="admin-nav-right">
        <span class="admin-nav-email">{{ admin.email }}</span>
        <button class="admin-logout-btn" @click="handleLogout">
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor"
               stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4"/>
            <polyline points="16 17 21 12 16 7"/>
            <line x1="21" y1="12" x2="9" y2="12"/>
          </svg>
          Sign out
        </button>
      </div>
    </header>

    <!-- Page content -->
    <main class="admin-main">
      <RouterView />
    </main>

  </div>
</template>

<script setup>
import { useRouter } from 'vue-router'
import { useAdminStore } from '../stores/admin'

const router = useRouter()
const admin  = useAdminStore()

function handleLogout() {
  admin.logout()
  router.push('/admin/login')
}
</script>

<style scoped>
.admin-shell {
  min-height: 100vh;
  background: #f1f3f4;
  display: flex;
  flex-direction: column;
}

/* ── Admin Navbar ─────────────────────────────────────────────────────────── */
.admin-nav {
  position: sticky;
  top: 0;
  z-index: 200;
  height: 60px;
  background: #1a2332;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 24px;
  gap: 16px;
  border-bottom: 1px solid rgba(255,255,255,0.08);
  box-shadow: 0 2px 8px rgba(0,0,0,0.25);
}

.admin-nav-left {
  display: flex;
  align-items: center;
  gap: 10px;
}

.admin-nav-logo {
  height: 26px;
  width: auto;
  filter: brightness(0) invert(1);
  opacity: 0.9;
}

.admin-nav-sep {
  color: rgba(255,255,255,0.3);
  font-size: 18px;
  font-weight: 300;
  line-height: 1;
}

.admin-nav-label {
  font-size: 13px;
  font-weight: 600;
  color: rgba(255,255,255,0.55);
  letter-spacing: 0.06em;
  text-transform: uppercase;
}

/* Nav links */
.admin-nav-links {
  display: flex;
  align-items: center;
  gap: 4px;
}

.admin-nav-link {
  display: flex;
  align-items: center;
  gap: 7px;
  padding: 7px 14px;
  border-radius: 6px;
  font-size: 13.5px;
  font-weight: 500;
  color: rgba(255,255,255,0.65);
  text-decoration: none;
  transition: background 0.15s, color 0.15s;
}
.admin-nav-link:hover {
  background: rgba(255,255,255,0.08);
  color: #fff;
  text-decoration: none;
}
.admin-nav-link.router-link-active {
  background: rgba(255,255,255,0.12);
  color: #fff;
}

/* Right section */
.admin-nav-right {
  display: flex;
  align-items: center;
  gap: 14px;
}

.admin-nav-email {
  font-size: 13px;
  color: rgba(255,255,255,0.5);
  max-width: 220px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.admin-logout-btn {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 6px 12px;
  border-radius: 6px;
  background: rgba(255,255,255,0.07);
  border: 1px solid rgba(255,255,255,0.12);
  color: rgba(255,255,255,0.7);
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  font-family: inherit;
  transition: background 0.15s, color 0.15s;
}
.admin-logout-btn:hover {
  background: rgba(255,255,255,0.14);
  color: #fff;
}

/* ── Main content ─────────────────────────────────────────────────────────── */
.admin-main {
  flex: 1;
  padding: 32px 32px 48px;
  max-width: 1200px;
  width: 100%;
  margin: 0 auto;
  box-sizing: border-box;
}
</style>
