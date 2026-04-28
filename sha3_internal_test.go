// SPDX-License-Identifier: EUPL-1.2

package core

func TestSha3_keccakF1600_Good(t *T) {
	var state [25]uint64

	keccakF1600(&state)

	AssertEqual(t, uint64(0xf1258f7940e1dde7), state[0])
}
func TestSha3_keccakF1600_Bad(t *T) {
	AssertPanics(t, func() {
		keccakF1600(nil)
	})
}
func TestSha3_keccakF1600_Ugly(t *T) {
	left := [25]uint64{0: 1, 24: 1 << 63}
	right := left

	keccakF1600(&left)
	keccakF1600(&right)

	AssertEqual(t, left, right)
}
