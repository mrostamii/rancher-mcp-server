package rancher

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNewNormanClient_BaseURLTrim(t *testing.T) {
	c := NewNormanClient("https://rancher.example.com/", "tok", true)
	if !strings.HasSuffix(c.baseURL, "rancher.example.com") || strings.HasSuffix(c.baseURL, "/") {
		t.Fatalf("baseURL = %q", c.baseURL)
	}
}

func TestNormanClient_Do_UsesV3URLPrefix(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v3/schemas" {
			t.Errorf("path = %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer tok" {
			t.Errorf("missing bearer")
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"type":"collection"}`))
	}))
	defer srv.Close()

	c := NewNormanClient(srv.URL, "tok", true)
	b, code, err := c.Do(context.Background(), http.MethodGet, "schemas", nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if code != 200 {
		t.Fatalf("code %d", code)
	}
	if !strings.Contains(string(b), "collection") {
		t.Fatalf("body %s", b)
	}
}

func TestRedactNormanSecrets(t *testing.T) {
	raw := []byte(`{"token":"secret","nested":{"accessKey":"k"}}`)
	out, err := RedactNormanSecrets(false, raw)
	if err != nil {
		t.Fatal(err)
	}
	s := string(out)
	if strings.Contains(s, "secret") || strings.Contains(s, `"k"`) {
		t.Fatalf("not redacted: %s", s)
	}
	out2, err := RedactNormanSecrets(true, raw)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(out2), "secret") {
		t.Fatalf("should not redact when show sensitive")
	}
}
