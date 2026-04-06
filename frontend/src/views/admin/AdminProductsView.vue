<template>
  <div class="ap-page">

    <!-- Page header -->
    <div class="ap-header">
      <div class="ap-header-left">
        <h1 class="ap-title">Products</h1>
        <p class="ap-sub">Manage the Gour product registry. Changes take effect immediately.</p>
      </div>
      <button class="btn-primary" @click="openCreate">
        <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor"
             stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round">
          <line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/>
        </svg>
        New Product
      </button>
    </div>

    <!-- Stats row — cards are clickable filters -->
    <div class="ap-stats">
      <button :class="['stat-card', activeFilter === 'active' && 'stat-active']" @click="activeFilter = 'active'">
        <span class="stat-num">{{ activeCount }}</span>
        <span class="stat-label">Active</span>
      </button>
      <button :class="['stat-card', activeFilter === 'inactive' && 'stat-active']" @click="activeFilter = 'inactive'">
        <span class="stat-num">{{ inactiveCount }}</span>
        <span class="stat-label">Inactive</span>
      </button>
      <button :class="['stat-card', activeFilter === 'all' && 'stat-active']" @click="activeFilter = 'all'">
        <span class="stat-num">{{ products.length }}</span>
        <span class="stat-label">Total</span>
      </button>
    </div>

    <!-- Search + filter bar -->
    <div class="ap-toolbar">
      <div class="search-wrap">
        <svg class="search-icon" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round">
          <circle cx="11" cy="11" r="8"/><line x1="21" y1="21" x2="16.65" y2="16.65"/>
        </svg>
        <input v-model="search" class="search-input" type="text" placeholder="Search by name or description…" />
        <button v-if="search" class="search-clear" @click="search = ''" title="Clear">
          <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
        </button>
      </div>
      <span class="ap-count">{{ filteredProducts.length }} of {{ products.length }}</span>
    </div>

    <!-- Loading -->
    <div v-if="loading" class="ap-loading">
      <div class="spinner"></div>
      <span>Loading products…</span>
    </div>

    <!-- Error -->
    <div v-else-if="loadError" class="alert alert-error">
      <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor"
           stroke-width="2" stroke-linecap="round" stroke-linejoin="round" style="flex-shrink:0">
        <circle cx="12" cy="12" r="10"/><line x1="12" y1="8" x2="12" y2="12"/>
        <line x1="12" y1="16" x2="12.01" y2="16"/>
      </svg>
      {{ loadError }}
      <button class="link-btn" style="margin-left:auto" @click="fetchProducts">Retry</button>
    </div>

    <!-- Products table -->
    <div v-else class="ap-table-wrap">
      <table class="ap-table">
        <thead>
          <tr>
            <th>Product</th>
            <th>Redirect URLs</th>
            <th>Status</th>
            <th>API Key</th>
            <th>Actions</th>
          </tr>
        </thead>
        <tbody>
          <tr v-if="filteredProducts.length === 0">
            <td colspan="5" class="ap-empty">
              <svg width="32" height="32" viewBox="0 0 24 24" fill="none" stroke="#dadce0" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round" style="display:block;margin:0 auto 10px"><rect x="2" y="3" width="20" height="14" rx="2"/><line x1="8" y1="21" x2="16" y2="21"/><line x1="12" y1="17" x2="12" y2="21"/></svg>
              <span v-if="search">No products match <strong>"{{ search }}"</strong></span>
              <span v-else-if="activeFilter === 'active'">No active products yet.</span>
              <span v-else-if="activeFilter === 'inactive'">No inactive products.</span>
              <span v-else>No products registered yet.</span>
            </td>
          </tr>
          <tr v-for="p in filteredProducts" :key="p.id" :class="{ 'row-inactive': !p.is_active }">
            <td>
              <div class="product-name">
                {{ p.name }}
                <span v-if="!p.is_active && (!p.redirect_urls || !p.redirect_urls.length)" class="badge-legacy">legacy</span>
              </div>
              <div class="product-desc">{{ p.description || '—' }}</div>
            </td>
            <td>
              <div v-if="p.redirect_urls && p.redirect_urls.length" class="url-list">
                <span v-for="(u, i) in p.redirect_urls.slice(0, 2)" :key="i" class="url-chip">{{ u }}</span>
                <span v-if="p.redirect_urls.length > 2" class="url-more">+{{ p.redirect_urls.length - 2 }} more</span>
              </div>
              <span v-else class="text-muted">—</span>
            </td>
            <td>
              <span :class="['status-badge', p.is_active ? 'badge-active' : 'badge-inactive']">
                <span class="badge-dot"></span>
                {{ p.is_active ? 'Active' : 'Inactive' }}
              </span>
            </td>
            <!-- API Key column -->
            <td class="apikey-cell">
              <template v-if="p.api_key">
                <code class="apikey-chip">{{ p.api_key.slice(0, 18) }}…</code>
                <button class="btn-icon-copy" @click="copyKey(p.api_key, p.id)" :title="copied === p.id ? 'Copied!' : 'Copy full key'">
                  <svg v-if="copied !== p.id" width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="9" y="9" width="13" height="13" rx="2" ry="2"/><path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"/></svg>
                  <svg v-else width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="#137333" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><polyline points="20 6 9 17 4 12"/></svg>
                </button>
              </template>
              <span v-else class="text-muted">—</span>
            </td>
            <td class="actions-cell">
              <button class="btn-action btn-edit" @click="openEdit(p)" :disabled="toggling === p.id">
                <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor"
                     stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                  <path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"/>
                  <path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"/>
                </svg>
                Edit
              </button>
              <button
                :class="['btn-action', p.is_active ? 'btn-disable' : 'btn-enable']"
                @click="toggleActive(p)"
                :disabled="toggling === p.id"
              >
                <svg v-if="toggling === p.id" class="spin-icon" width="12" height="12" viewBox="0 0 24 24"
                     fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round">
                  <path d="M12 2v4M12 18v4M4.93 4.93l2.83 2.83M16.24 16.24l2.83 2.83M2 12h4M18 12h4M4.93 19.07l2.83-2.83M16.24 7.76l2.83-2.83"/>
                </svg>
                {{ p.is_active ? 'Disable' : 'Enable' }}
              </button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- Toast notification -->
    <transition name="toast-fade">
      <div v-if="toast.show" :class="['toast', 'toast-' + toast.type]">
        <svg v-if="toast.type === 'success'" width="15" height="15" viewBox="0 0 24 24" fill="none"
             stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round">
          <path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"/><polyline points="22 4 12 14.01 9 11.01"/>
        </svg>
        <svg v-else width="15" height="15" viewBox="0 0 24 24" fill="none"
             stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
          <circle cx="12" cy="12" r="10"/><line x1="12" y1="8" x2="12" y2="12"/>
          <line x1="12" y1="16" x2="12.01" y2="16"/>
        </svg>
        {{ toast.message }}
      </div>
    </transition>

    <!-- ── Create / Edit Modal ─────────────────────────────────────────────── -->
    <Teleport to="body">
      <transition name="modal-fade">
        <div v-if="modal.open" class="modal-overlay" @click.self="closeModal">
          <div class="modal-card" role="dialog" aria-modal="true">

            <!-- Modal header -->
            <div class="modal-header">
              <h2 class="modal-title">{{ modal.isEdit ? `Edit: ${modal.form.name}` : 'New Product' }}</h2>
              <button class="modal-close" @click="closeModal" aria-label="Close">
                <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor"
                     stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                  <line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/>
                </svg>
              </button>
            </div>

            <!-- Modal error -->
            <div v-if="modal.error" class="modal-alert">
              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor"
                   stroke-width="2" stroke-linecap="round" stroke-linejoin="round" style="flex-shrink:0">
                <circle cx="12" cy="12" r="10"/><line x1="12" y1="8" x2="12" y2="12"/>
                <line x1="12" y1="16" x2="12.01" y2="16"/>
              </svg>
              {{ modal.error }}
            </div>

            <!-- Form body -->
            <div class="modal-body">

              <!-- Name — read-only on edit -->
              <div class="mfield">
                <label>Product name (slug) <span class="req">*</span></label>
                <input
                  v-model="modal.form.name"
                  type="text"
                  placeholder="e.g. my-product"
                  :class="{ 'is-error': modal.errors.name }"
                  :readonly="modal.isEdit"
                  :disabled="modal.saving"
                />
                <p v-if="modal.errors.name" class="field-err">{{ modal.errors.name }}</p>
                <p v-if="!modal.isEdit" class="field-hint">Lowercase letters, numbers, hyphens only (e.g. <code>my-product</code>). Cannot be changed after creation.</p>
              </div>

              <!-- Description -->
              <div class="mfield">
                <label>Description</label>
                <textarea
                  v-model="modal.form.description"
                  rows="2"
                  placeholder="Short description shown to users"
                  :disabled="modal.saving"
                ></textarea>
              </div>

              <!-- Redirect URLs -->
              <div class="mfield">
                <label>Redirect URLs <span class="field-hint-inline">(one per line)</span></label>
                <textarea
                  v-model="modal.form.redirectUrlsText"
                  rows="4"
                  placeholder="http://localhost:3000/callback&#10;https://app.example.com/callback"
                  :disabled="modal.saving"
                  class="mono-input"
                ></textarea>
                <p class="field-hint">Full callback URLs that this product is allowed to redirect to after launch. Scheme + host is used for CORS validation.</p>
              </div>

              <!-- API Key (edit mode only) -->
              <div v-if="modal.isEdit" class="mfield mfield-apikey">
                <label>API Key</label>
                <div class="apikey-display">
                  <code class="apikey-full">{{ modal.form.api_key }}</code>
                  <button class="btn-icon-copy" @click="copyKey(modal.form.api_key, 'modal')" :title="copied === 'modal' ? 'Copied!' : 'Copy'">
                    <svg v-if="copied !== 'modal'" width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="9" y="9" width="13" height="13" rx="2" ry="2"/><path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"/></svg>
                    <svg v-else width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="#137333" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><polyline points="20 6 9 17 4 12"/></svg>
                  </button>
                  <button class="btn-regen" @click="regenKey" :disabled="modal.regenerating || modal.saving">
                    <svg v-if="modal.regenerating" class="spin-icon" width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round"><path d="M12 2v4M12 18v4M4.93 4.93l2.83 2.83M16.24 16.24l2.83 2.83M2 12h4M18 12h4M4.93 19.07l2.83-2.83M16.24 7.76l2.83-2.83"/></svg>
                    <svg v-else width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="23 4 23 10 17 10"/><path d="M20.49 15a9 9 0 1 1-2.12-9.36L23 10"/></svg>
                    {{ modal.regenerating ? 'Regenerating…' : 'Regenerate' }}
                  </button>
                </div>
                <p class="field-hint apikey-hint">
                  <strong>Put in product .env:</strong>
                  <code class="env-snippet">ACCOUNTS_API_KEY={{ modal.form.api_key }}</code>
                  Regenerating immediately invalidates the old key.
                </p>
              </div>


            </div><!-- end modal-body -->

            <!-- Modal footer -->
            <div class="modal-footer">
              <div v-if="modal.isEdit" class="modal-footer-danger">
                <template v-if="!modal.deleteConfirm">
                  <button
                    class="btn-danger-ghost"
                    @click="confirmDeactivate"
                    :disabled="modal.saving || modal.deleting || !modal.form.is_active"
                    title="Deactivate — keeps subscription history"
                  >
                    {{ modal.form.is_active ? 'Deactivate' : 'Inactive' }}
                  </button>
                  <button
                    class="btn-danger-ghost"
                    @click="modal.deleteConfirm = true"
                    :disabled="modal.saving || modal.deleting"
                    title="Permanently delete this product and all its subscriptions"
                  >
                    <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                      <polyline points="3 6 5 6 21 6"/>
                      <path d="M19 6l-1 14a2 2 0 0 1-2 2H8a2 2 0 0 1-2-2L5 6"/>
                      <path d="M10 11v6M14 11v6"/>
                      <path d="M9 6V4a1 1 0 0 1 1-1h4a1 1 0 0 1 1 1v2"/>
                    </svg>
                    Delete
                  </button>
                </template>
                <template v-else>
                  <span class="modal-delete-confirm">
                    <span class="modal-delete-label">Delete <strong>"{{ modal.form.name }}"</strong>?</span>
                    <button class="btn-confirm-delete-sm" @click="permanentDeleteFromModal" :disabled="modal.deleting">
                      <svg v-if="modal.deleting" class="spin-icon" width="11" height="11" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round">
                        <path d="M12 2v4M12 18v4M4.93 4.93l2.83 2.83M16.24 16.24l2.83 2.83M2 12h4M18 12h4M4.93 19.07l2.83-2.83M16.24 7.76l2.83-2.83"/>
                      </svg>
                      {{ modal.deleting ? 'Deleting…' : 'Yes, delete' }}
                    </button>
                    <button class="btn-cancel-sm" @click="modal.deleteConfirm = false" :disabled="modal.deleting">Cancel</button>
                  </span>
                </template>
              </div>
              <div class="modal-footer-right">
                <button class="btn-ghost" @click="closeModal" :disabled="modal.saving || modal.deleting">Cancel</button>
                <button class="btn-primary" @click="saveProduct" :disabled="modal.saving || modal.deleting">
                  <svg v-if="modal.saving" class="spin-icon" width="14" height="14" viewBox="0 0 24 24"
                       fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round">
                    <path d="M12 2v4M12 18v4M4.93 4.93l2.83 2.83M16.24 16.24l2.83 2.83M2 12h4M18 12h4M4.93 19.07l2.83-2.83M16.24 7.76l2.83-2.83"/>
                  </svg>
                  {{ modal.saving ? 'Saving…' : (modal.isEdit ? 'Save changes' : 'Create product') }}
                </button>
              </div>
            </div>

          </div>
        </div>
      </transition>
    </Teleport>

  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { adminAPI } from '../../services/api'

