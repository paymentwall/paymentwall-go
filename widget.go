// Package paymentwall provides a Go SDK for interacting with the Paymentwall APIs.
package paymentwall

import (
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

// Widget builds Paymentwall widget URLs and HTML.
type Widget struct {
	Client      *Client
	UserID      string
	WidgetCode  string
	Products    []*Product
	ExtraParams map[string]any
}

// NewWidget initializes a new Widget.
func NewWidget(client *Client, userID, widgetCode string, products []*Product, extra map[string]any) *Widget {
	if extra == nil {
		extra = make(map[string]any)
	}
	return &Widget{
		Client:      client,
		UserID:      userID,
		WidgetCode:  widgetCode,
		Products:    products,
		ExtraParams: extra,
	}
}

// getDefaultSignatureVersion returns the default signature version (v3 except v2 for cart).
func (w *Widget) getDefaultSignatureVersion() SignatureVersion {
	if w.Client.APIType == APICart {
		return SigV2
	}
	return SigV3
}

// GetParams constructs the parameter map for the widget, including signature.
func (w *Widget) GetParams() (map[string]any, error) {
	params := map[string]any{
		"key":    w.Client.AppKey,
		"uid":    w.UserID,
		"widget": w.WidgetCode,
	}

	switch w.Client.APIType {
	case APIGoods:
		// Expect exactly one product
		if len(w.Products) > 1 {
			w.Client.AppendError("only one product allowed for API Checkout and empty products for API Goods")
			return params, fmt.Errorf("invalid product count: %d", len(w.Products))
		}

		if len(w.Products) == 1 { 
            prod := w.Products[0]
			var postTrialProduct *Product
			if prod.TrialProduct != nil {
				postTrialProduct = prod
				prod = prod.TrialProduct
			}

			// Basic product fields
			params["amount"] = prod.Amount
			params["currencyCode"] = prod.CurrencyCode
			params["ag_name"] = prod.Name
			params["ag_external_id"] = prod.ID
			params["ag_type"] = prod.Type

			// Subscription fields
			if prod.Type == ProductTypeSubscription {
				params["ag_period_length"] = prod.PeriodLength
				params["ag_period_type"] = prod.PeriodType
				if prod.Recurring {
					params["ag_recurring"] = 1
				} else {
					params["ag_recurring"] = 0
				}

				// Trial product fields
				if postTrialProduct != nil {
					params["ag_trial"] = 1
					params["ag_post_trial_external_id"] = postTrialProduct.ID
					params["ag_post_trial_period_length"] = postTrialProduct.PeriodLength
					params["ag_post_trial_period_type"] = postTrialProduct.PeriodType
					params["ag_post_trial_name"] = postTrialProduct.Name
					params["post_trial_amount"] = postTrialProduct.Amount
					params["post_trial_currencyCode"] = postTrialProduct.CurrencyCode
				}
			}
		}
		
	case APICart:
		// Multiple products
		for i, prod := range w.Products {
			params[fmt.Sprintf("external_ids[%d]", i)] = prod.ID
			if prod.Amount > 0 {
				params[fmt.Sprintf("prices[%d]", i)] = prod.Amount
			}
			if prod.CurrencyCode != "" {
				params[fmt.Sprintf("currencies[%d]", i)] = prod.CurrencyCode
			}
		}
	default:
		// APIVC: no product fields
	}

	// Merge extra params
	for k, v := range w.ExtraParams {
		params[k] = v
	}

	// Determine signature version
	sigVer := w.getDefaultSignatureVersion()
	if sv, ok := w.ExtraParams["sign_version"]; ok {
		if i, err := strconv.Atoi(fmt.Sprint(sv)); err == nil {
			sigVer = SignatureVersion(i)
		}
	}
	params["sign_version"] = int(sigVer)

	// Calculate signature
	sig, err := w.Client.CalculateSignature(params, sigVer)
	if err != nil {
		return params, err
	}
	params["sign"] = sig

	return params, nil
}

// GetURL builds the full widget URL.
func (w *Widget) GetURL() (string, error) {
	params, err := w.GetParams()
	if err != nil {
		return "", err
	}
	controller := w.buildController(w.WidgetCode)
	vals := url.Values{}
	for k, v := range params {
		vals.Set(k, fmt.Sprint(v))
	}
	return fmt.Sprintf("%s/%s?%s", BaseURL, controller, vals.Encode()), nil
}

// GetHTMLCode returns the iframe HTML code for the widget.
func (w *Widget) GetHTMLCode(attrs map[string]string) (string, error) {
	iframeURL, err := w.GetURL()
	if err != nil {
		return "", err
	}
	// Default attributes
	defaultAttrs := map[string]string{"frameborder": "0", "width": "750", "height": "800"}
	// Merge attrs
	for k, v := range attrs {
		defaultAttrs[k] = v
	}
	// Build attribute string
	var parts []string
	for k, v := range defaultAttrs {
		parts = append(parts, fmt.Sprintf(`%s="%s"`, k, v))
	}
	return fmt.Sprintf("<iframe src=\"%s\" %s></iframe>", iframeURL, strings.Join(parts, " ")), nil
}

// buildController selects the appropriate controller path.
func (w *Widget) buildController(code string) string {
	pattern := regexp.MustCompile(`^(w|s|mw)`)
	switch w.Client.APIType {
	case APIVC:
		if !pattern.MatchString(code) {
			return VCController
		}
	case APIGoods:
		if !pattern.MatchString(code) {
			return GoodsController
		}
	}
	return CartController
}
