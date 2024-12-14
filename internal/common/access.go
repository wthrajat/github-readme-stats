package common

import (
	"os"
	"strings"
)

// GetAllowedUsernames returns the lowercased allowlist from ALLOWED_USERNAMES,
// defaulting to "wthrajat". Ported from src/common/access.js.
func GetAllowedUsernames() []string {
	raw := os.Getenv("ALLOWED_USERNAMES")
	if raw == "" {
		raw = "wthrajat"
	}
	var out []string
	for _, name := range strings.Split(raw, ",") {
		if n := LowercaseTrim(name); n != "" {
			out = append(out, n)
		}
	}
	return out
}

// IsUsernameAllowed reports whether u is on the allowlist.
func IsUsernameAllowed(u string) bool {
	if u == "" {
		return false
	}
	for _, allowed := range GetAllowedUsernames() {
		if strings.ToLower(u) == allowed {
			return true
		}
	}
	return false
}

// IsPrivateInstance reports whether access control is enforced,
// i.e. the GRS_TOKEN environment variable is set.
func IsPrivateInstance() bool {
	return os.Getenv("GRS_TOKEN") != ""
}

// IsTokenValid reports whether provided matches the GRS_TOKEN secret.
func IsTokenValid(provided string) bool {
	required := os.Getenv("GRS_TOKEN")
	return required != "" && provided == required
}

// IsAuthorized reports whether a request is authorized. Public instances
// (no GRS_TOKEN) allow everything. On private instances the username
// allowlist is enforced when requireUsername is set, and a valid token
// (query token or header token) is required.
func IsAuthorized(username, token, headerToken string, requireUsername bool) bool {
	if !IsPrivateInstance() {
		return true
	}
	if requireUsername && !IsUsernameAllowed(username) {
		return false
	}
	provided := token
	if provided == "" {
		provided = headerToken
	}
	return IsTokenValid(provided)
}

// RenderAccessDenied renders the private-instance error card.
func RenderAccessDenied(opts map[string]string) string {
	return RenderError("Access denied", "This is a private instance", opts)
}
