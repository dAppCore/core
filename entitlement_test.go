package core_test

import (
	"context"
	"testing"

	. "dappco.re/go/core"
	"github.com/stretchr/testify/assert"
)

// --- Entitled ---

func TestEntitlement_Entitled_Good_DefaultPermissive(t *testing.T) {
	c := New()
	e := c.Entitled("anything")
	assert.True(t, e.Allowed, "default checker permits everything")
	assert.True(t, e.Unlimited)
}

func TestEntitlement_Entitled_Good_BooleanGate(t *testing.T) {
	c := New()
	c.SetEntitlementChecker(func(action string, qty int, ctx context.Context) Entitlement {
		if action == "premium.feature" {
			return Entitlement{Allowed: true}
		}
		return Entitlement{Allowed: false, Reason: "not in package"}
	})

	assert.True(t, c.Entitled("premium.feature").Allowed)
	assert.False(t, c.Entitled("other.feature").Allowed)
	assert.Equal(t, "not in package", c.Entitled("other.feature").Reason)
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
	assert.True(t, e.Allowed)
	assert.Equal(t, 5, e.Limit)
	assert.Equal(t, 3, e.Used)
	assert.Equal(t, 2, e.Remaining)

	// Can't create 3 more
	e = c.Entitled("social.accounts", 3)
	assert.False(t, e.Allowed)
	assert.Equal(t, "limit exceeded", e.Reason)
}

func TestEntitlement_Entitled_Bad_Denied(t *testing.T) {
	c := New()
	c.SetEntitlementChecker(func(action string, qty int, ctx context.Context) Entitlement {
		return Entitlement{Allowed: false, Reason: "locked by M1"}
	})

	e := c.Entitled("product.create")
	assert.False(t, e.Allowed)
	assert.Equal(t, "locked by M1", e.Reason)
}

func TestEntitlement_Entitled_Ugly_DefaultQuantityIsOne(t *testing.T) {
	c := New()
	var receivedQty int
	c.SetEntitlementChecker(func(action string, qty int, ctx context.Context) Entitlement {
		receivedQty = qty
		return Entitlement{Allowed: true}
	})

	c.Entitled("test")
	assert.Equal(t, 1, receivedQty, "default quantity should be 1")
}

// --- Action.Run Entitlement Enforcement ---

func TestEntitlement_ActionRun_Good_Permitted(t *testing.T) {
	c := New()
	c.Action("work", func(_ context.Context, _ Options) Result {
		return Result{Value: "done", OK: true}
	})

	r := c.Action("work").Run(context.Background(), NewOptions())
	assert.True(t, r.OK)
	assert.Equal(t, "done", r.Value)
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
	assert.False(t, r.OK, "denied action must not execute")
	err, ok := r.Value.(error)
	assert.True(t, ok)
	assert.Contains(t, err.Error(), "not entitled")
	assert.Contains(t, err.Error(), "tier too low")
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

	assert.True(t, c.Action("allowed").Run(context.Background(), NewOptions()).OK)
	assert.False(t, c.Action("blocked").Run(context.Background(), NewOptions()).OK)
}

// --- NearLimit ---

func TestEntitlement_NearLimit_Good(t *testing.T) {
	e := Entitlement{Allowed: true, Limit: 100, Used: 85, Remaining: 15}
	assert.True(t, e.NearLimit(0.8))
	assert.False(t, e.NearLimit(0.9))
}

func TestEntitlement_NearLimit_Bad_Unlimited(t *testing.T) {
	e := Entitlement{Allowed: true, Unlimited: true}
	assert.False(t, e.NearLimit(0.8), "unlimited should never be near limit")
}

func TestEntitlement_NearLimit_Ugly_ZeroLimit(t *testing.T) {
	e := Entitlement{Allowed: true, Limit: 0}
	assert.False(t, e.NearLimit(0.8), "boolean gate (limit=0) should not report near limit")
}

// --- UsagePercent ---

func TestEntitlement_UsagePercent_Good(t *testing.T) {
	e := Entitlement{Limit: 100, Used: 75}
	assert.Equal(t, 75.0, e.UsagePercent())
}

func TestEntitlement_UsagePercent_Ugly_ZeroLimit(t *testing.T) {
	e := Entitlement{Limit: 0, Used: 5}
	assert.Equal(t, 0.0, e.UsagePercent(), "zero limit = boolean gate, no percentage")
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
	assert.Equal(t, "ai.credits", recorded)
	assert.Equal(t, 10, recordedQty)
}

func TestEntitlement_RecordUsage_Good_NoRecorder(t *testing.T) {
	c := New()
	// No recorder set — should not panic
	assert.NotPanics(t, func() {
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
	assert.True(t, e.Allowed)

	// AI credits near limit
	e = c.Entitled("ai.credits", 1)
	assert.True(t, e.Allowed)
	assert.True(t, e.NearLimit(0.8))
	assert.Equal(t, 96.0, e.UsagePercent())

	// Feature not in package
	e = c.Entitled("premium.feature")
	assert.False(t, e.Allowed)
}
