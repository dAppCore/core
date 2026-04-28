package core_test

import (
	. "dappco.re/go"
)

// --- I18n ---

func TestI18n_Good(t *T) {
	c := New()
	AssertNotNil(t, c.I18n())
}

func TestI18n_AddLocales_Good(t *T) {
	c := New()
	r := c.Data().New(NewOptions(
		Option{Key: "name", Value: "lang"},
		Option{Key: "source", Value: testFS},
		Option{Key: "path", Value: "testdata"},
	))
	if r.OK {
		c.I18n().AddLocales(r.Value.(*Embed))
	}
	r2 := c.I18n().Locales()
	AssertTrue(t, r2.OK)
	AssertLen(t, r2.Value.([]*Embed), 1)
}

func TestI18n_Locales_Empty_Good(t *T) {
	c := New()
	r := c.I18n().Locales()
	AssertTrue(t, r.OK)
	AssertEmpty(t, r.Value.([]*Embed))
}

// --- Translator (no translator registered) ---

func TestI18n_Translate_NoTranslator_Good(t *T) {
	c := New()
	// Without a translator, Translate returns the key as-is
	r := c.I18n().Translate("greeting.hello")
	AssertTrue(t, r.OK)
	AssertEqual(t, "greeting.hello", r.Value)
}

func TestI18n_SetLanguage_NoTranslator_Good(t *T) {
	c := New()
	r := c.I18n().SetLanguage("de")
	AssertTrue(t, r.OK) // no-op without translator
}

func TestI18n_Language_NoTranslator_Good(t *T) {
	c := New()
	AssertEqual(t, "en", c.I18n().Language())
}

func TestI18n_AvailableLanguages_NoTranslator_Good(t *T) {
	c := New()
	langs := c.I18n().AvailableLanguages()
	AssertEqual(t, []string{"en"}, langs)
}

func TestI18n_Translator_Nil_Good(t *T) {
	c := New()
	AssertFalse(t, c.I18n().Translator().OK)
}

// --- Translator (with mock) ---

type mockTranslator struct {
	lang string
}

func (m *mockTranslator) Translate(id string, args ...any) Result {
	return Result{Concat("translated:", id), true}
}
func (m *mockTranslator) SetLanguage(lang string) error { m.lang = lang; return nil }
func (m *mockTranslator) Language() string              { return m.lang }
func (m *mockTranslator) AvailableLanguages() []string  { return []string{"en", "de", "fr"} }

func TestI18n_WithTranslator_Good(t *T) {
	c := New()
	tr := &mockTranslator{lang: "en"}
	c.I18n().SetTranslator(tr)

	AssertEqual(t, tr, c.I18n().Translator().Value)
	AssertEqual(t, "translated:hello", c.I18n().Translate("hello").Value)
	AssertEqual(t, "en", c.I18n().Language())
	AssertEqual(t, []string{"en", "de", "fr"}, c.I18n().AvailableLanguages())

	c.I18n().SetLanguage("de")
	AssertEqual(t, "de", c.I18n().Language())
}