// ── State ─────────────────────────────────────────────────────────────────────
const products  = ref([])
const loading   = ref(false)
const loadError = ref('')
const toggling = ref(null)  // product id currently being toggled
const copied   = ref(null)  // id or 'modal' when copy-flash is active

// Filter & search
const activeFilter = ref('active')  // 'active' | 'inactive' | 'all' — default hides legacy inactive
const search       = ref('')

const activeCount   = computed(() => products.value.filter(p => p.is_active).length)
const inactiveCount = computed(() => products.value.filter(p => !p.is_active).length)

const filteredProducts = computed(() => {
  let list = products.value
  if (activeFilter.value === 'active')   list = list.filter(p => p.is_active)
  if (activeFilter.value === 'inactive') list = list.filter(p => !p.is_active)
  if (search.value.trim()) {
    const q = search.value.trim().toLowerCase()
    list = list.filter(p =>
      p.name.toLowerCase().includes(q) ||
      (p.description ?? '').toLowerCase().includes(q)
    )
  }
  return list
})

// ── Toast ─────────────────────────────────────────────────────────────────────
const toast = reactive({ show: false, message: '', type: 'success', _t: null })

function showToast(message, type = 'success') {
  clearTimeout(toast._t)
  toast.message = message
  toast.type    = type
  toast.show    = true
  toast._t = setTimeout(() => { toast.show = false }, 3500)
}

