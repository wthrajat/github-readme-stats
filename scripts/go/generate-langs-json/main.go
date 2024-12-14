// Command generate-langs-json ports scripts/generate-langs-json.js: it
// fetches the GitHub linguist languages.yml, extracts each language's
// color without any YAML dependency (minimal line-based parsing), and
// writes src/common/languageColors.json plus a copy at
// internal/common/languageColors.json. Run from the repository root:
//
//	go run ./scripts/go/generate-langs-json
package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

const linguistURL = "https://raw.githubusercontent.com/github/linguist/master/lib/linguist/languages.yml"

var hexColorPattern = regexp.MustCompile(`^#[0-9A-Fa-f]+$`)

type langColor struct {
	name  string
	color string
}

// parseLanguageColors scans languages.yml for top-level `Language:` keys
// and their `color:` entries. Languages without a color are skipped,
// mirroring the JS version (JSON.stringify drops undefined values).
func parseLanguageColors(yml string) []langColor {
	var out []langColor
	var current string
	var hasCurrent, hasColor bool

	flush := func() {
		hasCurrent = false
		hasColor = false
		current = ""
	}

	for _, line := range strings.Split(yml, "\n") {
		if line == "" {
			continue
		}
		if line[0] != ' ' && line[0] != '\t' {
			// Top-level key: `Name:` (values live on indented lines).
			trimmed := strings.TrimSpace(line)
			if !strings.HasSuffix(trimmed, ":") {
				flush()
				continue
			}
			name := strings.TrimSpace(strings.TrimSuffix(trimmed, ":"))
			name = strings.Trim(name, `"'`)
			if name == "" {
				flush()
				continue
			}
			current, hasCurrent, hasColor = name, true, false
			continue
		}
		if !hasCurrent || hasColor {
			continue
		}
		trimmed := strings.TrimSpace(line)
		rest, ok := strings.CutPrefix(trimmed, "color:")
		if !ok {
			continue
		}
		value := strings.TrimSpace(rest)
		// Strip inline comments and surrounding quotes.
		if i := strings.Index(value, " #"); i >= 0 {
			value = strings.TrimSpace(value[:i])
		}
		value = strings.Trim(value, `"'`)
		if !hexColorPattern.MatchString(value) {
			continue
		}
		out = append(out, langColor{name: current, color: value})
		hasColor = true
	}
	return out
}

// escapeJSONString escapes s the way JSON.stringify does for the
// characters that can appear here (quotes, backslashes, controls).
func escapeJSONString(s string) string {
	var b strings.Builder
	for _, r := range s {
		switch r {
		case '"':
			b.WriteString(`\"`)
		case '\\':
			b.WriteString(`\\`)
		case '\n':
			b.WriteString(`\n`)
		case '\r':
			b.WriteString(`\r`)
		case '\t':
			b.WriteString(`\t`)
		case '\b':
			b.WriteString(`\b`)
		case '\f':
			b.WriteString(`\f`)
		default:
			if r < 0x20 {
				fmt.Fprintf(&b, `\u%04x`, r)
			} else {
				b.WriteRune(r)
			}
		}
	}
	return b.String()
}

// marshalKeepOrder renders the pairs in YAML order with 4-space indent,
// matching JSON.stringify(colors, null, "    ") without trailing newline.
func marshalKeepOrder(pairs []langColor) string {
	var b strings.Builder
	b.WriteString("{\n")
	for i, p := range pairs {
		fmt.Fprintf(&b, `    "%s": "%s"`, escapeJSONString(p.name), escapeJSONString(p.color))
		if i+1 < len(pairs) {
			b.WriteString(",")
		}
		b.WriteString("\n")
	}
	b.WriteString("}")
	return b.String()
}

func writeFile(path, content string) {
	if dir := filepath.Dir(path); dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			log.Fatalf("create dir %s: %v", dir, err)
		}
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		log.Fatalf("write %s: %v", path, err)
	}
}

func main() {
	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Get(linguistURL)
	if err != nil {
		log.Fatalf("fetch %s: %v", linguistURL, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		log.Fatalf("fetch %s: status %s", linguistURL, resp.Status)
	}
	raw, err := io.ReadAll(io.LimitReader(resp.Body, 32<<20))
	if err != nil {
		log.Fatalf("read response: %v", err)
	}

	pairs := parseLanguageColors(string(raw))
	if len(pairs) == 0 {
		log.Fatal("no language colors parsed")
	}
	content := marshalKeepOrder(pairs)
	writeFile(filepath.Join("internal", "common", "languageColors.json"), content)
	log.Printf("wrote %d language colors", len(pairs))
}
