package http

import (
	"crypto/tls"
	"net/http"
	"testing"
)

func TestBuildPublicURL(t *testing.T) {
	tests := []struct {
		name   string
		base   string
		value  string
		expect string
	}{
		{name: "empty value", base: "https://example.com", value: "", expect: ""},
		{name: "absolute url", base: "https://example.com", value: "https://cdn.example.com/a.jpg", expect: "https://cdn.example.com/a.jpg"},
		{name: "path without base", base: "", value: "uploads/pets/a.jpg", expect: "/uploads/pets/a.jpg"},
		{name: "path with base", base: "https://example.com/", value: "/uploads/pets/a.jpg", expect: "https://example.com/uploads/pets/a.jpg"},
		{name: "trims spaces", base: " https://example.com/api/ ", value: " uploads/pets/a.jpg ", expect: "https://example.com/api/uploads/pets/a.jpg"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := buildPublicURL(test.base, test.value); got != test.expect {
				t.Fatalf("buildPublicURL() = %q, want %q", got, test.expect)
			}
		})
	}
}

func TestPublicBaseURLFromRequest(t *testing.T) {
	t.Run("keeps configured base", func(t *testing.T) {
		request := &http.Request{Host: "api.example.com"}
		got := publicBaseURLFromRequest("https://configured.example.com", request)
		if got != "https://configured.example.com" {
			t.Fatalf("publicBaseURLFromRequest() = %q", got)
		}
	})

	t.Run("uses forwarded proto and host", func(t *testing.T) {
		request := &http.Request{
			Host:   "smart-pet-feeder.alfian-gading.my.id",
			Header: http.Header{"X-Forwarded-Proto": []string{"https"}},
		}
		got := publicBaseURLFromRequest("", request)
		if got != "https://smart-pet-feeder.alfian-gading.my.id" {
			t.Fatalf("publicBaseURLFromRequest() = %q", got)
		}
	})

	t.Run("uses tls when no proxy header is present", func(t *testing.T) {
		request := &http.Request{Host: "api.example.com", TLS: &tls.ConnectionState{}}
		got := publicBaseURLFromRequest("", request)
		if got != "https://api.example.com" {
			t.Fatalf("publicBaseURLFromRequest() = %q", got)
		}
	})
}
