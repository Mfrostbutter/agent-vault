package cmd

import (
	"testing"

	"github.com/Infisical/agent-vault/internal/broker"
)

func TestHostWarnings(t *testing.T) {
	tests := []struct {
		host     string
		wantLen  int
		contains string
	}{
		{"api.stripe.com", 0, ""},
		{"*.github.com", 0, ""},
		{"stripe", 1, "no dots"},
		{"api stripe com", 2, "whitespace"}, // no dots + whitespace
		{"api.stripe .com", 1, "whitespace"},
	}

	for _, tt := range tests {
		warnings := hostWarnings(tt.host)
		if len(warnings) != tt.wantLen {
			t.Errorf("hostWarnings(%q) returned %d warnings, want %d: %v", tt.host, len(warnings), tt.wantLen, warnings)
		}
		if tt.contains != "" {
			found := false
			for _, w := range warnings {
				if containsStr(w, tt.contains) {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("hostWarnings(%q) = %v, want a warning containing %q", tt.host, warnings, tt.contains)
			}
		}
	}
}

func TestFindDuplicateHosts(t *testing.T) {
	rules := []broker.Rule{
		{Host: "api.stripe.com", Auth: broker.Auth{Type: "bearer", Token: "A"}},
		{Host: "api.github.com", Auth: broker.Auth{Type: "bearer", Token: "A"}},
		{Host: "api.stripe.com", Auth: broker.Auth{Type: "bearer", Token: "B"}},
	}

	dups := findDuplicateHosts(rules)
	if len(dups) != 1 || dups[0] != "api.stripe.com" {
		t.Errorf("findDuplicateHosts() = %v, want [api.stripe.com]", dups)
	}

	noDups := findDuplicateHosts(rules[:2])
	if len(noDups) != 0 {
		t.Errorf("findDuplicateHosts() = %v, want empty", noDups)
	}
}

func TestFindUnresolvedCredentials(t *testing.T) {
	rules := []broker.Rule{
		{
			Host: "api.stripe.com",
			Auth: broker.Auth{Type: "bearer", Token: "STRIPE_KEY"},
		},
		{
			Host: "api.github.com",
			Auth: broker.Auth{Type: "custom", Headers: map[string]string{
				"Authorization": "token {{ GITHUB_TOKEN }}",
				"X-Extra":       "{{ MISSING_SECRET }}",
			}},
		},
	}

	known := []string{"STRIPE_KEY", "GITHUB_TOKEN"}
	unresolved := findUnresolvedCredentials(rules, known)
	if len(unresolved) != 1 || unresolved[0] != "MISSING_SECRET" {
		t.Errorf("findUnresolvedCredentials() = %v, want [MISSING_SECRET]", unresolved)
	}

	allKnown := []string{"STRIPE_KEY", "GITHUB_TOKEN", "MISSING_SECRET"}
	unresolved2 := findUnresolvedCredentials(rules, allKnown)
	if len(unresolved2) != 0 {
		t.Errorf("findUnresolvedCredentials() = %v, want empty", unresolved2)
	}
}

func TestMergeRules(t *testing.T) {
	existing := []broker.Rule{
		{Host: "api.stripe.com", Auth: broker.Auth{Type: "bearer", Token: "A"}},
	}
	newRules := []broker.Rule{
		{Host: "api.github.com", Auth: broker.Auth{Type: "bearer", Token: "B"}},
	}

	// Append
	appended := mergeRules(existing, newRules, mergeAppend)
	if len(appended) != 2 {
		t.Errorf("mergeRules(append) = %d rules, want 2", len(appended))
	}
	if appended[0].Host != "api.stripe.com" || appended[1].Host != "api.github.com" {
		t.Errorf("mergeRules(append) hosts = [%s, %s], want [api.stripe.com, api.github.com]",
			appended[0].Host, appended[1].Host)
	}

	// Replace
	replaced := mergeRules(existing, newRules, mergeReplace)
	if len(replaced) != 1 {
		t.Errorf("mergeRules(replace) = %d rules, want 1", len(replaced))
	}
	if replaced[0].Host != "api.github.com" {
		t.Errorf("mergeRules(replace) host = %s, want api.github.com", replaced[0].Host)
	}
}

func TestPolicySetInteractive_NonTTY(t *testing.T) {
	// When running in tests, stdin is not a TTY, so the interactive mode
	// should fail with the TTY-required message.
	output, err := executeCommand("vault", "policy", "set")
	if err == nil {
		t.Fatal("expected error for non-TTY interactive mode, got nil")
	}
	errMsg := err.Error()
	if !containsStr(errMsg, "interactive mode requires a terminal") {
		t.Errorf("expected TTY error, got: %s (output: %s)", errMsg, output)
	}
}

func containsStr(s, substr string) bool {
	return len(s) >= len(substr) && searchStr(s, substr)
}

func searchStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
