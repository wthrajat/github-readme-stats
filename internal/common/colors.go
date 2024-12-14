package common

import (
	_ "embed"
	"encoding/json"
)

//go:embed languageColors.json
var languageColorsData []byte

// LanguageColors maps language names to their display colors.
// Loaded from languageColors.json (copied from src/common/languageColors.json).
var LanguageColors = map[string]string{}

func init() {
	_ = json.Unmarshal(languageColorsData, &LanguageColors)
}

// GetLanguageColor returns the display color for a language name,
// or "" when unknown.
func GetLanguageColor(name string) string {
	return LanguageColors[name]
}

// LanguageColor is an alias for GetLanguageColor.
func LanguageColor(name string) string {
	return GetLanguageColor(name)
}
