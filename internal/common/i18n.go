package common

import "fmt"

// FallbackLocale is used when NewI18n is given an empty locale.
const FallbackLocale = "en"

// I18n is a translation helper.
// Ported from src/common/I18n.js.
type I18n struct {
	Locale       string
	Translations map[string]map[string]string
}

// NewI18n creates an I18n, defaulting an empty locale to "en".
func NewI18n(locale string, translations map[string]map[string]string) *I18n {
	if locale == "" {
		locale = FallbackLocale
	}
	return &I18n{Locale: locale, Translations: translations}
}

// T returns the translation for key in the configured locale,
// falling back to "en". It panics with a descriptive error when the
// key or locale is missing, mirroring the JS implementation.
func (i *I18n) T(key string) string {
	strs, ok := i.Translations[key]
	if !ok {
		panic(fmt.Sprintf("%s Translation string not found", key))
	}
	if s, ok := strs[i.Locale]; ok {
		return s
	}
	if s, ok := strs[FallbackLocale]; ok {
		return s
	}
	panic(fmt.Sprintf("'%s' translation not found for locale '%s'", key, i.Locale))
}
