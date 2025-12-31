package constants

import "time"

// Server configuration
const (
	// DefaultServerPort is the default port the server listens on
	DefaultServerPort = ":8080"

	// ShutdownGracePeriod is how long to wait for ongoing requests to complete
	ShutdownGracePeriod = 100 * time.Millisecond
)

// Authentication constants
const (
	// JWTExpirationHours is how long a JWT token is valid
	JWTExpirationHours = 72 * time.Hour

	// MinPasswordLength is the minimum required password length
	MinPasswordLength = 8
)

// Database timeouts
const (
	// DefaultDBTimeout is the default timeout for database operations
	DefaultDBTimeout = 10 * time.Second

	// DBRetryBackoffBase is the base backoff duration for database retries
	DBRetryBackoffBase = 100 * time.Millisecond

	// MaxDBRetries is the maximum number of retry attempts for database operations
	MaxDBRetries = 3
)

// Validation limits
const (
	// MaxAlertThreshold is the maximum alert threshold percentage
	MaxAlertThreshold = 100

	// MinAlertThreshold is the minimum alert threshold percentage
	MinAlertThreshold = 0

	// MaxAmount is the maximum allowed transaction/budget amount
	MaxAmount = 1000000000 // 1 billion

	// MaxCategoryNameLength is the maximum length for category names
	MaxCategoryNameLength = 100

	// MaxDescriptionLength is the maximum length for descriptions
	MaxDescriptionLength = 500

	// MaxPaginationLimit is the maximum number of records per page
	MaxPaginationLimit = 1000

	// DefaultPaginationLimit is the default number of records per page
	DefaultPaginationLimit = 20
)

// Pre-allocation capacities (for memory optimization)
const (
	// TypicalTransactionCount is a reasonable pre-allocation for transaction lists
	TypicalTransactionCount = 100

	// TypicalCategoryCount is a reasonable pre-allocation for category lists
	TypicalCategoryCount = 10

	// TypicalBudgetCount is a reasonable pre-allocation for budget lists
	TypicalBudgetCount = 5

	// TypicalRecurringCount is a reasonable pre-allocation for recurring transaction lists
	TypicalRecurringCount = 10

	// TypicalMonthlyDataPoints is a reasonable pre-allocation for monthly data
	TypicalMonthlyDataPoints = 12

	// TypicalWeeklyDataPoints is a reasonable pre-allocation for weekly data
	TypicalWeeklyDataPoints = 52

	// TypicalYearlyDataPoints is a reasonable pre-allocation for yearly data
	TypicalYearlyDataPoints = 3
)

// Recurring job constants
const (
	// RecurringJobInterval is how often the recurring job runs
	RecurringJobInterval = 1 * time.Hour

	// RecurringJobLockID is the PostgreSQL advisory lock ID for the recurring job
	RecurringJobLockID = 123456789

	// MaxRecurringIterations prevents infinite loops in recurring date calculations
	MaxRecurringIterations = 3650 // ~10 years of daily transactions
)

// Database connection pool settings
const (
	// MaxOpenConnections is the maximum number of open database connections
	MaxOpenConnections = 25

	// MaxIdleConnections is the maximum number of idle database connections
	MaxIdleConnections = 10

	// ConnectionMaxLifetime is the maximum lifetime of a database connection
	ConnectionMaxLifetime = 5 * time.Minute

	// ConnectionMaxIdleTime is the maximum idle time before closing a connection
	ConnectionMaxIdleTime = 2 * time.Minute
)

// Rate limiting
const (
	// AuthRateLimitPerMinute is the number of auth requests allowed per minute
	AuthRateLimitPerMinute = 5.0 / 60.0 // 5 requests per 60 seconds

	// AuthRateLimitBurst is the burst capacity for auth rate limiting
	AuthRateLimitBurst = 5

	// APIRateLimitPerSecond is the number of API requests allowed per second for authenticated endpoints
	APIRateLimitPerSecond = 100.0 / 60.0 // 100 requests per 60 seconds

	// APIRateLimitBurst is the burst capacity for API rate limiting
	APIRateLimitBurst = 20
)
