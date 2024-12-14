package handlers

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/wthrajat/github-readme-stats/internal/cards"
	"github.com/wthrajat/github-readme-stats/internal/common"
	"github.com/wthrajat/github-readme-stats/internal/fetchers"
	"github.com/wthrajat/github-readme-stats/internal/translations"
)

// validTopLangsLayouts mirrors the layout allowlist in api/top-langs.js.
var validTopLangsLayouts = map[string]bool{
	"compact": true, "normal": true, "donut": true,
	"donut-vertical": true, "pie": true,
}

// HandleTopLangs ports api/top-langs.js (top languages card).
func HandleTopLangs(w http.ResponseWriter, r *http.Request) {
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
		// Note: the JS handler renders this branch without color options.
		WriteSVG(w, common.RenderError("Something went wrong", "Locale not found", map[string]string{}), "")
		return
	}

	if queryHas(r, "layout") && !validTopLangsLayouts[GetQuery(r, "layout")] {
		WriteSVG(w, common.RenderError("Something went wrong", "Incorrect layout input", map[string]string{}), "")
		return
	}

	topLangs, err := fetchers.FetchTopLanguages(
		username,
		common.ParseArray(GetQuery(r, "exclude_repo")),
		parseFloatOr(GetQuery(r, "size_weight"), 1),
		parseFloatOr(GetQuery(r, "count_weight"), 0),
	)
	if err != nil {
		WriteSVG(w, common.RenderError(err.Error(), errSecondary(err), AccessDeniedOptions(r)),
			fmt.Sprintf("max-age=%d, s-maxage=%d, stale-while-revalidate=%d",
				common.ErrorCacheSeconds/2, common.ErrorCacheSeconds, common.OneDay))
		return
	}

	// Note: unlike the other card handlers, top-langs does not clamp the
	// cache duration; it only applies the CACHE_SECONDS env override.
	cacheSeconds := atoiOr(GetQuery(r, "cache_seconds"), common.TopLangsCacheSeconds)
	if env, ok := os.LookupEnv("CACHE_SECONDS"); ok && env != "" {
		if n, convErr := strconv.Atoi(env); convErr == nil && n != 0 {
			cacheSeconds = n
		}
	}

	WriteSVG(w, cards.RenderTopLanguages(topLangs, cards.TopLangOptions{
		CustomTitle:       GetQuery(r, "custom_title"),
		HideTitle:         common.ParseBoolean(GetQuery(r, "hide_title")),
		HideBorder:        common.ParseBoolean(GetQuery(r, "hide_border")),
		CardWidth:         floatPtr(GetQuery(r, "card_width")),
		Hide:              common.ParseArray(GetQuery(r, "hide")),
		TitleColor:        GetQuery(r, "title_color"),
		TextColor:         GetQuery(r, "text_color"),
		BgColor:           GetQuery(r, "bg_color"),
		Theme:             GetQuery(r, "theme"),
		Layout:            GetQuery(r, "layout"),
		LangsCount:        intPtr(GetQuery(r, "langs_count")),
		BorderRadius:      floatPtr(GetQuery(r, "border_radius")),
		BorderColor:       GetQuery(r, "border_color"),
		Locale:            strings.ToLower(locale),
		DisableAnimations: common.ParseBoolean(GetQuery(r, "disable_animations")),
		HideProgress:      common.ParseBoolean(GetQuery(r, "hide_progress")),
	}), fmt.Sprintf("max-age=%s, s-maxage=%d", halfHeader(cacheSeconds), cacheSeconds))
}
