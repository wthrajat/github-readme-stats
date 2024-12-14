package fetchers

import "github.com/wthrajat/github-readme-stats/internal/rank"

// Rank is a user's rank level and percentile.
type Rank struct {
	Level      string
	Percentile float64
}

// StatsData holds aggregated GitHub stats for a user.
type StatsData struct {
	Name                     string
	TotalPRs                 int
	TotalPRsMerged           int
	MergedPRsPercentage      float64
	TotalReviews             int
	TotalCommits             int
	TotalContributions       int
	TotalIssues              int
	TotalStars               int
	TotalDiscussionsStarted  int
	TotalDiscussionsAnswered int
	ContributedTo            int
	Rank                     rank.Rank
}

// Lang holds aggregated stats for a single programming language.
type Lang struct {
	Name  string
	Color string
	Size  float64
	Count float64
}

// TopLangData maps language names to their aggregated stats.
type TopLangData map[string]*Lang

// PrimaryLanguage holds the primary language of a repository.
type PrimaryLanguage struct {
	Color string
	ID    string
	Name  string
}

// RepositoryData holds repository information.
type RepositoryData struct {
	Name            string
	NameWithOwner   string
	Description     string
	IsPrivate       bool
	IsArchived      bool
	IsTemplate      bool
	StarCount       int
	ForkCount       int
	PrimaryLanguage *PrimaryLanguage
}

// GistLanguage holds the language of a gist file.
type GistLanguage struct {
	Name string
}

// GistFile holds a single file of a gist.
type GistFile struct {
	Name     string
	Language *GistLanguage
	Size     float64
}

// GistData holds gist information.
type GistData struct {
	Name          string
	NameWithOwner string
	Description   string
	Language      string
	StarsCount    int
	ForksCount    int
}

// WakaTimeLang holds stats for a single WakaTime language.
type WakaTimeLang struct {
	Name    string
	Text    string
	Percent float64
	Hours   int
	Minutes int
}

// WakaTimeData holds the subset of a WakaTime stats response used by the card.
type WakaTimeData struct {
	Languages               []WakaTimeLang
	Range                   string
	IsCodingActivityVisible bool
	IsOtherUsageVisible     bool
}
