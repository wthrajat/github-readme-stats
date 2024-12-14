package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strconv"

	"github.com/wthrajat/github-readme-stats/internal/common"
)

// GetQuery returns the first value of the given URL query parameter,
// or "" when absent. It mirrors `req.query[key]` in the JS handlers.
func GetQuery(r *http.Request, key string) string {
	return r.URL.Query().Get(key)
}

// queryHas reports whether the given query parameter was present in the
// URL at all (even with an empty value). This mirrors the JS
// `layout !== undefined` presence checks.
func queryHas(r *http.Request, key string) bool {
	_, ok := r.URL.Query()[key]
	return ok
}

// ParseBoolPtr parses "true"/"false" (case-insensitive) into a *bool and
// returns nil for anything else, reusing common.ParseBoolean.
func ParseBoolPtr(s string) *bool {
	return common.ParseBoolean(s)
}

// atoiOr parses s as an int, returning def when s is empty or invalid.
func atoiOr(s string, def int) int {
	if s == "" {
		return def
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return def
	}
	return n
}

// intPtr parses s as an int, returning nil when s is empty or invalid.
func intPtr(s string) *int {
	if s == "" {
		return nil
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return nil
	}
	return &n
}

// floatPtr parses s as a float, returning nil when s is empty or invalid.
func floatPtr(s string) *float64 {
	if s == "" {
		return nil
	}
	n, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return nil
	}
	return &n
}

// parseFloatOr parses s as a float, returning def when s is empty or invalid.
func parseFloatOr(s string, def float64) float64 {
	if s == "" {
		return def
	}
	n, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return def
	}
	return n
}

// CacheControl sets the Cache-Control header when v is non-empty.
func CacheControl(w http.ResponseWriter, v string) {
	if v != "" {
		w.Header().Set("Cache-Control", v)
	}
}

// WriteSVG writes svg with Content-Type image/svg+xml and the given
// Cache-Control value (skipped when empty).
func WriteSVG(w http.ResponseWriter, svg, cache string) {
	w.Header().Set("Content-Type", "image/svg+xml")
	CacheControl(w, cache)
	_, _ = w.Write([]byte(svg))
}

// WriteJSON writes v as JSON with Content-Type application/json and the
// given Cache-Control value (skipped when empty). No trailing newline is
// appended, matching express res.send(obj).
func WriteJSON(w http.ResponseWriter, v any, cache string) {
	w.Header().Set("Content-Type", "application/json")
	CacheControl(w, cache)
	_ = jsonEncode(w, v)
}

// BoolOr dereferences b, returning def when b is nil.
func BoolOr(b *bool, def bool) bool {
	if b == nil {
		return def
	}
	return *b
}

// FloatOr dereferences f, returning def when f is nil.
func FloatOr(f *float64, def float64) float64 {
	if f == nil {
		return def
	}
	return *f
}

// clampInt clamps n into [min, max].
func clampInt(n, min, max int) int {
	if n < min {
		return min
	}
	if n > max {
		return max
	}
	return n
}

// EffectiveCacheSeconds resolves the cache duration for a handler. It
// mirrors the JS pattern:
//
//	cacheSeconds = clampValue(parseInt(query || def), min, max)
//	cacheSeconds = process.env.CACHE_SECONDS
//	  ? parseInt(process.env.CACHE_SECONDS) || cacheSeconds
//	  : cacheSeconds
func EffectiveCacheSeconds(queryVal string, def, min, max int) int {
	resolved := clampInt(atoiOr(queryVal, def), min, max)
	if env, ok := os.LookupEnv("CACHE_SECONDS"); ok && env != "" {
		if n, err := strconv.Atoi(env); err == nil && n != 0 {
			return n
		}
		return resolved
	}
	return resolved
}

// AccessDeniedOptions extracts the card color options used for the
// access-denied / error cards, mirroring the JS handlers.
func AccessDeniedOptions(r *http.Request) map[string]string {
	return map[string]string{
		"title_color":  GetQuery(r, "title_color"),
		"text_color":   GetQuery(r, "text_color"),
		"bg_color":     GetQuery(r, "bg_color"),
		"border_color": GetQuery(r, "border_color"),
		"theme":        GetQuery(r, "theme"),
	}
}

// jsonEncode marshals v without HTML escaping or a trailing newline,
// matching express res.send(obj).
func jsonEncode(w io.Writer, v any) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	_, err = w.Write(data)
	return err
}

// halfHeader formats n/2 the way JS `${n / 2}` stringifies it
// (e.g. 518400 -> "259200", 5 -> "2.5").
func halfHeader(n int) string {
	return strconv.FormatFloat(float64(n)/2, 'f', -1, 64)
}

// isAuthorizedRequest adapts common.IsAuthorized to an HTTP request,
// passing the query username, query token and x-grs-token header through.
func isAuthorizedRequest(r *http.Request, requireUsername bool) bool {
	return common.IsAuthorized(
		GetQuery(r, "username"),
		GetQuery(r, "token"),
		r.Header.Get("X-Grs-Token"),
		requireUsername,
	)
}

// errSecondary extracts the secondary message for error cards,
// mirroring err.secondaryMessage in the JS handlers.
func errSecondary(err error) string {
	switch e := err.(type) {
	case *common.CustomError:
		return e.SecondaryMessage
	case *common.MissingParamError:
		return e.Secondary
	default:
		return ""
	}
}
