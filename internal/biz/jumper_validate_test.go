package biz

import (
	"testing"

	"loris-tunnel/internal/model"
)

func TestValidateJumperPayloadAllowsEmptyKeyPathForAutoMode(t *testing.T) {
	payload := model.JumperPayload{
		Name:      "auto",
		Host:      "example.com",
		User:      "ops",
		Port:      22,
		AuthType:  "ssh_key",
		KeyPath:   "",
		TimeoutMs: 5000,
	}
	if err := validateJumperPayload(normalizeJumperPayload(payload)); err != nil {
		t.Fatalf("empty keyPath should be allowed (auto mode), got: %v", err)
	}
}