// ── Modal state ───────────────────────────────────────────────────────────────
const emptyForm = () => ({
  id:               '',
  name:             '',
  description:      '',
  redirectUrlsText: '',
  is_active:        true,
  api_key:          '',
})

const modal = reactive({
  open:             false,
  isEdit:           false,
  saving:           false,
  regenerating:     false,
  deleteConfirm:    false,
  deleting:         false,
  error:            '',
  form:             emptyForm(),
  errors: { name: '' },
})

// ── Fetch products ────────────────────────────────────────────────────────────
async function fetchProducts() {
  loading.value   = true
  loadError.value = ''
  try {
    const { data } = await adminAPI.listProducts()
    products.value = data.products ?? []
  } catch (err) {
    loadError.value = err.response?.data?.error ?? 'Failed to load products. Please try again.'
  } finally {
    loading.value = false
  }
}

onMounted(fetchProducts)

// ── Open create modal ─────────────────────────────────────────────────────────
function openCreate() {
  Object.assign(modal, { open: true, isEdit: false, saving: false, error: '' })
  Object.assign(modal.form, emptyForm())
  Object.assign(modal.errors, { name: '' })
}

// ── Open edit modal ───────────────────────────────────────────────────────────
function openEdit(product) {
  Object.assign(modal, { open: true, isEdit: true, saving: false, regenerating: false, deleteConfirm: false, deleting: false, error: '' })
  Object.assign(modal.form, {
    id:               product.id,
    name:             product.name,
    description:      product.description ?? '',
    redirectUrlsText: (product.redirect_urls ?? []).join('\n'),
    is_active:        product.is_active,
    api_key:          product.api_key ?? '',
  })
  Object.assign(modal.errors, { name: '' })
}

