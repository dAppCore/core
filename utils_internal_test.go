// SPDX-License-Identifier: EUPL-1.2

package core

func TestUtils_shortRand_Good(t *T) {
	token := shortRand()

	AssertLen(t, token, 6)
	AssertTrue(t, HexDecode(token).OK)
}
func TestUtils_shortRand_Bad(t *T) {
	token := shortRand()

	AssertNotEmpty(t, token)
	AssertFalse(t, Contains(token, "/"))
}
func TestUtils_shortRand_Ugly(t *T) {
	for i := 0; i < 5; i++ {
		token := shortRand()
		decoded := HexDecode(token)
		RequireTrue(t, decoded.OK)
		AssertLen(t, decoded.Value.([]byte), 3)
	}
}
