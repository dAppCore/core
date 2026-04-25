package core_test

import (
	"bufio"
	"strings"
	"testing"

	. "dappco.re/go/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Scanner ---

func TestScanner_NewLineScanner_Good(t *testing.T) {
	longLine := strings.Repeat("x", bufio.MaxScanTokenSize+1)
	scanner := NewLineScanner(strings.NewReader(longLine + "\nnext"))

	require.True(t, scanner.Scan())
	assert.Equal(t, longLine, scanner.Text())

	require.True(t, scanner.Scan())
	assert.Equal(t, "next", scanner.Text())

	assert.False(t, scanner.Scan())
	assert.NoError(t, scanner.Err())
}

func TestScanner_NewLineScanner_Bad(t *testing.T) {
	scanner := NewLineScannerWithSize(strings.NewReader("too long\n"), 4)

	assert.False(t, scanner.Scan())
	assert.Error(t, scanner.Err())
}

func TestScanner_NewLineScanner_Ugly(t *testing.T) {
	scanner := NewLineScanner(strings.NewReader("first\r\nsecond\n\nthird"))

	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	assert.Equal(t, []string{"first", "second", "", "third"}, lines)
	assert.NoError(t, scanner.Err())
}
