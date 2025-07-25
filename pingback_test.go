package paymentwall

import (
	"fmt"
	"testing"
)

func TestPingbackValidate_VC_SkipIP(t *testing.T) {
	client := NewClient("k", "s", APIVC)
	// prepare params without sig_version; default uses SigV1
	params := map[string]any{"uid": "u1", "currency": "42", "type": "0", "ref": "r1"}
	// calculate expected sig: md5("u1"+"s")
	expSig, err := client.CalculateSignature(map[string]any{"uid": "u1"}, SigV1)
	if err != nil {
		t.Fatalf("calc sig error: %v", err)
	}
	params["sig"] = expSig
	pb := NewPingback(client, params, "203.0.113.1")
	if !pb.Validate(true) {
		t.Fatalf("Validate failed: %v", pb.Errors)
	}
	// deliverable types: type=0 => deliverable
	if !pb.IsDeliverable() {
		t.Errorf("IsDeliverable=false, want true")
	}
	if pb.IsCancelable() {
		t.Errorf("IsCancelable=true, want false")
	}
}

func TestPingbackValidate_MissingParams(t *testing.T) {
	client := NewClient("k", "s", APIVC)
	// omit currency and sig
	params := map[string]any{"uid": "u2", "type": "1", "ref": "r2"}
	pb := NewPingback(client, params, "174.36.92.186") // valid IP
	ok := pb.Validate(false)
	if ok {
		t.Fatal("expected Validate to fail due to missing params")
	}
	// Expect two missing errors: currency and sig
	missingCount := 0
	for _, e := range pb.Errors {
		if e == "Parameter currency is missing" || e == "Parameter sig is missing" {
			missingCount++
		}
	}
	if missingCount != 2 {
		t.Errorf("expected 2 missing param errors, got %v: %v", missingCount, pb.Errors)
	}
}

func TestPingbackValidate_WrongSignature(t *testing.T) {
	client := NewClient("k", "s", APIVC)
	params := map[string]any{"uid": "u3", "currency": "1", "type": "0", "ref": "r3", "sig": "bad"}
	pb := NewPingback(client, params, "174.36.92.186")
	if pb.Validate(false) {
		t.Fatal("expected Validate to fail due to wrong signature")
	}
	found := false
	for _, e := range pb.Errors {
		if e == "Wrong signature" {
			found = true
		}
	}
	if !found {
		t.Errorf("Wrong signature error not found in %v", pb.Errors)
	}
}

func TestPingbackValidate_IPWhitelist(t *testing.T) {
	client := NewClient("k", "s", APIVC)
	params := map[string]any{"uid": "u4", "currency": "5", "type": "0", "ref": "r4"}
	expSig, _ := client.CalculateSignature(map[string]any{"uid": "u4"}, SigV1)
	params["sig"] = expSig
	// invalid IP
	pb := NewPingback(client, params, "1.2.3.4")
	if pb.Validate(false) {
		t.Fatal("expected Validate to fail due to IP not whitelisted")
	}
	if len(pb.Errors) == 0 || pb.Errors[0] != "IP address is not whitelisted" {
		t.Errorf("wrong IP error not recorded: %v", pb.Errors)
	}
}

func TestPingbackHelpers_andProductReconstruction(t *testing.T) {
	client := NewClient("k", "s", APIGoods)
	// Create pingback for Goods API: supply goodsid, slength, speriod, type, ref
	params := map[string]any{
		"uid":     "u5",
		"goodsid": "g1",
		"slength": "2",
		"speriod": "month",
		"type":    "1",
		"ref":     "r5",
	}
	// sig using SigV1 (default)
	expSig, _ := client.CalculateSignature(map[string]any{"uid": "u5", "goodsid": "g1"}, SigV1)
	params["sig"] = expSig
	pb := NewPingback(client, params, "174.36.92.186")
	if !pb.Validate(false) {
		t.Fatalf("expected Validate to succeed, got errors: %v", pb.Errors)
	}
	// Helper getters
	if pb.GetUserID() != "u5" {
		t.Errorf("GetUserID = %s, want u5", pb.GetUserID())
	}
	typeInt, err := pb.GetType()
	if err != nil || typeInt != 1 {
		t.Errorf("GetType = %d, err %v, want 1", typeInt, err)
	}
	if pb.GetProductID() != "g1" {
		t.Errorf("GetProductID = %s, want g1", pb.GetProductID())
	}
	// GetProduct
	prod, err := pb.GetProduct()
	if err != nil {
		t.Fatalf("GetProduct error: %v", err)
	}
	if prod.ID != "g1" || prod.Type != ProductTypeSubscription || prod.PeriodLength != 2 || prod.PeriodType != "month" {
		t.Errorf("Product reconstructed = %+v, want matching fields", prod)
	}
}

func TestPingbackGetProducts_Cart(t *testing.T) {
	client := NewClient("k", "s", APICart)
	params := map[string]any{
		"uid":     "u6",
		"goodsid": []any{"i1", "i2"},
		"type":    "0",
		"ref":     "r6",
	}
	// sig using SigV1 and goodsid filtered for uid only
	expSig, _ := client.CalculateSignature(params, SigV2)
	params["sig"] = expSig
	pb := NewPingback(client, params, "174.36.92.186")
	if !pb.Validate(false) {
		t.Fatalf("Validate failed: %v", pb.Errors)
	}
	items, err := pb.GetProducts()
	if err != nil {
		t.Fatalf("GetProducts error: %v", err)
	}
	if len(items) != 2 || items[0].ID != "i1" || items[1].ID != "i2" {
		t.Errorf("GetProducts = %+v, want IDs i1,i2", items)
	}
}

func TestPingbackStatusCheckers(t *testing.T) {
	client := NewClient("k", "s", APIVC)
	for _, ts := range []struct {
		code                    int
		deliver, cancel, review bool
	}{
		{0, true, false, false},
		{1, true, false, false},
		{2, false, true, false},
		{200, false, false, true},
		{201, true, false, false},
		{202, false, true, false},
		{203, false, false, false},
	} {
		params := map[string]any{"uid": "u7", "currency": "0", "type": fmt.Sprint(ts.code), "ref": "r7"}
		expSig, _ := client.CalculateSignature(map[string]any{"uid": "u7"}, SigV1)
		params["sig"] = expSig
		pb := NewPingback(client, params, "174.36.92.186")
		if !pb.Validate(false) {
			t.Fatalf("Validate failed for code %d: %v", ts.code, pb.Errors)
		}
		if pb.IsDeliverable() != ts.deliver {
			t.Errorf("IsDeliverable(%d) = %v, want %v", ts.code, pb.IsDeliverable(), ts.deliver)
		}
		if pb.IsCancelable() != ts.cancel {
			t.Errorf("IsCancelable(%d) = %v, want %v", ts.code, pb.IsCancelable(), ts.cancel)
		}
		if pb.IsUnderReview() != ts.review {
			t.Errorf("IsUnderReview(%d) = %v, want %v", ts.code, pb.IsUnderReview(), ts.review)
		}
	}
}
