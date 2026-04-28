package core_test

import . "dappco.re/go"

type exampleTranslator struct {
	lang string
}

func (t *exampleTranslator) Translate(messageID string, args ...any) Result {
	if len(args) == 0 {
		return Result{Value: messageID, OK: true}
	}
	return Result{Value: Sprintf(messageID, args...), OK: true}
}

func (t *exampleTranslator) SetLanguage(lang string) error {
	t.lang = lang
	return nil
}

func (t *exampleTranslator) Language() string {
	return t.lang
}

func (t *exampleTranslator) AvailableLanguages() []string {
	return []string{"en", "fr"}
}

type exampleLocaleProvider struct {
	mount *Embed
}

func (p exampleLocaleProvider) Locales() *Embed {
	return p.mount
}

// ExampleTranslator declares a translator through `Translator` for operator-facing
// localisation. Language selection and translation stay behind the I18n service.
func ExampleTranslator() {
	var tr Translator = &exampleTranslator{lang: "en"}
	Println(tr.Language())
	// Output: en
}

// ExampleLocaleProvider declares a locale provider through `LocaleProvider` for
// operator-facing localisation. Language selection and translation stay behind the I18n
// service.
func ExampleLocaleProvider() {
	var provider LocaleProvider = exampleLocaleProvider{}
	Println(provider.Locales() == nil)
	// Output: true
}

// ExampleI18n_AddLocales adds locale tables through `I18n.AddLocales` for operator-facing
// localisation. Language selection and translation stay behind the I18n service.
func ExampleI18n_AddLocales() {
	fs := (&Fs{}).New("/")
	dir := fs.TempDir("core-i18n-example")
	defer fs.DeleteAll(dir)
	fs.Write(Path(dir, "locales", "en.yaml"), "hello: Hello")

	mount := Mount(DirFS(dir), "locales").Value.(*Embed)
	i := &I18n{}
	i.AddLocales(mount)

	r := i.Locales()
	Println(len(r.Value.([]*Embed)))
	// Output: 1
}

// ExampleI18n_Locales lists locale tables through `I18n.Locales` for operator-facing
// localisation. Language selection and translation stay behind the I18n service.
func ExampleI18n_Locales() {
	i := &I18n{}
	r := i.Locales()
	Println(r.OK)
	Println(len(r.Value.([]*Embed)))
	// Output:
	// true
	// 0
}

// ExampleI18n_SetTranslator installs a translator through `I18n.SetTranslator` for
// operator-facing localisation. Language selection and translation stay behind the I18n
// service.
func ExampleI18n_SetTranslator() {
	i := &I18n{}
	tr := &exampleTranslator{}
	i.SetLanguage("fr")
	i.SetTranslator(tr)

	Println(tr.Language())
	// Output: fr
}

// ExampleI18n_Translator declares a translator through `I18n.Translator` for
// operator-facing localisation. Language selection and translation stay behind the I18n
// service.
func ExampleI18n_Translator() {
	i := &I18n{}
	i.SetTranslator(&exampleTranslator{lang: "en"})
	Println(i.Translator().OK)
	// Output: true
}

// ExampleI18n_Translate translates a key through `I18n.Translate` for operator-facing
// localisation. Language selection and translation stay behind the I18n service.
func ExampleI18n_Translate() {
	i := &I18n{}
	i.SetTranslator(&exampleTranslator{lang: "en"})
	r := i.Translate("hello %s", "codex")
	Println(r.Value)
	// Output: hello codex
}

// ExampleI18n_SetLanguage sets the active language through `I18n.SetLanguage` for
// operator-facing localisation. Language selection and translation stay behind the I18n
// service.
func ExampleI18n_SetLanguage() {
	i := &I18n{}
	i.SetLanguage("fr")
	Println(i.Language())
	// Output: fr
}

// ExampleI18n_Language reads the active language through `I18n.Language` for
// operator-facing localisation. Language selection and translation stay behind the I18n
// service.
func ExampleI18n_Language() {
	i := &I18n{}
	Println(i.Language())
	// Output: en
}

// ExampleI18n_AvailableLanguages lists available languages through
// `I18n.AvailableLanguages` for operator-facing localisation. Language selection and
// translation stay behind the I18n service.
func ExampleI18n_AvailableLanguages() {
	i := &I18n{}
	i.SetTranslator(&exampleTranslator{})
	Println(i.AvailableLanguages())
	// Output: [en fr]
}
