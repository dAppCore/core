package core_test

import (
	"context"

	. "dappco.re/go"
)

// --- Entitled ---

func TestEntitlement_Entitled_Good_DefaultPermissive(t *T) {
	c := New()
	e := c.Entitled("anything")
	AssertTrue(t, e.Allowed, "default checker permits everything")
	AssertTrue(t, e.Unlimited)
}

func TestEntitlement_Entitled_Good_BooleanGate(t *T) {
	c := New()
	c.SetEntitlementChecker(func(action string, qty int, ctx context.Context) Entitlement {
		if action == "premium.feature" {
			return Entitlement{Allowed: true}
		}
		return Entitlement{Allowed: false, Reason: "not in package"}
	})

	AssertTrue(t, c.Entitled("premium.feature").Allowed)
	AssertFalse(t, c.Entitled("other.feature").Allowed)
	AssertEqual(t, "not in package", c.Entitled("other.feature").Reason)
}

func TestEntitlement_Entitled_Good_QuantityCheck(t *T) {
	c := New()
	c.SetEntitlementChecker(func(action string, qty int, ctx context.Context) Entitlement {
		if action == "social.accounts" {
			limit := 5
			used := 3
			remaining := limit - used
			if qty > remaining {
				return Entitlement{Allowed: false, Limit: limit, Used: used, Remaining: remaining, Reason: "limit exceeded"}
			}
			return Entitlement{Allowed: true, Limit: limit, Used: used, Remaining: remaining}
		}
		return Entitlement{Allowed: true, Unlimited: true}
	})

	// Can create 2 more (3 used of 5)
	e := c.Entitled("social.accounts", 2)
	AssertTrue(t, e.Allowed)
	AssertEqual(t, 5, e.Limit)
	AssertEqual(t, 3, e.Used)
	AssertEqual(t, 2, e.Remaining)

	// Can't create 3 more
	e = c.Entitled("social.accounts", 3)
	AssertFalse(t, e.Allowed)
	AssertEqual(t, "limit exceeded", e.Reason)
}

func TestEntitlement_Entitled_Bad_Denied(t *T) {
	c := New()
	c.SetEntitlementChecker(func(action string, qty int, ctx context.Context) Entitlement {
		return Entitlement{Allowed: false, Reason: "locked by M1"}
	})

	e := c.Entitled("product.create")
	AssertFalse(t, e.Allowed)
	AssertEqual(t, "locked by M1", e.Reason)
}

func TestEntitlement_Entitled_Ugly_DefaultQuantityIsOne(t *T) {
	c := New()
	var receivedQty int
	c.SetEntitlementChecker(func(action string, qty int, ctx context.Context) Entitlement {
		receivedQty = qty
		return Entitlement{Allowed: true}
	})

	c.Entitled("test")
	AssertEqual(t, 1, receivedQty, "default quantity should be 1")
}

// --- Action.Run Entitlement Enforcement ---

func TestEntitlement_ActionRun_Good_Permitted(t *T) {
	c := New()
	c.Action("work", func(_ context.Context, _ Options) Result {
		return Result{Value: "done", OK: true}
	})

	r := c.Action("work").Run(context.Background(), NewOptions())
	AssertTrue(t, r.OK)
	AssertEqual(t, "done", r.Value)
}

func TestEntitlement_ActionRun_Bad_Denied(t *T) {
	c := New()
	c.Action("restricted", func(_ context.Context, _ Options) Result {
		return Result{Value: "should not reach", OK: true}
	})
	c.SetEntitlementChecker(func(action string, qty int, ctx context.Context) Entitlement {
		if action == "restricted" {
			return Entitlement{Allowed: false, Reason: "tier too low"}
		}
		return Entitlement{Allowed: true, Unlimited: true}
	})

	r := c.Action("restricted").Run(context.Background(), NewOptions())
	AssertFalse(t, r.OK, "denied action must not execute")
	err, ok := r.Value.(error)
	AssertTrue(t, ok)
	AssertContains(t, err.Error(), "not entitled")
	AssertContains(t, err.Error(), "tier too low")
}

func TestEntitlement_ActionRun_Good_OtherActionsStillWork(t *T) {
	c := New()
	c.Action("allowed", func(_ context.Context, _ Options) Result {
		return Result{Value: "ok", OK: true}
	})
	c.Action("blocked", func(_ context.Context, _ Options) Result {
		return Result{Value: "nope", OK: true}
	})
	c.SetEntitlementChecker(func(action string, qty int, ctx context.Context) Entitlement {
		if action == "blocked" {
			return Entitlement{Allowed: false, Reason: "nope"}
		}
		return Entitlement{Allowed: true, Unlimited: true}
	})

	AssertTrue(t, c.Action("allowed").Run(context.Background(), NewOptions()).OK)
	AssertFalse(t, c.Action("blocked").Run(context.Background(), NewOptions()).OK)
}

