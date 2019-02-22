package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"runtime"
	"strings"
	"testing"
)

func TestServe(t *testing.T) {
	username := "foo"
	password := "bar"

	fakeFedora := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u, p, ok := r.BasicAuth()
		if !ok || u != username || p != password {
			t.Fatalf("Basic auth is not correct")
		}
		w.Header().Set("Content-Type", "application/ld+json")
		w.Write([]byte(`{
			"@context": "http://example.org/foo",
			"schemas": ["http://example.org/schemas/test"],
			"foo": "bar"
		}`))
	}))
	defer fakeFedora.Close()

	os.Args = []string{"schemas", "serve",
		"-u", username,
		"-p", password,
		"testdata/schema.json"}

	go main()

	for serveOpts.port == 0 {
		runtime.Gosched()
	}

	requestBody := fmt.Sprintf(`["%s"]`, fakeFedora.URL)

	resp, err := http.Post(
		fmt.Sprintf("http://localhost:%d", serveOpts.port), "application/json", strings.NewReader(requestBody))
	if err != nil {
		t.Fatalf("schema service error %+v", err)
	}

	if resp.StatusCode != 200 {
		t.Fatalf("schema service request failed!")
	}

	defer resp.Body.Close()

	var schemas []interface{}

	_ = json.NewDecoder(resp.Body).Decode(&schemas)

	if len(schemas) == 0 {
		t.Fatalf("Did not get any schemas!")
	}
}