function closeModal() {
  if (modal.saving || modal.deleting) return
  modal.open = false
}

// ── Validate modal form ───────────────────────────────────────────────────────
function validateModal() {
  let ok = true
  modal.errors.name = ''

  if (!modal.isEdit) {
    const nameVal = modal.form.name.trim()
    if (!nameVal) {
      modal.errors.name = 'Product name is required'
      ok = false
    } else if (!/^[a-z0-9_-]+$/.test(nameVal)) {
      modal.errors.name = 'Only lowercase letters, numbers, hyphens and underscores allowed'
      ok = false
    }
  }

  return ok
}

// ── Parse redirect URLs from textarea ────────────────────────────────────────
function parseRedirectURLs(text) {
  return text
    .split('\n')
    .map(u => u.trim())
    .filter(u => u.length > 0)
}

// ── Save (create or update) ───────────────────────────────────────────────────
async function saveProduct() {
  modal.error = ''
  if (!validateModal()) return

  modal.saving = true
  const redirectUrls = parseRedirectURLs(modal.form.redirectUrlsText)

  try {
    if (modal.isEdit) {
      const payload = {
        description:   modal.form.description.trim(),
        redirect_urls: redirectUrls,
      }
      await adminAPI.updateProduct(modal.form.id, payload)
      showToast(`"${modal.form.name}" updated successfully`)
    } else {
      const payload = {
        name:          modal.form.name.trim().toLowerCase(),
        description:   modal.form.description.trim(),
        redirect_urls: redirectUrls,
      }
      await adminAPI.createProduct(payload)
      showToast(`"${payload.name}" created successfully`)
    }
    modal.open = false
    await fetchProducts()
  } catch (err) {
    const status = err.response?.status
    modal.error = err.response?.data?.error
      ?? (status === 409 ? 'A product with that name already exists.' : 'Failed to save. Please try again.')
  } finally {
    modal.saving = false
  }
}

// ── Toggle active / inactive ──────────────────────────────────────────────────
async function toggleActive(product) {
  if (toggling.value) return
  toggling.value = product.id
  try {
    if (product.is_active) {
      await adminAPI.deactivateProduct(product.id)
      showToast(`"${product.name}" has been disabled`)
    } else {
      await adminAPI.updateProduct(product.id, { is_active: true })
      showToast(`"${product.name}" has been enabled`)
    }
    await fetchProducts()
  } catch (err) {
    showToast(err.response?.data?.error ?? 'Failed to update product status', 'error')
  } finally {
    toggling.value = null
  }
}

