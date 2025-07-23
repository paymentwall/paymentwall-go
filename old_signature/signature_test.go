package paymentwall

import "testing"

func TestCalculateSignature(t *testing.T) {
	privateKey := "my-private-key"
	data := make(map[string]interface{})

	data["key1"] = "Hello"
	data["key2"] = 2
	data["key3"] = map[string]string{"key4": "hello", "key5": "world", "abc": "def"}
	data["key0"] = "yolo"
	data["key11"] = nil
	data["key111"] = 1.2

	correct := "d8aa8cf7e9233345745902cc0278d587c1750c2faa5fe09ac99ff5fd161049e1"
	sign := CalculateSignature(privateKey, data, 3)
	if sign != correct {
		t.Error("Expected: ", correct, " got ", sign)
	}
}
