// Package paymentwall defines common error variables for the SDK.
package paymentwall

import "errors"

// Sentinel errors for SDK operations.
var (
	// ErrEmptySecretKey indicates that the secret key was not provided.
	ErrEmptySecretKey = errors.New("secret key cannot be empty")

	// ErrInvalidProductCount indicates an incorrect number of products for the current API.
	ErrInvalidProductCount = errors.New("invalid product count")
)
