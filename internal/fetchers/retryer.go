package fetchers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/wthrajat/github-readme-stats/internal/common"
)

var patKeyRe = regexp.MustCompile(`PAT_\d*$`)

var httpClient = &http.Client{Timeout: 30 * time.Second}

// PATCount returns the number of GitHub API tokens available in the
// environment (keys matching ^PAT_\d*$ semantics of the JS retryer).
func PATCount() int {
	count := 0
	for _, kv := range os.Environ() {
		key, _, _ := strings.Cut(kv, "=")
		if patKeyRe.MatchString(key) {
			count++
		}
	}
	return count
}

// FetcherFunc fetches one page of data given variables and a token,
// returning the decoded JSON body, the HTTP status code and any transport error.
type FetcherFunc func(vars map[string]any, token string) (body map[string]any, statusCode int, err error)

// Retryer executes fetcher until it succeeds or tokens are exhausted.
// It rotates through PAT_{n} tokens on rate limiting, bad credentials or
// suspended accounts, mirroring src/common/retryer.js.
func Retryer(fetcher FetcherFunc, vars map[string]any, retries ...int) (map[string]any, int, error) {
	attempt := 0
	if len(retries) > 0 {
		attempt = retries[0]
	}
	for {
		total := PATCount()
		if total == 0 {
			return nil, 0, common.NewCustomError("No GitHub API tokens found", common.ErrNoTokens)
		}
		if attempt > total {
			return nil, 0, common.NewCustomError("Downtime due to GitHub API rate limiting", common.ErrMaxRetry)
		}
		token := os.Getenv(fmt.Sprintf("PAT_%d", attempt+1))
		body, status, err := fetcher(vars, token)
		if err != nil {
			return nil, status, err
		}
		if firstGLErrorType(body) == "RATE_LIMITED" {
			log.Printf("PAT_%d Failed", attempt+1)
			attempt++
			continue
		}
		if msg := responseMessage(body); msg == "Bad credentials" ||
			msg == "Sorry. Your account was suspended." {
			log.Printf("PAT_%d Failed", attempt+1)
			attempt++
			continue
		}
		return body, status, nil
	}
}

// doGraphQL posts a GraphQL request to the GitHub API via
// common.GraphQLRequest and returns the decoded JSON body and HTTP status.
// auth is the full Authorization header value
// (e.g. "bearer <token>" or "token <token>").
func doGraphQL(query string, vars map[string]any, auth string) (map[string]any, int, error) {
	raw, status, err := common.GraphQLRequest(
		map[string]any{"query": query, "variables": vars},
		map[string]string{"Authorization": auth},
	)
	if err != nil {
		return nil, status, err
	}
	var body map[string]any
	if len(raw) > 0 {
		if err := json.Unmarshal(raw, &body); err != nil {
			return nil, status, err
		}
	}
	if body == nil {
		body = map[string]any{}
	}
	return body, status, nil
}

func postJSON(url string, payload any, headers map[string]string) (map[string]any, int, error) {
	raw, err := json.Marshal(payload)
	if err != nil {
		return nil, 0, err
	}
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(raw))
	if err != nil {
		return nil, 0, err
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	return doRequest(req)
}

func getJSON(url string, headers map[string]string) (map[string]any, int, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, 0, err
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	return doRequest(req)
}

func doRequest(req *http.Request) (map[string]any, int, error) {
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, err
	}
	var body map[string]any
	if len(raw) > 0 {
		if err := json.Unmarshal(raw, &body); err != nil {
			return nil, resp.StatusCode, err
		}
	}
	if body == nil {
		body = map[string]any{}
	}
	return body, resp.StatusCode, nil
}

func childValue(m map[string]any, keys ...string) any {
	var cur any = m
	for _, k := range keys {
		cm, ok := cur.(map[string]any)
		if !ok {
			return nil
		}
		cur, ok = cm[k]
		if !ok {
			return nil
		}
	}
	return cur
}

func asMap(v any) map[string]any {
	m, _ := v.(map[string]any)
	return m
}

func childMap(m map[string]any, keys ...string) map[string]any {
	return asMap(childValue(m, keys...))
}

func childSlice(m map[string]any, keys ...string) []any {
	s, _ := childValue(m, keys...).([]any)
	return s
}

func childString(m map[string]any, keys ...string) string {
	s, _ := childValue(m, keys...).(string)
	return s
}

func childFloat(m map[string]any, keys ...string) float64 {
	switch n := childValue(m, keys...).(type) {
	case float64:
		return n
	case float32:
		return float64(n)
	case int:
		return float64(n)
	case int64:
		return float64(n)
	case json.Number:
		f, _ := n.Float64()
		return f
	default:
		return 0
	}
}

func childInt(m map[string]any, keys ...string) int {
	return int(childFloat(m, keys...))
}

func childBool(m map[string]any, keys ...string) bool {
	b, _ := childValue(m, keys...).(bool)
	return b
}

func graphQLErrors(body map[string]any) []any {
	if body == nil {
		return nil
	}
	errs, _ := body["errors"].([]any)
	return errs
}

func firstGLError(body map[string]any) map[string]any {
	errs := graphQLErrors(body)
	if len(errs) == 0 {
		return nil
	}
	return asMap(errs[0])
}

func firstGLErrorType(body map[string]any) string {
	return childString(firstGLError(body), "type")
}

func firstGLErrorMessage(body map[string]any) string {
	return childString(firstGLError(body), "message")
}

func responseMessage(body map[string]any) string {
	return childString(body, "message")
}

// firstWrappedLine wraps text like wrapTextMultiline(text, width, 1)[0] in
// utils.js and returns the first line.
func firstWrappedLine(text string, width int) string {
	lines := common.WrapTextMultiline(text, width, 1)
	if len(lines) == 0 {
		return text
	}
	return lines[0]
}
