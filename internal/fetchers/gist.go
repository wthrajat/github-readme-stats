package fetchers

import (
	"errors"
	"fmt"
	"sort"

	"github.com/wthrajat/github-readme-stats/internal/common"
)

const gistQuery = `
query gistInfo($gistName: String!) {
    viewer {
        gist(name: $gistName) {
            description
            owner {
                login
            }
            stargazerCount
            forks {
                totalCount
            }
            files {
                name
                language {
                    name
                }
                size
            }
        }
    }
}
`

// CalculatePrimaryLanguage returns the language with the largest total file
// size. Ties resolve alphabetically for determinism.
func CalculatePrimaryLanguage(files map[string]GistFile) string {
	sizes := map[string]float64{}
	for _, f := range files {
		if f.Language != nil {
			sizes[f.Language.Name] += f.Size
		}
	}
	names := make([]string, 0, len(sizes))
	for name := range sizes {
		names = append(names, name)
	}
	sort.Strings(names)
	primary := ""
	for _, name := range names {
		if primary == "" || sizes[name] > sizes[primary] {
			primary = name
		}
	}
	return primary
}

func toGistFile(v any) GistFile {
	m := asMap(v)
	f := GistFile{Name: childString(m, "name"), Size: childFloat(m, "size")}
	switch lang := m["language"].(type) {
	case map[string]any:
		f.Language = &GistLanguage{Name: childString(lang, "name")}
	case string:
		f.Language = &GistLanguage{Name: lang}
	}
	return f
}

func parseGistFiles(v any) map[string]GistFile {
	out := map[string]GistFile{}
	switch t := v.(type) {
	case map[string]any:
		for k, fv := range t {
			out[k] = toGistFile(fv)
		}
	case []any:
		for i, fv := range t {
			f := toGistFile(fv)
			key := f.Name
			if key == "" {
				key = fmt.Sprintf("%d", i)
			}
			out[key] = f
		}
	}
	return out
}

// FetchGist fetches GitHub gist information by ID.
func FetchGist(id string) (*GistData, error) {
	if id == "" {
		return nil, &common.MissingParamError{MissedParams: []string{"id"}, Secondary: "/api/gist?id=GIST_ID"}
	}

	fetcher := func(vars map[string]any, token string) (map[string]any, int, error) {
		return doGraphQL(gistQuery, vars, "token "+token)
	}

	body, _, err := Retryer(fetcher, map[string]any{"gistName": id})
	if err != nil {
		return nil, err
	}
	if firstGLError(body) != nil {
		return nil, errors.New(firstGLErrorMessage(body))
	}
	gist := childMap(body, "data", "viewer", "gist")
	if gist == nil {
		return nil, errors.New("Gist not found")
	}
	files := parseGistFiles(gist["files"])
	if len(files) == 0 {
		return nil, errors.New("Gist not found")
	}

	keys := make([]string, 0, len(files))
	for k := range files {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	first := files[keys[0]]
	owner := childString(gist, "owner", "login")

	return &GistData{
		Name:          first.Name,
		NameWithOwner: owner + "/" + first.Name,
		Description:   childString(gist, "description"),
		Language:      CalculatePrimaryLanguage(files),
		StarsCount:    childInt(gist, "stargazerCount"),
		ForksCount:    childInt(gist, "forks", "totalCount"),
	}, nil
}
