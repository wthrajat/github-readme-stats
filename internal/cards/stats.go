package cards

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/wthrajat/github-readme-stats/internal/common"
	"github.com/wthrajat/github-readme-stats/internal/fetchers"
	"github.com/wthrajat/github-readme-stats/internal/translations"
)

// StatCardOptions mirrors the JS StatCardOptions for the stats card.
type StatCardOptions struct {
	Hide                                                                                                     []string
	ShowIcons, HideTitle, HideBorder                                                                         *bool
	CardWidth                                                                                                *float64
	HideRank                                                                                                 *bool
	IncludeAllCommits                                                                                        *bool
	LineHeight                                                                                               int
	TitleColor, RingColor, IconColor, TextColor, BgColor, Theme, CustomTitle, NumberFormat, Locale, RankIcon string
	BorderRadius                                                                                             *float64
	BorderColor                                                                                              string
	TextBold                                                                                                 *bool
	DisableAnimations                                                                                        *bool
	Show                                                                                                     []string
	Lowercase                                                                                                *bool
}

const (
	statCardMinWidth       = 287
	statCardDefaultWidth   = 287
	rankCardMinWidth       = 420
	rankCardDefaultWidth   = 450
	rankOnlyCardMinWidth   = 290
	rankOnlyCardDefaultWid = 290
)

// statLongLocales mirrors the LONG_LOCALES set in stats-card.js.
var statLongLocales = map[string]bool{
	"cn": true, "es": true, "fr": true, "pt-br": true, "ru": true,
	"uk-ua": true, "id": true, "ml": true, "my": true, "pl": true,
	"de": true, "nl": true, "zh-tw": true, "uz": true,
}

// statEntry is a single row of the stats card.
type statEntry struct {
	icon       string
	label      string
	value      float64
	display    string // preformatted display value (percentage stats); "" formats from value
	id         string
	unitSymbol string
	a11yValue  string
}

// statTextNode creates a stats card text item.
func statTextNode(icon, label string, value float64, display, id, unitSymbol string, index int, showIcons bool, shiftValuePos float64, bold bool, numberFormat string) string {
	var kValue string
	if display != "" {
		if strings.ToLower(numberFormat) == "long" {
			kValue = display
		} else {
			var f float64
			fmt.Sscanf(display, "%f", &f)
			kValue = common.KFormatter(f)
		}
	} else if strings.ToLower(numberFormat) == "long" {
		kValue = formatLong(value)
	} else {
		kValue = common.KFormatter(value)
	}
	staggerDelay := (index + 3) * 150

	labelOffset := ""
	if showIcons {
		labelOffset = `x="25"`
	}
	iconSvg := ""
	if showIcons {
		iconSvg = `
    <svg data-testid="icon" data-icon-for="` + id + `" class="icon" viewBox="0 0 16 16" version="1.1" width="16" height="16">
      ` + icon + `
    </svg>
  `
	}
	boldClass := "not_bold"
	if bold {
		boldClass = " bold"
	}
	x := 120.0 + shiftValuePos
	if showIcons {
		x = 140.0 + shiftValuePos
	}
	unit := ""
	if unitSymbol != "" {
		unit = " " + unitSymbol
	}
	return `
    <g class="stagger" style="animation-delay: ` + fmt.Sprint(staggerDelay) + `ms" transform="translate(25, 0)">
      ` + iconSvg + `
      <text class="stat ` + boldClass + `" ` + labelOffset + ` y="12.5">` + label + `:</text>
      <text
        class="stat stat-value ` + boldClass + `"
        x="` + fmtNum(x) + `"
        y="12.5"
        data-testid="` + id + `"
      >` + kValue + unit + `</text>
    </g>
  `
}

// calculateCircleProgress calculates progress along the boundary of the
// circle, i.e. its circumference.
func calculateCircleProgress(value float64) float64 {
	radius := 40.0
	c := math.Pi * (radius * 2)

	if value < 0 {
		value = 0
	}
	if value > 100 {
		value = 100
	}

	return ((100 - value) / 100) * c
}

