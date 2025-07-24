# About Paymentwall

[Paymentwall](https://paymentwall.com?source=gh-go) is the leading digital payments platform for globally monetizing digital goods and services. Paymentwall assists game publishers, dating sites, rewards sites, SaaS companies and many other verticals to monetize their digital content and services. Merchants can plug in Paymentwall’s API to accept payments from over 100 different methods including credit cards, debit cards, bank transfers, SMS/Mobile payments, prepaid cards, eWallets, landline payments and others.

To sign up for a Paymentwall Merchant Account, [click here](https://paymentwall.com/signup/merchant?source=gh-go).

# Paymentwall Go Library

This Go module provides a thin, idiomatic wrapper around the Paymentwall APIs:

* **Virtual Currency** (`APIVC`)
* **Digital Goods** (`APIGoods`)
* **Cart / Shopping Cart** (`APICart`)

It mirrors the behavior of our official Python and Node.js SDKs, with Go-style APIs and error handling.

---

## Installation

Requires Go 1.16+ (modules)

```bash
go get github.com/SamoySamoy/paymentwall-go
```

In your code:

```go
import "github.com/SamoySamoy/paymentwall-go"
```

---

# Code Samples

## Checkout / Digital Goods API

[Web API details](https://docs.paymentwall.com/apis#section-checkout-onetime)

### Initializing Paymentwall

```go
client := paymentwall.NewClient(
  "YOUR_APPLICATION_KEY",
  "YOUR_SECRET_KEY",
  paymentwall.APIGoods, // Digital Goods API
)
```

### Widget Call

```go
// 1) Create a one-time or subscription Product
prod, err := paymentwall.NewProduct(
  "product301",               // external ID
  12.12,                      // amount
  "USD",                      // currency
  "Test Product",             // name
  paymentwall.ProductTypeFixed,
  0, "", false, nil,
)
if err != nil {
  panic(err)
}

// 2) Build the Widget
widget := paymentwall.NewWidget(
  client,
  "user4522",                 // your end-user ID
  "pw",                       // widget code from Merchant Area
  []*paymentwall.Product{prod},
  map[string]any{"email":"user@hostname.com"},
)

// 3) Get iframe HTML
html, err := widget.GetHTMLCode(nil)
if err != nil {
  panic(err)
}
fmt.Println(html)
```

### Pingback Processing

```go
// Suppose you’ve parsed all query params into a map[string]any:
params := map[string]any{
  "uid":    "user4522",
  "goodsid":"product301",
  "slength":"0",
  "speriod":"",
  "type":   "0",
  "ref":    "REF123",
}

// Calculate signature (SigV1 default for Goods)
sig, _ := client.calculateSignature(
  map[string]any{"uid":"user4522","goodsid":"product301"},
  paymentwall.SigV1,
)
params["sig"] = sig

pb := paymentwall.NewPingback(client, params, "174.36.92.186")
if !pb.Validate(false) {
  // handle validation errors
  fmt.Println(pb.ErrorSummary())
  return
}

if pb.IsDeliverable() {
  // deliver
} else if pb.IsCancelable() {
  // cancel
}
```

---

## Virtual Currency API

[Web API details](https://www.paymentwall.com/en/documentation/Virtual-Currency-API/711)

```go
client := paymentwall.NewClient("APP_KEY", "SECRET_KEY", paymentwall.APIVC)

widget := paymentwall.NewWidget(
  client,
  "user40012",
  "vc_widget_code",
  []*paymentwall.Product{},              // always empty
  map[string]any{"email":"user@hostname.com"},
)
fmt.Println(widget.GetHTMLCode(nil))
```

```go
// Pingback
params := map[string]any{
  "uid":      "user40012",
  "currency": "100",   // virtual currency amount
  "type":     "0",
  "ref":      "REF456",
}
sig, _ := client.calculateSignature(
  map[string]any{"uid":"user40012"},
  paymentwall.SigV1,
)
params["sig"] = sig

pb := paymentwall.NewPingback(client, params, "174.36.92.186")
if pb.Validate(false) && pb.IsDeliverable() {
  // grant virtual currency
}
```

---

## Cart API

[Web API details](https://www.paymentwall.com/en/documentation/Shopping-Cart-API/1098)

```go
client := paymentwall.NewClient("APP_KEY", "SECRET_KEY", paymentwall.APICart)

prod1, _ := paymentwall.NewProduct("product301", 3.33, "EUR", "Item A", paymentwall.ProductTypeFixed, 0, "", false, nil)
prod2, _ := paymentwall.NewProduct("product607", 7.77, "EUR", "Item B", paymentwall.ProductTypeFixed, 0, "", false, nil)

widget := paymentwall.NewWidget(
  client,
  "user40012",
  "cart_widget_code",
  []*paymentwall.Product{prod1, prod2},
  map[string]any{"email":"user@hostname.com"},
)
fmt.Println(widget.GetHTMLCode(nil))
```

```go
// Cart pingback
params := map[string]any{
  "uid":      "user40012",
  "goodsid":  []any{"product301","product607"},
  "type":     "0",
  "ref":      "REF789",
}
sig, _ := client.calculateSignature(params, paymentwall.SigV2) // Cart defaults to SigV2
params["sig"] = sig

pb := paymentwall.NewPingback(client, params, "174.36.92.186")
if pb.Validate(false) {
  for _, p := range pb.GetProducts() {
    fmt.Println("Deliver product:", p.ID)
  }
}
```

---

## Contributing & Support

* See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.
* Report issues at [https://github.com/SamoySamoy/paymentwall-go/issues](https://github.com/SamoySamoy/paymentwall-go/issues).

---
