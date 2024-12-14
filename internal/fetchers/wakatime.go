package fetchers

import (
	"fmt"
	"strings"

	"github.com/wthrajat/github-readme-stats/internal/common"
)

// FetchWakatimeStats fetches WakaTime stats for username from apiDomain
// (defaulting to wakatime.com).
func FetchWakatimeStats(username, apiDomain string) (*WakaTimeData, error) {
	if username == "" {
		return nil, &common.MissingParamError{MissedParams: []string{"username"}}
	}

	domain := "wakatime.com"
	if apiDomain != "" {
		domain = strings.TrimSuffix(apiDomain, "/")
	}

	body, status, err := getJSON(
		fmt.Sprintf("https://%s/api/v1/users/%s/stats?is_including_today=true", domain, username),
		nil,
	)
	if err != nil {
		return nil, err
	}
	if status < 200 || status > 299 {
		return nil, common.NewCustomError(
			fmt.Sprintf("Could not resolve to a User with the login of '%s'", username),
			common.ErrWakatime,
		)
	}

	data := childMap(body, "data")
	var langs []WakaTimeLang
	for _, lv := range childSlice(data, "languages") {
		lm := asMap(lv)
		langs = append(langs, WakaTimeLang{
			Name:    childString(lm, "name"),
			Text:    childString(lm, "text"),
			Percent: childFloat(lm, "percent"),
			Hours:   childInt(lm, "hours"),
			Minutes: childInt(lm, "minutes"),
		})
	}

	return &WakaTimeData{
		Languages:               langs,
		Range:                   childString(data, "range"),
		IsCodingActivityVisible: childBool(data, "is_coding_activity_visible"),
		IsOtherUsageVisible:     childBool(data, "is_other_usage_visible"),
	}, nil
}
