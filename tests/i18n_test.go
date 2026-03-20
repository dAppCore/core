package core_test

import (
	"testing"

	. "forge.lthn.ai/core/go/pkg/core"
	"github.com/stretchr/testify/assert"
)

func TestI18n_Good(t *testing.T) {
	c := New()
	assert.NotNil(t, c.I18n())
}

func TestI18n_AddLocales_Good(t *testing.T) {
	c := New()
	// AddLocales takes *Embed mounts — mount testdata and add it
	r := c.Data().New(Options{
		{K: "name", V: "lang"},
		{K: "source", V: testFS},
		{K: "path", V: "testdata"},
	})
	if r.OK {
		c.I18n().AddLocales(r.Value)
	}
}
