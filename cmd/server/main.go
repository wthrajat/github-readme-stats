// Command server runs the GitHub Readme Stats cards as a Go HTTP server,
// porting express.js (local dev) and the Vercel routing in vercel.json.
package main

import (
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/wthrajat/github-readme-stats/internal/handlers"
)

// loadDotEnv parses KEY=VALUE lines from path and sets variables that are
// not already present in the environment. It mirrors `import
// "dotenv/config"` in express.js without any external dependency. A
// missing file is silently ignored.
func loadDotEnv(path string) {
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		line = strings.TrimSpace(strings.TrimPrefix(line, "export "))
		key, value, ok := strings.Cut(line, "=")
		key = strings.TrimSpace(key)
		if !ok || key == "" {
			continue
		}
		value = strings.TrimSpace(value)
		if len(value) >= 2 {
			if (value[0] == '"' && value[len(value)-1] == '"') ||
				(value[0] == '\'' && value[len(value)-1] == '\'') {
				value = value[1 : len(value)-1]
			}
		}
		if _, exists := os.LookupEnv(key); !exists {
			_ = os.Setenv(key, value)
		}
	}
}

func main() {
	loadDotEnv(".env")
	mux := handlers.NewMux()

	port := os.Getenv("PORT")
	if port == "" {
		port = os.Getenv("port")
	}
	if port == "" {
		port = "9000"
	}

	log.Printf("github-readme-stats Go server listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}
