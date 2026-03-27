package proposal

import (
	"testing"

	"github.com/Infisical/agent-vault/internal/broker"
)

func mergeBearer(token string) *broker.Auth {
	return &broker.Auth{Type: "bearer", Token: token}
}

func TestMergeRulesSetAppend(t *testing.T) {
	existing := []broker.Rule{
		{Host: "api.github.com", Auth: broker.Auth{Type: "bearer", Token: "GH"}},
	}
	proposed := []Rule{
		{Action: ActionSet, Host: "api.stripe.com", Description: "Stripe", Auth: mergeBearer("SK")},
	}

	merged, warnings := MergeRules(existing, proposed)
	if len(warnings) != 0 {
		t.Fatalf("expected no warnings, got %v", warnings)
	}
	if len(merged) != 2 {
		t.Fatalf("expected 2 rules, got %d", len(merged))
	}
	if merged[1].Host != "api.stripe.com" {
		t.Fatalf("expected appended host api.stripe.com, got %s", merged[1].Host)
	}
	if merged[1].Description == nil || *merged[1].Description != "Stripe" {
		t.Fatalf("expected description 'Stripe'")
	}
}

func TestMergeRulesSetReplacesExisting(t *testing.T) {
	existing := []broker.Rule{
		{Host: "api.stripe.com", Auth: broker.Auth{Type: "bearer", Token: "OLD"}},
	}
	proposed := []Rule{
		{Action: ActionSet, Host: "api.stripe.com", Auth: mergeBearer("NEW")},
	}

	merged, warnings := MergeRules(existing, proposed)
	if len(warnings) != 0 {
		t.Fatalf("expected no warnings, got %v", warnings)
	}
	if len(merged) != 1 {
		t.Fatalf("expected 1 rule, got %d", len(merged))
	}
	if merged[0].Auth.Token != "NEW" {
		t.Fatalf("expected replaced rule with token NEW, got %s", merged[0].Auth.Token)
	}
}

func TestMergeRulesDelete(t *testing.T) {
	existing := []broker.Rule{
		{Host: "api.github.com", Auth: broker.Auth{Type: "bearer", Token: "GH"}},
		{Host: "api.stripe.com", Auth: broker.Auth{Type: "bearer", Token: "SK"}},
	}
	proposed := []Rule{
		{Action: ActionDelete, Host: "api.stripe.com"},
	}

	merged, warnings := MergeRules(existing, proposed)
	if len(warnings) != 0 {
		t.Fatalf("expected no warnings, got %v", warnings)
	}
	if len(merged) != 1 {
		t.Fatalf("expected 1 rule after delete, got %d", len(merged))
	}
	if merged[0].Host != "api.github.com" {
		t.Fatalf("expected remaining host api.github.com, got %s", merged[0].Host)
	}
}

func TestMergeRulesDeleteNonExistent(t *testing.T) {
	existing := []broker.Rule{
		{Host: "api.github.com", Auth: broker.Auth{Type: "bearer", Token: "GH"}},
	}
	proposed := []Rule{
		{Action: ActionDelete, Host: "api.stripe.com"},
	}

	merged, warnings := MergeRules(existing, proposed)
	if len(warnings) != 1 {
		t.Fatalf("expected 1 warning, got %d", len(warnings))
	}
	if len(merged) != 1 {
		t.Fatalf("expected 1 rule unchanged, got %d", len(merged))
	}
}

func TestMergeRulesMixed(t *testing.T) {
	existing := []broker.Rule{
		{Host: "api.github.com", Auth: broker.Auth{Type: "bearer", Token: "GH"}},
		{Host: "api.slack.com", Auth: broker.Auth{Type: "bearer", Token: "SLACK"}},
	}
	proposed := []Rule{
		{Action: ActionSet, Host: "api.stripe.com", Auth: mergeBearer("SK")},
		{Action: ActionDelete, Host: "api.slack.com"},
		{Action: ActionSet, Host: "api.github.com", Auth: mergeBearer("GH_NEW")},
	}

	merged, warnings := MergeRules(existing, proposed)
	if len(warnings) != 0 {
		t.Fatalf("expected no warnings, got %v", warnings)
	}
	if len(merged) != 2 {
		t.Fatalf("expected 2 rules (1 added, 1 updated, 1 deleted), got %d", len(merged))
	}
	if merged[0].Auth.Token != "GH_NEW" {
		t.Fatalf("expected updated github rule")
	}
	if merged[1].Host != "api.stripe.com" {
		t.Fatalf("expected stripe appended")
	}
}

func TestMergeRulesEmpty(t *testing.T) {
	merged, warnings := MergeRules(nil, []Rule{
		{Action: ActionSet, Host: "example.com", Auth: mergeBearer("KEY")},
	})
	if len(warnings) != 0 {
		t.Fatalf("expected no warnings, got %v", warnings)
	}
	if len(merged) != 1 {
		t.Fatalf("expected 1 rule, got %d", len(merged))
	}
}

func TestMergeRulesNoDescription(t *testing.T) {
	merged, _ := MergeRules(nil, []Rule{
		{Action: ActionSet, Host: "example.com", Auth: mergeBearer("KEY")},
	})
	if merged[0].Description != nil {
		t.Fatalf("expected nil description, got %v", merged[0].Description)
	}
}

func TestMergeRulesBasicAuth(t *testing.T) {
	existing := []broker.Rule{
		{Host: "api.ashby.com", Auth: broker.Auth{Type: "bearer", Token: "OLD"}},
	}
	proposed := []Rule{
		{Action: ActionSet, Host: "api.ashby.com", Auth: &broker.Auth{Type: "basic", Username: "ASHBY_KEY"}},
	}
	merged, _ := MergeRules(existing, proposed)
	if merged[0].Auth.Type != "basic" {
		t.Fatalf("expected basic auth type, got %s", merged[0].Auth.Type)
	}
	if merged[0].Auth.Username != "ASHBY_KEY" {
		t.Fatalf("expected username ASHBY_KEY, got %s", merged[0].Auth.Username)
	}
}
