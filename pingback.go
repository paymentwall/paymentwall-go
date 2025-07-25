// Package paymentwall provides a Go SDK for interacting with the Paymentwall APIs.
package paymentwall

import (
	"fmt"
	"strconv"
	"strings"
)

// Pingback represents a Paymentwall webhook notification validator.
type Pingback struct {
	Client    *Client
	Params    map[string]any
	IPAddress string
	Errors    []string
}

// NewPingback constructs a Pingback with given params and source IP.
func NewPingback(client *Client, params map[string]any, ip string) *Pingback {
	return &Pingback{
		Client:    client,
		Params:    params,
		IPAddress: ip,
		Errors:    []string{},
	}
}

// Validate runs parameter, IP whitelist, and signature checks.
// skipIPWhitelist allows bypassing the IP check (for testing).
func (p *Pingback) Validate(skipIPWhitelist bool) bool {
	if !p.isParamsValid() {
		p.Errors = append(p.Errors, "Missing parameters")
		return false
	}
	if !skipIPWhitelist && !p.isIPAddressValid() {
		p.Errors = append(p.Errors, "IP address is not whitelisted")
		return false
	}
	if !p.isSignatureValid() {
		p.Errors = append(p.Errors, "Wrong signature")
		return false
	}
	return true
}

// isParamsValid checks for required fields and records missing ones.
// Matches Python: VC needs ["uid","currency","type","ref","sig"];
// Goods/Cart need ["uid","goodsid","type","ref","sig"].
func (p *Pingback) isParamsValid() bool {
	var required []string
	if p.Client.APIType == APIVC {
		required = []string{"uid", "currency", "type", "ref", "sig"}
	} else {
		required = []string{"uid", "goodsid", "type", "ref", "sig"}
	}

	valid := true
	for _, key := range required {
		if _, ok := p.Params[key]; !ok {
			p.Errors = append(p.Errors, fmt.Sprintf("Parameter %s is missing", key))
			valid = false
		}
	}
	return valid
}

// isIPAddressValid checks if the source IP is in Paymentwall's whitelist.
func (p *Pingback) isIPAddressValid() bool {
	// Static whitelist entries
	whitelist := []string{
		"174.36.92.186",
		"174.36.96.66",
		"174.36.92.187",
		"174.36.92.192",
		"174.37.14.28",
	}
	// Add 216.127.71.0/24
	for i := 0; i < 256; i++ {
		whitelist = append(whitelist, fmt.Sprintf("216.127.71.%d", i))
	}

	for _, ip := range whitelist {
		if p.IPAddress == ip {
			return true
		}
	}
	return false
}

// isSignatureValid recalculates and compares the signature using Python-style rules.
func (p *Pingback) isSignatureValid() bool {
	// 1) Determine sign_version
	sv := SigV1
	if val, ok := p.Params["sign_version"]; ok {
		if i, err := strconv.Atoi(fmt.Sprint(val)); err == nil {
			sv = SignatureVersion(i)
		}
	} else if p.Client.APIType == APICart {
		// default for Cart
		sv = SigV2
	}

	// 2) Build a copy of params without "sig"
	signedParams := make(map[string]any, len(p.Params))
	for k, v := range p.Params {
		if k == "sig" {
			continue
		}
		signedParams[k] = v
	}

	if sv == SigV1 {
		var fields []string
		switch p.Client.APIType {
		case APIVC:
			fields = []string{"uid", "currency", "type", "ref"}
		case APIGoods:
			fields = []string{"uid", "goodsid", "slength", "speriod", "type", "ref"}
		default: // Cart
			fields = []string{"uid", "goodsid", "type", "ref"}
		}
		filtered := make(map[string]any, len(fields))
		for _, f := range fields {
			if v, ok := signedParams[f]; ok {
				filtered[f] = v
			}
		}
		signedParams = filtered
	}

	// 4) Delegate to Client.CalculateSignature (handles V1, V2, V3 hashing)
	sigCalc, err := p.Client.CalculateSignature(signedParams, sv)
	if err != nil {
		return false
	}

	// 5) Compare to the original
	return fmt.Sprint(p.Params["sig"]) == sigCalc
}


// GetUserID returns the "uid" parameter.
func (p *Pingback) GetUserID() string {
	return fmt.Sprint(p.Params["uid"])
}

// GetType returns the pingback type as int.
func (p *Pingback) GetType() (int, error) {
	return strconv.Atoi(fmt.Sprint(p.Params["type"]))
}

// GetVCAmount returns the "currency" parameter for virtual currency.
func (p *Pingback) GetVCAmount() string {
	return fmt.Sprint(p.Params["currency"])
}

// GetProductID returns the product ID from "goodsid".
func (p *Pingback) GetProductID() string {
	return fmt.Sprint(p.Params["goodsid"])
}

// GetProduct reconstructs a Product from pingback parameters.
func (p *Pingback) GetProduct() (*Product, error) {
	length, _ := strconv.Atoi(fmt.Sprint(p.Params["slength"]))
	periodType := fmt.Sprint(p.Params["speriod"])
	t := ProductTypeFixed
	if length > 0 {
		t = ProductTypeSubscription
	}
	return NewProduct(
		fmt.Sprint(p.Params["goodsid"]),
		0, "", "", t, length, periodType, false, nil,
	)
}

// GetProducts returns a slice of Products for Cart API.
func (p *Pingback) GetProducts() ([]*Product, error) {
	var prods []*Product
	if vals, ok := p.Params["goodsid"].([]any); ok {
		for _, v := range vals {
			id := fmt.Sprint(v)
			if prod, err := NewProduct(id, 0, "", "", ProductTypeFixed, 0, "", false, nil); err == nil {
				prods = append(prods, prod)
			}
		}
	}
	return prods, nil
}

// IsDeliverable returns true if pingback type indicates delivery.
func (p *Pingback) IsDeliverable() bool {
	t, err := p.GetType()
	return err == nil && (t == 0 || t == 1 || t == 201)
}

// IsCancelable returns true if pingback type indicates cancellation.
func (p *Pingback) IsCancelable() bool {
	t, err := p.GetType()
	return err == nil && (t == 2 || t == 202)
}

// IsUnderReview returns true if pingback type indicates under review.
func (p *Pingback) IsUnderReview() bool {
	t, err := p.GetType()
	return err == nil && t == 200
}

// ErrorSummary returns accumulated errors.
func (p *Pingback) ErrorSummary() string {
	return strings.Join(p.Errors, "\n")
}

// GetReferenceID returns the "ref" parameter.
func (p *Pingback) GetReferenceID() string {
	return fmt.Sprint(p.Params["ref"])
}

// GetPingbackUniqueID returns a unique ID composed of ref and type, e.g. "REF123_0".
func (p *Pingback) GetPingbackUniqueID() string {
	ref := p.GetReferenceID()
	t, err := p.GetType()
	if err != nil {
		return ref
	}
	return fmt.Sprintf("%s_%d", ref, t)
}
