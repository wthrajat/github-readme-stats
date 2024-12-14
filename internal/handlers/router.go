package handlers

import "net/http"

// redirectHome mirrors the vercel.json redirect of "/" to the repository.
func redirectHome(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "https://github.com/wthrajat/github-readme-stats", http.StatusFound)
}

// NewMux returns the router with all card, legacy, status, and home
// routes registered. It is shared by the standalone server (cmd/server)
// and the Vercel Go function (api/index.go).
func NewMux() *http.ServeMux {
	mux := http.NewServeMux()
	// Primary API routes (Vercel: /api/*.js).
	mux.HandleFunc("GET /api", HandleStats)
	mux.HandleFunc("GET /api/pin", HandlePin)
	mux.HandleFunc("GET /api/top-langs", HandleTopLangs)
	mux.HandleFunc("GET /api/wakatime", HandleWakatime)
	mux.HandleFunc("GET /api/gist", HandleGist)
	mux.HandleFunc("GET /api/status/up", HandleStatusUp)
	mux.HandleFunc("GET /api/status/pat-info", HandlePatInfo)
	// Legacy express.js paths, kept for compatibility.
	mux.HandleFunc("GET /pin", HandlePin)
	mux.HandleFunc("GET /top-langs", HandleTopLangs)
	mux.HandleFunc("GET /wakatime", HandleWakatime)
	mux.HandleFunc("GET /gist", HandleGist)
	// vercel.json redirect.
	mux.HandleFunc("GET /{$}", redirectHome)
	return mux
}
