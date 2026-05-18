package http

import "strings"

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
