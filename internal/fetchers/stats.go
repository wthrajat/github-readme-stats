package fetchers

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"regexp"

	"github.com/wthrajat/github-readme-stats/internal/common"
	"github.com/wthrajat/github-readme-stats/internal/rank"
)

const graphQLReposField = `
  repositories(first: 100, ownerAffiliations: OWNER, orderBy: {direction: DESC, field: STARGAZERS}, after: $after) {
    totalCount
    nodes {
      name
      stargazers {
        totalCount
      }
    }
    pageInfo {
      hasNextPage
      endCursor
    }
  }
`

const graphQLReposQuery = `
  query userInfo($login: String!, $after: String) {
    user(login: $login) {
      ` + graphQLReposField + `
    }
  }
`

const graphQLStatsQuery = `
  query userInfo($login: String!, $after: String, $includeMergedPullRequests: Boolean!, $includeDiscussions: Boolean!, $includeDiscussionsAnswers: Boolean!) {
    user(login: $login) {
      name
      login
      contributionsCollection {
        totalCommitContributions,
        totalIssueContributions,
        totalPullRequestContributions,
        totalPullRequestReviewContributions
      }
      repositoriesContributedTo(first: 1, contributionTypes: [COMMIT, ISSUE, PULL_REQUEST, REPOSITORY]) {
        totalCount
      }
      pullRequests(first: 1) {
        totalCount
      }
      mergedPullRequests: pullRequests(states: MERGED) @include(if: $includeMergedPullRequests) {
        totalCount
      }
      openIssues: issues(states: OPEN) {
        totalCount
      }
      closedIssues: issues(states: CLOSED) {
        totalCount
      }
      followers {
        totalCount
      }
      repositoryDiscussions @include(if: $includeDiscussions) {
        totalCount
      }
      repositoryDiscussionComments(onlyAnswers: true) @include(if: $includeDiscussionsAnswers) {
        totalCount
      }
      ` + graphQLReposField + `
    }
  }
`

// githubUsernameRe matches valid GitHub usernames (github-username-regex,
// rewritten without lookahead since Go's regexp has no lookahead support:
// hyphens are only allowed between alphanumerics, length capped at 39).
var githubUsernameRe = regexp.MustCompile(`(?i)^[a-z0-9]+(?:-[a-z0-9]+)*$`)

func isValidGitHubUsername(username string) bool {
	return len(username) >= 1 && len(username) <= 39 && githubUsernameRe.MatchString(username)
}

func statsFetcher(username string, includeMergedPRs, includeDiscussions, includeDiscussionsAnswers bool) (map[string]any, int, error) {
	fetcher := func(vars map[string]any, token string) (map[string]any, int, error) {
		query := graphQLStatsQuery
		if after, ok := vars["after"].(string); ok && after != "" {
			query = graphQLReposQuery
		}
		return doGraphQL(query, vars, "bearer "+token)
	}

	var stats map[string]any
	firstStatus := 0
	hasNextPage := true
	var endCursor any
	for hasNextPage {
		vars := map[string]any{
			"login":                     username,
			"first":                     100,
			"after":                     endCursor,
			"includeMergedPullRequests": includeMergedPRs,
			"includeDiscussions":        includeDiscussions,
			"includeDiscussionsAnswers": includeDiscussionsAnswers,
		}
		body, status, err := Retryer(fetcher, vars)
		if err != nil {
			return nil, status, err
		}
		if firstGLError(body) != nil {
			return body, status, nil
		}
		if stats == nil {
			stats = body
			firstStatus = status
		} else {
			statsNodes := childSlice(childMap(stats, "data", "user", "repositories"), "nodes")
			combined := append(statsNodes, childSlice(childMap(body, "data", "user", "repositories"), "nodes")...)
			childMap(stats, "data", "user", "repositories")["nodes"] = combined
		}

		repos := childMap(body, "data", "user", "repositories")
		nodes := childSlice(repos, "nodes")
		starred := 0
		for _, n := range nodes {
			if childInt(asMap(n), "stargazers", "totalCount") != 0 {
				starred++
			}
		}
		pageInfo := childMap(repos, "pageInfo")
		hasNextPage = os.Getenv("FETCH_MULTI_PAGE_STARS") == "true" &&
			len(nodes) == starred &&
			childBool(pageInfo, "hasNextPage")
		endCursor, _ = pageInfo["endCursor"]
	}
	return stats, firstStatus, nil
}

