package middleware

import (
	"net/http"
	"time"

	"meueu/internal/storage/postgres"
)

// Container aggregates all middleware instances for clean dependency injection.
type Container struct {
	Whitelist     *WhitelistMiddleware
	ContentFilter *ContentFilterMiddleware
	Banlist       *BanlistMiddleware
	RateLimiter   *RateLimiter
}

// NewContainer creates a fully-wired middleware container from the admin repository.
func NewContainer(adminRepo *postgres.AdminRepository) *Container {
	return &Container{
		Whitelist:     NewWhitelistMiddleware(adminRepo),
		ContentFilter: NewContentFilterMiddleware(adminRepo),
		Banlist:       NewBanlistMiddleware(adminRepo),
		RateLimiter:   NewRateLimiter(100, 30, 1*time.Minute),
	}
}

// StandardSecurity builds the governance middleware chain for the Publish endpoint.
// Chain order: RateLimiter → Banlist → ContentFilter → Whitelist → handler
func (c *Container) StandardSecurity(handler http.Handler) http.Handler {
	return c.RateLimiter.Middleware(
		c.Banlist.Middleware(
			c.ContentFilter.Middleware(
				c.Whitelist.Middleware(handler),
			),
		),
	)
}

// QuerySecurity applies whitelist-only filtering to query endpoints.
func (c *Container) QuerySecurity(handler http.Handler) http.Handler {
	return c.Whitelist.Middleware(handler)
}

// CORS wraps a handler with CORS headers (re-exported for convenience).
func (c *Container) CORS(handler http.Handler) http.Handler {
	return CORS(handler)
}