// getProgressAnimation retrieves the animation to display progress along the
// circumference of circle from the beginning to the given value in a
// clockwise direction.
func getProgressAnimation(progress float64) string {
	return `
    @keyframes rankAnimation {
      from {
        stroke-dashoffset: ` + fmtNum(calculateCircleProgress(0)) + `;
      }
      to {
        stroke-dashoffset: ` + fmtNum(calculateCircleProgress(progress)) + `;
      }
    }
  `
}

// statCardStyles retrieves CSS styles for the stats card.
func statCardStyles(textColor, iconColor, ringColor string, showIcons bool, progress float64) string {
	display := "none"
	if showIcons {
		display = "block"
	}
	return `
    .stat {
      font: 600 14px 'Segoe UI', Ubuntu, "Helvetica Neue", Sans-Serif; fill: ` + textColor + `;
      letter-spacing: 0.15px;
    }
    @supports(-moz-appearance: auto) {
      /* Selector detects Firefox */
      .stat { font-size:12px; }
    }
    .stat-value {
      filter: drop-shadow(0 0 6px rgba(96, 165, 250, 0.28));
    }
    .stagger {
      opacity: 0;
      animation: fadeInAnimation 0.3s ease-in-out forwards;
    }
    .rank-text {
      font: 800 22px 'Segoe UI', Ubuntu, Sans-Serif; fill: ` + textColor + `;
      animation: scaleInAnimation 0.3s ease-in-out forwards;
      filter: drop-shadow(0 0 10px rgba(34, 211, 238, 0.3));
    }
    .rank-percentile-header {
      font-size: 13px;
    }
    .rank-percentile-text {
      font-size: 15px;
    }

    .not_bold { font-weight: 400 }
    .bold { font-weight: 700 }
    .icon {
      fill: ` + iconColor + `;
      display: ` + display + `;
    }
    [data-icon-for="stars"] {
      filter: drop-shadow(0 0 4px rgba(248, 216, 71, 0.35));
    }
    [data-icon-for="commits"] {
      filter: drop-shadow(0 0 4px rgba(88, 166, 255, 0.4));
    }
    [data-icon-for="prs"] {
      filter: drop-shadow(0 0 4px rgba(191, 145, 243, 0.4));
    }
    [data-icon-for="issues"] {
      filter: drop-shadow(0 0 4px rgba(121, 255, 151, 0.3));
    }
    [data-icon-for="contribs"] {
      filter: drop-shadow(0 0 4px rgba(34, 211, 238, 0.38));
    }
    [data-icon-for="reviews"],
    [data-icon-for="contributions"],
    [data-icon-for="discussions_started"],
    [data-icon-for="discussions_answered"],
    [data-icon-for="prs_merged"],
    [data-icon-for="prs_merged_percentage"] {
      filter: drop-shadow(0 0 4px rgba(96, 165, 250, 0.3));
    }

    .rank-circle-rim {
      stroke: ` + ringColor + `;
      fill: none;
      stroke-width: 5.5;
      opacity: 0.15;
    }
    .rank-circle {
      stroke: url(#rank-gradient);
      stroke-dasharray: 250;
      fill: none;
      stroke-width: 5.5;
      stroke-linecap: round;
      opacity: 0.95;
      transform-origin: -10px 8px;
      transform: rotate(-90deg);
      animation: rankAnimation 1s forwards ease-in-out;
      filter: url(#rank-glow);
    }
    ` + getProgressAnimation(progress) + `
  `
}