// --- NearLimit ---

func TestEntitlement_NearLimit_Good(t *T) {
	e := Entitlement{Allowed: true, Limit: 100, Used: 85, Remaining: 15}
	AssertTrue(t, e.NearLimit(0.8))
	AssertFalse(t, e.NearLimit(0.9))
}

func TestEntitlement_NearLimit_Bad_Unlimited(t *T) {
	e := Entitlement{Allowed: true, Unlimited: true}
	AssertFalse(t, e.NearLimit(0.8), "unlimited should never be near limit")
}

func TestEntitlement_NearLimit_Ugly_ZeroLimit(t *T) {
	e := Entitlement{Allowed: true, Limit: 0}
	AssertFalse(t, e.NearLimit(0.8), "boolean gate (limit=0) should not report near limit")
}

// --- UsagePercent ---

func TestEntitlement_UsagePercent_Good(t *T) {
	e := Entitlement{Limit: 100, Used: 75}
	AssertEqual(t, 75.0, e.UsagePercent())
}

func TestEntitlement_UsagePercent_Ugly_ZeroLimit(t *T) {
	e := Entitlement{Limit: 0, Used: 5}
	AssertEqual(t, 0.0, e.UsagePercent(), "zero limit = boolean gate, no percentage")
}

// --- RecordUsage ---

func TestEntitlement_RecordUsage_Good(t *T) {
	c := New()
	var recorded string
	var recordedQty int

	c.SetUsageRecorder(func(action string, qty int, ctx context.Context) {
		recorded = action
		recordedQty = qty
	})

	c.RecordUsage("ai.credits", 10)
	AssertEqual(t, "ai.credits", recorded)
	AssertEqual(t, 10, recordedQty)
}

func TestEntitlement_RecordUsage_Good_NoRecorder(t *T) {
	c := New()
	// No recorder set — should not panic
	AssertNotPanics(t, func() {
		c.RecordUsage("anything", 5)
	})
}

// --- Permission Model Integration ---

func TestEntitlement_Ugly_SaaSGatingPattern(t *T) {
	c := New()

	// Simulate RFC-004 entitlement service
	packages := map[string]int{
		"social.accounts":        5,
		"social.posts.scheduled": 100,
		"ai.credits":             50,
	}
	usage := map[string]int{
		"social.accounts":        3,
		"social.posts.scheduled": 45,
		"ai.credits":             48,
	}

	c.SetEntitlementChecker(func(action string, qty int, ctx context.Context) Entitlement {
		limit, hasFeature := packages[action]
		if !hasFeature {
			return Entitlement{Allowed: false, Reason: "feature not in package"}
		}
		used := usage[action]
		remaining := limit - used
		if qty > remaining {
			return Entitlement{Allowed: false, Limit: limit, Used: used, Remaining: remaining, Reason: "limit exceeded"}
		}
		return Entitlement{Allowed: true, Limit: limit, Used: used, Remaining: remaining}
	})

	// Can create 2 social accounts
	e := c.Entitled("social.accounts", 2)
	AssertTrue(t, e.Allowed)

	// AI credits near limit
	e = c.Entitled("ai.credits", 1)
	AssertTrue(t, e.Allowed)
	AssertTrue(t, e.NearLimit(0.8))
	AssertEqual(t, 96.0, e.UsagePercent())

	// Feature not in package
	e = c.Entitled("premium.feature")
	AssertFalse(t, e.Allowed)
}

// --- AX-7 canonical triplets ---

func TestEntitlement_Entitlement_NearLimit_Good(t *T) {
	e := Entitlement{Allowed: true, Limit: 100, Used: 85}
	AssertTrue(t, e.NearLimit(0.8))
}

func TestEntitlement_Entitlement_NearLimit_Bad(t *T) {
	e := Entitlement{Allowed: true, Unlimited: true, Limit: 100, Used: 99}
	AssertFalse(t, e.NearLimit(0.8))
}

func TestEntitlement_Entitlement_NearLimit_Ugly(t *T) {
	e := Entitlement{Allowed: true, Limit: 0, Used: 10}
	AssertFalse(t, e.NearLimit(0.1))
}

func TestEntitlement_Entitlement_UsagePercent_Good(t *T) {
	e := Entitlement{Limit: 200, Used: 50}
	AssertEqual(t, 25.0, e.UsagePercent())
}

