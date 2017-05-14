// funcs_test.go
package main

import (
	"fmt"
	"testing"
)

func Test_fetch(t *testing.T) {
	var err error

	results, err := Fetch("https://graph.facebook.com/v2.9/OneInTenBand/events?access_token=280016349127894|eUaDof1f47j6uoTn_tgMsIE5B58")
	if err != nil {
		t.Fatal(err)
	}

	t.Log(results)

	for key, value := range results["data"].([]interface{}) {
		t.Log(fmt.Sprintf("%v: %+v", key, value))
	}
}
