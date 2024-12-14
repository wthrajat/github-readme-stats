package common

// Blacklist holds usernames excluded from rank calculations.
// Ported from src/common/blacklist.js.
var Blacklist = []string{"renovate-bot", "technote-space", "sw-yx"}

// IsBlacklisted reports whether u is on the blacklist.
func IsBlacklisted(u string) bool {
	for _, b := range Blacklist {
		if b == u {
			return true
		}
	}
	return false
}
