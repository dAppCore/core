// SPDX-License-Identifier: EUPL-1.2

package core

func TestEntitlement_defaultChecker_Good(t *T) {
	entitlement := defaultChecker("process.run", 1, Background())

	AssertTrue(t, entitlement.Allowed)
	AssertTrue(t, entitlement.Unlimited)
}
func TestEntitlement_defaultChecker_Bad(t *T) {
	entitlement := defaultChecker("process.delete", -1, Background())

	AssertTrue(t, entitlement.Allowed)
	AssertTrue(t, entitlement.Unlimited)
}
func TestEntitlement_defaultChecker_Ugly(t *T) {
	entitlement := defaultChecker("", 0, nil)

	AssertTrue(t, entitlement.Allowed)
	AssertTrue(t, entitlement.Unlimited)
}
