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
	return &Pingback{Client: client, Params: params, IPAddress: ip, Errors: []string{}}
}

// Validate runs parameter, IP whitelist, and signature checks.
// skipIPWhitelist allows bypassing the IP check (for testing).
func (p *Pingback) Validate(skipIPWhitelist bool) bool {
	if !p.isParamsValid() {
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
func (p *Pingback) isParamsValid() bool {
	required := []string{"uid", "sig", "type", "ref"}
	switch p.Client.APIType {
	case APIVC:
		required = []string{"uid", "currency", "type", "ref", "sig"}
	case APIGoods:
		required = []string{"uid", "goodsid", "slength", "speriod", "type", "ref", "sig"}
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
	// Whitelisted ranges: Hardcoded list
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

// isSignatureValid recalculates and compares the signature.
func (p *Pingback) isSignatureValid() bool {
	// Determine sign_version
	sv := SigV1
	if val, ok := p.Params["sign_version"]; ok {
		if i, err := strconv.Atoi(fmt.Sprint(val)); err == nil {
			sv = SignatureVersion(i)
		}
	} else if p.Client.APIType == APICart {
		// default for cart if not present
		sv = SigV2
	}

	// Remove "sig" from params copy
	signedParams := make(map[string]any)
	for k, v := range p.Params {
		if k != "sig" {
			signedParams[k] = v
		}
	}

	// For v1 and goodsvc, filter only required fields
	if sv == SigV1 {
		// fields depend on API type
		var keys []string
		if p.Client.APIType == APIVC {
			keys = []string{"uid"}
		} else {
			keys = []string{"uid", "goodsid"}
		}
		p400 := make(map[string]any)
		for _, k := range keys {
			if v, ok := signedParams[k]; ok {
				p400[k] = v
			}
		}
		signedParams = p400
	}

	// Calculate signature
	sigCalc, err := p.Client.calculateSignature(signedParams, sv)
	if err != nil {
		return false
	}
	// Compare
	sigOrig := fmt.Sprint(p.Params["sig"])
	return sigOrig == sigCalc
}

// Helper getters

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
	prod, err := NewProduct(
		fmt.Sprint(p.Params["goodsid"]),
		0, "", "", t, length, periodType, false, nil,
	)
	return prod, err
}

// GetProducts returns a slice of Products for Cart API.
func (p *Pingback) GetProducts() ([]*Product, error) {
	var prods []*Product
	if vals, ok := p.Params["goodsid"].([]any); ok {
		for _, v := range vals {
			id := fmt.Sprint(v)
			prod, err := NewProduct(id, 0, "", "", ProductTypeFixed, 0, "", false, nil)
			if err != nil {
				continue
			}
			prods = append(prods, prod)
		}
	}
	return prods, nil
}

// IsDeliverable returns true if pingback type indicates delivery.
func (p *Pingback) IsDeliverable() bool {
	t, err := p.GetType()
	if err != nil {
		return false
	}
	return t == 0 || t == 1 || t == 201
}

// IsCancelable returns true if pingback type indicates cancellation.
func (p *Pingback) IsCancelable() bool {
	t, err := p.GetType()
	if err != nil {
		return false
	}
	return t == 2 || t == 202
}

// IsUnderReview returns true if pingback type indicates under review.
func (p *Pingback) IsUnderReview() bool {
	t, err := p.GetType()
	if err != nil {
		return false
	}
	return t == 200
}

// ErrorSummary returns accumulated errors.
func (p *Pingback) ErrorSummary() string {
	return strings.Join(p.Errors, "\n")
}