// ── Permanent delete — triggered from inside the Edit modal ─────────────────
async function permanentDeleteFromModal() {
  if (modal.deleting) return
  modal.deleting = true
  modal.error    = ''
  try {
    await adminAPI.permanentDeleteProduct(modal.form.id)
    showToast(`"${modal.form.name}" permanently deleted`)
    modal.open = false
    await fetchProducts()
  } catch (err) {
    modal.error         = err.response?.data?.error ?? 'Failed to delete product. Try again.'
    modal.deleteConfirm = false
  } finally {
    modal.deleting = false
  }
}

// ── Copy API key to clipboard ─────────────────────────────────────────────────
function copyKey(key, id) {
  navigator.clipboard.writeText(key).then(() => {
    copied.value = id
    setTimeout(() => { if (copied.value === id) copied.value = null }, 2000)
  })
}

// ── Regenerate API key from inside the edit modal ────────────────────────────
async function regenKey() {
  if (!modal.form.id || modal.regenerating) return
  modal.regenerating = true
  modal.error = ''
  try {
    const { data } = await adminAPI.regenerateProductKey(modal.form.id)
    modal.form.api_key = data.product.api_key
    // Also refresh the table row in background
    await fetchProducts()
    showToast('API key regenerated — update your product .env now', 'success')
  } catch (err) {
    modal.error = err.response?.data?.error ?? 'Failed to regenerate API key'
  } finally {
    modal.regenerating = false
  }
}

// ── Deactivate from inside the edit modal ────────────────────────────────────
async function confirmDeactivate() {
  if (!modal.form.is_active) return
  modal.saving = true
  modal.error  = ''
  try {
    await adminAPI.deactivateProduct(modal.form.id)
    showToast(`"${modal.form.name}" has been disabled`)
    modal.open = false
    await fetchProducts()
  } catch (err) {
    modal.error = err.response?.data?.error ?? 'Failed to deactivate product'
  } finally {
    modal.saving = false
  }
}


</script>

<style scoped>
/* ── Page ─────────────────────────────────────────────────────────────────── */
.ap-page { padding: 0; }

