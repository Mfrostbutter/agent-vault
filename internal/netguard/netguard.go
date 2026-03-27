package netguard

import (
	"context"
	"fmt"
	"net"
	"os"
	"strings"
	"time"
)

// Mode controls which network ranges the proxy is allowed to reach.
type Mode string

const (
	// ModePrivate allows all outbound connections including private ranges.
	// This is the default for local/private deployments.
	ModePrivate Mode = "private"
	// ModePublic blocks connections to private/reserved IP ranges.
	// Use this when Agent Vault is deployed on a public network or cloud.
	ModePublic Mode = "public"
)

// ModeFromEnv reads AGENT_VAULT_NETWORK_MODE and returns the corresponding Mode.
// Returns ModePublic if unset or unrecognized.
func ModeFromEnv() Mode {
	switch strings.ToLower(os.Getenv("AGENT_VAULT_NETWORK_MODE")) {
	case "private":
		return ModePrivate
	default:
		return ModePublic
	}
}

// alwaysBlocked contains IP ranges that are blocked regardless of mode.
// These are metadata service endpoints and other dangerous destinations.
var alwaysBlocked = []net.IPNet{
	// AWS/GCP/Azure IMDS
	parseCIDR("169.254.169.254/32"),
	// AWS IMDSv2 IPv6
	parseCIDR("fd00:ec2::254/128"),
}

// privateRanges contains RFC-1918 and other private/reserved ranges,
// blocked only in "public" mode.
var privateRanges = []net.IPNet{
	// IPv4 private
	parseCIDR("10.0.0.0/8"),
	parseCIDR("172.16.0.0/12"),
	parseCIDR("192.168.0.0/16"),
	// IPv4 loopback
	parseCIDR("127.0.0.0/8"),
	// IPv4 link-local
	parseCIDR("169.254.0.0/16"),
	// IPv4 shared address space (CGN)
	parseCIDR("100.64.0.0/10"),
	// IPv6 loopback
	parseCIDR("::1/128"),
	// IPv6 link-local
	parseCIDR("fe80::/10"),
	// IPv6 unique local
	parseCIDR("fc00::/7"),
	// 0.0.0.0 (often routes to localhost)
	parseCIDR("0.0.0.0/32"),
}

func parseCIDR(s string) net.IPNet {
	_, ipNet, err := net.ParseCIDR(s)
	if err != nil {
		panic("netguard: bad CIDR: " + s)
	}
	return *ipNet
}

// isBlockedIP checks if an IP is blocked for the given mode.
func isBlockedIP(ip net.IP, mode Mode) bool {
	// Always block metadata endpoints.
	for _, n := range alwaysBlocked {
		if n.Contains(ip) {
			return true
		}
	}

	// In public mode, also block private/reserved ranges.
	if mode == ModePublic {
		for _, n := range privateRanges {
			if n.Contains(ip) {
				return true
			}
		}
	}

	return false
}

// SafeDialContext returns a DialContext function that blocks connections to
// forbidden IP ranges based on the network mode.
func SafeDialContext(mode Mode) func(ctx context.Context, network, addr string) (net.Conn, error) {
	dialer := &net.Dialer{
		Timeout:   10 * time.Second,
		KeepAlive: 30 * time.Second,
	}

	return func(ctx context.Context, network, addr string) (net.Conn, error) {
		host, port, err := net.SplitHostPort(addr)
		if err != nil {
			return nil, fmt.Errorf("netguard: invalid address %q: %w", addr, err)
		}

		// Resolve the hostname to IP addresses.
		ips, err := net.DefaultResolver.LookupIPAddr(ctx, host)
		if err != nil {
			return nil, fmt.Errorf("netguard: DNS lookup failed for %q: %w", host, err)
		}

		// Check all resolved IPs before connecting.
		for _, ipAddr := range ips {
			if isBlockedIP(ipAddr.IP, mode) {
				return nil, fmt.Errorf("netguard: connection to %s (%s) blocked by network policy (mode=%s)",
					host, ipAddr.IP.String(), mode)
			}
		}

		// All IPs are safe — connect directly to a validated IP to prevent
		// DNS rebinding (TOCTOU: a second resolution could return a different IP).
		return dialer.DialContext(ctx, network, net.JoinHostPort(ips[0].IP.String(), port))
	}
}
