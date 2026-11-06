package main

import "strings"

// Backend service endpoints shared by AI debug and cloud sync.
const (
	devBackendAPIBaseURL  = "http://127.0.0.1:8000/api/v1"
	prodBackendAPIBaseURL = "https://loris-tunnel-prod.flyml.net/api/v1"
)

func backendAPIBaseURL(buildType string) string {
	if strings.EqualFold(strings.TrimSpace(buildType), "dev") {
		return devBackendAPIBaseURL
	}
	return prodBackendAPIBaseURL
}

func backendEnvironment(buildType string) string {
	if strings.EqualFold(strings.TrimSpace(buildType), "dev") {
		return "dev"
	}
	return "prod"
}

func backendServerBaseURL() string {
	return strings.TrimSuffix(prodBackendAPIBaseURL, "/api/v1")
}
