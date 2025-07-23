package paymentwall

import (
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"hash"
	"reflect"
	"sort"
)

func sortKeys(arr interface{}) []string {
	v := reflect.ValueOf(arr)
	if v.Kind() != reflect.Map {
		return make([]string, 0)
	}

	allKeys := reflect.ValueOf(arr).MapKeys()
	keys := make([]string, len(allKeys))

	for _, v := range allKeys {
		k := v.String()
		if len(k) > 0 {
			keys = append(keys, k)
		}
	}

	sort.Strings(keys)
	return keys
}

func prepareParameters(data map[string]interface{}, baseString string) string {
	base := ""
	keys := sortKeys(data)
	for _, key := range keys {
		value := data[key]
		switch t := value.(type) {
		case nil:
			continue
		case map[string]string:
			ks := sortKeys(t)
			for _, k := range ks {
				if len(k) == 0 {
					continue
				}
				v := t[k]
				base += key + "[" + k + "]" + "=" + v
			}
		default:
			base += key + "=" + fmt.Sprint(t)
		}

	}

	return base + baseString
}

func CalculateSignature(privateKey string, data map[string]interface{}, version int) string {
	str := prepareParameters(data, privateKey)
	var hasher hash.Hash
	switch version {
	case 1, 2:
		hasher = md5.New()
	case 3:
		hasher = sha256.New()
	default:
		panic("invalid signature version")
	}

	hasher.Write([]byte(str))
	return hex.EncodeToString(hasher.Sum(nil))
}
