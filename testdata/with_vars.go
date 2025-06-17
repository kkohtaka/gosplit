package testdata

// Package-level comment for MaxRetries
// This is a multi-line comment
// explaining the purpose of MaxRetries
const MaxRetries = 3

const DefaultTimeout = 30 // Inline comment for DefaultTimeout

// Group of error messages
// Each constant represents a specific error case
const (
	// ErrNotFound is returned when a resource is not found
	ErrNotFound = "not found"

	ErrInvalidData = "invalid data" // Inline comment for ErrInvalidData

	// ErrTimeout represents a timeout error
	// It includes the timeout duration in the message
	ErrTimeout = "operation timed out"
)

// Numeric constants with different types
const (
	Pi         = 3.14159
	MaxInt32   = 1<<31 - 1
	MinInt32   = -1 << 31
	MaxUint32  = 1<<32 - 1
	MaxFloat32 = 3.402823e+38
	MinFloat32 = 1.401298e-45
)

// Boolean flags
const (
	IsProduction = false
	EnableCache  = true
	UseSSL       = true
)

// Config holds application configuration
// It contains basic server settings
var Config = struct {
	Host string // Server hostname
	Port int    // Server port number
}{
	Host: "localhost", // Default host
	Port: 8080,        // Default port
}

var Debug = false // Global debug flag

// Version information
// Contains build metadata
var (
	// Version represents the current release version
	Version = "1.0.0"

	BuildTime = "2024-03-20" // Build timestamp

	// CommitHash stores the git commit hash
	// Used for version tracking
	CommitHash = "abc123"
)

// Database configuration
var DBConfig = struct {
	Host     string
	Port     int
	User     string
	Password string
	Database string
	SSL      bool
}{
	Host:     "db.example.com",
	Port:     5432,
	User:     "admin",
	Password: "secret",
	Database: "my_db",
	SSL:      true,
}

// Feature flags with different comment styles
var (
	// EnableNewUI controls the new user interface
	// When true, the new UI is shown
	// When false, the legacy UI is used
	EnableNewUI = true

	EnableAnalytics = false // Simple inline comment

	/* EnableLogging is a multi-line
	   comment using block style
	   for better readability */
	EnableLogging = true

	// EnableMetrics controls metrics collection
	EnableMetrics = true // Additional inline detail
)

// Cache settings with mixed comment styles
var CacheSettings = struct {
	// MaxSize defines the maximum cache size in bytes
	MaxSize int64
	TTL     int // Time-to-live in seconds
	/* Compression enables data compression
	   when storing cache entries */
	Compression bool
	Algorithm   string // Cache replacement algorithm
}{
	MaxSize:     1024 * 1024 * 100, // 100MB
	TTL:         3600,              // 1 hour
	Compression: true,              // Enable compression
	Algorithm:   "lru",             // Least Recently Used
}

// unexported variables
var (
	internalCounter = 0
	debugLevel      = 2
	secretKey       = "internal-secret-key"
)

// API endpoints
const (
	APIVersion = "v1"
	BaseURL    = "https://api.example.com"
)

// HTTP methods with different comment positions
const (
	// MethodGet represents HTTP GET method
	MethodGet = "GET"

	MethodPost = "POST" // HTTP POST method

	/* MethodPut represents HTTP PUT method
	   Used for updating resources */
	MethodPut = "PUT"

	// MethodDelete represents HTTP DELETE method
	// Used for removing resources
	MethodDelete = "DELETE"
)

// Status codes with various comment styles
const (
	StatusOK = 200 // Success

	// StatusCreated indicates successful resource creation
	StatusCreated = 201

	/* StatusNotFound indicates that the requested
	   resource was not found on the server */
	StatusNotFound = 404

	// StatusInternal represents server errors
	// Should be logged for investigation
	StatusInternal = 500
)
