// Package helpers ports scripts/helpers.js: small helpers for automation
// scripts (stdlib only, no @actions/core dependency).
package helpers

import (
	"errors"
	"os"
	"strings"
)

// RepoInfo identifies the repository running the action, mirroring
// getRepoInfo in scripts/helpers.js (owner/repo fallback).
type RepoInfo struct {
	Owner string
	Repo  string
}

// GetRepoInfo returns the owner/repo from GITHUB_REPOSITORY ("owner/repo"),
// falling back to wthrajat/github-readme-stats when unset or malformed. The
// JS version reads the GitHub Actions context; the env variable carries
// the same value in Actions runners.
func GetRepoInfo() RepoInfo {
	if parts := strings.Split(os.Getenv("GITHUB_REPOSITORY"), "/"); len(parts) == 2 && parts[0] != "" && parts[1] != "" {
		return RepoInfo{Owner: parts[0], Repo: parts[1]}
	}
	return RepoInfo{Owner: "wthrajat", Repo: "github-readme-stats"}
}

// GetGithubToken returns the GitHub token, mirroring getGithubToken in
// scripts/helpers.js (getInput("github_token") reads INPUT_GITHUB_TOKEN).
func GetGithubToken() (string, error) {
	if token := os.Getenv("INPUT_GITHUB_TOKEN"); token != "" {
		return token, nil
	}
	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		return token, nil
	}
	return "", errors.New("Could not find github token")
}
