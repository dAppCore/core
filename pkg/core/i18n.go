// SPDX-License-Identifier: EUPL-1.2

// Internationalisation for the Core framework.
// I18n collects locale mounts from services and delegates
// translation to a registered Translator implementation (e.g., go-i18n).

package core

import (
	"sync"
)

// Translator defines the interface for translation services.
// Implemented by go-i18n's Srv.
type Translator interface {
	// T translates a message by its ID with optional arguments.
	T(messageID string, args ...any) string
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
type I18n struct {
	mu         sync.RWMutex
	locales    []*Embed     // collected from LocaleProvider services
	translator Translator // registered implementation (nil until set)
}


// AddLocales adds locale mounts (called during service registration).
func (i *I18n) AddLocales(mounts ...*Embed) {
	i.mu.Lock()
	i.locales = append(i.locales, mounts...)
	i.mu.Unlock()
}

// Locales returns all collected locale mounts.
func (i *I18n) Locales() []*Embed {
	i.mu.RLock()
	out := make([]*Embed, len(i.locales))
	copy(out, i.locales)
	i.mu.RUnlock()
	return out
}

// SetTranslator registers the translation implementation.
// Called by go-i18n's Srv during startup.
func (i *I18n) SetTranslator(t Translator) {
	i.mu.Lock()
	i.translator = t
	i.mu.Unlock()
}

// Translator returns the registered translation implementation, or nil.
func (i *I18n) Translator() Translator {
	i.mu.RLock()
	t := i.translator
	i.mu.RUnlock()
	return t
}

// T translates a message. Returns the key as-is if no translator is registered.
func (i *I18n) T(messageID string, args ...any) string {
	i.mu.RLock()
	t := i.translator
	i.mu.RUnlock()
	if t != nil {
		return t.T(messageID, args...)
	}
	return messageID
}

// SetLanguage sets the active language. No-op if no translator is registered.
func (i *I18n) SetLanguage(lang string) error {
	i.mu.RLock()
	t := i.translator
	i.mu.RUnlock()
	if t != nil {
		return t.SetLanguage(lang)
	}
	return nil
}

// Language returns the current language code, or "en" if no translator.
func (i *I18n) Language() string {
	i.mu.RLock()
	t := i.translator
	i.mu.RUnlock()
	if t != nil {
		return t.Language()
	}
	return "en"
}

// AvailableLanguages returns all loaded language codes.
func (i *I18n) AvailableLanguages() []string {
	i.mu.RLock()
	t := i.translator
	i.mu.RUnlock()
	if t != nil {
		return t.AvailableLanguages()
	}
	return []string{"en"}
}
