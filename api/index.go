// Command api is the Vercel Go runtime entrypoint. Vercel requires an
// api/*.go file exporting Handler; vercel.json rewrites all paths here.
package main

import (
	"net/http"

	"github.com/wthrajat/github-readme-stats/internal/handlers"
)

var mux = handlers.NewMux()

// Handler serves every request routed to the function by vercel.json.
func Handler(w http.ResponseWriter, r *http.Request) {
	mux.ServeHTTP(w, r)
}

// main is intentionally empty: Vercel's Go builder invokes Handler directly,
// but a main package still needs the symbol to compile with plain go build.
func main() {}
