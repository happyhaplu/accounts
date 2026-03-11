import axios from 'axios'

const api = axios.create({
  baseURL: '/api/v1',
  headers: { 'Content-Type': 'application/json' },
})

// Attach JWT to every outgoing request
api.interceptors.request.use((config) => {
  const token = localStorage.getItem('oc_token')
  if (token) config.headers.Authorization = `Bearer ${token}`
  return config
})

// On 401 — clear stored credentials and redirect to login
// But skip redirect for auth endpoints themselves (login, register, forgot-password)
// so that wrong-credential errors can be shown in the UI.
const AUTH_PATHS = ['/auth/login', '/auth/register', '/auth/forgot-password', '/auth/reset-password', '/auth/resend-verification', '/auth/verify-email-otp', '/auth/verify-reset-otp']
api.interceptors.response.use(
  (res) => res,
  (err) => {
    const url = err.config?.url ?? ''
    const isAuthEndpoint = AUTH_PATHS.some((p) => url.includes(p))
    if (err.response?.status === 401 && !isAuthEndpoint) {
      localStorage.removeItem('oc_token')
      localStorage.removeItem('oc_user')
      // Preserve redirect_uri for product-launch flows so the user is sent
      // back to the product app after re-authenticating.
      const params = new URLSearchParams(window.location.search)
      const redirectUri = params.get('redirect_uri')
      window.location.href = redirectUri
        ? '/login?redirect_uri=' + encodeURIComponent(redirectUri)
        : '/login'
    }
    return Promise.reject(err)
  },
)

// ── Named endpoint helpers ──────────────────────────────────────────────────
export const authAPI = {
  resendVerification: (data) => api.post('/auth/resend-verification', data),
  register:           (data) => api.post('/auth/register',            data),
  verifyEmailOTP:     (data) => api.post('/auth/verify-email-otp',    data),
  login:              (data) => api.post('/auth/login',               data),
  logout:             ()     => api.post('/auth/logout'),
  forgotPassword:     (data) => api.post('/auth/forgot-password',     data),
  verifyResetOTP:     (data) => api.post('/auth/verify-reset-otp',    data),
  resetPassword:      (data) => api.post('/auth/reset-password',      data),
  getProfile:      ()     => api.get('/profile'),
  setupProfile:    (data) => api.post('/profile',               data),
  changePassword:  (data) => api.post('/auth/change-password', data),

  // Workspace
  listWorkspaces:  ()           => api.get('/workspaces'),
  getWorkspace:    (id)         => id ? api.get(`/workspaces/${id}`) : api.get('/workspace'),
  createWorkspace: (data)       => api.post('/workspaces',                         data),
  addMember:       (id, data)   => api.post(`/workspaces/${id}/members`,           data),
  removeMember:    (id, userID) => api.delete(`/workspaces/${id}/members/${userID}`),

  // Invites
  sendInvite:    (id, data)          => api.post(`/workspaces/${id}/invites`,              data),
  listInvites:   (id)                => api.get(`/workspaces/${id}/invites`),
  revokeInvite:  (wsId, inviteId)    => api.delete(`/workspaces/${wsId}/invites/${inviteId}`),
  previewInvite: (token)             => api.get('/invites/preview',  { params: { token } }),
  acceptInvite:  (token)             => api.post('/invites/accept',  { token }),

  // Subscriptions
  listSubscriptions:   (wsId)         => api.get(`/workspaces/${wsId}/subscriptions`),
  checkAccess:         (wsId, prodId) => api.get(`/workspaces/${wsId}/subscriptions/access`, { params: { product_id: prodId } }),
  cancelSubscription:  (wsId, subId)  => api.delete(`/workspaces/${wsId}/subscriptions/${subId}`),

  // Products
  listProducts:  () => api.get('/products'),
  // Launch a product — backend validates subscription and returns the callback URL.
  // Call launchProduct(name) from the dashboard (uses DB-configured URL).
  // Call launchProduct(name, 'https://warmup.outcraftly.com/callback') when
  // the product app redirected the user here with a specific redirect_uri.
  launchProduct: (name, redirectUri) =>
    api.get(`/products/${name}/launch`, {
      params: redirectUri ? { redirect_uri: redirectUri } : {},
    }),

  // Billing (Stripe)
  getBillingStatus:       (wsId)       => api.get(`/workspaces/${wsId}/billing`),
  createCheckoutSession:  (wsId, data) => api.post(`/workspaces/${wsId}/billing/checkout`, data),
  createPortalSession:    (wsId, data) => api.post(`/workspaces/${wsId}/billing/portal`, data),
  syncBilling:            (wsId)       => api.post(`/workspaces/${wsId}/billing/sync`),
}

export default api
