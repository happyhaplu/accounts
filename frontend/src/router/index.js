import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '../stores/auth'

const routes = [
  { path: '/', redirect: '/dashboard' },
  {
    path: '/login',
    name: 'Login',
    component: () => import('../views/LoginView.vue'),
    meta: { guest: true },
  },
  {
    path: '/register',
    name: 'Register',
    component: () => import('../views/RegisterView.vue'),
    meta: { guest: true },
  },
  {
    path: '/forgot-password',
    name: 'ForgotPassword',
    component: () => import('../views/ForgotPasswordView.vue'),
    meta: { guest: true },
  },
  {
    path: '/reset-password',
    name: 'ResetPassword',
    component: () => import('../views/ResetPasswordView.vue'),
    meta: { guest: true },
  },
  // Public — handles verification token from email link
  {
    path: '/verify-email',
    name: 'VerifyEmail',
    component: () => import('../views/VerifyEmailView.vue'),
  },
  // Protected — required after email verification
  {
    path: '/profile-setup',
    name: 'ProfileSetup',
    component: () => import('../views/ProfileSetupView.vue'),
    meta: { auth: true },
  },
  {
    path: '/dashboard',
    name: 'Dashboard',
    component: () => import('../views/DashboardView.vue'),
    meta: { auth: true },
  },
  {
    path: '/settings',
    name: 'Settings',
    component: () => import('../views/SettingsView.vue'),
    meta: { auth: true },
  },
  {
    path: '/workspaces',
    name: 'Workspaces',
    component: () => import('../views/WorkspacesView.vue'),
    meta: { auth: true },
  },
  {
    path: '/billing',
    name: 'Billing',
    component: () => import('../views/BillingView.vue'),
    meta: { auth: true },
  },
  // Product launch bridge — orchestrates auth check → API call → redirect to
  // the product app.  Handles auth internally (no meta.auth) so it can redirect
  // to /login?redirect_uri=… and preserve the callback URL.
  {
    path: '/products/:slug/launch',
    name: 'ProductLaunch',
    component: () => import('../views/ProductLaunchView.vue'),
  },
  // Logout — clears auth state and optionally redirects back to a product.
  // Used by external products: /logout?redirect_uri=https://warmup.outcraftly.com
  {
    path: '/logout',
    name: 'Logout',
    component: () => import('../views/LogoutView.vue'),
  },
  // Public — handles workspace invite accept link from email
  {
    path: '/invite',
    name: 'InviteAccept',
    component: () => import('../views/InviteAcceptView.vue'),
  },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
})

router.beforeEach((to) => {
  const auth = useAuthStore()

  // Auth-required route — must be logged in
  if (to.meta.auth && !auth.isAuthenticated) {
    return { name: 'Login' }
  }

  // Guest-only route — logged-in users get redirected
  if (to.meta.guest && auth.isAuthenticated) {
    if (!auth.user?.profile_complete) return { name: 'ProfileSetup' }
    return { name: 'Dashboard' }
  }

  // Logged in but profile incomplete — force profile setup
  // (except when already going to profile-setup)
  if (auth.isAuthenticated && !auth.user?.profile_complete && to.meta.auth && to.name !== 'ProfileSetup') {
    return { name: 'ProfileSetup' }
  }
})

export default router
