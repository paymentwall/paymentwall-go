// pingback_test.go
package paymentwall

import (
	"strconv"
	"testing"
	"strings"
)

func TestPingback_MissingParams(t *testing.T) {
	pb := NewPingback(NewClient("k","s",APIGoods), map[string]any{}, "1.2.3.4")
	if ok := pb.Validate(true); ok {
		t.Errorf("Validate missing params = true; want false")
	}
	if !strings.Contains(pb.ErrorSummary(), "Parameter uid is missing") {
		t.Errorf("ErrorSummary missing uid error: %s", pb.ErrorSummary())
	}
}

func TestPingback_IPWhitelist(t *testing.T) {
	cl := NewClient("k","s",APIGoods)
	params := map[string]any{"uid":"u","goodsid":"g","type":"0","ref":"r","sig":"x"}
	pb := NewPingback(cl, params, "8.8.8.8")
	// skip whitelist
	if ok := pb.Validate(true); ok {
		t.Errorf("Validate skip whitelist = true; want false (sig wrong)")
	}
	// enforce whitelist
	if ok := pb.Validate(false); ok {
		t.Errorf("Validate non-whitelisted IP = true; want false")
	}
	// whitelisted
	pb2 := NewPingback(cl, params, "174.36.92.186")
	if ok := pb2.Validate(false); ok {
		t.Errorf("Wrong signature but passes whitelisted IP")
	}
}

func TestPingback_SigV2_Valid(t *testing.T) {
	secret := "sec2"
	cl := NewClient("k", secret, APICart)
	// build Cart pingback with sign_version=2
	params := map[string]any{
		"uid":"u2","goodsid": []any{"p1","p2"},
		"type":"1","ref":"refX","sign_version": strconv.Itoa(int(SigV2)),
	}
	// calculate sig over all params except "sig"
	signed := map[string]any{
		"uid":"u2",
		"goodsid": []any{"p1","p2"},
		"type":"1","ref":"refX","sign_version": int(SigV2),
	}
	sig, _ := cl.CalculateSignature(signed, SigV2)
	params["sig"] = sig
	pb := NewPingback(cl, params, "174.36.92.186")
	if !pb.Validate(false) {
		t.Errorf("Validate valid SigV2 = false; %s", pb.ErrorSummary())
	}
}
