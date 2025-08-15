package models

import (
	"time"

	"github.com/google/uuid"
)

// PlanType represents the subscription plan type
type PlanType string

const (
	PlanFree       PlanType = "free"
	PlanStarter    PlanType = "starter"
	PlanPro        PlanType = "pro"
	PlanEnterprise PlanType = "enterprise"
)

// User represents a QUIVer user
type User struct {
	ID               uuid.UUID    `json:"id" db:"id"`
	Email            string       `json:"email" db:"email"`
	StripeCustomerID string       `json:"stripe_customer_id,omitempty" db:"stripe_customer_id"`
	Plan             PlanType     `json:"plan" db:"plan"`
	APIKey           string       `json:"api_key,omitempty" db:"api_key"`
	MonthlyRequests  int          `json:"monthly_requests" db:"monthly_requests"`
	UsedRequests     int          `json:"used_requests" db:"used_requests"`
	ResetAt          time.Time    `json:"reset_at" db:"reset_at"`
	CreatedAt        time.Time    `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time    `json:"updated_at" db:"updated_at"`
	Subscription     *Subscription `json:"subscription,omitempty"`
}

// Subscription represents a user's subscription
type Subscription struct {
	ID                 uuid.UUID  `json:"id" db:"id"`
	UserID             uuid.UUID  `json:"user_id" db:"user_id"`
	StripeSubscriptionID string   `json:"stripe_subscription_id" db:"stripe_subscription_id"`
	Plan               PlanType   `json:"plan" db:"plan"`
	Status             string     `json:"status" db:"status"`
	CurrentPeriodStart time.Time  `json:"current_period_start" db:"current_period_start"`
	CurrentPeriodEnd   time.Time  `json:"current_period_end" db:"current_period_end"`
	CancelAtPeriodEnd  bool       `json:"cancel_at_period_end" db:"cancel_at_period_end"`
	CreatedAt          time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at" db:"updated_at"`
}

// Usage represents API usage statistics
type Usage struct {
	ID           uuid.UUID `json:"id" db:"id"`
	UserID       uuid.UUID `json:"user_id" db:"user_id"`
	Date         time.Time `json:"date" db:"date"`
	Model        string    `json:"model" db:"model"`
	Requests     int       `json:"requests" db:"requests"`
	InputTokens  int64     `json:"input_tokens" db:"input_tokens"`
	OutputTokens int64     `json:"output_tokens" db:"output_tokens"`
	Cost         float64   `json:"cost" db:"cost"`
}

// PlanLimits defines the limits for each plan
var PlanLimits = map[PlanType]struct {
	MonthlyRequests int
	RateLimit       int    // requests per minute
	MaxModel        string // maximum model size
	Priority        int    // 0=lowest, 3=highest
}{
	PlanFree: {
		MonthlyRequests: 1000,
		RateLimit:       10,
		MaxModel:        "3b",
		Priority:        0,
	},
	PlanStarter: {
		MonthlyRequests: 10000,
		RateLimit:       60,
		MaxModel:        "7b",
		Priority:        1,
	},
	PlanPro: {
		MonthlyRequests: 100000,
		RateLimit:       300,
		MaxModel:        "32b",
		Priority:        2,
	},
	PlanEnterprise: {
		MonthlyRequests: -1, // unlimited
		RateLimit:       -1, // custom
		MaxModel:        "*", // all models
		Priority:        3,
	},
}

// PlanPricing defines the pricing for each plan (in cents)
var PlanPricing = map[PlanType]int{
	PlanFree:       0,
	PlanStarter:    999,   // $9.99
	PlanPro:        4999,  // $49.99
	PlanEnterprise: 29999, // $299.99 base
}

// CanUseModel checks if a user can use a specific model
func (u *User) CanUseModel(model string) bool {
	limits := PlanLimits[u.Plan]
	if limits.MaxModel == "*" {
		return true
	}
	
	// Extract model size (e.g., "7b" from "qwen3:7b")
	// Simplified logic - in production would be more robust
	return true // TODO: Implement model size checking
}

// HasRequestsRemaining checks if user has requests remaining
func (u *User) HasRequestsRemaining() bool {
	if u.Plan == PlanEnterprise {
		return true
	}
	
	limits := PlanLimits[u.Plan]
	return u.UsedRequests < limits.MonthlyRequests
}

// GetRateLimit returns the rate limit for the user's plan
func (u *User) GetRateLimit() int {
	return PlanLimits[u.Plan].RateLimit
}

// GetPriority returns the priority level for the user's plan
func (u *User) GetPriority() int {
	return PlanLimits[u.Plan].Priority
}