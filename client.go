// Package paymentwall provides a Go SDK for interacting with the Paymentwall APIs.
package paymentwall

import (
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
)

// APIType identifies the Paymentwall API mode.
type APIType int

const (
	APIVC    APIType = 1 // Virtual Currency
	APIGoods APIType = 2 // Digital Goods
	APICart  APIType = 3 // Cart API
)

// SignatureVersion denotes the version of the signature algorithm.
type SignatureVersion int

const (
	SigV1 SignatureVersion = 1 // MD5(params + secret)
	SigV2 SignatureVersion = 2 // MD5(sorted params + secret)
	SigV3 SignatureVersion = 3 // SHA256(sorted params + secret)
)

const (
	VCController    = "ps"
	GoodsController = "subscription"
	CartController  = "cart"
	BaseURL         = "https://api.paymentwall.com/api"
	VersionString   = "0.1.1"
)

// Client holds global configuration and error state for the SDK.
type Client struct {
	APIType   APIType
	AppKey    string
	SecretKey string
	Errors    []string
}

// NewClient initializes a Paymentwall Client with the given keys and API type.
func NewClient(appKey, secretKey string, api APIType) *Client {
	return &Client{
		APIType:   api,
		AppKey:    appKey,
		SecretKey: secretKey,
		Errors:    []string{},
	}
}

// SetAPIType allows overriding the API type on an existing client.
func (c *Client) SetAPIType(api APIType) {
	c.APIType = api
}

// AppendError records an error message in the client's error list.
func (c *Client) AppendError(err string) {
	c.Errors = append(c.Errors, err)
}

// ErrorSummary returns all recorded errors as a single string.
func (c *Client) ErrorSummary() string {
	return strings.Join(c.Errors, "\n")
}

// hashMD5 computes the MD5 hash of the input string and returns its hex encoding.
func hashMD5(s string) string {
	h := md5.Sum([]byte(s))
	return hex.EncodeToString(h[:])
}

// hashSHA256 computes the SHA256 hash of the input string and returns its hex encoding.
func hashSHA256(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:])
}

// CalculateSignature builds a signature string based on the provided parameters and version.
// It implements SigV1, SigV2, and SigV3 signature algorithms as per Paymentwall's documentation.
func (c *Client) CalculateSignature(
	params map[string]any,
	version SignatureVersion,
) (string, error) {
	if c.SecretKey == "" {
		return "", fmt.Errorf("secret key cannot be empty")
	}

	var base strings.Builder
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}

	switch version {
	case SigV1:
		// v1: MD5 of parameters in provided order
		// Maintain order as provided by pingback.go
		for _, k := range keys {
			v := params[k]
			switch val := v.(type) {
			case []any:
				for i, item := range val {
					base.WriteString(fmt.Sprintf("%s[%d]=%v", k, i, item))
				}
			default:
				// Handle nil or missing values as empty string, matching Python's None
				if val == nil {
					base.WriteString(fmt.Sprintf("%s=", k))
				} else {
					base.WriteString(fmt.Sprintf("%s=%v", k, val))
				}
			}
		}
		base.WriteString(c.SecretKey)
		return hashMD5(base.String()), nil

	case SigV2, SigV3:
		// v2/v3: sorted key=value pairs + secret
		sort.Strings(keys)
		for _, k := range keys {
			v := params[k]
			switch val := v.(type) {
			case []any:
				for i, item := range val {
					base.WriteString(fmt.Sprintf("%s[%d]=%v", k, i, item))
				}
			default:
				// Handle nil or missing values as empty string
				if val == nil {
					base.WriteString(fmt.Sprintf("%s=", k))
				} else {
					base.WriteString(fmt.Sprintf("%s=%v", k, val))
				}
			}
		}
		base.WriteString(c.SecretKey)
		if version == SigV2 {
			return hashMD5(base.String()), nil
		}
		return hashSHA256(base.String()), nil

	default:
		return "", fmt.Errorf("unsupported signature version: %d", version)
	}
}