// SPDX-License-Identifier: EUPL-1.2

// Permission primitive for the Core framework.
// Entitlement answers "can [subject] do [action] with [quantity]?"
// Default: everything permitted (trusted conclave).
// With go-entitlements: checks workspace packages, features, usage, boosts.
// With commerce-matrix: checks entity hierarchy, lock cascade.
//
// Usage:
//
//	e := c.Entitled("process.run")           // boolean gate
//	e := c.Entitled("social.accounts", 3)    // quantity check
//	if e.Allowed { proceed() }
//	if e.NearLimit(0.8) { showUpgradePrompt() }
//
// Registration:
//
//	c.SetEntitlementChecker(myChecker)
//	c.SetUsageRecorder(myRecorder)
package core

import "context"

// Entitlement is the result of a permission check.
// Carries context for both boolean gates (Allowed) and usage limits (Limit/Used/Remaining).
//
//	e := c.Entitled("social.accounts", 3)
//	e.Allowed     // true
//	e.Limit       // 5
//	e.Used        // 2
//	e.Remaining   // 3
//	e.NearLimit(0.8) // false
type Entitlement struct {
	Allowed   bool   // permission granted
	Unlimited bool   // no cap (agency tier, admin, trusted conclave)
	Limit     int    // total allowed (0 = boolean gate)
	Used      int    // current consumption
	Remaining int    // Limit - Used
	Reason    string // denial reason — for UI and audit logging
}

// NearLimit returns true if usage exceeds the threshold percentage.
//
//	if e.NearLimit(0.8) { showUpgradePrompt() }
func (e Entitlement) NearLimit(threshold float64) bool {
	if e.Unlimited || e.Limit == 0 {
		return false
	}
	return float64(e.Used)/float64(e.Limit) >= threshold
}

// UsagePercent returns current usage as a percentage of the limit.
//
//	pct := e.UsagePercent() // 75.0
func (e Entitlement) UsagePercent() float64 {
	if e.Limit == 0 {
		return 0
	}
	return float64(e.Used) / float64(e.Limit) * 100
}

// EntitlementChecker answers "can [subject] do [action] with [quantity]?"
// Subject comes from context (workspace, entity, user — consumer's concern).
type EntitlementChecker func(action string, quantity int, ctx context.Context) Entitlement

// UsageRecorder records consumption after a gated action succeeds.
// Consumer packages provide the implementation (database, cache, etc).
type UsageRecorder func(action string, quantity int, ctx context.Context)

// defaultChecker — trusted conclave, everything permitted.
func defaultChecker(_ string, _ int, _ context.Context) Entitlement {
	return Entitlement{Allowed: true, Unlimited: true}
}

// Entitled checks if an action is permitted in the current context.
// Default: always returns Allowed=true, Unlimited=true.
// Denials are logged via core.Security().
//
//	e := c.Entitled("process.run")
//	e := c.Entitled("social.accounts", 3)
func (c *Core) Entitled(action string, quantity ...int) Entitlement {
	qty := 1
	if len(quantity) > 0 {
		qty = quantity[0]
	}

	e := c.entitlementChecker(action, qty, c.Context())

	if !e.Allowed {
		Security("entitlement.denied", "action", action, "quantity", qty, "reason", e.Reason)
	}

	return e
}

// SetEntitlementChecker replaces the default (permissive) checker.
// Called by go-entitlements or commerce-matrix during OnStartup.
//
//	func (s *EntitlementService) OnStartup(ctx context.Context) core.Result {
//	    s.Core().SetEntitlementChecker(s.check)
//	    return core.Result{OK: true}
//	}
func (c *Core) SetEntitlementChecker(checker EntitlementChecker) {
	c.entitlementChecker = checker
}

// RecordUsage records consumption after a gated action succeeds.
// Delegates to the registered UsageRecorder. No-op if none registered.
//
//	e := c.Entitled("ai.credits", 10)
//	if e.Allowed {
//	    doWork()
//	    c.RecordUsage("ai.credits", 10)
//	}
func (c *Core) RecordUsage(action string, quantity ...int) {
	if c.usageRecorder == nil {
		return
	}
	qty := 1
	if len(quantity) > 0 {
		qty = quantity[0]
	}
	c.usageRecorder(action, qty, c.Context())
}

// SetUsageRecorder registers a usage tracking function.
// Called by go-entitlements during OnStartup.
func (c *Core) SetUsageRecorder(recorder UsageRecorder) {
	c.usageRecorder = recorder
}
