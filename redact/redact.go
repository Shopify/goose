package redact

import (
	"net/http"
	"strings"
)

const placeholderText = "[FILTERED]"

var GlobalSensitiveSubstrings = []string{
	"authorization",
	"cookie",
	"token",
	"password",
	"secret",
}

func AddSensitiveSubstring(substrings ...string) {
	lowerSubstrings := make([]string, len(substrings))
	for i, s := range substrings {
		lowerSubstrings[i] = strings.ToLower(s)
	}
	GlobalSensitiveSubstrings = append(GlobalSensitiveSubstrings, lowerSubstrings...)
}

func IsSensitive(key string) bool {
	for _, sensitiveKey := range GlobalSensitiveSubstrings {
		if strings.Contains(strings.ToLower(key), sensitiveKey) {
			return true
		}
	}
	return false
}

func Map(data map[string]interface{}) map[string]interface{} {
	redactedData := make(map[string]interface{}, len(data))

	for key, value := range data {
		if IsSensitive(key) {
			redactedData[key] = placeholderText
		} else {
			redactedData[key] = value
		}
	}

	return redactedData
}

func Headers(headers http.Header) map[string]string {
	redactedData := make(map[string]string, len(headers))

	for key, value := range headers {
		if IsSensitive(key) {
			redactedData[key] = placeholderText
		} else {
			redactedData[key] = strings.Join(value, ",")
		}
	}

	return redactedData
}
