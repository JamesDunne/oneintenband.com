// funcs_test.go
package main

import (
	"fmt"
	"testing"
)

func Test_fetch(t *testing.T) {
	var err error

	results, err := Fetch("https://graph.facebook.com/v2.9/OneInTenBand/events?access_token=EAACEdEose0cBAP1JrZBu1mSzVpk5OpYUFJqRA0qnGCN9wAWHbWduXdFPDABLKQIhQQ45bmfhlQ1fInhCpZBxU0yqJCF1OJo6NQZB9o0eHIyXB3cEYHZB0hRR30ee8g6eqaN4ZAEbPqpogfggdbm5ZAJpAHFhRTPAXdPEprWOllkZAgEsfQaJZAF8EBhHohh2RGoZD")
	if err != nil {
		t.Fatal(err)
	}

	t.Log(results)

	for key, value := range results["data"].([]interface{}) {
		t.Log(fmt.Sprintf("%v: %+v", key, value))
	}
}