func TestEntitlement_Entitlement_UsagePercent_Bad(t *T) {
	e := Entitlement{Limit: 0, Used: 50}
	AssertEqual(t, 0.0, e.UsagePercent())
}

func TestEntitlement_Entitlement_UsagePercent_Ugly(t *T) {
	e := Entitlement{Limit: 10, Used: 15}
	AssertEqual(t, 150.0, e.UsagePercent())
}

func TestEntitlement_Core_Entitled_Good(t *T) {
	c := New()
	e := c.Entitled("agent.dispatch")
	AssertTrue(t, e.Allowed)
	AssertTrue(t, e.Unlimited)
}

func TestEntitlement_Core_Entitled_Bad(t *T) {
	c := New()
	c.SetEntitlementChecker(func(action string, quantity int, ctx context.Context) Entitlement {
		return Entitlement{Allowed: false, Reason: "agent quota exhausted"}
	})
	e := c.Entitled("agent.dispatch")
	AssertFalse(t, e.Allowed)
	AssertEqual(t, "agent quota exhausted", e.Reason)
}

func TestEntitlement_Core_Entitled_Ugly(t *T) {
	c := New()
	seenQuantity := 0
	c.SetEntitlementChecker(func(action string, quantity int, ctx context.Context) Entitlement {
		seenQuantity = quantity
		return Entitlement{Allowed: true}
	})
	c.Entitled("agent.dispatch")
	AssertEqual(t, 1, seenQuantity)
}

func TestEntitlement_Core_SetEntitlementChecker_Good(t *T) {
	c := New()
	c.SetEntitlementChecker(func(action string, quantity int, ctx context.Context) Entitlement {
		return Entitlement{Allowed: action == "agent.dispatch"}
	})
	AssertTrue(t, c.Entitled("agent.dispatch").Allowed)
}

func TestEntitlement_Core_SetEntitlementChecker_Bad(t *T) {
	c := New()
	c.SetEntitlementChecker(nil)
	AssertPanics(t, func() {
		c.Entitled("agent.dispatch")
	})
}

func TestEntitlement_Core_SetEntitlementChecker_Ugly(t *T) {
	c := New()
	c.SetEntitlementChecker(func(action string, quantity int, ctx context.Context) Entitlement {
		return Entitlement{Allowed: false, Reason: "first checker"}
	})
	c.SetEntitlementChecker(func(action string, quantity int, ctx context.Context) Entitlement {
		return Entitlement{Allowed: true, Reason: "replacement checker"}
	})
	e := c.Entitled("agent.dispatch")
	AssertTrue(t, e.Allowed)
	AssertEqual(t, "replacement checker", e.Reason)
}

func TestEntitlement_Core_RecordUsage_Good(t *T) {
	c := New()
	recorded := ""
	recordedQuantity := 0
	c.SetUsageRecorder(func(action string, quantity int, ctx context.Context) {
		recorded = action
		recordedQuantity = quantity
	})
	c.RecordUsage("agent.dispatch", 3)
	AssertEqual(t, "agent.dispatch", recorded)
	AssertEqual(t, 3, recordedQuantity)
}

func TestEntitlement_Core_RecordUsage_Bad(t *T) {
	c := New()
	AssertNotPanics(t, func() {
		c.RecordUsage("agent.dispatch", 3)
	})
}

func TestEntitlement_Core_RecordUsage_Ugly(t *T) {
	c := New()
	recordedQuantity := 0
	c.SetUsageRecorder(func(action string, quantity int, ctx context.Context) {
		recordedQuantity = quantity
	})
	c.RecordUsage("agent.dispatch")
	AssertEqual(t, 1, recordedQuantity)
}

func TestEntitlement_Core_SetUsageRecorder_Good(t *T) {
	c := New()
	recorded := false
	c.SetUsageRecorder(func(action string, quantity int, ctx context.Context) {
		recorded = true
	})
	c.RecordUsage("agent.dispatch")
	AssertTrue(t, recorded)
}

func TestEntitlement_Core_SetUsageRecorder_Bad(t *T) {
	c := New()
	c.SetUsageRecorder(nil)
	AssertNotPanics(t, func() {
		c.RecordUsage("agent.dispatch")
	})
}

func TestEntitlement_Core_SetUsageRecorder_Ugly(t *T) {
	c := New()
	recorded := "none"
	c.SetUsageRecorder(func(action string, quantity int, ctx context.Context) {
		recorded = "first"
	})
	c.SetUsageRecorder(func(action string, quantity int, ctx context.Context) {
		recorded = "second"
	})
	c.RecordUsage("agent.dispatch")
	AssertEqual(t, "second", recorded)
}
