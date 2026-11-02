package plutus

import (
	"context"
	"encoding/json"
	"sort"
	"unicode/utf8"

	"github.com/VynnLockwood/kuronuma/apps/api/internal/core/hermes"
	"github.com/VynnLockwood/kuronuma/apps/api/internal/core/themis"
)

const (
	msgRuleAppliesLater = "Just so you know, this applies to new transactions only; saved records are unchanged."
	msgRuleLimit        = "Sorry, the user already has %d payee rules, the maximum. Delete one with delete_payee_rule first."
)

// registerPayeeTools adds the payee rule tools. They use the existing plutus scopes.
func (m *Module) registerPayeeTools(s *hermes.Server) {
	s.AddTool(hermes.Tool{
		Name:  "set_payee_rule",
		Title: "Remember what a payee is",
		Description: "Teach Kuronuma what a payee is, for example 7-Eleven is a convenience store in category food. From then on, new spending whose merchant " +
			"contains this text is filed under that category when you give none. Matching ignores case and spacing and uses contains, so a short payee " +
			"can match unrelated merchants: choose a distinctive name. Ask the user before saving a rule. Rules affect only future transactions; saved " +
			"records are not changed. Saving the same payee again updates its rule.",
		Scope: ScopeWrite,
		InputSchema: hermes.Schema(map[string]any{
			"payee":       hermes.Prop("string", "Payee or part of the merchant name, at least 2 characters (max 100)."),
			"category":    hermes.Prop("string", "A category from the user's list (list_categories shows it), or uncategorized."),
			"description": hermes.Prop("string", "Optional note about the payee, for example Thai food store (max 200 characters)."),
		}, "payee", "category"),
		Handler: m.setPayeeRule,
	})
	s.AddTool(hermes.Tool{
		Name:        "list_payee_rules",
		Title:       "List payee rules",
		Description: "List the payee rules the user has saved, sorted by payee.",
		Scope:       ScopeRead,
		ReadOnly:    true,
		InputSchema: hermes.Schema(map[string]any{}),
		Handler:     m.listPayeeRules,
	})
	s.AddTool(hermes.Tool{
		Name:  "delete_payee_rule",
		Title: "Delete a payee rule",
		Description: "Delete the payee rule for one payee (matched ignoring case and spacing). Only do this when the user asks, " +
			"and confirm which rule first (list_payee_rules shows them). Saved transactions are not changed.",
		Scope:       ScopeWrite,
		Destructive: true,
		InputSchema: hermes.Schema(map[string]any{"payee": hermes.Prop("string", "The payee of the rule to delete.")}, "payee"),
		Handler:     m.deletePayeeRule,
	})
}

type ruleOut struct {
	Payee       string `json:"payee"`
	Category    string `json:"category"`
	Description string `json:"description"`
}

func toRuleOut(r PayeeRule) ruleOut {
	return ruleOut{
		Payee:       hermes.Untrusted(r.Payee, maxPayeeLen),
		Category:    hermes.Untrusted(r.Category, 40),
		Description: hermes.Untrusted(r.Description, maxPayeeDescription),
	}
}

func (m *Module) setPayeeRule(ctx context.Context, c hermes.Caller, raw json.RawMessage) (string, error) {
	var a struct {
		Payee       string `json:"payee"`
		Category    string `json:"category"`
		Description string `json:"description"`
	}
	if err := hermes.DecodeArgs(raw, &a); err != nil {
		return "", err
	}
	norm := NormalizePayee(a.Payee)
	switch {
	case norm == "":
		return "", hermes.Errorf("payee is required.")
	case utf8.RuneCountInString(norm) < minPayeeLen:
		return "", hermes.Errorf("payee must be at least %d characters.", minPayeeLen)
	}
	category := themis.NormalizeName(a.Category)
	if category == "" {
		return "", hermes.Errorf("category is required.")
	}
	if err := m.checkCategory(ctx, c.UserID, category); err != nil {
		return "", err
	}
	rules, err := m.store.PayeeRules(ctx, c.UserID)
	if err != nil {
		return "", err
	}
	exists := false
	for _, r := range rules {
		if r.PayeeNorm == norm {
			exists = true
			break
		}
	}
	if !exists && len(rules) >= maxPayeeRules {
		return "", hermes.Errorf(msgRuleLimit, maxPayeeRules)
	}
	rule, created, err := m.store.SetPayeeRule(ctx, c.UserID, PayeeRule{
		Payee:       hermes.Untrusted(a.Payee, maxPayeeLen),
		PayeeNorm:   norm,
		Category:    category,
		Description: hermes.Untrusted(a.Description, maxPayeeDescription),
	})
	if err != nil {
		return "", err
	}
	status := "updated"
	if created {
		status = "created"
	}
	return hermes.Respond(map[string]any{
		"status":  status,
		"rule":    toRuleOut(rule),
		"message": msgRuleAppliesLater,
	})
}

func (m *Module) listPayeeRules(ctx context.Context, c hermes.Caller, raw json.RawMessage) (string, error) {
	if len(raw) > 0 {
		var a struct{}
		if err := hermes.DecodeArgs(raw, &a); err != nil {
			return "", err
		}
	}
	rules, err := m.store.PayeeRules(ctx, c.UserID)
	if err != nil {
		return "", err
	}
	sort.Slice(rules, func(i, j int) bool { return rules[i].PayeeNorm < rules[j].PayeeNorm })
	out := make([]ruleOut, len(rules))
	for i, r := range rules {
		out[i] = toRuleOut(r)
	}
	return hermes.Respond(map[string]any{"rules": out, "count": len(out)})
}

func (m *Module) deletePayeeRule(ctx context.Context, c hermes.Caller, raw json.RawMessage) (string, error) {
	var a struct {
		Payee string `json:"payee"`
	}
	if err := hermes.DecodeArgs(raw, &a); err != nil {
		return "", err
	}
	norm := NormalizePayee(a.Payee)
	if norm == "" {
		return "", hermes.Errorf("payee is required.")
	}
	deleted, err := m.store.DeletePayeeRule(ctx, c.UserID, norm)
	if err != nil {
		return "", err
	}
	status := "not_found"
	if deleted {
		status = "deleted"
	}
	return hermes.Respond(map[string]any{"status": status})
}
