package cards

import (
	"fmt"
	"math"
	"strings"

	"github.com/wthrajat/github-readme-stats/internal/common"
	"github.com/wthrajat/github-readme-stats/internal/fetchers"
	"github.com/wthrajat/github-readme-stats/internal/translations"
)

// WakatimeOptions mirrors the JS WakaTimeOptions for the wakatime card.
type WakatimeOptions struct {
	HideTitle, HideBorder                                                                                     *bool
	Hide                                                                                                      []string
	LineHeight                                                                                                int
	TitleColor, IconColor, TextColor, BgColor, Theme, CustomTitle, Locale, Layout, BorderColor, DisplayFormat string
	BorderRadius                                                                                              *float64
	HideProgress                                                                                              *bool
	LangsCount                                                                                                *int
	DisableAnimations                                                                                         *bool
}

// noCodingActivityNode creates the no coding activity SVG node.
func noCodingActivityNode(color, text string) string {
	return `
    <text x="25" y="11" class="stat bold" fill="` + color + `">` + text + `</text>
  `
}

// formatWakaLanguageValue formats a wakatime language value.
func formatWakaLanguageValue(displayFormat string, lang fetchers.WakaTimeLang, percent float64) string {
	if displayFormat == "percent" {
		return fmt.Sprintf("%.2f", percent) + " %"
	}
	return lang.Text
}

// wakaCompactLangNode creates a compact WakaTime language node.
func wakaCompactLangNode(lang fetchers.WakaTimeLang, x, y float64, displayFormat string, percent float64) string {
	color := languageColorFor(lang.Name)
	value := formatWakaLanguageValue(displayFormat, lang, percent)

	return `
    <g transform="translate(` + fmtNum(x) + `, ` + fmtNum(y) + `)">
      <circle cx="5" cy="6" r="5" fill="` + color + `" />
      <text data-testid="lang-name" x="15" y="10" class='lang-name'>
        ` + lang.Name + ` - ` + value + `
      </text>
    </g>
  `
}

// wakaLanguage pairs a language with its recalculated percentage.
type wakaLanguage struct {
	lang    fetchers.WakaTimeLang
	percent float64
}

// wakaLanguageTextNode creates WakaTime language text node items in two columns.
func wakaLanguageTextNode(langs []wakaLanguage, y float64, displayFormat string) []string {
	items := make([]wakaLanguage, len(langs))
	copy(items, langs)
	nodes := make([]string, 0, len(items))
	for index, item := range items {
		if index%2 == 0 {
			nodes = append(nodes, wakaCompactLangNode(item.lang, 25, 12.5*float64(index)+y, displayFormat, item.percent))
		} else {
			nodes = append(nodes, wakaCompactLangNode(item.lang, 230, 12.5+12.5*float64(index), displayFormat, item.percent))
		}
	}
	return nodes
}

// wakaTextNode creates a WakaTime text item.
func wakaTextNode(id, label, value string, index int, percent float64, hideProgress bool, progressBarColor, progressBarBackgroundColor string) string {
	staggerDelay := (index + 3) * 150

	cardProgress := ""
	if !hideProgress {
		cardProgress = common.CreateProgressNode(110, 4, 220, progressBarColor, percent, progressBarBackgroundColor, staggerDelay+300)
	}

	valueX := "350"
	if hideProgress {
		valueX = "170"
	}

	return `
    <g class="stagger" style="animation-delay: ` + fmtNum(float64(staggerDelay)) + `ms" transform="translate(25, 0)">
      <text class="stat bold" y="12.5" data-testid="` + id + `">` + label + `:</text>
      <text
        class="stat"
        x="` + valueX + `"
        y="12.5"
      >` + value + `</text>
      ` + cardProgress + `
    </g>
  `
}

