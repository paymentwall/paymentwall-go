# About Paymentwall

[Paymentwall](https://paymentwall.com) is the leading digital payments platform for globally monetizing digital goods and services. Paymentwall assists game publishers, dating sites, rewards sites, SaaS companies and many other verticals to monetize their digital content and services. Merchants can plug in Paymentwall’s API to accept payments from over 100 different methods including credit cards, debit cards, bank transfers, SMS/Mobile payments, prepaid cards, eWallets, landline payments and others.

To sign up for a Paymentwall Merchant Account, [click here](https://paymentwall.com/signup/merchant).

# Paymentwall Golang Library
This library allows developers to use [Paymentwall APIs](https://docs.paymentwall.com/) (Virtual Currency, Digital Goods featuring recurring billing, and Virtual Cart).

To use Paymentwall, all you need to do is to sign up for a Paymentwall Merchant Account so you can setup an Application designed for your site.
To open your merchant account and set up an application, you can [sign up here](http://paymentwall.com/signup/merchant).


---

## Installation

Requires Go 1.18+ (modules)

```bash
go get github.com/paymentwall/paymentwall-go
```

In your code:

```go
import "github.com/paymentwall/paymentwall-go"
```

---

# Code Samples

## Checkout API & Digital Goods API

[Web API details for API Checkout](https://docs.paymentwall.com/apis#section-checkout-onetime) </br>
[Web API details for API Goods](https://docs.paymentwall.com/apis#section-widget-dg)

### Initializing Paymentwall

```go
client := paymentwall.NewClient(
  "YOUR_APPLICATION_KEY",
  "YOUR_SECRET_KEY",
  paymentwall.APIGoods, // Digital Goods API
)
```

### Widget Call
[Web API details for Checkout](https://docs.paymentwall.com/apis#section-checkout-onetime) </br>
[Web API details for Goods](https://docs.paymentwall.com/apis#section-widget-dg)

The widget is a payment page hosted by Paymentwall that embeds the entire payment flow: selecting the payment method, completing the billing details, and providing customer support via the Help section. You can redirect the users to this page or embed it via iframe. Below is an example that renders an iframe with Paymentwall Widget.

```go
// 1) Create a Product (for checkout API)
prod, err := paymentwall.NewProduct(
  "product301",               // external ID
  12.12,                      // amount
  "USD",                      // currency
  "Test Product",             // name
  paymentwall.ProductTypeFixed, // type
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
  []*paymentwall.Product{prod}, // Checkout API require 1 product, let empty for digital good API
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
The Pingback is a webhook notifying about a payment being made. Pingbacks are sent via HTTP/HTTPS to your servers. To process pingbacks use the following guide:

```go
// pingback
pb := paymentwall.NewPingback(client, "query params", "remote_address")
if !pb.Validate(false) {
  // handle validation errors
  fmt.Println(pb.ErrorSummary())
  return
}

if pb.IsDeliverable() {
  // deliver the product
} else if pb.IsCancelable() {
  // withdraw the product
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
// pingback
pb := paymentwall.NewPingback(client, "query params", "remote_address")
if !pb.Validate(false) {
  // handle validation errors
  fmt.Println(pb.ErrorSummary())
  return
}

if pb.IsDeliverable() {
  // deliver the product
} else if pb.IsCancelable() {
  // withdraw the product
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
// pingback
pb := paymentwall.NewPingback(client, "query params", "remote_address")
if !pb.Validate(false) {
  // handle validation errors
  fmt.Println(pb.ErrorSummary())
  return
}

if pb.IsDeliverable() {
  // deliver the product
} else if pb.IsCancelable() {
  // withdraw the product
}
```

---

## Contributing & Support

* Report issues at [https://github.com/paymentwall/paymentwall-go/issues](https://github.com/paymentwall/paymentwall-go/issues).

---
