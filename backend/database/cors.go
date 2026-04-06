package database

import (
	"log"
	"net/url"
	"strings"
	"sync"
	"time"

	"outcraftly/accounts/models"
)

// ─────────────────────────────────────────────────────────────────────────────
// Dynamic CORS origin allowlist — driven by the product registry in the DB.
//
// Every product has redirect_urls (e.g. ["http://localhost:3000/callback",
// "https://app.example.com/callback"]).  We extract the origin
// (scheme://host) from each URL and cache the set.  The cache refreshes every
// 60 seconds so that newly registered products (via admin API) are picked up
// without a restart.
//
// Usage in main.go:
//
//	cors.New(cors.Config{
//	    AllowOriginsFunc: database.IsAllowedOrigin,
//	    ...
//	})
// ─────────────────────────────────────────────────────────────────────────────

var (
	cachedOrigins   map[string]bool
	cacheMu         sync.RWMutex
	cacheExpiry     time.Time
	cacheTTL        = 60 * time.Second
	// staticOrigins holds origins from the ALLOW_ORIGINS env var (e.g.
	// the accounts frontend itself).  Set once at startup via SetStaticOrigins.
	staticOrigins   map[string]bool
)

// SetStaticOrigins records origins that should always be allowed regardless of
// the product registry — typically the Accounts frontend URL and any dev
// localhost ports.  Call once at startup before the Fiber app starts listening.
//
// Input: comma-separated origin list, e.g.
//
//	"http://localhost:5173,http://localhost:3000"
func SetStaticOrigins(commaSeparated string) {
	staticOrigins = make(map[string]bool)
	for _, entry := range strings.Split(commaSeparated, ",") {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}
		origin := extractOrigin(entry)
		if origin != "" {
			staticOrigins[origin] = true
		}
	}
	log.Printf("[cors] static origins: %v", staticOrigins)
}

// IsAllowedOrigin is the callback for Fiber's cors.Config.AllowOriginsFunc.
// It returns true if the browser origin matches:
//   - a static origin (ALLOW_ORIGINS env — the accounts frontend)
//   - any origin derived from a product's redirect_urls in the DB
func IsAllowedOrigin(origin string) bool {
	origin = strings.TrimRight(strings.ToLower(origin), "/")

	// 1. Check static (env) origins — always allowed.
	if staticOrigins[origin] {
		return true
	}

	// 2. Check product-derived origins (cached, refreshed periodically).
	cacheMu.RLock()
	if time.Now().Before(cacheExpiry) {
		allowed := cachedOrigins[origin]
		cacheMu.RUnlock()
		return allowed
	}
	cacheMu.RUnlock()

	// Cache expired — refresh.
	refreshOriginCache()

	cacheMu.RLock()
	defer cacheMu.RUnlock()
	return cachedOrigins[origin]
}

// refreshOriginCache queries all active products, extracts origins from their
// redirect_urls, and rebuilds the cache.
func refreshOriginCache() {
	cacheMu.Lock()
	defer cacheMu.Unlock()

	// Double-check: another goroutine may have refreshed while we waited.
	if time.Now().Before(cacheExpiry) {
		return
	}

	origins := make(map[string]bool)

	var products []models.Product
	if err := DB.Where("is_active = true").Find(&products).Error; err != nil {
		log.Printf("[cors] WARNING: could not load products for CORS cache: %v", err)
		// Keep stale cache, extend TTL briefly to avoid hammering DB.
		cacheExpiry = time.Now().Add(5 * time.Second)
		return
	}

	for _, p := range products {
		for _, rawURL := range p.RedirectURLs {
			if o := extractOrigin(rawURL); o != "" {
				origins[o] = true
			}
		}
	}

	cachedOrigins = origins
	cacheExpiry = time.Now().Add(cacheTTL)
	log.Printf("[cors] refreshed product origin cache: %d origins from %d products", len(origins), len(products))
}

// extractOrigin returns "scheme://host" from a URL string, lowercased.
// Returns "" for unparseable or empty URLs.
func extractOrigin(rawURL string) string {
	u, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil || u.Host == "" || u.Scheme == "" {
		return ""
	}
	return strings.ToLower(u.Scheme + "://" + u.Host)
}
