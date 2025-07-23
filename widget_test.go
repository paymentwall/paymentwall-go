package paymentwall

import (
	"net/url"
	"strings"
	"testing"
)

// Helper to remove fields from map and stringify values
func copyParams(m map[string]any, remove ...string) map[string]any {
	out := make(map[string]any, len(m))
	for k, v := range m {
		skip := false
		for _, r := range remove {
			if k == r {
				skip = true
				break
			}
		}
		if !skip {
			out[k] = v
		}
	}
	return out
}

func TestGetParamsAPIVC(t *testing.T) {
	client := NewClient("appkey", "secret", APIVC)
	widget := NewWidget(client, "user123", "vc1", nil, map[string]any{"email": "u@example.com"})
	params, err := widget.GetParams()
	if err != nil {
		t.Fatalf("GetParams error: %v", err)
	}
	// Basic fields
	if params["key"] != "appkey" {
		t.Errorf("key = %v, want appkey", params["key"])
	}
	if params["uid"] != "user123" {
		t.Errorf("uid = %v, want user123", params["uid"])
	}
	// Extra param merged
	if params["email"] != "u@example.com" {
		t.Errorf("email = %v, want u@example.com", params["email"])
	}
	// Signature version should default to 3
	svFloat, ok := params["sign_version"].(int)
	if !ok || SignatureVersion(svFloat) != SigV3 {
		t.Errorf("sign_version = %v, want %d", params["sign_version"], SigV3)
	}
	// Signature should match calculateSignature output
	sv := SignatureVersion(svFloat)
	// Remove sign from params for recalculation
	core := copyParams(params, "sign")
	expected, serr := client.calculateSignature(core, sv)
	if serr != nil {
		t.Fatalf("recalculate signature error: %v", serr)
	}
	if params["sign"] != expected {
		t.Errorf("sign = %v, want %v", params["sign"], expected)
	}
}

func TestGetParamsAPIGoods_SingleProduct(t *testing.T) {
	client := NewClient("k", "s", APIGoods)
	prod, _ := NewProduct("p1", 2.5, "USD", "Name", ProductTypeFixed, 0, "", false, nil)
	widget := NewWidget(client, "u1", "w1", []*Product{prod}, nil)
	params, err := widget.GetParams()
	if err != nil {
		t.Fatalf("GetParams error: %v", err)
	}
	// Check product fields
	if params["ag_external_id"] != "p1" {
		t.Errorf("ag_external_id = %v, want p1", params["ag_external_id"])
	}
	if amt, ok := params["amount"].(float64); !ok || amt != 2.5 {
		t.Errorf("amount = %v, want 2.5", params["amount"])
	}
}

func TestGetParamsAPIGoods_InvalidCount(t *testing.T) {
	client := NewClient("k", "s", APIGoods)
	p1, _ := NewProduct("p1", 1, "USD", "A", ProductTypeFixed, 0, "", false, nil)
	p2, _ := NewProduct("p2", 2, "USD", "B", ProductTypeFixed, 0, "", false, nil)
	widget := NewWidget(client, "u2", "w2", []*Product{p1, p2}, nil)
	_, err := widget.GetParams()
	if err == nil {
		t.Fatal("expected error for two products in API_GOODS, got none")
	}
	// The widget still returns params map; check that AppendError was called
	summary := client.ErrorSummary()
	if !strings.Contains(summary, "only one product allowed for API_GOODS") {
		t.Errorf("error summary = %q, want mention of product count", summary)
	}
}

func TestGetParamsAPICart(t *testing.T) {
	client := NewClient("k", "s", APICart)
	p1, _ := NewProduct("a", 1.1, "EUR", "A", ProductTypeFixed, 0, "", false, nil)
	p2, _ := NewProduct("b", 0, "", "B", ProductTypeFixed, 0, "", false, nil)
	widget := NewWidget(client, "u3", "c1", []*Product{p1, p2}, nil)
	params, err := widget.GetParams()
	if err != nil {
		t.Fatalf("GetParams error: %v", err)
	}
	// external_ids and prices
	if params["external_ids[0]"] != "a" || params["external_ids[1]"] != "b" {
		t.Errorf("external_ids = %v and %v, want a and b", params["external_ids[0]"], params["external_ids[1]"])
	}
	if params["prices[0]"] != 1.1 {
		t.Errorf("prices[0] = %v, want 1.1", params["prices[0]"])
	}
	// No currency for second
	if _, ok := params["currencies[1]"]; ok {
		t.Errorf("currencies[1] should be absent or empty, got %v", params["currencies[1]"])
	}
}

func TestGetURLAndHTML(t *testing.T) {
	client := NewClient("appk", "sec", APIVC)
	widget := NewWidget(client, "u4", "v1", nil, map[string]any{"foo": "bar"})
	urlStr, err := widget.GetURL()
	if err != nil {
		t.Fatalf("GetURL error: %v", err)
	}
	// Parse URL and ensure query params present
	u, err := url.Parse(urlStr)
	if err != nil {
		t.Fatalf("invalid URL: %v", err)
	}
	q := u.Query()
	if q.Get("uid") != "u4" || q.Get("foo") != "bar" {
		t.Errorf("query params = %v, want uid=u4, foo=bar", q)
	}

	html, err := widget.GetHTMLCode(map[string]string{"width": "600"})
	if err != nil {
		t.Fatalf("GetHTMLCode error: %v", err)
	}
	if !strings.Contains(html, "iframe") || !strings.Contains(html, "width=\"600\"") {
		t.Errorf("iframe HTML = %s, missing expected parts", html)
	}
}
