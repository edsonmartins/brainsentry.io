package mcp

// Error categories as specified in MCP_SERVER_API.md
const (
	ErrCategoryValidation    = "validation"
	ErrCategoryAuthorization = "authorization"
	ErrCategoryNotFound      = "not_found"
	ErrCategoryInternal      = "internal"
	ErrCategoryTenant        = "tenant"
	ErrCategoryRateLimit     = "rate_limit"
	ErrCategoryTimeout       = "timeout"
)

// MCPError represents a structured MCP error with category.
type MCPError struct {
	Code     int    `json:"code"`
	Message  string `json:"message"`
	Category string `json:"errorCategory"`
	Details  any    `json:"details,omitempty"`
}

// NewMCPError creates a new categorized MCP error response.
func NewMCPError(code int, message, category string) *RPCError {
	return &RPCError{
		Code:    code,
		Message: message,
		Data: map[string]any{
			"errorCategory": category,
			"errorType":     category,
		},
	}
}

// Error constructors for common error types.

func ErrValidation(message string) *RPCError {
	return NewMCPError(-32602, message, ErrCategoryValidation)
}

func ErrNotFound(message string) *RPCError {
	return NewMCPError(-32602, message, ErrCategoryNotFound)
}

func ErrAuthorization(message string) *RPCError {
	return NewMCPError(-32600, message, ErrCategoryAuthorization)
}

func ErrInternal(message string) *RPCError {
	return NewMCPError(-32603, message, ErrCategoryInternal)
}

func ErrTenant(message string) *RPCError {
	return NewMCPError(-32600, message, ErrCategoryTenant)
}

func ErrRateLimit(message string) *RPCError {
	return NewMCPError(-32600, message, ErrCategoryRateLimit)
}

func ErrTimeout(message string) *RPCError {
	return NewMCPError(-32603, message, ErrCategoryTimeout)
}
