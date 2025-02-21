package middlewares

import (
	"github.com/go-go-golems/go-go-mcp/pkg"
)

// ChainBuilder helps construct middleware chains in a fluent manner
type ChainBuilder struct {
	chain        []ToolProviderMiddleware
	errorHandler func(error) error
}

// NewChainBuilder creates a new chain builder instance
func NewChainBuilder() *ChainBuilder {
	return &ChainBuilder{
		chain: make([]ToolProviderMiddleware, 0),
	}
}

// With adds a middleware to the chain
func (b *ChainBuilder) With(middleware ToolProviderMiddleware) *ChainBuilder {
	b.chain = append(b.chain, middleware)
	return b
}

// WithErrorHandler adds an error handling function to the chain
func (b *ChainBuilder) WithErrorHandler(handler func(error) error) *ChainBuilder {
	b.errorHandler = handler
	return b
}

// Build constructs the final ToolProvider with all middleware applied
func (b *ChainBuilder) Build(base pkg.ToolProvider) pkg.ToolProvider {
	provider := base
	// Apply middlewares in reverse order to maintain the expected execution order
	for i := len(b.chain) - 1; i >= 0; i-- {
		provider = b.chain[i](provider)
	}
	return provider
}