func totalCommitsFetcher(username string) (int, error) {
	if !isValidGitHubUsername(username) {
		return 0, errors.New("Invalid username provided.")
	}

	fetcher := func(vars map[string]any, token string) (map[string]any, int, error) {
		login, _ := vars["login"].(string)
		return getJSON("https://api.github.com/search/commits?q=author:"+login,
			map[string]string{
				"Content-Type":  "application/json",
				"Accept":        "application/vnd.github.cloak-preview",
				"Authorization": "token " + token,
			})
	}

	body, _, err := Retryer(fetcher, map[string]any{"login": username})
	if err != nil {
		return 0, fmt.Errorf("%v", err)
	}
	total := childInt(body, "total_count")
	if total == 0 {
		return 0, common.NewCustomError("Could not fetch total commits.", common.ErrGithubRestError)
	}
	return total, nil
}

// FetchStats fetches aggregated GitHub stats for username.
func FetchStats(username string, includeAllCommits bool, excludeRepo []string, includeMergedPRs, includeDiscussions, includeDiscussionsAnswers bool) (*StatsData, error) {
	if username == "" {
		return nil, &common.MissingParamError{MissedParams: []string{"username"}}
	}

	stats := &StatsData{Rank: rank.Rank{Level: "C", Percentile: 100}}

	body, status, err := statsFetcher(username, includeMergedPRs, includeDiscussions, includeDiscussionsAnswers)
	if err != nil {
		return nil, err
	}

	if firstGLError(body) != nil {
		if firstGLErrorType(body) == "NOT_FOUND" {
			msg := firstGLErrorMessage(body)
			if msg == "" {
				msg = "Could not fetch user."
			}
			return nil, common.NewCustomError(msg, common.ErrUserNotFound)
		}
		if msg := firstGLErrorMessage(body); msg != "" {
			return nil, common.NewCustomError(firstWrappedLine(msg, 90), http.StatusText(status))
		}
		return nil, common.NewCustomError(
			"Something went wrong while trying to retrieve the stats data using the GraphQL API.",
			common.ErrGraphQLError,
		)
	}

	user := childMap(body, "data", "user")
	if user == nil {
		user = map[string]any{}
	}

	stats.Name = childString(user, "name")
	if stats.Name == "" {
		stats.Name = childString(user, "login")
	}

	if includeAllCommits {
		total, err := totalCommitsFetcher(username)
		if err != nil {
			return nil, err
		}
		stats.TotalCommits = total
	} else {
		stats.TotalCommits = childInt(user, "contributionsCollection", "totalCommitContributions")
	}

	stats.TotalPRs = childInt(user, "pullRequests", "totalCount")
	if includeMergedPRs {
		stats.TotalPRsMerged = childInt(user, "mergedPullRequests", "totalCount")
		stats.MergedPRsPercentage = float64(stats.TotalPRsMerged) / float64(stats.TotalPRs) * 100
	}
	stats.TotalReviews = childInt(user, "contributionsCollection", "totalPullRequestReviewContributions")
	stats.TotalContributions =
		childInt(user, "contributionsCollection", "totalCommitContributions") +
			childInt(user, "contributionsCollection", "totalIssueContributions") +
			childInt(user, "contributionsCollection", "totalPullRequestContributions") +
			childInt(user, "contributionsCollection", "totalPullRequestReviewContributions")
	stats.TotalIssues = childInt(user, "openIssues", "totalCount") + childInt(user, "closedIssues", "totalCount")
	if includeDiscussions {
		stats.TotalDiscussionsStarted = childInt(user, "repositoryDiscussions", "totalCount")
	}
	if includeDiscussionsAnswers {
		stats.TotalDiscussionsAnswered = childInt(user, "repositoryDiscussionComments", "totalCount")
	}
	stats.ContributedTo = childInt(user, "repositoriesContributedTo", "totalCount")

	hidden := map[string]bool{}
	for _, r := range excludeRepo {
		hidden[r] = true
	}
	totalStars := 0
	for _, n := range childSlice(childMap(user, "repositories"), "nodes") {
		node := asMap(n)
		if !hidden[childString(node, "name")] {
			totalStars += childInt(node, "stargazers", "totalCount")
		}
	}
	stats.TotalStars = totalStars

	stats.Rank = rank.CalculateRank(rank.Params{
		AllCommits: includeAllCommits,
		Commits:    float64(stats.TotalCommits),
		PRs:        float64(stats.TotalPRs),
		Issues:     float64(stats.TotalIssues),
		Reviews:    float64(stats.TotalReviews),
		Repos:      float64(childInt(user, "repositories", "totalCount")),
		Stars:      float64(stats.TotalStars),
		Followers:  float64(childInt(user, "followers", "totalCount")),
	})

	return stats, nil
}
