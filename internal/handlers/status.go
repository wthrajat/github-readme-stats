package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/wthrajat/github-readme-stats/internal/common"
	"github.com/wthrajat/github-readme-stats/internal/fetchers"
)

// status.go ports api/status/up.js and api/status/pat-info.js on top of
// common.GraphQLRequest and fetchers.Retryer.

const uptimeQuery = `
        query {
          rateLimit {
              remaining
          }
        }
        `

const patInfoQuery = `
        query {
          rateLimit {
            remaining
            resetAt
          },
        }`

var patKeyPattern = regexp.MustCompile(`PAT_\d*$`)

// uptimeFetcher mirrors the uptimeFetcher in up.js/pat-info.js, issuing
// the rateLimit query with the given token via common.GraphQLRequest.
func uptimeFetcher(query string) fetchers.FetcherFunc {
	return func(vars map[string]any, token string) (map[string]any, int, error) {
		raw, status, err := common.GraphQLRequest(
			map[string]any{"query": query, "variables": vars},
			map[string]string{"Authorization": "bearer " + token},
		)
		if err != nil {
			return nil, status, err
		}
		var body map[string]any
		if err := json.Unmarshal(raw, &body); err != nil {
			return nil, status, fmt.Errorf("could not decode GitHub API response: %w", err)
		}
		return body, status, nil
	}
}

// patEnvKeys returns the names of all PAT_* environment variables,
// mirroring getAllPATs (/PAT_\d*$/ test) in pat-info.js.
func patEnvKeys() []string {
	var out []string
	for _, kv := range os.Environ() {
		if key, _, ok := strings.Cut(kv, "="); ok && patKeyPattern.MatchString(key) {
			out = append(out, key)
		}
	}
	return out
}

// firstErrType returns data.errors[0].type, or "" when absent.
func firstErrType(body map[string]any) string {
	errs, _ := body["errors"].([]any)
	if len(errs) == 0 {
		return ""
	}
	first, _ := errs[0].(map[string]any)
	if first == nil {
		return ""
	}
	t, _ := first["type"].(string)
	return t
}

// firstErrMessage returns data.errors[0].message, or "" when absent.
func firstErrMessage(body map[string]any) string {
	errs, _ := body["errors"].([]any)
	if len(errs) == 0 {
		return ""
	}
	first, _ := errs[0].(map[string]any)
	if first == nil {
		return ""
	}
	msg, _ := first["message"].(string)
	return msg
}

// childMap walks nested maps, returning nil when any level is missing.
func childMap(m map[string]any, keys ...string) map[string]any {
	cur := m
	for _, k := range keys {
		next, _ := cur[k].(map[string]any)
		if next == nil {
			return nil
		}
		cur = next
	}
	return cur
}

// rateLimitRemaining returns data.data.rateLimit.remaining.
func rateLimitRemaining(body map[string]any) (float64, bool) {
	remaining, ok := childMap(body, "data", "rateLimit")["remaining"].(float64)
	return remaining, ok
}

// rateLimitResetAt returns data.data.rateLimit.resetAt.
func rateLimitResetAt(body map[string]any) string {
	resetAt, _ := childMap(body, "data", "rateLimit")["resetAt"].(string)
	return resetAt
}

// shieldsBadge mirrors the shields.io response built by
// shieldsUptimeBadge in api/status/up.js.
type shieldsBadge struct {
	SchemaVersion int    `json:"schemaVersion"`
	Label         string `json:"label"`
	Message       string `json:"message"`
	Color         string `json:"color"`
	IsError       bool   `json:"isError"`
}

func newShieldsBadge(up bool) shieldsBadge {
	message, color := "down", "red"
	if up {
		message, color = "up", "brightgreen"
	}
	return shieldsBadge{
		SchemaVersion: 1,
		Label:         "Public Instance",
		Message:       message,
		Color:         color,
		IsError:       true,
	}
}

// HandleStatusUp ports api/status/up.js. Query `type` selects the output:
// "shields" (shields.io JSON), "json" ({"up": bool}), or anything else
// (bare true/false JSON literal).
func HandleStatusUp(w http.ResponseWriter, r *http.Request) {
	responseType := GetQuery(r, "type")
	if responseType == "" {
		responseType = "boolean"
	} else {
		responseType = strings.ToLower(responseType)
	}

	up := true
	if _, _, err := fetchers.Retryer(uptimeFetcher(uptimeQuery), map[string]any{}); err != nil {
		log.Printf("status/up check failed: %v", err)
		up = false
	}

	if up {
		CacheControl(w, fmt.Sprintf("max-age=0, s-maxage=%d", common.FiveMinutes))
	} else {
		CacheControl(w, "no-store")
	}

	switch responseType {
	case "shields":
		WriteJSON(w, newShieldsBadge(up), "")
	case "json":
		WriteJSON(w, map[string]bool{"up": up}, "")
	default:
		w.Header().Set("Content-Type", "application/json")
		if up {
			_, _ = w.Write([]byte("true"))
		} else {
			_, _ = w.Write([]byte("false"))
		}
	}
}