// RenderStatsCard renders the stats card SVG.
func RenderStatsCard(stats *fetchers.StatsData, o StatCardOptions) string {
	showIcons := boolOpt(o.ShowIcons)
	hideBorder := boolOpt(o.HideBorder)
	hideRank := boolOpt(o.HideRank)
	includeAllCommits := boolOpt(o.IncludeAllCommits)
	disableAnimations := boolOpt(o.DisableAnimations)
	// Hide the title altogether so the card starts directly with stats.
	// An explicit hide_title=false is the only escape hatch to show a title.
	effectiveHideTitle := o.HideTitle == nil || *o.HideTitle
	bold := o.TextBold == nil || *o.TextBold
	doLowercase := boolOpt(o.Lowercase)

	lheight := o.LineHeight
	if lheight == 0 {
		lheight = 25
	}
	numberFormat := o.NumberFormat
	if numberFormat == "" {
		numberFormat = "short"
	}
	theme := o.Theme
	if theme == "" {
		theme = "default"
	}
	rankIconKind := o.RankIcon
	if rankIconKind == "" {
		rankIconKind = "default"
	}
	currentYear := time.Now().Year()

	colors := common.GetCardColors(o.TitleColor, o.TextColor, o.IconColor, o.BgColor, o.BorderColor, o.RingColor, theme, "default")

	apostrophe := "s"
	if runes := []rune(stats.Name); len(runes) > 0 {
		if last := strings.ToLower(string(runes[len(runes)-1])); last == "x" || last == "s" {
			apostrophe = ""
		}
	}
	i18n := common.NewI18n(o.Locale, translations.StatCardLocales(stats.Name, apostrophe))

	// Meta data for creating text nodes with statTextNode.
	commitsLabel := i18n.T("statcard.commits")
	if !includeAllCommits {
		commitsLabel += fmt.Sprintf(" (%d)", currentYear)
	}
	entries := []statEntry{
		{icon: common.Icons["star"], label: i18n.T("statcard.totalstars"), value: float64(stats.TotalStars), id: "stars", a11yValue: fmtNum(float64(stats.TotalStars))},
		{icon: common.Icons["commits"], label: commitsLabel, value: float64(stats.TotalCommits), id: "commits", a11yValue: fmtNum(float64(stats.TotalCommits))},
		{icon: common.Icons["prs"], label: i18n.T("statcard.prs-merged"), value: float64(stats.TotalPRs), id: "prs", a11yValue: fmtNum(float64(stats.TotalPRs))},
	}

	showSet := map[string]bool{}
	for _, s := range o.Show {
		showSet[s] = true
	}

	// Default row already carries the merged label, so show=prs_merged is
	// redundant and intentionally a no-op.
	if showSet["prs_merged_percentage"] {
		pct := float64(stats.MergedPRsPercentage)
		entries = append(entries, statEntry{
			icon: common.Icons["prs_merged_percentage"], label: i18n.T("statcard.prs-merged-percentage"),
			value: pct, display: fmt.Sprintf("%.2f", pct), id: "prs_merged_percentage", unitSymbol: "%",
			a11yValue: fmt.Sprintf("%.2f", pct),
		})
	}
	if showSet["reviews"] {
		entries = append(entries, statEntry{icon: common.Icons["reviews"], label: i18n.T("statcard.reviews"), value: float64(stats.TotalReviews), id: "reviews", a11yValue: fmtNum(float64(stats.TotalReviews))})
	}
	if showSet["contributions"] {
		entries = append(entries, statEntry{icon: common.Icons["star"], label: i18n.T("statcard.contributions"), value: float64(stats.TotalContributions), id: "contributions", a11yValue: fmtNum(float64(stats.TotalContributions))})
	}

	entries = append(entries, statEntry{icon: common.Icons["issues"], label: i18n.T("statcard.issues"), value: float64(stats.TotalIssues), id: "issues", a11yValue: fmtNum(float64(stats.TotalIssues))})

	if showSet["discussions_started"] {
		entries = append(entries, statEntry{icon: common.Icons["discussions_started"], label: i18n.T("statcard.discussions-started"), value: float64(stats.TotalDiscussionsStarted), id: "discussions_started", a11yValue: fmtNum(float64(stats.TotalDiscussionsStarted))})
	}
	if showSet["discussions_answered"] {
		entries = append(entries, statEntry{icon: common.Icons["discussions_answered"], label: i18n.T("statcard.discussions-answered"), value: float64(stats.TotalDiscussionsAnswered), id: "discussions_answered", a11yValue: fmtNum(float64(stats.TotalDiscussionsAnswered))})
	}

	entries = append(entries, statEntry{icon: common.Icons["contribs"], label: i18n.T("statcard.contribs"), value: float64(stats.ContributedTo), id: "contribs", a11yValue: fmtNum(float64(stats.ContributedTo))})

	if doLowercase {
		for i := range entries {
			entries[i].label = strings.ToLower(entries[i].label)
		}
	}

	hideSet := map[string]bool{}
	for _, h := range o.Hide {
		hideSet[h] = true
	}
	visible := make([]statEntry, 0, len(entries))
	for _, e := range entries {
		if !hideSet[e.id] {
			visible = append(visible, e)
		}
	}

	isLongLocale := statLongLocales[o.Locale]
	shiftValuePos := 79.01
	if isLongLocale {
		shiftValuePos += 50
	}

	// Filter out hidden stats defined by user & create the text nodes.
	statItems := make([]string, 0, len(visible))
	for i, e := range visible {
		statItems = append(statItems, statTextNode(e.icon, e.label, e.value, e.display, e.id, e.unitSymbol, i, showIcons, shiftValuePos, bold, numberFormat))
	}

	if len(statItems) == 0 && hideRank {
		panic(common.NewCustomError(
			"Could not render stats card.",
			"Either stats or rank are required.",
		))
	}

	// Calculate the card height depending on how many items there are
	// but if rank circle is visible clamp the minimum height to `150`.
	height := 45 + (len(statItems)+1)*lheight
	if !hideRank {
		minH := 180
		if len(statItems) > 0 {
			minH = 150
		}
		if height < minH {
			height = minH
		}
	}

	// The lower the user's percentile the better.
	progress := 100 - float64(stats.Rank.Percentile)
	css := statCardStyles(colors.TextColor, colors.IconColor, colors.RingColor, showIcons, progress)

	var measureTitle string
	if o.CustomTitle != "" {
		measureTitle = o.CustomTitle
	} else if len(statItems) > 0 {
		measureTitle = i18n.T("statcard.title")
	} else {
		measureTitle = i18n.T("statcard.ranktitle")
	}

	// When hide_rank=true, the minimum card width is 270 px + the title length and padding.
	// When hide_rank=false, the minimum card_width is 340 px + the icon width (if show_icons=true).
	iconWidth := 0.0
	if showIcons && len(statItems) > 0 {
		iconWidth = 16 + 1 // icon + padding
	}
	var minCardWidth, defaultCardWidth float64
	if hideRank {
		minCardWidth = common.ClampValue(50+common.MeasureText(measureTitle, 10)*2, statCardMinWidth, math.Inf(1)) + iconWidth
		defaultCardWidth = statCardDefaultWidth + iconWidth
	} else if len(statItems) > 0 {
		minCardWidth = rankCardMinWidth + iconWidth
		defaultCardWidth = rankCardDefaultWidth + iconWidth
	} else {
		minCardWidth = rankOnlyCardMinWidth + iconWidth
		defaultCardWidth = rankOnlyCardDefaultWid + iconWidth
	}
	width := defaultCardWidth
	if o.CardWidth != nil && !math.IsNaN(*o.CardWidth) && *o.CardWidth != 0 {
		width = *o.CardWidth
	}
	if width < minCardWidth {
		width = minCardWidth
	}

	customTitle := o.CustomTitle
	defaultTitle := i18n.T("statcard.title")
	if len(statItems) == 0 {
		defaultTitle = i18n.T("statcard.ranktitle")
	}
	if doLowercase {
		customTitle = strings.ToLower(customTitle)
		defaultTitle = strings.ToLower(defaultTitle)
	}

	card := common.NewCard(width, float64(height), cardBorderRadius(o.BorderRadius), colors, customTitle, defaultTitle, "")

	card.SetHideBorder(hideBorder)
	card.SetHideTitle(effectiveHideTitle)
	card.SetCSS(css)

	if disableAnimations {
		card.DisableAnimations()
	}

	// Calculates the right rank circle translation values such that the rank
	// circle keeps respecting the padding rules from stats-card.js.
	calculateRankXTranslation := func() float64 {
		if len(statItems) > 0 {
			minXTranslation := float64(rankCardMinWidth) + iconWidth - 70
			if width > rankCardDefaultWidth {
				xMaxExpansion := minXTranslation + (450-minCardWidth)/2
				return xMaxExpansion + width - rankCardDefaultWidth
			}
			return minXTranslation + (width-minCardWidth)/2
		}
		return width/2 + 20 - 10
	}

	// Conditionally rendered elements.
	var rankDefs, rankCircle string
	if !hideRank {
		rankDefs = `<defs>
        <linearGradient id="rank-gradient" x1="0" y1="0" x2="1" y2="1">
          <stop offset="0%" stop-color="` + colors.RingColor + `" />
          <stop offset="100%" stop-color="#22d3ee" />
        </linearGradient>
        <radialGradient id="rank-ambient" cx="0.5" cy="0.5" r="0.5">
          <stop offset="0%" stop-color="#3b82f6" stop-opacity="0" />
          <stop offset="62%" stop-color="#3b82f6" stop-opacity="0" />
          <stop offset="100%" stop-color="#38bdf8" stop-opacity="0.14" />
        </radialGradient>
        <filter
          id="rank-glow"
          x="-40%"
          y="-40%"
          width="180%"
          height="180%"
        >
          <feDropShadow
            dx="0"
            dy="0"
            stdDeviation="3.5"
            flood-color="#22d3ee"
            flood-opacity="0.35"
          />
        </filter>
      </defs>`
		rankCircle = `<g data-testid="rank-circle"
          transform="translate(` + fmtNum(calculateRankXTranslation()) + `, ` + fmtNum(float64(height)/2-50) + `)">
        <circle cx="-10" cy="8" r="58" fill="url(#rank-ambient)" />
        <circle class="rank-circle-rim" cx="-10" cy="8" r="40" />
        <circle class="rank-circle" cx="-10" cy="8" r="40" />
        <path
          d="M -38 2 A 40 40 0 0 1 -10 -32"
          stroke="#ffffff"
          stroke-opacity="0.32"
          stroke-width="2.5"
          fill="none"
          stroke-linecap="round"
        />
        <g class="rank-text">
          ` + common.RankIcon(rankIconKind, stats.Rank.Level, float64(stats.Rank.Percentile)) + `
        </g>
      </g>`
	}

	// Accessibility labels.
	labels := make([]string, 0, len(visible))
	for _, e := range visible {
		if e.id == "commits" {
			commitsA11y := i18n.T("statcard.commits")
			if doLowercase {
				commitsA11y = strings.ToLower(commitsA11y)
			}
			yearBit := fmt.Sprintf("in %d", currentYear)
			if includeAllCommits {
				yearBit = ""
			}
			labels = append(labels, commitsA11y+" "+yearBit+" : "+e.a11yValue)
		} else {
			labels = append(labels, e.label+": "+e.a11yValue)
		}
	}

	card.SetAccessibilityLabel(card.Title+", Rank: "+stats.Rank.Level, strings.Join(labels, ", "))

	return card.Render(`
    ` + rankDefs + `
    ` + rankCircle + `
    <svg x="0" y="0">
      ` + strings.Join(common.FlexLayout(statItems, lheight, "column", nil), "") + `
    </svg>
  `)
}
