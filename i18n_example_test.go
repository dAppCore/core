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

func ExampleTranslator() {
	var tr Translator = &exampleTranslator{lang: "en"}
	Println(tr.Language())
	// Output: en
}

func ExampleLocaleProvider() {
	var provider LocaleProvider = exampleLocaleProvider{}
	Println(provider.Locales() == nil)
	// Output: true
}

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

func ExampleI18n_Locales() {
	i := &I18n{}
	r := i.Locales()
	Println(r.OK)
	Println(len(r.Value.([]*Embed)))
	// Output:
	// true
	// 0
}

func ExampleI18n_SetTranslator() {
	i := &I18n{}
	tr := &exampleTranslator{}
	i.SetLanguage("fr")
	i.SetTranslator(tr)

	Println(tr.Language())
	// Output: fr
}

func ExampleI18n_Translator() {
	i := &I18n{}
	i.SetTranslator(&exampleTranslator{lang: "en"})
	Println(i.Translator().OK)
	// Output: true
}

func ExampleI18n_Translate() {
	i := &I18n{}
	i.SetTranslator(&exampleTranslator{lang: "en"})
	r := i.Translate("hello %s", "codex")
	Println(r.Value)
	// Output: hello codex
}

func ExampleI18n_SetLanguage() {
	i := &I18n{}
	i.SetLanguage("fr")
	Println(i.Language())
	// Output: fr
}

func ExampleI18n_Language() {
	i := &I18n{}
	Println(i.Language())
	// Output: en
}

func ExampleI18n_AvailableLanguages() {
	i := &I18n{}
	i.SetTranslator(&exampleTranslator{})
	Println(i.AvailableLanguages())
	// Output: [en fr]
}
