// widget_test.go
package paymentwall

import (
	"net/url"
	"strings"
	"testing"
)


func TestWidget_GetURL_HTMLCode(t *testing.T) {
	client := NewClient("k","s",APIGoods)
	// single product
	prod, _ := NewProduct("p", 5.0, "USD", "N", ProductTypeFixed, 0, "", false, nil)
	w := NewWidget(client, "u2", "pw", []*Product{prod}, nil)
	urlStr, err := w.GetURL()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(urlStr, BaseURL+"/") {
		t.Errorf("GetURL = %q; want prefix %q", urlStr, BaseURL+"/")
	}
	if _, err := url.Parse(urlStr); err != nil {
		t.Errorf("URL invalid: %v", err)
	}
	iframe, err := w.GetHTMLCode(map[string]string{"onload": `alert("X")`})
	if err != nil {
		t.Fatal(err)
	}
	// must HTML-escape attribute value
	if !strings.Contains(iframe, `onload="alert(&#34;X&#34;)"`) {
		t.Errorf("HTMLCode did not escape: %s", iframe)
	}
}