// recalculateWakaPercentages recalculates percentages so that the compact
// layout's progress bar does not break when hiding languages. It returns the
// adjusted percentages aligned with the input languages.
func recalculateWakaPercentages(languages []fetchers.WakaTimeLang) []float64 {
	percents := make([]float64, len(languages))
	for i, l := range languages {
		percents[i] = float64(l.Percent)
	}
	var totalSum float64
	for _, p := range percents {
		totalSum += p
	}
	if totalSum <= 0 {
		return percents
	}
	weight := round2(100 / totalSum)
	for i, p := range percents {
		percents[i] = round2(p * weight)
	}
	return percents
}

// wakaCardStyles retrieves CSS styles for the wakatime card.
func wakaCardStyles(textColor string) string {
	return `
    .stat {
      font: 600 14px 'Segoe UI', Ubuntu, "Helvetica Neue", Sans-Serif; fill: ` + textColor + `;
    }
    @supports(-moz-appearance: auto) {
      /* Selector detects Firefox */
      .stat { font-size:12px; }
    }
    .stagger {
      opacity: 0;
      animation: fadeInAnimation 0.3s ease-in-out forwards;
    }
    .not_bold { font-weight: 400 }
    .bold { font-weight: 700 }
  `
}

// wakaNoActivityText resolves the no-activity message from visibility flags.
func wakaNoActivityText(i18n interface{ T(string) string }, codingVisible, otherVisible bool) string {
	if !codingVisible {
		return i18n.T("wakatimecard.notpublic")
	}
	if !otherVisible {
		return i18n.T("wakatimecard.nocodedetails")
	}
	return i18n.T("wakatimecard.nocodingactivity")
}

