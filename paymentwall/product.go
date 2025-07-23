// Package paymentwall provides a Go SDK for interacting with the Paymentwall APIs.
package paymentwall

import (
	"fmt"
	"math"
)

// ProductType constants
const (
	ProductTypeFixed        = "fixed"
	ProductTypeSubscription = "subscription"
)

// PeriodType constants
const (
	PeriodDay   = "day"
	PeriodWeek  = "week"
	PeriodMonth = "month"
	PeriodYear  = "year"
)

// Product represents a one-time or subscription-based item in the Paymentwall SDK.
type Product struct {
	ID           string   // Unique identifier for the product
	Amount       float64  // Price, rounded to two decimals
	CurrencyCode string   // ISO currency code, e.g., "USD"
	Name         string   // Human-readable name
	Type         string   // Product type: fixed or subscription
	PeriodLength int      // Subscription period length
	PeriodType   string   // Subscription period unit: day, week, month, year
	Recurring    bool     // Whether subscription auto-renews
	TrialProduct *Product // Trial period product; non-nil only for recurring subscriptions
}

// NewProduct constructs and validates a Product, rounding amount to two decimals.
// Returns an error if prodType or periodType is invalid.
func NewProduct(
	id string,
	amount float64,
	currencyCode, name, prodType string,
	periodLength int,
	periodType string,
	recurring bool,
	trial *Product,
) (*Product, error) {
	// Validate product type
	if prodType != ProductTypeFixed && prodType != ProductTypeSubscription {
		return nil, fmt.Errorf("invalid product type: %s", prodType)
	}
	// Validate period type if provided
	if periodType != "" {
		switch periodType {
			case PeriodDay, PeriodWeek, PeriodMonth, PeriodYear:
				// valid
			default:
				return nil, fmt.Errorf("invalid period type: %s", periodType)
		}
	}
	// Round amount to two decimals
	rounded := math.Round(amount*100) / 100
	p := &Product{
		ID:           id,
		Amount:       rounded,
		CurrencyCode: currencyCode,
		Name:         name,
		Type:         prodType,
		PeriodLength: periodLength,
		PeriodType:   periodType,
		Recurring:    recurring,
	}
	// Assign trial product only for recurring subscriptions
	if prodType == ProductTypeSubscription && recurring {
		p.TrialProduct = trial
	}
	return p, nil
}

// IsRecurring returns true if the product is a recurring subscription.
func (p *Product) IsRecurring() bool {
	return p.Recurring
}
