package handlers

import (
	"fmt"
	"net/http"

	"github.com/wthrajat/github-readme-stats/internal/cards"
	"github.com/wthrajat/github-readme-stats/internal/common"
	"github.com/wthrajat/github-readme-stats/internal/fetchers"
	"github.com/wthrajat/github-readme-stats/internal/translations"
)

// HandleGist ports api/gist.js (gist card). Authorization does not require
// a username here, and there is no blacklist check, mirroring the JS.
func HandleGist(w http.ResponseWriter, r *http.Request) {
	if !isAuthorizedRequest(r, false) {
		WriteSVG(w, common.RenderAccessDenied(AccessDeniedOptions(r)), "")
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

	gistData, err := fetchers.FetchGist(GetQuery(r, "id"))
	if err != nil {
		WriteSVG(w, common.RenderError(err.Error(), errSecondary(err), AccessDeniedOptions(r)),
			fmt.Sprintf("max-age=%d, s-maxage=%d, stale-while-revalidate=%d",
				common.ErrorCacheSeconds/2, common.ErrorCacheSeconds, common.OneDay))
		return
	}

	cacheSeconds := EffectiveCacheSeconds(GetQuery(r, "cache_seconds"), common.TwoDay, common.TwoDay, common.SixDay)

	// Note: the Go GistCardOptions carries no locale field.
	WriteSVG(w, cards.RenderGistCard(gistData, cards.GistCardOptions{
		TitleColor:   GetQuery(r, "title_color"),
		IconColor:    GetQuery(r, "icon_color"),
		TextColor:    GetQuery(r, "text_color"),
		BgColor:      GetQuery(r, "bg_color"),
		Theme:        GetQuery(r, "theme"),
		BorderRadius: floatPtr(GetQuery(r, "border_radius")),
		BorderColor:  GetQuery(r, "border_color"),
		ShowOwner:    common.ParseBoolean(GetQuery(r, "show_owner")),
		HideBorder:   common.ParseBoolean(GetQuery(r, "hide_border")),
	}), fmt.Sprintf("max-age=%d, s-maxage=%d", cacheSeconds, cacheSeconds))
}