/* Header */
.ap-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
  margin-bottom: 24px;
}
.ap-title { font-size: 22px; font-weight: 600; color: #202124; letter-spacing: -0.3px; margin-bottom: 4px; }
.ap-sub   { font-size: 13.5px; color: #5f6368; }

/* Stats */
.ap-stats {
  display: flex;
  gap: 12px;
  margin-bottom: 16px;
}
.stat-card {
  background: #fff;
  border: 1.5px solid #dadce0;
  border-radius: 8px;
  padding: 14px 22px;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 2px;
  min-width: 90px;
  cursor: pointer;
  font-family: inherit;
  transition: border-color 0.15s, box-shadow 0.15s, background 0.15s;
}
.stat-card:hover { border-color: #1a73e8; background: #f8f9ff; }
.stat-card.stat-active {
  border-color: #1a73e8;
  background: #e8f0fe;
  box-shadow: 0 0 0 3px rgba(26,115,232,0.1);
}
.stat-card.stat-active .stat-num   { color: #1557b0; }
.stat-card.stat-active .stat-label { color: #1a73e8; }
.stat-num   { font-size: 26px; font-weight: 700; color: #1a73e8; line-height: 1; }
.stat-label { font-size: 11.5px; font-weight: 500; color: #5f6368; text-transform: uppercase; letter-spacing: 0.05em; }

/* Toolbar */
.ap-toolbar {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 16px;
}
.search-wrap {
  position: relative;
  flex: 1;
  max-width: 380px;
}
.search-icon {
  position: absolute;
  left: 11px;
  top: 50%;
  transform: translateY(-50%);
  color: #9aa0a6;
  pointer-events: none;
}
.search-input {
  width: 100%;
  padding: 8px 32px 8px 34px;
  border: 1.5px solid #dadce0;
  border-radius: 7px;
  font-size: 13.5px;
  color: #202124;
  font-family: inherit;
  outline: none;
  background: #fff;
  transition: border-color 0.15s, box-shadow 0.15s;
}
.search-input:focus { border-color: #1a73e8; box-shadow: 0 0 0 3px rgba(26,115,232,0.1); }
.search-clear {
  position: absolute;
  right: 8px;
  top: 50%;
  transform: translateY(-50%);
  background: none;
  border: none;
  cursor: pointer;
  color: #9aa0a6;
  display: flex;
  padding: 3px;
  border-radius: 4px;
}
.search-clear:hover { color: #5f6368; background: #f1f3f4; }
.ap-count {
  font-size: 12.5px;
  color: #9aa0a6;
  white-space: nowrap;
  margin-left: auto;
}

/* Legacy badge */
.badge-legacy {
  display: inline-block;
  font-size: 10px;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  background: #fef3cd;
  color: #856404;
  border: 1px solid #fde68a;
  border-radius: 4px;
  padding: 1px 5px;
  margin-left: 6px;
  vertical-align: middle;
  font-family: inherit;
}

/* Loading */
.ap-loading {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 40px;
  color: #5f6368;
  font-size: 14px;
  justify-content: center;
}
.spinner {
  width: 22px; height: 22px;
  border: 2.5px solid #dadce0;
  border-top-color: #1a73e8;
  border-radius: 50%;
  animation: spin 0.75s linear infinite;
}
@keyframes spin { to { transform: rotate(360deg); } }

/* Alert */
.alert {
  display: flex; align-items: flex-start; gap: 10px;
  padding: 12px 14px; border-radius: 6px;
  font-size: 13.5px; line-height: 1.5; margin-bottom: 20px;
}
.alert-error  { background: #fce8e6; color: #c5221f; border-left: 3px solid #d93025; }
.link-btn { font-size: 13px; font-weight: 500; color: #1a73e8; background: none; border: none; cursor: pointer; padding: 0; }

/* Table */
.ap-table-wrap {
  background: #fff;
  border: 1px solid #dadce0;
  border-radius: 10px;
  overflow: hidden;
}
.ap-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 13.5px;
}
.ap-table thead tr {
  background: #f8f9fa;
  border-bottom: 1px solid #dadce0;
}
.ap-table th {
  padding: 11px 16px;
  text-align: left;
  font-size: 11.5px;
  font-weight: 600;
  color: #5f6368;
  text-transform: uppercase;
  letter-spacing: 0.06em;
  white-space: nowrap;
}
.ap-table td {
  padding: 14px 16px;
  border-bottom: 1px solid #f1f3f4;
  vertical-align: top;
}
.ap-table tbody tr:last-child td { border-bottom: none; }
.ap-table tbody tr:hover { background: #fafbff; }
.ap-table tbody tr.row-inactive { opacity: 0.55; }

.ap-empty { text-align: center; color: #5f6368; padding: 48px 32px !important; font-size: 14px; }

/* Product name/desc */
.product-name { font-weight: 600; color: #202124; font-family: 'SFMono-Regular', Consolas, monospace; font-size: 13px; margin-bottom: 3px; }
.product-desc { font-size: 12px; color: #5f6368; line-height: 1.4; }

.text-muted { color: #9aa0a6; }

/* API Key — table cell */
.apikey-cell { white-space: nowrap; }
.apikey-chip {
  font-family: 'SFMono-Regular', Consolas, monospace;
  font-size: 11px; color: #5f6368; background: #f1f3f4;
  padding: 2px 6px; border-radius: 4px;
  user-select: all;
}
.btn-icon-copy {
  display: inline-flex; align-items: center; justify-content: center;
  background: none; border: none; cursor: pointer;
  padding: 3px; border-radius: 4px; color: #5f6368;
  transition: background 0.15s, color 0.15s;
  vertical-align: middle; margin-left: 4px;
}
.btn-icon-copy:hover { background: #e8f0fe; color: #1a73e8; }

/* API Key — modal display */
.mfield-apikey { border-top: 1px solid #e8eaed; padding-top: 14px; }
.apikey-display {
  display: flex; align-items: center; gap: 8px;
  background: #f8f9fa; border: 1px solid #dadce0;
  border-radius: 6px; padding: 8px 12px; flex-wrap: wrap;
}
.apikey-full {
  font-family: 'SFMono-Regular', Consolas, monospace;
  font-size: 12px; color: #202124; flex: 1; word-break: break-all;
  user-select: all;
}
.btn-regen {
  display: inline-flex; align-items: center; gap: 5px;
  background: #fff; border: 1px solid #dadce0; color: #3c4043;
  border-radius: 5px; padding: 5px 11px; font-size: 12.5px;
  font-weight: 500; cursor: pointer; font-family: inherit;
  transition: background 0.15s, border-color 0.15s;
  white-space: nowrap;
}
.btn-regen:hover:not(:disabled) { background: #fce8e6; border-color: #c5221f; color: #c5221f; }
.btn-regen:disabled { opacity: 0.5; cursor: not-allowed; }
.apikey-hint { margin-top: 8px !important; display: flex; flex-direction: column; gap: 5px; }
.env-snippet {
  display: block; background: #1e1e2e; color: #a6e3a1;
  border-radius: 5px; padding: 7px 12px;
  font-family: 'SFMono-Regular', Consolas, monospace;
  font-size: 12px; word-break: break-all; user-select: all;
}

/* URL chips */
.url-list { display: flex; flex-direction: column; gap: 4px; }
.url-chip { font-size: 11px; color: #1a73e8; background: #e8f0fe; padding: 2px 7px; border-radius: 4px; font-family: monospace; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; max-width: 240px; }
.url-more { font-size: 11px; color: #5f6368; }

/* Status badge */
.status-badge {
  display: inline-flex; align-items: center; gap: 5px;
  font-size: 12px; font-weight: 600; padding: 3px 9px;
  border-radius: 20px; white-space: nowrap;
}
.badge-active   { background: #e6f4ea; color: #137333; }
.badge-inactive { background: #f1f3f4; color: #5f6368; }
.badge-dot { width: 6px; height: 6px; border-radius: 50%; background: currentColor; }

/* Action buttons */
.actions-cell { white-space: nowrap; }
.btn-action {
  display: inline-flex; align-items: center; gap: 5px;
  padding: 5px 11px; border-radius: 5px;
  font-size: 12.5px; font-weight: 500; cursor: pointer;
  font-family: inherit; border: 1px solid;
  transition: background 0.15s;
  margin-right: 6px;
}
.btn-action:last-child { margin-right: 0; }
.btn-action:disabled { opacity: 0.5; cursor: not-allowed; }
.btn-edit    { background: #fff; border-color: #dadce0; color: #3c4043; }
.btn-edit:hover:not(:disabled)    { background: #f1f3f4; }
.btn-disable { background: #fff; border-color: #f5c6c4; color: #c5221f; }
.btn-disable:hover:not(:disabled) { background: #fce8e6; }
.btn-enable  { background: #fff; border-color: #a8d5b5; color: #137333; }
.btn-enable:hover:not(:disabled)  { background: #e6f4ea; }
.btn-delete  { background: #fff; border-color: #f5c6c4; color: #b31412; }
.btn-delete:hover:not(:disabled)  { background: #fce8e6; border-color: #b31412; }
.btn-confirm-delete { background: #b31412; border-color: #b31412; color: #fff; }
.btn-confirm-delete:hover:not(:disabled) { background: #8c0f0d; }
.btn-cancel-delete  { background: #fff; border-color: #dadce0; color: #3c4043; }
.btn-cancel-delete:hover:not(:disabled)  { background: #f1f3f4; }

.delete-confirm-inline {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  animation: fade-in-fast 0.12s ease;
}
.delete-confirm-label {
  font-size: 11.5px;
  font-weight: 600;
  color: #b31412;
  white-space: nowrap;
}
@keyframes fade-in-fast { from { opacity: 0; transform: scale(0.95); } to { opacity: 1; transform: scale(1); } }

/* Buttons */
.btn-primary {
  display: inline-flex; align-items: center; justify-content: center; gap: 7px;
  background: #1a73e8; color: #fff; border: none;
  border-radius: 6px; height: 38px; padding: 0 18px;
  font-size: 13.5px; font-weight: 500; cursor: pointer;
  font-family: inherit; white-space: nowrap;
  transition: background 0.18s, box-shadow 0.18s;
}
.btn-primary:hover:not(:disabled) { background: #1557b0; box-shadow: 0 2px 8px rgba(26,115,232,0.3); }
.btn-primary:disabled { opacity: 0.6; cursor: not-allowed; }

.btn-ghost {
  background: none; border: 1px solid #dadce0; color: #3c4043;
  border-radius: 6px; height: 36px; padding: 0 16px;
  font-size: 13.5px; font-weight: 500; cursor: pointer; font-family: inherit;
  transition: background 0.15s;
}
.btn-ghost:hover:not(:disabled) { background: #f1f3f4; }
.btn-ghost:disabled { opacity: 0.5; cursor: not-allowed; }

.btn-danger-ghost {
  background: none; border: 1px solid #f5c6c4; color: #c5221f;
  border-radius: 6px; height: 36px; padding: 0 16px;
  font-size: 13px; font-weight: 500; cursor: pointer; font-family: inherit;
  transition: background 0.15s;
}
.btn-danger-ghost:hover:not(:disabled) { background: #fce8e6; }
.btn-danger-ghost:disabled { opacity: 0.4; cursor: not-allowed; }

/* Spin icon */
.spin-icon { animation: spin 1s linear infinite; }

/* Toast */
.toast {
  position: fixed; bottom: 28px; left: 50%; transform: translateX(-50%);
  display: flex; align-items: center; gap: 9px;
  padding: 11px 20px; border-radius: 8px;
  font-size: 13.5px; font-weight: 500; z-index: 9999;
  box-shadow: 0 4px 16px rgba(0,0,0,0.2);
  white-space: nowrap;
}
.toast-success { background: #1e8e3e; color: #fff; }
.toast-error   { background: #d93025; color: #fff; }
.toast-fade-enter-active, .toast-fade-leave-active { transition: opacity 0.25s, transform 0.25s; }
.toast-fade-enter-from, .toast-fade-leave-to { opacity: 0; transform: translateX(-50%) translateY(10px); }

/* ── Modal ─────────────────────────────────────────────────────────────────── */
.modal-overlay {
  position: fixed; inset: 0; z-index: 500;
  background: rgba(0,0,0,0.45);
  display: flex; align-items: center; justify-content: center;
  padding: 24px;
}
.modal-card {
  background: #fff; border-radius: 12px;
  box-shadow: 0 20px 60px rgba(0,0,0,0.25);
  width: 100%; max-width: 540px;
  max-height: 90vh; display: flex; flex-direction: column;
  overflow: hidden;
}
.modal-header {
  display: flex; align-items: center; justify-content: space-between;
  padding: 20px 24px 16px; border-bottom: 1px solid #f1f3f4;
  flex-shrink: 0;
}
.modal-title { font-size: 17px; font-weight: 600; color: #202124; }
.modal-close {
  width: 32px; height: 32px; border-radius: 6px;
  background: none; border: none; cursor: pointer; color: #5f6368;
  display: flex; align-items: center; justify-content: center;
  transition: background 0.15s;
}
.modal-close:hover { background: #f1f3f4; }

.modal-alert {
  display: flex; align-items: flex-start; gap: 8px;
  margin: 12px 24px 0;
  padding: 10px 13px;
  background: #fce8e6; color: #c5221f;
  border-radius: 6px; font-size: 13px; line-height: 1.5;
  border-left: 3px solid #d93025;
  flex-shrink: 0;
}

.modal-body {
  padding: 20px 24px;
  overflow-y: auto;
  flex: 1;
}

/* Modal form fields */
.mfield { margin-bottom: 20px; }
.mfield label {
  display: block; font-size: 13px; font-weight: 500;
  color: #3c4043; margin-bottom: 6px;
}
.req { color: #d93025; }
.mfield input, .mfield textarea {
  width: 100%; padding: 9px 12px;
  border: 1.5px solid #dadce0; border-radius: 6px;
  font-size: 14px; color: #202124; font-family: inherit;
  outline: none; resize: vertical;
  transition: border-color 0.15s, box-shadow 0.15s;
  background: #fff;
}
.mfield input:focus, .mfield textarea:focus {
  border-color: #1a73e8; box-shadow: 0 0 0 3px rgba(26,115,232,0.1);
}
.mfield input.is-error, .mfield textarea.is-error { border-color: #d93025; }
.mfield input[readonly] { background: #f8f9fa; color: #5f6368; cursor: not-allowed; }
.mfield input:disabled, .mfield textarea:disabled { opacity: 0.6; cursor: not-allowed; }
.mono-input { font-family: 'SFMono-Regular', Consolas, monospace; font-size: 13px; }
.field-err  { font-size: 12px; color: #d93025; margin-top: 5px; }
.field-hint { font-size: 11.5px; color: #5f6368; margin-top: 5px; line-height: 1.5; }
.field-hint code { background: #f1f3f4; padding: 1px 4px; border-radius: 3px; font-size: 11px; }
.field-hint-inline { font-size: 12px; color: #9aa0a6; font-weight: 400; }

/* Modal footer */
.modal-footer {
  display: flex; align-items: center; justify-content: space-between;
  padding: 16px 24px; border-top: 1px solid #f1f3f4;
  gap: 12px; flex-shrink: 0;
  background: #fafbff;
  flex-wrap: wrap;
}
.modal-footer-right { display: flex; gap: 10px; align-items: center; }
.modal-footer-danger { display: flex; align-items: center; gap: 8px; }
.modal-delete-confirm {
  display: flex; align-items: center; gap: 8px;
  animation: fade-in-fast 0.12s ease;
  flex-wrap: wrap;
}
.modal-delete-label {
  font-size: 12.5px; color: #b31412; white-space: nowrap;
}
.btn-confirm-delete-sm {
  display: inline-flex; align-items: center; gap: 5px;
  background: #b31412; color: #fff; border: none;
  border-radius: 5px; padding: 5px 12px;
  font-size: 12.5px; font-weight: 500; cursor: pointer; font-family: inherit;
  transition: background 0.15s;
}
.btn-confirm-delete-sm:hover:not(:disabled) { background: #8c0f0d; }
.btn-confirm-delete-sm:disabled { opacity: 0.6; cursor: not-allowed; }
.btn-cancel-sm {
  background: none; border: 1px solid #dadce0; color: #5f6368;
  border-radius: 5px; padding: 5px 11px;
  font-size: 12.5px; font-weight: 500; cursor: pointer; font-family: inherit;
  transition: background 0.15s;
}
.btn-cancel-sm:hover:not(:disabled) { background: #f1f3f4; }
.btn-cancel-sm:disabled { opacity: 0.5; cursor: not-allowed; }

/* Modal transition */
.modal-fade-enter-active, .modal-fade-leave-active { transition: opacity 0.2s; }
.modal-fade-enter-from, .modal-fade-leave-to { opacity: 0; }
.modal-fade-enter-active .modal-card, .modal-fade-leave-active .modal-card { transition: transform 0.2s; }
.modal-fade-enter-from .modal-card { transform: scale(0.96); }
.modal-fade-leave-to .modal-card   { transform: scale(0.96); }


</style>
