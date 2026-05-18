package http

import (
	"net/http"
	"strings"
)

func buildPublicURL(baseURL, value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	if strings.HasPrefix(value, "http://") || strings.HasPrefix(value, "https://") {
		return value
	}
	if !strings.HasPrefix(value, "/") {
		value = "/" + value
	}
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if baseURL == "" {
		return value
	}
	return baseURL + value
}

func publicBaseURLFromRequest(configuredBaseURL string, request *http.Request) string {
	configuredBaseURL = strings.TrimSpace(configuredBaseURL)
	if configuredBaseURL != "" || request == nil {
		return configuredBaseURL
	}

	host := strings.TrimSpace(request.Host)
	if host == "" {
		return ""
	}

	scheme := strings.TrimSpace(request.Header.Get("X-Forwarded-Proto"))
	if scheme == "" {
		scheme = strings.TrimSpace(request.Header.Get("X-Forwarded-Scheme"))
	}
	if scheme == "" {
		if request.TLS != nil {
			scheme = "https"
		} else {
			scheme = "http"
		}
	}

	if comma := strings.Index(scheme, ","); comma >= 0 {
		scheme = strings.TrimSpace(scheme[:comma])
	}
	if scheme != "http" && scheme != "https" {
		scheme = "https"
	}

	return scheme + "://" + host
}
