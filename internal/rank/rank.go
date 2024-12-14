package rank

import "math"

// Rank is the result of CalculateRank.
type Rank struct {
	Level      string
	Percentile float64
}

// Params holds the user statistics on which the rank depends.
type Params struct {
	AllCommits bool
	Commits    float64
	PRs        float64
	Issues     float64
	Reviews    float64
	Repos      float64
	Stars      float64
	Followers  float64
}

const (
	commitsMedianDefault = 250
	commitsMedianAll     = 1000
	commitsWeight        = 2
	prsMedian            = 50
	prsWeight            = 3
	issuesMedian         = 25
	issuesWeight         = 1
	reviewsMedian        = 2
	reviewsWeight        = 1
	starsMedian          = 50
	starsWeight          = 4
	followersMedian      = 10
	followersWeight      = 1
)

const totalWeight = commitsWeight +
	prsWeight +
	issuesWeight +
	reviewsWeight +
	starsWeight +
	followersWeight

var rankThresholds = [9]float64{1, 12.5, 25, 37.5, 50, 62.5, 75, 87.5, 100}
var rankLevels = [9]string{"S+", "A+", "A+", "A+", "B+", "B", "B-", "C+", "C"}

func exponentialCDF(x float64) float64 {
	return 1 - math.Pow(2, -x)
}

func logNormalCDF(x float64) float64 {
	return x / (1 + x)
}

// CalculateRank computes the user's rank from their GitHub statistics.
// Ported from src/calculateRank.js.
func CalculateRank(p Params) Rank {
	commitsMedian := float64(commitsMedianDefault)
	if p.AllCommits {
		commitsMedian = commitsMedianAll
	}

	rank := 1 - (commitsWeight*exponentialCDF(p.Commits/commitsMedian)+
		prsWeight*exponentialCDF(p.PRs/prsMedian)+
		issuesWeight*exponentialCDF(p.Issues/issuesMedian)+
		reviewsWeight*exponentialCDF(p.Reviews/reviewsMedian)+
		starsWeight*logNormalCDF(p.Stars/starsMedian)+
		followersWeight*logNormalCDF(p.Followers/followersMedian))/
		totalWeight

	percentile := rank * 100
	level := "C"
	for i, t := range rankThresholds {
		if percentile <= t {
			level = rankLevels[i]
			break
		}
	}

	return Rank{Level: level, Percentile: percentile}
}
