package paymentwall

import (
	"testing"
)

func TestNewProduct_FixedRoundingAndType(t *testing.T) {
	p, err := NewProduct("p1", 9.999, "USD", "Product", ProductTypeFixed, 0, "", false, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Amount should round to two decimals: 9.999 -> 10.00
	if p.Amount != 10.00 {
		t.Errorf("Amount rounded = %v, want %v", p.Amount, 10.00)
	}
	if p.Type != ProductTypeFixed {
		t.Errorf("Type = %q, want %q", p.Type, ProductTypeFixed)
	}
	if p.IsRecurring() {
		t.Errorf("IsRecurring = true, want false for fixed product")
	}
}

func TestNewProduct_SubscriptionAndRecurring(t *testing.T) {
	// Create a trial product
	trial, err := NewProduct("trial", 0.99, "EUR", "Trial", ProductTypeSubscription, 1, PeriodWeek, false, nil)
	if err != nil {
		t.Fatalf("unexpected error creating trial: %v", err)
	}
	// Main subscription with recurring and trial
	main, err := NewProduct("sub1", 12.345, "USD", "Subs", ProductTypeSubscription, 1, PeriodMonth, true, trial)
	if err != nil {
		t.Fatalf("unexpected error creating subscription: %v", err)
	}
	// Amount should round 12.345 -> 12.35
	if main.Amount != 12.35 {
		t.Errorf("Amount rounded = %v, want %v", main.Amount, 12.35)
	}
	if main.Type != ProductTypeSubscription {
		t.Errorf("Type = %q, want %q", main.Type, ProductTypeSubscription)
	}
	if !main.IsRecurring() {
		t.Errorf("IsRecurring = false, want true for recurring subscription")
	}
	if main.TrialProduct == nil {
		t.Errorf("TrialProduct is nil, want non-nil for recurring subscription with trial")
	} else if main.TrialProduct.ID != trial.ID {
		t.Errorf("TrialProduct.ID = %q, want %q", main.TrialProduct.ID, trial.ID)
	}
}

func TestNewProduct_InvalidType(t *testing.T) {
	_, err := NewProduct("x", 1.23, "USD", "Bad", "invalidType", 0, "", false, nil)
	if err == nil {
		t.Fatal("expected error for invalid product type, got nil")
	}
}

func TestNewProduct_InvalidPeriod(t *testing.T) {
	_, err := NewProduct("x", 1.23, "USD", "Bad", ProductTypeSubscription, 1, "invalidPeriod", false, nil)
	if err == nil {
		t.Fatal("expected error for invalid period type, got nil")
	}
}

func TestNewProduct_TrialIgnoredWhenNotRecurring(t *testing.T) {
	trial, _ := NewProduct("trial", 0.10, "USD", "Trial", ProductTypeSubscription, 1, PeriodWeek, false, nil)
	p, err := NewProduct("p2", 5.00, "USD", "NonRec", ProductTypeSubscription, 1, PeriodWeek, false, trial)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.TrialProduct != nil {
		t.Errorf("TrialProduct not nil, want nil when recurring=false")
	}
}
