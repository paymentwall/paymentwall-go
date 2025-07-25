package paymentwall

import (
	"testing"
)

// Known hashes for "user"+"secret"
const (
	uid    = "user"
	secret = "secret"
	md5V1  = "d67673f506f955d7c61821867a3a41dc" // md5("usersecret")
	// For v2 and v3, the base string is "alice=bobsecret"
	md5V2    = "ac0fd25828991a28d9a92293ad92fa9a"                                 // md5("alice=bobsecret")
	sha256V3 = "9e859376d049e42d1882807589ce20b5ccaab97fce0528c071c076831acaf485" // sha256("alice=bobsecret")
)

func TestCalculateSignature_V1(t *testing.T) {
	client := NewClient("appKey", secret, APIVC)
	sig, err := client.CalculateSignature(map[string]any{"uid": uid}, SigV1)
	if err != nil {
		t.Fatalf("unexpected error for SigV1: %v", err)
	}
	if sig != md5V1 {
		t.Errorf("SigV1: got %s, want %s", sig, md5V1)
	}
}

func TestCalculateSignature_V2(t *testing.T) {
	client := NewClient("appKey", secret, APIVC)
	params := map[string]any{"alice": "bob"}
	sig, err := client.CalculateSignature(params, SigV2)
	if err != nil {
		t.Fatalf("unexpected error for SigV2: %v", err)
	}
	if sig != md5V2 {
		t.Errorf("SigV2: got %s, want %s", sig, md5V2)
	}
}

func TestCalculateSignature_V3(t *testing.T) {
	client := NewClient("appKey", secret, APIVC)
	params := map[string]any{"alice": "bob"}
	sig, err := client.CalculateSignature(params, SigV3)
	if err != nil {
		t.Fatalf("unexpected error for SigV3: %v", err)
	}
	if sig != sha256V3 {
		t.Errorf("SigV3: got %s, want %s", sig, sha256V3)
	}
}

func TestCalculateSignature_EmptySecret(t *testing.T) {
	client := NewClient("appKey", "", APIVC)
	_, err := client.CalculateSignature(map[string]any{"uid": uid}, SigV1)
	if err == nil {
		t.Fatal("expected error when secret key is empty, got nil")
	}
}

func TestErrorAggregation(t *testing.T) {
	client := NewClient("appKey", "secret", APIVC)
	client.Errors = nil
	client.AppendError("first error")
	client.AppendError("second error")
	want := "first error\nsecond error"
	if got := client.ErrorSummary(); got != want {
		t.Errorf("ErrorSummary = %q, want %q", got, want)
	}
}
