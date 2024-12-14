package cards

import (
	"strings"
	"testing"

	"github.com/wthrajat/github-readme-stats/internal/fetchers"
	"github.com/wthrajat/github-readme-stats/internal/rank"
)

func testStats() *fetchers.StatsData {
	return &fetchers.StatsData{
		Name: "Test", TotalPRs: 100, TotalCommits: 500, TotalIssues: 50,
		TotalStars: 200, ContributedTo: 30, TotalReviews: 10,
		TotalContributions: 600, Rank: rank.Rank{Level: "A+", Percentile: 10},
	}
}

func TestRenderStatsCard(t *testing.T) {
	svg := RenderStatsCard(testStats(), StatCardOptions{})
	for _, want := range []string{"<svg", "rank-circle", "stars", "commits"} {
		if !strings.Contains(svg, want) {
			t.Errorf("stats card missing %q", want)
		}
	}
}

func TestRenderRepoCard(t *testing.T) {
	repo := &fetchers.RepositoryData{
		Name: "myrepo", NameWithOwner: "user/myrepo", Description: "A test repo",
		StarCount: 42, ForkCount: 7,
		PrimaryLanguage: &fetchers.PrimaryLanguage{Name: "Go", Color: "#00ADD8"},
	}
	svg := RenderRepoCard(repo, RepoCardOptions{})
	if !strings.Contains(svg, "myrepo") || !strings.Contains(svg, "stargazers") {
		t.Error("repo card missing content")
	}
}

func TestRenderTopLanguages(t *testing.T) {
	data := fetchers.TopLangData{
		"Go":   {Name: "Go", Color: "#00ADD8", Size: 1000, Count: 3},
		"Rust": {Name: "Rust", Color: "#dea584", Size: 500, Count: 1},
	}
	for _, layout := range []string{"", "compact", "donut", "donut-vertical", "pie"} {
		svg := RenderTopLanguages(data, TopLangOptions{Layout: layout})
		if !strings.Contains(svg, "<svg") || !strings.Contains(svg, "Go") {
			t.Errorf("layout %q missing content", layout)
		}
	}
	empty := RenderTopLanguages(fetchers.TopLangData{}, TopLangOptions{})
	if !strings.Contains(empty, "No languages") {
		t.Error("empty langs should show nodata")
	}
}

func TestRenderGistCard(t *testing.T) {
	g := &fetchers.GistData{
		Name: "test.go", NameWithOwner: "user/test.go", Description: "desc",
		Language: "Go", StarsCount: 5, ForksCount: 1,
	}
	svg := RenderGistCard(g, GistCardOptions{})
	if !strings.Contains(svg, "test.go") {
		t.Error("gist card missing name")
	}
}

func TestRenderWakatimeCard(t *testing.T) {
	w := &fetchers.WakaTimeData{
		Languages: []fetchers.WakaTimeLang{
			{Name: "Go", Text: "5 hrs", Percent: 80, Hours: 5},
			{Name: "Python", Text: "1 hr", Percent: 20, Hours: 1},
		},
		Range:                   "last_7_days",
		IsCodingActivityVisible: true, IsOtherUsageVisible: true,
	}
	for _, layout := range []string{"", "compact"} {
		svg := RenderWakatimeCard(w, WakatimeOptions{Layout: layout})
		if !strings.Contains(svg, "WakaTime") || !strings.Contains(svg, "Go") {
			t.Errorf("wakatime %q missing content", layout)
		}
	}
}
