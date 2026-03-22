package core_test

import (
	"testing"

	. "dappco.re/go/core"
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
		{Key: "name", Value: "lang"},
		{Key: "source", Value: testFS},
		{Key: "path", Value: "testdata"},
	})
	if r.OK {
		c.I18n().AddLocales(r.Value.(*Embed))
	}
	r2 := c.I18n().Locales()
	assert.True(t, r2.OK)
	assert.Len(t, r2.Value.([]*Embed), 1)
}

func TestI18n_Locales_Empty_Good(t *testing.T) {
	c := New()
	r := c.I18n().Locales()
	assert.True(t, r.OK)
	assert.Empty(t, r.Value.([]*Embed))
}

// --- Translator (no translator registered) ---

func TestI18n_Translate_NoTranslator_Good(t *testing.T) {
	c := New()
	// Without a translator, Translate returns the key as-is
	r := c.I18n().Translate("greeting.hello")
	assert.True(t, r.OK)
	assert.Equal(t, "greeting.hello", r.Value)
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
	assert.False(t, c.I18n().Translator().OK)
}

// --- Translator (with mock) ---

type mockTranslator struct {
	lang string
}

func (m *mockTranslator) Translate(id string, args ...any) Result {
	return Result{"translated:" + id, true}
}
func (m *mockTranslator) SetLanguage(lang string) error { m.lang = lang; return nil }
func (m *mockTranslator) Language() string              { return m.lang }
func (m *mockTranslator) AvailableLanguages() []string  { return []string{"en", "de", "fr"} }

func TestI18n_WithTranslator_Good(t *testing.T) {
	c := New()
	tr := &mockTranslator{lang: "en"}
	c.I18n().SetTranslator(tr)

	assert.Equal(t, tr, c.I18n().Translator().Value)
	assert.Equal(t, "translated:hello", c.I18n().Translate("hello").Value)
	assert.Equal(t, "en", c.I18n().Language())
	assert.Equal(t, []string{"en", "de", "fr"}, c.I18n().AvailableLanguages())

	c.I18n().SetLanguage("de")
	assert.Equal(t, "de", c.I18n().Language())
}
