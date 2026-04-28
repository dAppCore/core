package core_test

import (
	"context"
	"testing"

	. "dappco.re/go/core"
)

// --- Entitled ---

func TestEntitlement_Entitled_Good_DefaultPermissive(t *testing.T) {
	c := New()
	e := c.Entitled("anything")
	AssertTrue(t, e.Allowed, "default checker permits everything")
	AssertTrue(t, e.Unlimited)
}

func TestEntitlement_Entitled_Good_BooleanGate(t *testing.T) {
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

func TestEntitlement_Entitled_Good_QuantityCheck(t *testing.T) {
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

func TestEntitlement_Entitled_Bad_Denied(t *testing.T) {
	c := New()
	c.SetEntitlementChecker(func(action string, qty int, ctx context.Context) Entitlement {
		return Entitlement{Allowed: false, Reason: "locked by M1"}
	})

	e := c.Entitled("product.create")
	AssertFalse(t, e.Allowed)
	AssertEqual(t, "locked by M1", e.Reason)
}

func TestEntitlement_Entitled_Ugly_DefaultQuantityIsOne(t *testing.T) {
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

func TestEntitlement_ActionRun_Good_Permitted(t *testing.T) {
	c := New()
	c.Action("work", func(_ context.Context, _ Options) Result {
		return Result{Value: "done", OK: true}
	})

	r := c.Action("work").Run(context.Background(), NewOptions())
	AssertTrue(t, r.OK)
	AssertEqual(t, "done", r.Value)
}

func TestEntitlement_ActionRun_Bad_Denied(t *testing.T) {
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

func TestEntitlement_ActionRun_Good_OtherActionsStillWork(t *testing.T) {
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

func TestEntitlement_NearLimit_Good(t *testing.T) {
	e := Entitlement{Allowed: true, Limit: 100, Used: 85, Remaining: 15}
	AssertTrue(t, e.NearLimit(0.8))
	AssertFalse(t, e.NearLimit(0.9))
}

func TestEntitlement_NearLimit_Bad_Unlimited(t *testing.T) {
	e := Entitlement{Allowed: true, Unlimited: true}
	AssertFalse(t, e.NearLimit(0.8), "unlimited should never be near limit")
}

func TestEntitlement_NearLimit_Ugly_ZeroLimit(t *testing.T) {
	e := Entitlement{Allowed: true, Limit: 0}
	AssertFalse(t, e.NearLimit(0.8), "boolean gate (limit=0) should not report near limit")
}

// --- UsagePercent ---

func TestEntitlement_UsagePercent_Good(t *testing.T) {
	e := Entitlement{Limit: 100, Used: 75}
	AssertEqual(t, 75.0, e.UsagePercent())
}

func TestEntitlement_UsagePercent_Ugly_ZeroLimit(t *testing.T) {
	e := Entitlement{Limit: 0, Used: 5}
	AssertEqual(t, 0.0, e.UsagePercent(), "zero limit = boolean gate, no percentage")
}

// --- RecordUsage ---

func TestEntitlement_RecordUsage_Good(t *testing.T) {
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

func TestEntitlement_RecordUsage_Good_NoRecorder(t *testing.T) {
	c := New()
	// No recorder set — should not panic
	AssertNotPanics(t, func() {
		c.RecordUsage("anything", 5)
	})
}

// --- Permission Model Integration ---

func TestEntitlement_Ugly_SaaSGatingPattern(t *testing.T) {
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