// patDetail is a single entry of the pat-info details map.
type patDetail struct {
	Status    string         `json:"status"`
	Remaining *float64       `json:"remaining,omitempty"`
	ResetIn   string         `json:"resetIn,omitempty"`
	Error     map[string]any `json:"error,omitempty"`
}

// patInfo mirrors the PATInfo object in api/status/pat-info.js.
type patInfo struct {
	ValidPATs     []string             `json:"validPATs"`
	ExpiredPATs   []string             `json:"expiredPATs"`
	ExhaustedPATs []string             `json:"exhaustedPATs"`
	SuspendedPATs []string             `json:"suspendedPATs"`
	ErrorPATs     []string             `json:"errorPATs"`
	Details       map[string]patDetail `json:"details"`
}

// getPATInfo mirrors getPATInfo in api/status/pat-info.js, querying every
// PAT_* token directly (no retryer, as in the JS) and classifying it as
// valid/expired/exhausted/suspended/error.
func getPATInfo() (*patInfo, error) {
	pats := patEnvKeys()
	details := map[string]patDetail{}

	for _, pat := range pats {
		raw, status, err := common.GraphQLRequest(
			map[string]any{"query": patInfoQuery, "variables": map[string]any{}},
			map[string]string{"Authorization": "bearer " + os.Getenv(pat)},
		)
		if err != nil {
			return nil, err
		}
		var body map[string]any
		if err := json.Unmarshal(raw, &body); err != nil {
			return nil, fmt.Errorf("could not decode GitHub API response: %w", err)
		}

		if status < 200 || status > 299 {
			msg, _ := body["message"].(string)
			switch strings.ToLower(msg) {
			case "bad credentials":
				details[pat] = patDetail{Status: "expired"}
			case "sorry. your account was suspended.":
				details[pat] = patDetail{Status: "suspended"}
			default:
				return nil, fmt.Errorf("GitHub API responded with status %d", status)
			}
			continue
		}

		errType := firstErrType(body)
		// Mirrors `Boolean(errors)`: any non-null errors value counts,
		// even an empty array.
		errVal, hasErrors := body["errors"]
		hasErrors = hasErrors && errVal != nil
		remaining, hasRemaining := rateLimitRemaining(body)
		rateLimited := (hasErrors && errType == "RATE_LIMITED") || (hasRemaining && remaining == 0)

		switch {
		case hasErrors && errType != "RATE_LIMITED":
			details[pat] = patDetail{
				Status: "error",
				Error:  map[string]any{"type": errType, "message": firstErrMessage(body)},
			}
		case rateLimited:
			resetIn := "unknown"
			if resetAt, parseErr := time.Parse(time.RFC3339, rateLimitResetAt(body)); parseErr == nil {
				resetIn = fmt.Sprintf("%d minutes", common.DateDiff(resetAt, time.Now()))
			}
			zero := 0.0
			details[pat] = patDetail{Status: "exhausted", Remaining: &zero, ResetIn: resetIn}
		default:
			remainingCopy := remaining
			details[pat] = patDetail{Status: "valid", Remaining: &remainingCopy}
		}
	}

	byStatus := func(status string) []string {
		out := []string{}
		for _, pat := range pats {
			if details[pat].Status == status {
				out = append(out, pat)
			}
		}
		return out
	}

	keys := make([]string, 0, len(details))
	for k := range details {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	sorted := make(map[string]patDetail, len(details))
	for _, k := range keys {
		sorted[k] = details[k]
	}

	return &patInfo{
		ValidPATs:     byStatus("valid"),
		ExpiredPATs:   byStatus("expired"),
		ExhaustedPATs: byStatus("exhausted"),
		SuspendedPATs: byStatus("suspended"),
		ErrorPATs:     byStatus("error"),
		Details:       sorted,
	}, nil
}

// HandlePatInfo ports api/status/pat-info.js.
func HandlePatInfo(w http.ResponseWriter, r *http.Request) {
	info, err := getPATInfo()
	if err != nil {
		log.Printf("status/pat-info failed: %v", err)
		CacheControl(w, "no-store")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte("Something went wrong: " + err.Error()))
		return
	}
	CacheControl(w, fmt.Sprintf("max-age=0, s-maxage=%d", common.FiveMinutes))
	w.Header().Set("Content-Type", "application/json")
	data, err := json.MarshalIndent(info, "", "  ")
	if err != nil {
		_, _ = w.Write([]byte("Something went wrong: " + err.Error()))
		return
	}
	_, _ = w.Write(data)
}
