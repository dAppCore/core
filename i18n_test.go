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
		Option{Key: "path", Value: "tests/data"},
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

type rejectingTranslator struct {
	lang      string
	languages []string
}

func (r *rejectingTranslator) Translate(id string, args ...any) Result {
	return Result{Value: NewError(Concat("missing translation: ", id)), OK: false}
}
func (r *rejectingTranslator) SetLanguage(lang string) error {
	r.lang = lang
	return NewError(Concat("unsupported language: ", lang))
}
func (r *rejectingTranslator) Language() string { return r.lang }
func (r *rejectingTranslator) AvailableLanguages() []string {
	return r.languages
}

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

// --- AX-7 canonical triplets ---

func TestI18n_I18n_AddLocales_Good(t *T) {
	c := New()
	r := Mount(testFS, "tests/data")
	AssertTrue(t, r.OK)
	c.I18n().AddLocales(r.Value.(*Embed))

	locales := c.I18n().Locales()
	AssertTrue(t, locales.OK)
	AssertLen(t, locales.Value.([]*Embed), 1)
}

func TestI18n_I18n_AddLocales_Bad(t *T) {
	c := New()
	c.I18n().AddLocales()
	locales := c.I18n().Locales()
	AssertTrue(t, locales.OK)
	AssertEmpty(t, locales.Value.([]*Embed))
}

func TestI18n_I18n_AddLocales_Ugly(t *T) {
	c := New()
	c.I18n().AddLocales(nil)
	locales := c.I18n().Locales()
	AssertTrue(t, locales.OK)
	AssertLen(t, locales.Value.([]*Embed), 1)
	AssertNil(t, locales.Value.([]*Embed)[0])
}

func TestI18n_I18n_Locales_Good(t *T) {
	c := New()
	r := Mount(testFS, "tests/data")
	AssertTrue(t, r.OK)
	c.I18n().AddLocales(r.Value.(*Embed))

	locales := c.I18n().Locales()
	AssertTrue(t, locales.OK)
	AssertLen(t, locales.Value.([]*Embed), 1)
}

func TestI18n_I18n_Locales_Bad(t *T) {
	c := New()
	locales := c.I18n().Locales()
	AssertTrue(t, locales.OK)
	AssertEmpty(t, locales.Value.([]*Embed))
}

func TestI18n_I18n_Locales_Ugly(t *T) {
	c := New()
	r := Mount(testFS, "tests/data")
	AssertTrue(t, r.OK)
	c.I18n().AddLocales(r.Value.(*Embed))
	locales := c.I18n().Locales().Value.([]*Embed)
	locales[0] = nil
	AssertNotNil(t, c.I18n().Locales().Value.([]*Embed)[0])
}

func TestI18n_I18n_SetTranslator_Good(t *T) {
	c := New()
	tr := &mockTranslator{lang: "en"}
	c.I18n().SetTranslator(tr)
	AssertEqual(t, tr, c.I18n().Translator().Value)
}

func TestI18n_I18n_SetTranslator_Bad(t *T) {
	c := New()
	c.I18n().SetTranslator(nil)
	AssertFalse(t, c.I18n().Translator().OK)
}

func TestI18n_I18n_SetTranslator_Ugly(t *T) {
	c := New()
	c.I18n().SetLanguage("de")
	tr := &mockTranslator{lang: "en"}
	c.I18n().SetTranslator(tr)
	AssertEqual(t, "de", tr.Language())
}

func TestI18n_I18n_Translator_Good(t *T) {
	c := New()
	tr := &mockTranslator{lang: "en"}
	c.I18n().SetTranslator(tr)
	r := c.I18n().Translator()
	AssertTrue(t, r.OK)
	AssertEqual(t, tr, r.Value)
}

func TestI18n_I18n_Translator_Bad(t *T) {
	c := New()
	r := c.I18n().Translator()
	AssertFalse(t, r.OK)
}

func TestI18n_I18n_Translator_Ugly(t *T) {
	c := New()
	c.I18n().SetTranslator(&mockTranslator{lang: "en"})
	c.I18n().SetTranslator(nil)
	AssertFalse(t, c.I18n().Translator().OK)
}

func TestI18n_I18n_Translate_Good(t *T) {
	c := New()
	c.I18n().SetTranslator(&mockTranslator{})
	r := c.I18n().Translate("cmd.deploy.description")
	AssertTrue(t, r.OK)
	AssertEqual(t, "translated:cmd.deploy.description", r.Value)
}

func TestI18n_I18n_Translate_Bad(t *T) {
	c := New()
	c.I18n().SetTranslator(&rejectingTranslator{})
	r := c.I18n().Translate("cmd.missing.description")
	AssertFalse(t, r.OK)
	AssertContains(t, r.Error(), "missing translation")
}

func TestI18n_I18n_Translate_Ugly(t *T) {
	c := New()
	r := c.I18n().Translate("")
	AssertTrue(t, r.OK)
	AssertEqual(t, "", r.Value)
}

func TestI18n_I18n_SetLanguage_Good(t *T) {
	c := New()
	tr := &mockTranslator{}
	c.I18n().SetTranslator(tr)
	r := c.I18n().SetLanguage("cy")
	AssertTrue(t, r.OK)
	AssertEqual(t, "cy", tr.Language())
}

func TestI18n_I18n_SetLanguage_Bad(t *T) {
	c := New()
	c.I18n().SetTranslator(&rejectingTranslator{})
	r := c.I18n().SetLanguage("xx")
	AssertFalse(t, r.OK)
}

func TestI18n_I18n_SetLanguage_Ugly(t *T) {
	c := New()
	r := c.I18n().SetLanguage("")
	AssertTrue(t, r.OK)
	AssertEqual(t, "en", c.I18n().Language())
}

func TestI18n_I18n_Language_Good(t *T) {
	c := New()
	c.I18n().SetLanguage("de")
	AssertEqual(t, "de", c.I18n().Language())
}

func TestI18n_I18n_Language_Bad(t *T) {
	c := New()
	AssertEqual(t, "en", c.I18n().Language())
}

func TestI18n_I18n_Language_Ugly(t *T) {
	c := New()
	c.I18n().SetTranslator(&rejectingTranslator{})
	c.I18n().SetLanguage("xx")
	AssertEqual(t, "xx", c.I18n().Language())
}

func TestI18n_I18n_AvailableLanguages_Good(t *T) {
	c := New()
	c.I18n().SetTranslator(&mockTranslator{})
	AssertEqual(t, []string{"en", "de", "fr"}, c.I18n().AvailableLanguages())
}

func TestI18n_I18n_AvailableLanguages_Bad(t *T) {
	c := New()
	AssertEqual(t, []string{"en"}, c.I18n().AvailableLanguages())
}

func TestI18n_I18n_AvailableLanguages_Ugly(t *T) {
	c := New()
	c.I18n().SetTranslator(&rejectingTranslator{languages: []string{}})
	AssertEmpty(t, c.I18n().AvailableLanguages())
}
