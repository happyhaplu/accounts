<template>
  <div class="auth-shell">
    <!-- ── Left: Branding panel ──────────────────────────── -->
    <aside class="auth-panel">
      <div class="panel-inner">
        <div class="panel-logo">
          <div class="logo-badge">
            <img class="logo-img" src="/icon.svg" alt="Gour" />
          </div>
          <span>Outcraftly</span>
        </div>

        <div class="panel-body">
          <h2>One account.<br />Everything<br />connected.</h2>
          <p>Access all Outcraftly products and services with a single, secure sign-in.</p>

          <ul class="feature-list">
            <li v-for="f in features" :key="f">
              <span class="feat-check">
                <svg width="11" height="11" viewBox="0 0 12 12" fill="none">
                  <path d="M1.5 6L4.5 9L10.5 3" stroke="white" stroke-width="2"
                        stroke-linecap="round" stroke-linejoin="round"/>
                </svg>
              </span>
              {{ f }}
            </li>
          </ul>
        </div>

        <div class="panel-footer">© 2026 Outcraftly, Inc.</div>
      </div>
    </aside>

    <!-- ── Right: Form area ──────────────────────────────── -->
    <main class="auth-main">
      <div class="auth-form-box">
        <slot />
      </div>
    </main>
  </div>
</template>

<script setup>
const features = [
  'Centralized identity & access',
  'Workspace & team management',
  'Billing & subscriptions',
]
</script>

<style scoped>
.auth-shell {
  display: flex;
  min-height: 100vh;
}

/* ── Left panel ──────────────────────────────────────────── */
.auth-panel {
  width: 420px;
  flex-shrink: 0;
  background: linear-gradient(155deg, #1565c0 0%, #0d47a1 50%, #082d7a 100%);
  padding: 52px 44px;
  display: flex;
  flex-direction: column;
  position: relative;
  overflow: hidden;
}
/* Decorative circles */
.auth-panel::before {
  content: '';
  position: absolute;
  top: -130px; right: -80px;
  width: 360px; height: 360px;
  border-radius: 50%;
  background: rgba(255,255,255,0.045);
  pointer-events: none;
}
.auth-panel::after {
  content: '';
  position: absolute;
  bottom: -100px; left: -60px;
  width: 300px; height: 300px;
  border-radius: 50%;
  background: rgba(255,255,255,0.035);
  pointer-events: none;
}

.panel-inner {
  position: relative;
  z-index: 1;
  display: flex;
  flex-direction: column;
  height: 100%;
}

/* Logo row */
.panel-logo {
  display: flex;
  align-items: center;
  gap: 11px;
  margin-bottom: 80px;
}
.logo-badge {
  width: 32px;
  height: 32px;
  border-radius: 8px;
  overflow: hidden;
  flex-shrink: 0;
  background: #fff;
  box-shadow: 0 2px 8px rgba(0,0,0,0.18);
}
.logo-img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  display: block;
}
.panel-logo span {
  font-size: 17px;
  font-weight: 700;
  color: #fff;
  letter-spacing: -0.3px;
}

/* Body */
.panel-body {
  flex: 1;
  display: flex;
  flex-direction: column;
  justify-content: center;
}
.panel-body h2 {
  font-size: 2.15rem;
  font-weight: 700;
  color: #fff;
  line-height: 1.18;
  letter-spacing: -0.6px;
  margin-bottom: 18px;
}
.panel-body p {
  font-size: 14px;
  color: rgba(255,255,255,0.68);
  line-height: 1.65;
  margin-bottom: 40px;
  max-width: 280px;
}

/* Features */
.feature-list {
  list-style: none;
  padding: 0;
  margin: 0;
  display: flex;
  flex-direction: column;
  gap: 15px;
}
.feature-list li {
  display: flex;
  align-items: center;
  gap: 11px;
  font-size: 13.5px;
  color: rgba(255,255,255,0.82);
}
.feat-check {
  width: 20px; height: 20px;
  background: rgba(255,255,255,0.16);
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

/* Footer */
.panel-footer {
  margin-top: 48px;
  font-size: 12px;
  color: rgba(255,255,255,0.36);
}

/* ── Right: form area ────────────────────────────────────── */
.auth-main {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #fff;
  padding: 48px 32px;
  min-height: 100vh;
}
.auth-form-box {
  width: 100%;
  max-width: 400px;
}

/* ── Responsive ──────────────────────────────────────────── */
@media (max-width: 820px) {
  .auth-shell { flex-direction: column; }
  .auth-panel {
    width: 100%;
    padding: 24px 20px 20px;
    min-height: unset;
  }
  .panel-logo { margin-bottom: 0; }
  .panel-body, .panel-footer { display: none; }
  .auth-main { padding: 32px 20px; min-height: unset; }
}
</style>
