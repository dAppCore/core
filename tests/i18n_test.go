package core_test

import (
	"testing"

	. "forge.lthn.ai/core/go/pkg/core"
	"github.com/stretchr/testify/assert"
)

// --- I18n ---

func TestI18n_Good(t *testing.T) {
	c := New()
	assert.NotNil(t, c.I18n())
}

func TestI18n_AddLocales_Good(t *testing.T) {
	c := New()
	r := c.Data().New(Options{
		{K: "name", V: "lang"},
		{K: "source", V: testFS},
		{K: "path", V: "testdata"},
	})
	if r.OK {
		c.I18n().AddLocales(r.Value.(*Embed))
	}
	locales := c.I18n().Locales()
	assert.Len(t, locales, 1)
}

func TestI18n_Locales_Empty_Good(t *testing.T) {
	c := New()
	locales := c.I18n().Locales()
	assert.Empty(t, locales)
}

// --- Translator (no translator registered) ---

func TestI18n_T_NoTranslator_Good(t *testing.T) {
	c := New()
	// Without a translator, T returns the key as-is
	result := c.I18n().T("greeting.hello")
	assert.Equal(t, "greeting.hello", result)
}

func TestI18n_SetLanguage_NoTranslator_Good(t *testing.T) {
	c := New()
	r := c.I18n().SetLanguage("de")
	assert.True(t, r.OK) // no-op without translator
}

func TestI18n_Language_NoTranslator_Good(t *testing.T) {
	c := New()
	assert.Equal(t, "en", c.I18n().Language())
}

func TestI18n_AvailableLanguages_NoTranslator_Good(t *testing.T) {
	c := New()
	langs := c.I18n().AvailableLanguages()
	assert.Equal(t, []string{"en"}, langs)
}

func TestI18n_Translator_Nil_Good(t *testing.T) {
	c := New()
	assert.Nil(t, c.I18n().Translator())
}

// --- Translator (with mock) ---

type mockTranslator struct {
	lang string
}

func (m *mockTranslator) T(id string, args ...any) string     { return "translated:" + id }
func (m *mockTranslator) SetLanguage(lang string) error        { m.lang = lang; return nil }
func (m *mockTranslator) Language() string                     { return m.lang }
func (m *mockTranslator) AvailableLanguages() []string         { return []string{"en", "de", "fr"} }

func TestI18n_WithTranslator_Good(t *testing.T) {
	c := New()
	tr := &mockTranslator{lang: "en"}
	c.I18n().SetTranslator(tr)

	assert.Equal(t, tr, c.I18n().Translator())
	assert.Equal(t, "translated:hello", c.I18n().T("hello"))
	assert.Equal(t, "en", c.I18n().Language())
	assert.Equal(t, []string{"en", "de", "fr"}, c.I18n().AvailableLanguages())

	c.I18n().SetLanguage("de")
	assert.Equal(t, "de", c.I18n().Language())
}
