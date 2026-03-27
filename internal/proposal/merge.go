package proposal

import (
	"fmt"

	"github.com/Infisical/agent-vault/internal/broker"
)

// MergeRules applies proposed rule changes to existing rules.
// Set-action rules upsert (add or replace); delete-action rules remove.
// Returns the merged slice and a list of warnings for no-op operations.
func MergeRules(existing []broker.Rule, proposed []Rule) ([]broker.Rule, []string) {
	// Index existing rules by host for O(1) lookup.
	hostIndex := make(map[string]int, len(existing))
	for i, r := range existing {
		hostIndex[r.Host] = i
	}

	merged := make([]broker.Rule, len(existing))
	copy(merged, existing)

	// Track which indices to remove (from delete actions).
	removeSet := make(map[int]bool)

	var warnings []string
	for _, p := range proposed {
		switch p.Action {
		case ActionDelete:
			idx, exists := hostIndex[p.Host]
			if !exists {
				warnings = append(warnings, fmt.Sprintf("skipped delete for %q: host not found", p.Host))
				continue
			}
			removeSet[idx] = true
			delete(hostIndex, p.Host)

		default: // ActionSet: upsert
			toBrokerRule := toBrokerRule(p)
			if idx, exists := hostIndex[p.Host]; exists {
				// Replace existing rule in place.
				merged[idx] = toBrokerRule
			} else {
				// Append new rule.
				hostIndex[p.Host] = len(merged)
				merged = append(merged, toBrokerRule)
			}
		}
	}

	// Remove deleted rules (iterate in reverse-stable order).
	if len(removeSet) > 0 {
		result := make([]broker.Rule, 0, len(merged)-len(removeSet))
		for i, r := range merged {
			if !removeSet[i] {
				result = append(result, r)
			}
		}
		merged = result
	}

	return merged, warnings
}

func toBrokerRule(p Rule) broker.Rule {
	var desc *string
	if p.Description != "" {
		d := p.Description
		desc = &d
	}
	rule := broker.Rule{
		Host:        p.Host,
		Description: desc,
	}
	if p.Auth != nil {
		rule.Auth = *p.Auth
	}
	return rule
}
