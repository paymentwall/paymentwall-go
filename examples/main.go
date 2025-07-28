package examples

import (
    "fmt"
    paymentwall "github.com/paymentwall/paymentwall-go"
)

func main() {
    // ─────────────────
    // 1) Initialize client
    // ─────────────────
    client := paymentwall.NewClient(
        "YOUR_APPLICATION_KEY",
        "YOUR_SECRET_KEY",
        paymentwall.APIGoods, // Digital Goods API
    )

    // ────────────────────────────────
    // 2) Build & render a checkout widget
    // ────────────────────────────────
    prod, err := paymentwall.NewProduct(
        "product301",                  // external ID
        12.12,                         // amount
        "USD",                         // currency
        "Test Product",                // name
        paymentwall.ProductTypeFixed, // type
        0, "", false, nil,             // no subscription
    )
    if err != nil {
        panic(err)
    }

    widget := paymentwall.NewWidget(
        client,
        "user4522",                    // end-user ID
        "pw",                          // widget code
        []*paymentwall.Product{prod},  // one product
        map[string]any{"email": "user@hostname.com"},
    )

    html, err := widget.GetHTMLCode(nil)
    if err != nil {
        panic(err)
    }
    fmt.Println("=== Widget IFrame ===")
    fmt.Println(html)
    fmt.Println()

    // ───────────────────────────────────────────────
    // 3) Simulate a Pingback (Digital Goods / Goods API)
    // ───────────────────────────────────────────────
    //
    // Imagine Paymentwall calls you with:
    //    uid=u5&goodsid=g1&slength=0&speriod=&type=1&ref=r5
    // plus a correctly-calculated ‘sig’ and (optionally) sign_version.
    //
    params := map[string]any{
        "uid":     "u5",
        "goodsid": "g1",
        "slength": "0",
        "speriod": "",
        "type":    "1",
        "ref":     "r5",
    }
    // choose version 1 (default for Digital Goods) and sign only uid+goodsid:
    sig, err := client.CalculateSignature(
        map[string]any{
            "uid":     params["uid"],
            "goodsid": params["goodsid"],
        },
        paymentwall.SigV1,
    )
    if err != nil {
        panic(err)
    }
    params["sig"] = sig

    // Create pingback and validate
    pb := paymentwall.NewPingback(client, params, "174.36.92.186")
    if ok := pb.Validate(false); !ok {
        fmt.Println("Pingback validation failed:", pb.Errors)
        return
    }

    fmt.Println("=== Pingback OK ===")
    fmt.Printf("UserID: %s\n", pb.GetUserID())
    fmt.Printf("Type: %d\n", mustInt(pb.GetType()))
    fmt.Printf("Product ID: %s\n", pb.GetProductID())

    // Deliver or cancel
    if pb.IsDeliverable() {
        fmt.Println("→ deliver the product")
    } else if pb.IsCancelable() {
        fmt.Println("→ cancel the product")
    }
}

func mustInt(i int, _ error) int { return i }
