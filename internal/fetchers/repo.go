package fetchers

import (
	"errors"

	"github.com/wthrajat/github-readme-stats/internal/common"
)

const repoQuery = `
      fragment RepoInfo on Repository {
        name
        nameWithOwner
        isPrivate
        isArchived
        isTemplate
        stargazers {
          totalCount
        }
        description
        primaryLanguage {
          color
          id
          name
        }
        forkCount
      }
      query getRepo($login: String!, $repo: String!) {
        user(login: $login) {
          repository(name: $repo) {
            ...RepoInfo
          }
        }
        organization(login: $login) {
          repository(name: $repo) {
            ...RepoInfo
          }
        }
      }
    `

const repoURLExample = "/api/pin?username=USERNAME&amp;repo=REPO_NAME"

func repoDataFromMap(m map[string]any) *RepositoryData {
	d := &RepositoryData{
		Name:          childString(m, "name"),
		NameWithOwner: childString(m, "nameWithOwner"),
		Description:   childString(m, "description"),
		IsPrivate:     childBool(m, "isPrivate"),
		IsArchived:    childBool(m, "isArchived"),
		IsTemplate:    childBool(m, "isTemplate"),
		StarCount:     childInt(m, "stargazers", "totalCount"),
		ForkCount:     childInt(m, "forkCount"),
	}
	if pl := childMap(m, "primaryLanguage"); pl != nil {
		d.PrimaryLanguage = &PrimaryLanguage{
			Color: childString(pl, "color"),
			ID:    childString(pl, "id"),
			Name:  childString(pl, "name"),
		}
	}
	return d
}

// FetchRepo fetches repository data for username/reponame.
func FetchRepo(username, reponame string) (*RepositoryData, error) {
	if username == "" && reponame == "" {
		return nil, &common.MissingParamError{MissedParams: []string{"username", "repo"}, Secondary: repoURLExample}
	}
	if username == "" {
		return nil, &common.MissingParamError{MissedParams: []string{"username"}, Secondary: repoURLExample}
	}
	if reponame == "" {
		return nil, &common.MissingParamError{MissedParams: []string{"repo"}, Secondary: repoURLExample}
	}

	fetcher := func(vars map[string]any, token string) (map[string]any, int, error) {
		return doGraphQL(repoQuery, vars, "token "+token)
	}

	body, _, err := Retryer(fetcher, map[string]any{"login": username, "repo": reponame})
	if err != nil {
		return nil, err
	}

	data := childMap(body, "data")
	userNode := childMap(data, "user")
	orgNode := childMap(data, "organization")

	if userNode == nil && orgNode == nil {
		return nil, errors.New("Not found")
	}

	isUser := orgNode == nil && userNode != nil
	isOrg := userNode == nil && orgNode != nil

	if isUser {
		repo := childMap(userNode, "repository")
		if repo == nil || childBool(repo, "isPrivate") {
			return nil, errors.New("User Repository Not found")
		}
		return repoDataFromMap(repo), nil
	}

	if isOrg {
		repo := childMap(orgNode, "repository")
		if repo == nil || childBool(repo, "isPrivate") {
			return nil, errors.New("Organization Repository Not found")
		}
		return repoDataFromMap(repo), nil
	}

	return nil, errors.New("Unexpected behavior")
}
