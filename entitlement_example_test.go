package core_test

import (
	"context"

	. "dappco.re/go"
)

func ExampleEntitlement_UsagePercent() {
	e := Entitlement{Limit: 100, Used: 75}
	Println(e.UsagePercent())
	// Output: 75
}

func ExampleEntitlement_NearLimit_threshold() {
	e := Entitlement{Limit: 100, Used: 90}
	Println(e.NearLimit(0.8))
	// Output: true
}

func ExampleEntitlementChecker() {
	var checker EntitlementChecker = func(action string, quantity int, _ context.Context) Entitlement {
		return Entitlement{Allowed: action == "deploy" && quantity <= 1}
	}
	Println(checker("deploy", 1, context.Background()).Allowed)
	// Output: true
}

func ExampleUsageRecorder() {
	var recorded string
	var recorder UsageRecorder = func(action string, quantity int, _ context.Context) {
		recorded = Sprintf("%s:%d", action, quantity)
	}
	recorder("ai.credits", 3, context.Background())
	Println(recorded)
	// Output: ai.credits:3
}

func ExampleCore_Entitled() {
	c := New()
	e := c.Entitled("deploy")
	Println(e.Allowed)
	Println(e.Unlimited)
	// Output:
	// true
	// true
}

func ExampleCore_SetEntitlementChecker() {
	c := New()
	c.SetEntitlementChecker(func(action string, qty int, _ context.Context) Entitlement {
		limits := map[string]int{"social.accounts": 5, "ai.credits": 100}
		usage := map[string]int{"social.accounts": 3, "ai.credits": 95}

		limit, ok := limits[action]
		if !ok {
			return Entitlement{Allowed: false, Reason: "not in package"}
		}
		used := usage[action]
		remaining := limit - used
		if qty > remaining {
			return Entitlement{Allowed: false, Limit: limit, Used: used, Remaining: remaining, Reason: "limit exceeded"}
		}
		return Entitlement{Allowed: true, Limit: limit, Used: used, Remaining: remaining}
	})

	Println(c.Entitled("social.accounts", 2).Allowed)
	Println(c.Entitled("social.accounts", 5).Allowed)
	Println(c.Entitled("ai.credits").NearLimit(0.9))
	// Output:
	// true
	// false
	// true
}

func ExampleCore_RecordUsage() {
	c := New()
	var recorded string
	c.SetUsageRecorder(func(action string, qty int, _ context.Context) {
		recorded = Concat(action, ":", Sprint(qty))
	})

	c.RecordUsage("ai.credits", 10)
	Println(recorded)
	// Output: ai.credits:10
}
