package middlewares

import (
	"github.com/go-go-golems/go-go-mcp/pkg"
)

// ToolProviderMiddleware is a function that wraps a ToolProvider with additional functionality
type ToolProviderMiddleware func(pkg.ToolProvider) pkg.ToolProvider
