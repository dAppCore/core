// SPDX-License-Identifier: EUPL-1.2

// Internationalisation for the Core framework.
// I18n collects locale mounts from services and delegates
// translation to a registered Translator implementation (e.g., go-i18n).

package core

import (
)

// Translator defines the interface for translation services.
// Implemented by go-i18n's Srv.
//
//	c := core.New()
//	if r := c.I18n().Translator(); r.OK {
//	    tr := r.Value.(core.Translator)
//	    _ = tr.SetLanguage("en-GB")
//	}
type Translator interface {
	// Translate translates a message by its ID with optional arguments.
	Translate(messageID string, args ...any) Result
	// SetLanguage sets the active language (BCP47 tag, e.g., "en-GB", "de").
	SetLanguage(lang string) error
	// Language returns the current language code.
	Language() string
	// AvailableLanguages returns all loaded language codes.
	AvailableLanguages() []string
}

// LocaleProvider is implemented by services that ship their own translation files.
// Core discovers this interface during service registration and collects the
// locale mounts. The i18n service loads them during startup.
//
// Usage in a service package:
//
//	//go:embed locales
//	var localeFS embed.FS
//
//	func (s *MyService) Locales() *Embed {
//	    m, _ := Mount(localeFS, "locales")
//	    return m
//	}
type LocaleProvider interface {
	Locales() *Embed
}

// I18n manages locale collection and translation dispatch.
//
//	c := core.New()
//	r := c.I18n().Translate("cmd.deploy.description")
//	if r.OK { core.Println(r.Value.(string)) }
type I18n struct {
	mu         RWMutex
	locales    []*Embed // collected from LocaleProvider services
	locale     string
	translator Translator // registered implementation (nil until set)
}

// AddLocales adds locale mounts (called during service registration).
//
//	c := core.New()
//	r := core.Mount(core.DirFS("locales"), ".")
//	if r.OK { c.I18n().AddLocales(r.Value.(*core.Embed)) }
func (i *I18n) AddLocales(mounts ...*Embed) {
	i.mu.Lock()
	i.locales = append(i.locales, mounts...)
	i.mu.Unlock()
}

// Locales returns all collected locale mounts.
//
//	c := core.New()
//	r := c.I18n().Locales()
//	if r.OK { mounts := r.Value.([]*core.Embed); _ = mounts }
func (i *I18n) Locales() Result {
	i.mu.RLock()
	out := make([]*Embed, len(i.locales))
	copy(out, i.locales)
	i.mu.RUnlock()
	return Result{out, true}
}

// SetTranslator registers the translation implementation.
// Called by go-i18n's Srv during startup.
//
//	c := core.New()
//	c.I18n().SetTranslator(nil)
func (i *I18n) SetTranslator(t Translator) {
	i.mu.Lock()
	i.translator = t
	locale := i.locale
	i.mu.Unlock()
	if t != nil && locale != "" {
		_ = t.SetLanguage(locale)
	}
}

// Translator returns the registered translation implementation, or nil.
//
//	c := core.New()
//	r := c.I18n().Translator()
//	if r.OK { tr := r.Value.(core.Translator); core.Println(tr.Language()) }
func (i *I18n) Translator() Result {
	i.mu.RLock()
	t := i.translator
	i.mu.RUnlock()
	if t == nil {
		return Result{}
	}
	return Result{t, true}
}

// Translate translates a message. Returns the key as-is if no translator is registered.
//
//	c := core.New()
//	r := c.I18n().Translate("cmd.deploy.description")
//	if r.OK { core.Println(r.Value.(string)) }
func (i *I18n) Translate(messageID string, args ...any) Result {
	i.mu.RLock()
	t := i.translator
	i.mu.RUnlock()
	if t != nil {
		return t.Translate(messageID, args...)
	}
	return Result{messageID, true}
}

// SetLanguage sets the active language and forwards to the translator if registered.
//
//	c := core.New()
//	r := c.I18n().SetLanguage("en-GB")
//	if !r.OK { return r }
func (i *I18n) SetLanguage(lang string) Result {
	if lang == "" {
		return Result{OK: true}
	}
	i.mu.Lock()
	i.locale = lang
	t := i.translator
	i.mu.Unlock()
	if t != nil {
		if err := t.SetLanguage(lang); err != nil {
			return Result{err, false}
		}
	}
	return Result{OK: true}
}

// Language returns the current language code, or "en" if not set.
//
//	c := core.New()
//	c.I18n().SetLanguage("en-GB")
//	lang := c.I18n().Language()
//	core.Println(lang)
func (i *I18n) Language() string {
	i.mu.RLock()
	locale := i.locale
	i.mu.RUnlock()
	if locale != "" {
		return locale
	}
	return "en"
}

// AvailableLanguages returns all loaded language codes.
//
//	c := core.New()
//	languages := c.I18n().AvailableLanguages()
//	core.Println(core.Join(", ", languages...))
func (i *I18n) AvailableLanguages() []string {
	i.mu.RLock()
	t := i.translator
	i.mu.RUnlock()
	if t != nil {
		return t.AvailableLanguages()
	}
	return []string{"en"}
}
