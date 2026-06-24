package xray

import (
	"testing"

	"github.com/google/uuid"
	"vpn-client/internal/vless"
)

func TestValidateStealthGlobalConfig(t *testing.T) {
	cfg := &vless.Config{
		UUID:        uuid.MustParse("883e028c-4e41-43c0-8823-c2a107c40c60"),
		Host:        "example.com",
		Port:        443,
		Security:    vless.SecurityTLS,
		Encryption:  "none",
		SNI:         "example.com",
		Fingerprint: "chrome",
		Type:        "ws",
		Path:        "/ws",
	}
	rules := &RouteRules{
		DomainStrategy: "ipifnonmatch",
		GlobalProxy:    true,
	}

	if err := ValidateConfig(cfg, rules, 18080, false); err != nil {
		t.Fatalf("proxy config should validate: %v", err)
	}
	if err := ValidateConfig(cfg, rules, 18080, true); err != nil {
		t.Fatalf("TUN config should validate: %v", err)
	}
}
