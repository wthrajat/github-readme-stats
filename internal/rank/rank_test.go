package rank

import "testing"

func TestCalculateRankNewUser(t *testing.T) {
	r := CalculateRank(Params{})
	if r.Level != "C" {
		t.Fatalf("expected C for new user, got %s (%f)", r.Level, r.Percentile)
	}
	if r.Percentile < 99 || r.Percentile > 100 {
		t.Fatalf("expected ~100 percentile, got %f", r.Percentile)
	}
}

func TestCalculateRankActiveUser(t *testing.T) {
	r := CalculateRank(Params{
		Commits: 1000, PRs: 200, Issues: 100, Reviews: 50,
		Stars: 1000, Followers: 500,
	})
	if r.Percentile >= 20 {
		t.Fatalf("expected low percentile for active user, got %f", r.Percentile)
	}
	if r.Level != "S+" && r.Level != "A+" {
		t.Fatalf("expected top level, got %s", r.Level)
	}
}

func TestCalculateRankAllCommitsMedian(t *testing.T) {
	a := CalculateRank(Params{AllCommits: false, Commits: 500})
	b := CalculateRank(Params{AllCommits: true, Commits: 500})
	if a.Percentile == b.Percentile {
		t.Fatal("all_commits median should change percentile")
	}
	if a.Percentile > b.Percentile {
		t.Fatal("smaller median should give better (lower) percentile")
	}
}