// RenderWakatimeCard renders the WakaTime card SVG.
func RenderWakatimeCard(stats *fetchers.WakaTimeData, o WakatimeOptions) string {
	var languages []fetchers.WakaTimeLang
	var codingVisible, otherVisible bool
	var rng string
	if stats != nil {
		languages = stats.Languages
		codingVisible = stats.IsCodingActivityVisible
		otherVisible = stats.IsOtherUsageVisible
		rng = stats.Range
	}

	hideTitle := boolOpt(o.HideTitle)
	hideBorder := boolOpt(o.HideBorder)
	hideProgress := boolOpt(o.HideProgress)
	disableAnimations := boolOpt(o.DisableAnimations)

	theme := o.Theme
	if theme == "" {
		theme = "default"
	}
	displayFormat := o.DisplayFormat
	if displayFormat == "" {
		displayFormat = "time"
	}

	defaultLangsCount := len(languages)
	if len(o.Hide) > 0 {
		languagesToHide := map[string]bool{}
		for _, h := range o.Hide {
			languagesToHide[common.LowercaseTrim(h)] = true
		}
		filtered := make([]fetchers.WakaTimeLang, 0, len(languages))
		for _, lang := range languages {
			if !languagesToHide[common.LowercaseTrim(lang.Name)] {
				filtered = append(filtered, lang)
			}
		}
		languages = filtered
	}

	langsCountOpt := defaultLangsCount
	if o.LangsCount != nil {
		langsCountOpt = *o.LangsCount
	}
	// Since the percentages are sorted in descending order, we can just
	// slice from the beginning without sorting.
	if langsCountOpt < 0 {
		langsCountOpt = 0
	}
	if langsCountOpt < len(languages) {
		languages = languages[:langsCountOpt]
	}
	percents := recalculateWakaPercentages(languages)
	paired := make([]wakaLanguage, len(languages))
	for i, lang := range languages {
		paired[i] = wakaLanguage{lang: lang, percent: percents[i]}
	}

	i18n := common.NewI18n(o.Locale, translations.WakatimeCardLocales)

	lheight := o.LineHeight
	if lheight == 0 {
		lheight = 25
	}

	langsCount := int(common.ClampValue(float64(langsCountOpt), 1, float64(langsCountOpt)))

	colors := common.GetCardColors(o.TitleColor, o.TextColor, o.IconColor, o.BgColor, o.BorderColor, "", theme, "default")

	filtered := make([]wakaLanguage, 0, len(paired))
	for _, item := range paired {
		if float64(item.lang.Hours) != 0 || float64(item.lang.Minutes) != 0 {
			filtered = append(filtered, item)
		}
	}
	if langsCount < len(filtered) {
		filtered = filtered[:langsCount]
	}

	// Calculate the card height depending on how many items there are
	// but clamp the minimum height to `150`.
	height := 45 + (len(filtered)+1)*lheight
	if height < 150 {
		height = 150
	}

	cssStyles := wakaCardStyles(colors.TextColor)

	var finalLayout string

	width := 440.0

	if o.Layout == "compact" {
		width += 50
		height = 90 + int(math.Round(float64(len(filtered))/2))*25

		// progressOffset holds the previous language's width and is used to
		// offset the next language so that we can stack them one after another.
		progressOffset := 0.0
		bars := make([]string, 0, len(filtered))
		for _, item := range filtered {
			progress := (width - 25) * item.percent / 100

			languageColor := languageColorFor(item.lang.Name)

			bars = append(bars, `
          <rect
            mask="url(#rect-mask)"
            data-testid="lang-progress"
            x="`+fmtNum(progressOffset)+`"
            y="0"
            width="`+fmtNum(progress)+`"
            height="8"
            fill="`+languageColor+`"
          />
        `)
			progressOffset += progress
		}
		compactProgressBar := strings.Join(bars, "")

		body := ""
		if len(filtered) > 0 {
			body = strings.Join(wakaLanguageTextNode(filtered, 25, displayFormat), "")
		} else {
			body = noCodingActivityNode(colors.TextColor, wakaNoActivityText(i18n, codingVisible, otherVisible))
		}
		finalLayout = `
      <mask id="rect-mask">
      <rect x="25" y="0" width="` + fmtNum(width-50) + `" height="8" fill="white" rx="5" />
      </mask>
      ` + compactProgressBar + `
      ` + body + `
    `
	} else {
		var items []string
		if len(filtered) > 0 {
			items = make([]string, 0, len(filtered))
			for index, item := range filtered {
				items = append(items, wakaTextNode(
					item.lang.Name,
					item.lang.Name,
					formatWakaLanguageValue(displayFormat, item.lang, item.percent),
					index,
					item.percent,
					hideProgress,
					colors.TitleColor,
					colors.TextColor,
				))
			}
		} else {
			items = []string{noCodingActivityNode(colors.TextColor, wakaNoActivityText(i18n, codingVisible, otherVisible))}
		}
		finalLayout = strings.Join(common.FlexLayout(items, lheight, "column", nil), "")
	}

	// Get title range text.
	titleText := i18n.T("wakatimecard.title")
	switch rng {
	case "last_7_days":
		titleText += ` (` + i18n.T("wakatimecard.last7days") + `)`
	case "last_year":
		titleText += ` (` + i18n.T("wakatimecard.lastyear") + `)`
	}

	card := common.NewCard(495, float64(height), cardBorderRadius(o.BorderRadius), colors, o.CustomTitle, titleText, "")

	if disableAnimations {
		card.DisableAnimations()
	}

	card.SetHideBorder(hideBorder)
	card.SetHideTitle(hideTitle)
	card.SetCSS(
		`
    ` + cssStyles + `
    @keyframes slideInAnimation {
      from {
        width: 0;
      }
      to {
        width: calc(100%-100px);
      }
    }
    @keyframes growWidthAnimation {
      from {
        width: 0;
      }
      to {
        width: 100%;
      }
    }
    .lang-name { font: 400 11px 'Segoe UI', Ubuntu, Sans-Serif; fill: ` + colors.TextColor + ` }
    #rect-mask rect{
      animation: slideInAnimation 1s ease-in-out forwards;
    }
    .lang-progress{
      animation: growWidthAnimation 0.6s ease-in-out forwards;
    }
    `,
	)

	return card.Render(`
    <svg x="0" y="0" width="100%">
      ` + finalLayout + `
    </svg>
  `)
}
