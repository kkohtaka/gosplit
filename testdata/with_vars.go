package testdata

// MaxRetries defines the maximum number of retry attempts
const MaxRetries = 3

// DefaultTimeout specifies the default timeout in seconds
const DefaultTimeout = 30

// Error messages
const (
	ErrNotFound    = "not found"
	ErrInvalidData = "invalid data"
)

// Config holds application configuration
var Config = struct {
	Host string
	Port int
}{
	Host: "localhost",
	Port: 8080,
}

// Debug mode flag
var Debug = false

// Version information
var (
	Version    = "1.0.0"
	BuildTime  = "2024-03-20"
	CommitHash = "abc123"
)
