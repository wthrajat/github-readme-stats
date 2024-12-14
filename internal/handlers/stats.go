package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/wthrajat/github-readme-stats/internal/cards"
	"github.com/wthrajat/github-readme-stats/internal/common"
	"github.com/wthrajat/github-readme-stats/internal/fetchers"
	"github.com/wthrajat/github-readme-stats/internal/translations"
)

// HandleStats ports api/index.js (stats card).
func HandleStats(w http.ResponseWriter, r *http.Request) {
	username := GetQuery(r, "username")

	if !isAuthorizedRequest(r, true) {
		WriteSVG(w, common.RenderAccessDenied(AccessDeniedOptions(r)), "")
		return
	}

	if common.IsBlacklisted(username) {
		WriteSVG(w, common.RenderError(
			"Something went wrong",
			"This username is blacklisted",
			AccessDeniedOptions(r),
		), "")
		return
	}

	locale := GetQuery(r, "locale")
	if locale != "" && !translations.IsLocaleAvailable(locale) {
		WriteSVG(w, common.RenderError(
			"Something went wrong",
			"Language not found",
			AccessDeniedOptions(r),
		), "")
		return
	}

	show := common.ParseArray(GetQuery(r, "show"))
	showSet := map[string]bool{}
	for _, s := range show {
		showSet[s] = true
	}

	stats, err := fetchers.FetchStats(
		username,
		BoolOr(common.ParseBoolean(GetQuery(r, "include_all_commits")), false),
		common.ParseArray(GetQuery(r, "exclude_repo")),
		showSet["prs_merged"] || showSet["prs_merged_percentage"],
		showSet["discussions_started"],
		showSet["discussions_answered"],
	)
	if err != nil {
		WriteSVG(w, common.RenderError(err.Error(), errSecondary(err), AccessDeniedOptions(r)),
			fmt.Sprintf("public, max-age=%d, s-maxage=%d, stale-while-revalidate=%d, stale-if-error=%d",
				common.ErrorCacheSeconds/2, common.ErrorCacheSeconds, common.OneDay, common.OneDay))
		return
	}

	cacheSeconds := EffectiveCacheSeconds(GetQuery(r, "cache_seconds"), common.CardCacheSeconds, common.TwelveHours, common.TenDay)

	WriteSVG(w, cards.RenderStatsCard(stats, cards.StatCardOptions{
		Hide:              common.ParseArray(GetQuery(r, "hide")),
		ShowIcons:         common.ParseBoolean(GetQuery(r, "show_icons")),
		HideTitle:         common.ParseBoolean(GetQuery(r, "hide_title")),
		HideBorder:        common.ParseBoolean(GetQuery(r, "hide_border")),
		CardWidth:         floatPtr(GetQuery(r, "card_width")),
		HideRank:          common.ParseBoolean(GetQuery(r, "hide_rank")),
		IncludeAllCommits: common.ParseBoolean(GetQuery(r, "include_all_commits")),
		LineHeight:        atoiOr(GetQuery(r, "line_height"), 0),
		TitleColor:        GetQuery(r, "title_color"),
		RingColor:         GetQuery(r, "ring_color"),
		IconColor:         GetQuery(r, "icon_color"),
		TextColor:         GetQuery(r, "text_color"),
		TextBold:          common.ParseBoolean(GetQuery(r, "text_bold")),
		BgColor:           GetQuery(r, "bg_color"),
		Theme:             GetQuery(r, "theme"),
		CustomTitle:       GetQuery(r, "custom_title"),
		BorderRadius:      floatPtr(GetQuery(r, "border_radius")),
		BorderColor:       GetQuery(r, "border_color"),
		NumberFormat:      GetQuery(r, "number_format"),
		Locale:            strings.ToLower(locale),
		DisableAnimations: common.ParseBoolean(GetQuery(r, "disable_animations")),
		RankIcon:          GetQuery(r, "rank_icon"),
		Show:              show,
		Lowercase:         common.ParseBoolean(GetQuery(r, "lowercase")),
	}), fmt.Sprintf("public, max-age=%d, s-maxage=%d, stale-while-revalidate=%d, stale-if-error=%d",
		cacheSeconds, cacheSeconds, common.OneDay, common.TwoDay))
}
