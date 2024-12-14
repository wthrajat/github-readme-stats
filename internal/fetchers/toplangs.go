package fetchers

import (
	"math"
	"net/http"

	"github.com/wthrajat/github-readme-stats/internal/common"
)

const topLangsQuery = `
      query userInfo($login: String!) {
        user(login: $login) {
          # fetch only owner repos & not forks
          repositories(ownerAffiliations: OWNER, isFork: false, first: 100) {
            nodes {
              name
              languages(first: 10, orderBy: {field: SIZE, direction: DESC}) {
                edges {
                  size
                  node {
                    color
                    name
                  }
                }
              }
            }
          }
        }
      }
      `

// FetchTopLanguages fetches aggregated top language stats for username.
func FetchTopLanguages(username string, excludeRepo []string, sizeWeight, countWeight float64) (TopLangData, error) {
	if username == "" {
		return nil, &common.MissingParamError{MissedParams: []string{"username"}}
	}

	fetcher := func(vars map[string]any, token string) (map[string]any, int, error) {
		return doGraphQL(topLangsQuery, vars, "token "+token)
	}

	body, status, err := Retryer(fetcher, map[string]any{"login": username})
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
			"Something went wrong while trying to retrieve the language data using the GraphQL API.",
			common.ErrGraphQLError,
		)
	}

	nodes := childSlice(childMap(body, "data", "user", "repositories"), "nodes")

	hidden := map[string]bool{}
	for _, r := range excludeRepo {
		hidden[r] = true
	}

	type edge struct {
		size  float64
		name  string
		color string
	}
	var flat []edge
	for _, n := range nodes {
		node := asMap(n)
		if hidden[childString(node, "name")] {
			continue
		}
		langEdges := childSlice(childMap(node, "languages"), "edges")
		if len(langEdges) == 0 {
			continue
		}
		var repoEdges []edge
		for _, e := range langEdges {
			em := asMap(e)
			langNode := asMap(em["node"])
			repoEdges = append(repoEdges, edge{
				size:  childFloat(em, "size"),
				name:  childString(langNode, "name"),
				color: childString(langNode, "color"),
			})
		}
		flat = append(repoEdges, flat...)
	}

	agg := map[string]*Lang{}
	repoCount := 0
	for _, e := range flat {
		if cur, ok := agg[e.name]; ok && e.name == cur.Name {
			cur.Size = e.size + cur.Size
			repoCount++
			cur.Count = float64(repoCount)
			cur.Color = e.color
		} else {
			repoCount = 1
			agg[e.name] = &Lang{Name: e.name, Color: e.color, Size: e.size, Count: 1}
		}
	}

	for _, l := range agg {
		l.Size = math.Pow(l.Size, sizeWeight) * math.Pow(l.Count, countWeight)
	}

	return TopLangData(agg), nil
}
