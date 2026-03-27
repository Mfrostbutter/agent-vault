package netguard

import (
	"net"
	"testing"
)

func TestIsBlockedIP_AlwaysBlocked(t *testing.T) {
	// IMDS endpoints are blocked in both modes.
	imds := net.ParseIP("169.254.169.254")

	if !isBlockedIP(imds, ModePrivate) {
		t.Error("169.254.169.254 should be blocked in private mode")
	}
	if !isBlockedIP(imds, ModePublic) {
		t.Error("169.254.169.254 should be blocked in public mode")
	}

	// AWS IMDSv2 IPv6
	imdsV6 := net.ParseIP("fd00:ec2::254")
	if !isBlockedIP(imdsV6, ModePrivate) {
		t.Error("fd00:ec2::254 should be blocked in private mode")
	}
}

func TestIsBlockedIP_PrivateMode(t *testing.T) {
	// Private ranges should NOT be blocked in private mode.
	cases := []string{
		"10.0.0.1",
		"172.16.0.1",
		"192.168.1.1",
		"127.0.0.1",
	}
	for _, ip := range cases {
		if isBlockedIP(net.ParseIP(ip), ModePrivate) {
			t.Errorf("%s should NOT be blocked in private mode", ip)
		}
	}
}

func TestIsBlockedIP_PublicMode(t *testing.T) {
	// Private ranges should be blocked in public mode.
	blocked := []string{
		"10.0.0.1",
		"172.16.0.1",
		"192.168.1.1",
		"127.0.0.1",
		"0.0.0.0",
		"100.64.0.1",
	}
	for _, ip := range blocked {
		if !isBlockedIP(net.ParseIP(ip), ModePublic) {
			t.Errorf("%s should be blocked in public mode", ip)
		}
	}

	// Public IPs should NOT be blocked.
	allowed := []string{
		"8.8.8.8",
		"1.1.1.1",
		"104.18.0.1",
	}
	for _, ip := range allowed {
		if isBlockedIP(net.ParseIP(ip), ModePublic) {
			t.Errorf("%s should NOT be blocked in public mode", ip)
		}
	}
}

func TestModeFromEnv(t *testing.T) {
	// Default is public.
	t.Setenv("AGENT_VAULT_NETWORK_MODE", "")
	if ModeFromEnv() != ModePublic {
		t.Error("empty AGENT_VAULT_NETWORK_MODE should default to public")
	}

	t.Setenv("AGENT_VAULT_NETWORK_MODE", "private")
	if ModeFromEnv() != ModePrivate {
		t.Error("AGENT_VAULT_NETWORK_MODE=private should return ModePrivate")
	}

	t.Setenv("AGENT_VAULT_NETWORK_MODE", "PRIVATE")
	if ModeFromEnv() != ModePrivate {
		t.Error("AGENT_VAULT_NETWORK_MODE=PRIVATE should return ModePrivate (case-insensitive)")
	}

	t.Setenv("AGENT_VAULT_NETWORK_MODE", "public")
	if ModeFromEnv() != ModePublic {
		t.Error("AGENT_VAULT_NETWORK_MODE=public should return ModePublic")
	}

	t.Setenv("AGENT_VAULT_NETWORK_MODE", "garbage")
	if ModeFromEnv() != ModePublic {
		t.Error("unrecognized AGENT_VAULT_NETWORK_MODE should default to public")
	}
}
