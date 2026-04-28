package core_test

import (
	"testing"

	. "dappco.re/go/core"
	"github.com/stretchr/testify/assert"
)

func TestCmp_Compare_Good(t *testing.T) {
	assert.Equal(t, -1, Compare(1, 2))
	assert.Equal(t, 1, Compare(2, 1))
}

func TestCmp_Compare_Bad(t *testing.T) {
	assert.Equal(t, 0, Compare(7, 7))
}

func TestCmp_Compare_Ugly(t *testing.T) {
	assert.Equal(t, -1, Compare("alpha", "beta"))
	assert.Equal(t, 1, Compare("beta", "alpha"))
}
