// product_test.go
package paymentwall

import (
	"math"
	"testing"
)

func TestNewProduct_FixedOK(t *testing.T) {
	p, err := NewProduct("p1", 1.239, "USD", "Name", ProductTypeFixed, 0, "", false, nil)
	if err != nil {
		t.Fatal(err)
	}
	if p.ID != "p1" {
		t.Errorf("ID = %q; want p1", p.ID)
	}
	if d := math.Abs(p.Amount - 1.24); d > 1e-9 {
		t.Errorf("Amount = %f; want 1.24", p.Amount)
	}
	if p.CurrencyCode != "USD" || p.Name != "Name" || p.Type != ProductTypeFixed {
		t.Errorf("Fields incorrect: %+v", p)
	}
	if p.IsRecurring() {
		t.Error("IsRecurring fixed = true; want false")
	}
	if p.TrialProduct != nil {
		t.Errorf("TrialProduct = %v; want nil", p.TrialProduct)
	}
}

func TestNewProduct_InvalidType(t *testing.T) {
	if _, err := NewProduct("x", 0, "", "", "bad", 0, "", false, nil); err == nil {
		t.Error("Expected error on invalid product type")
	}
}

func TestNewProduct_InvalidPeriodType(t *testing.T) {
	_, err := NewProduct("x", 0, "", "", ProductTypeSubscription, 1, "bad", true, nil)
	if err == nil {
		t.Error("Expected error on invalid period type")
	}
}

func TestNewProduct_SubscriptionWithTrial(t *testing.T) {
	tr, _ := NewProduct("t", 0.5, "USD", "Trial", ProductTypeFixed, 0, "", false, nil)
	p, err := NewProduct("s", 2.0, "EUR", "Sub", ProductTypeSubscription, 1, PeriodMonth, true, tr)
	if err != nil {
		t.Fatal(err)
	}
	if !p.IsRecurring() {
		t.Error("IsRecurring sub = false; want true")
	}
	if p.TrialProduct != tr {
		t.Errorf("TrialProduct = %v; want %v", p.TrialProduct, tr)
	}
}

func TestNewProduct_Sub_NoTrial(t *testing.T) {
	tr := &Product{ID: "t2"}
	p, err := NewProduct("s2", 3.0, "USD", "No", ProductTypeSubscription, 1, PeriodWeek, false, tr)
	if err != nil {
		t.Fatal(err)
	}
	if p.TrialProduct != nil {
		t.Error("TrialProduct = non-nil; want nil")
	}
}
