package cards

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/wthrajat/github-readme-stats/internal/common"
)

// fmtNum renders a float the way JavaScript template interpolation renders
// numbers: integral values without a decimal point, otherwise the shortest
// round-trip decimal representation.
func fmtNum(f float64) string {
	return strconv.FormatFloat(f, 'f', -1, 64)
}

// boolOpt dereferences an optional bool flag, defaulting to false when nil.
func boolOpt(p *bool) bool {
	return p != nil && *p
}

// stringVal coerces string-like fetcher fields (string or *string) to string
// so the cards compile regardless of whether the fetchers package models
// nullable strings as values or pointers.
func stringVal(v any) string {
	switch t := v.(type) {
	case nil:
		return ""
	case string:
		return t
	case *string:
		if t != nil {
			return *t
		}
		return ""
	case fmt.Stringer:
		return t.String()
	default:
		return fmt.Sprint(t)
	}
}

// cardBorderRadius resolves the optional border radius, defaulting to 4.5
// like the JS Card implementation.
func cardBorderRadius(p *float64) float64 {
	if p == nil {
		return 4.5
	}
	return *p
}

// formatLong formats a number with en-US thousands separators, replicating
// Number.prototype.toLocaleString("en-US") for the values used on cards.
func formatLong(n float64) string {
	neg := n < 0
	if neg {
		n = -n
	}
	intPart := int64(n)
	frac := n - float64(intPart)
	digits := strconv.FormatInt(intPart, 10)
	var b strings.Builder
	rem := len(digits) % 3
	if rem == 0 {
		rem = 3
	}
	b.WriteString(digits[:rem])
	for i := rem; i < len(digits); i += 3 {
		b.WriteByte(',')
		b.WriteString(digits[i : i+3])
	}
	if frac != 0 {
		fs := strconv.FormatFloat(frac, 'f', -1, 64)
		b.WriteString(strings.TrimPrefix(fs, "0"))
	}
	if neg {
		return "-" + b.String()
	}
	return b.String()
}

// round2 rounds to two decimal places, replicating +(n.toFixed(2)).
func round2(f float64) float64 {
	return math.Round(f*100) / 100
}

// languageColorFor returns the display color for a language name,
// falling back to gray for unknown languages like the JS implementation.
func languageColorFor(name string) string {
	if color := common.GetLanguageColor(name); color != "" {
		return color
	}
	return "#858585"
}
