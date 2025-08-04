// client_test.go
package paymentwall

import (
	"testing"
)

func TestNewClient_SetAPIType_AppendError_ErrorSummary(t *testing.T) {
	c := NewClient("app", "sec", APIVC)
	if c.AppKey != "app" || c.SecretKey != "sec" || c.APIType != APIVC {
		t.Fatalf("NewClient fields incorrect: %+v", c)
	}
	c.SetAPIType(APIGoods)
	if c.APIType != APIGoods {
		t.Errorf("SetAPIType = %v; want %v", c.APIType, APIGoods)
	}
	c.AppendError("e1")
	c.AppendError("e2")
	if got := c.ErrorSummary(); got != "e1\ne2" {
		t.Errorf("ErrorSummary = %q; want %q", got, "e1\ne2")
	}
}

func TestCalculateSignature_EmptySecret(t *testing.T) {
	c := NewClient("app", "", APIVC)
	if _, err := c.CalculateSignature(map[string]any{"uid": "u"}, SigV1); err == nil {
		t.Fatal("Expected error when secret is empty")
	}
}

func TestCalculateSignature_SigV1(t *testing.T) {
	secret := "s"
	c := NewClient("app", secret, APIVC)
	params := map[string]any{"uid": "userX"}
	got, err := c.CalculateSignature(params, SigV1)
	if err != nil {
		t.Fatal(err)
	}
    // SigV1 builds "key=value" before the secret
    want := hashMD5("uid=userX" + secret)
	if got != want {
		t.Errorf("SigV1 = %s; want %s", got, want)
	}
}


func TestCalculateSignature_SigV2(t *testing.T) {
	secret := "sec2"
	c := NewClient("app", secret, APIVC)
	params := map[string]any{"b": "bee", "a": "aye"}
	got, err := c.CalculateSignature(params, SigV2)
	if err != nil {
		t.Fatal(err)
	}
	// sorted: a=aye b=bee+secret
	want := hashMD5("a=aye"+"b=bee"+secret)
	if got != want {
		t.Errorf("SigV2 = %s; want %s", got, want)
	}
}

func TestCalculateSignature_SigV3(t *testing.T) {
	secret := "sec3"
	c := NewClient("app", secret, APIVC)
	params := map[string]any{"x": "ex", "y": "why"}
	got, err := c.CalculateSignature(params, SigV3)
	if err != nil {
		t.Fatal(err)
	}
	want := hashSHA256("x=ex"+"y=why"+secret)
	if got != want {
		t.Errorf("SigV3 = %s; want %s", got, want)
	}
}

func TestCalculateSignature_Slice(t *testing.T) {
	secret := "s"
	c := NewClient("app", secret, APIVC)
	params := map[string]any{"list": []any{"one", "two"}}
	got, err := c.CalculateSignature(params, SigV2)
	if err != nil {
		t.Fatal(err)
	}
	// list[0]=one list[1]=two + secret
	want := hashMD5("list[0]=one"+"list[1]=two"+secret)
	if got != want {
		t.Errorf("SigV2 slice = %s; want %s", got, want)
	}
}
