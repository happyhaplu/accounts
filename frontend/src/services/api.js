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
  // Call launchProduct(name, 'https://app.example.com/callback') when
  // the product app redirected the user here with a specific redirect_uri.
  launchProduct: (name, redirectUri) =>
    api.get(`/products/${name}/launch`, {
      params: redirectUri ? { redirect_uri: redirectUri } : {},
    }),

  // Billing (Stripe)
  getBillingStatus:       (wsId)       => api.get(`/workspaces/${wsId}/billing`),
  createPortalSession:    (wsId, data) => api.post(`/workspaces/${wsId}/billing/portal`, data),
  syncBilling:            (wsId)       => api.post(`/workspaces/${wsId}/billing/sync`),

  // Public config
  getConfig: () => api.get('/config'),
}

export default api

// ── Admin API ────────────────────────────────────────────────────────────────
// Separate axios instance that automatically injects X-Admin-Secret header
// from sessionStorage. On 403 it clears the session and redirects to login.

const adminAxios = axios.create({ baseURL: '/api/v1' })

adminAxios.interceptors.request.use((config) => {
  const secret = sessionStorage.getItem('admin_secret')
  if (secret) config.headers['X-Admin-Secret'] = secret
  // Only set JSON content-type for non-FormData requests.
  // For FormData (e.g. logo upload) Axios must set it automatically so that
  // the correct multipart boundary is included in the header.
  if (!(config.data instanceof FormData)) {
    config.headers['Content-Type'] = 'application/json'
  }
  return config
})

adminAxios.interceptors.response.use(
  (res) => res,
  (err) => {
    if (err.response?.status === 403 || err.response?.status === 401) {
      // Session expired or secret invalid — force re-login
      sessionStorage.removeItem('admin_secret')
      sessionStorage.removeItem('admin_email')
      if (!window.location.pathname.startsWith('/admin/login')) {
        window.location.href = '/admin/login'
      }
    }
    return Promise.reject(err)
  },
)

export const adminAPI = {
  // Auth (public — no X-Admin-Secret needed)
  login: (data) => api.post('/admin/auth/login', data),

  // Products
  listProducts:           ()         => adminAxios.get('/admin/products'),
  createProduct:          (data)     => adminAxios.post('/admin/products', data),
  updateProduct:          (id, data) => adminAxios.patch(`/admin/products/${id}`, data),
  deactivateProduct:      (id)       => adminAxios.delete(`/admin/products/${id}`),
  permanentDeleteProduct: (id)       => adminAxios.delete(`/admin/products/${id}/permanent`),
  regenerateProductKey:   (id)       => adminAxios.post(`/admin/products/${id}/regenerate-key`),
  // Logo upload — send a FormData object with a "logo" file field.
  // Returns { product, logo_url }
  // Do NOT set Content-Type manually — Axios must auto-set it with the
  // multipart boundary (e.g. "multipart/form-data; boundary=----Xyz").
  // Setting it manually strips the boundary and Fiber can't parse the file.
  uploadProductLogo: (id, formData) =>
    adminAxios.post(`/admin/products/${id}/logo`, formData),

  // Users
  listUsers:             (params) => adminAxios.get('/admin/users',                  { params }),
  purgeUnverifiedUsers:  ()       => adminAxios.delete('/admin/users/purge-unverified'),

  // Workspaces
  listWorkspaces: (params) => adminAxios.get('/admin/workspaces', { params }),

  // Subscriptions (read-only for now)
  listSubscriptions: (params)   => adminAxios.get('/admin/subscriptions', { params }),

  // Billing overview
  billingOverview: () => adminAxios.get('/admin/billing'),
}
