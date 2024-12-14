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

// HandlePin ports api/pin.js (repo card).
func HandlePin(w http.ResponseWriter, r *http.Request) {
	username := GetQuery(r, "username")
	repo := GetQuery(r, "repo")

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

	repoData, err := fetchers.FetchRepo(username, repo)
	if err != nil {
		WriteSVG(w, common.RenderError(err.Error(), errSecondary(err), AccessDeniedOptions(r)),
			fmt.Sprintf("max-age=%d, s-maxage=%d, stale-while-revalidate=%d",
				common.ErrorCacheSeconds/2, common.ErrorCacheSeconds, common.OneDay))
		return
	}

	cacheSeconds := EffectiveCacheSeconds(GetQuery(r, "cache_seconds"), common.PinCardCacheSeconds, common.OneDay, common.TenDay)

	WriteSVG(w, cards.RenderRepoCard(repoData, cards.RepoCardOptions{
		HideBorder:            common.ParseBoolean(GetQuery(r, "hide_border")),
		TitleColor:            GetQuery(r, "title_color"),
		IconColor:             GetQuery(r, "icon_color"),
		TextColor:             GetQuery(r, "text_color"),
		BgColor:               GetQuery(r, "bg_color"),
		Theme:                 GetQuery(r, "theme"),
		BorderRadius:          floatPtr(GetQuery(r, "border_radius")),
		BorderColor:           GetQuery(r, "border_color"),
		ShowOwner:             common.ParseBoolean(GetQuery(r, "show_owner")),
		Locale:                strings.ToLower(locale),
		DescriptionLinesCount: intPtr(GetQuery(r, "description_lines_count")),
	}), fmt.Sprintf("max-age=%d, s-maxage=%d", cacheSeconds, cacheSeconds))
}
