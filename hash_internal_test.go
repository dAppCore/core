// SPDX-License-Identifier: EUPL-1.2

package core

func TestHash_hashFor_Good(t *T) {
	factory := hashFor("sha256")

	AssertNotNil(t, factory)
	AssertEqual(t, 32, factory().Size())
}
func TestHash_hashFor_Bad(t *T) {
	AssertNil(t, hashFor("blake2"))
}
func TestHash_hashFor_Ugly(t *T) {
	factory := hashFor("sha512")

	AssertNotNil(t, factory)
	AssertEqual(t, 64, factory().Size())
}
