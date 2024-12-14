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

// HandleWakatime ports api/wakatime.js (wakatime card). Note the JS
// handler performs no blacklist check.
func HandleWakatime(w http.ResponseWriter, r *http.Request) {
	username := GetQuery(r, "username")

	if !isAuthorizedRequest(r, true) {
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

	stats, err := fetchers.FetchWakatimeStats(username, GetQuery(r, "api_domain"))
	if err != nil {
		WriteSVG(w, common.RenderError(err.Error(), errSecondary(err), AccessDeniedOptions(r)),
			fmt.Sprintf("max-age=%d, s-maxage=%d, stale-while-revalidate=%d",
				common.ErrorCacheSeconds/2, common.ErrorCacheSeconds, common.OneDay))
		return
	}

	cacheSeconds := EffectiveCacheSeconds(GetQuery(r, "cache_seconds"), common.CardCacheSeconds, common.SixHours, common.TwoDay)

	WriteSVG(w, cards.RenderWakatimeCard(stats, cards.WakatimeOptions{
		HideTitle:         common.ParseBoolean(GetQuery(r, "hide_title")),
		HideBorder:        common.ParseBoolean(GetQuery(r, "hide_border")),
		Hide:              common.ParseArray(GetQuery(r, "hide")),
		LineHeight:        atoiOr(GetQuery(r, "line_height"), 0),
		TitleColor:        GetQuery(r, "title_color"),
		IconColor:         GetQuery(r, "icon_color"),
		TextColor:         GetQuery(r, "text_color"),
		BgColor:           GetQuery(r, "bg_color"),
		Theme:             GetQuery(r, "theme"),
		CustomTitle:       GetQuery(r, "custom_title"),
		Locale:            strings.ToLower(locale),
		Layout:            GetQuery(r, "layout"),
		BorderColor:       GetQuery(r, "border_color"),
		DisplayFormat:     GetQuery(r, "display_format"),
		BorderRadius:      floatPtr(GetQuery(r, "border_radius")),
		HideProgress:      common.ParseBoolean(GetQuery(r, "hide_progress")),
		LangsCount:        intPtr(GetQuery(r, "langs_count")),
		DisableAnimations: common.ParseBoolean(GetQuery(r, "disable_animations")),
	}), fmt.Sprintf("max-age=%s, s-maxage=%d, stale-while-revalidate=%d",
		halfHeader(cacheSeconds), cacheSeconds, common.OneDay))
}
