package cards

import (
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/wthrajat/github-readme-stats/internal/common"
	"github.com/wthrajat/github-readme-stats/internal/fetchers"
	"github.com/wthrajat/github-readme-stats/internal/translations"
)

// TopLangOptions mirrors the JS TopLangOptions for the top languages card.
type TopLangOptions struct {
	HideTitle, HideBorder                                                           *bool
	CardWidth                                                                       *float64
	TitleColor, TextColor, BgColor, Theme, CustomTitle, Locale, Layout, BorderColor string
	BorderRadius                                                                    *float64
	Hide                                                                            []string
	HideProgress                                                                    *bool
	LangsCount                                                                      *int
	DisableAnimations                                                               *bool
}

const (
	topLangDefaultCardWidth = 300
	// MinCardWidth is the minimum width of the top languages card.
	MinCardWidth          = 280
	topLangDefaultColor   = "#858585"
	topLangCardPadding    = 25
	compactLayoutBaseH    = 90
	maximumLangsCount     = 20
	normalDefaultLangs    = 5
	compactDefaultLangs   = 6
	donutDefaultLangs     = 5
	pieDefaultLangs       = 6
	donutVerticalDefaults = 6
)

// Point is a cartesian coordinate pair.
type Point struct {
	X, Y float64
}

// Polar is a polar coordinate pair.
type Polar struct {
	Radius, AngleInDegrees float64
}

// GetLongestLang retrieves the programming language whose name is the longest.
func GetLongestLang(langs []*fetchers.Lang) *fetchers.Lang {
	var longest *fetchers.Lang
	for _, l := range langs {
		if longest == nil || len(l.Name) > len(longest.Name) {
			longest = l
		}
	}
	return longest
}

// DegreesToRadians converts degrees to radians.
func DegreesToRadians(angleInDegrees float64) float64 {
	return angleInDegrees * (math.Pi / 180.0)
}

// RadiansToDegrees converts radians to degrees.
func RadiansToDegrees(angleInRadians float64) float64 {
	return angleInRadians / (math.Pi / 180.0)
}

// PolarToCartesian converts polar coordinates to cartesian coordinates.
func PolarToCartesian(centerX, centerY, radius, angleInDegrees float64) Point {
	rads := DegreesToRadians(angleInDegrees)
	return Point{
		X: centerX + radius*math.Cos(rads),
		Y: centerY + radius*math.Sin(rads),
	}
}

// CartesianToPolar converts cartesian coordinates to polar coordinates.
func CartesianToPolar(centerX, centerY, x, y float64) Polar {
	radius := math.Sqrt(math.Pow(x-centerX, 2) + math.Pow(y-centerY, 2))
	angleInDegrees := RadiansToDegrees(math.Atan2(y-centerY, x-centerX))
	if angleInDegrees < 0 {
		angleInDegrees += 360
	}
	return Polar{Radius: radius, AngleInDegrees: angleInDegrees}
}

// GetCircleLength calculates the length of a circle.
func GetCircleLength(radius float64) float64 {
	return 2 * math.Pi * radius
}

// CalculateCompactLayoutHeight calculates height for the compact layout.
func CalculateCompactLayoutHeight(totalLangs int) int {
	return compactLayoutBaseH + int(math.Round(float64(totalLangs)/2))*25
}

// CalculateNormalLayoutHeight calculates height for the normal layout.
func CalculateNormalLayoutHeight(totalLangs int) int {
	return 45 + (totalLangs+1)*40
}

// CalculateDonutLayoutHeight calculates height for the donut layout.
func CalculateDonutLayoutHeight(totalLangs int) int {
	extra := totalLangs - 5
	if extra < 0 {
		extra = 0
	}
	return 215 + extra*32
}

// CalculateDonutVerticalLayoutHeight calculates height for the donut vertical layout.
func CalculateDonutVerticalLayoutHeight(totalLangs int) int {
	return 300 + int(math.Round(float64(totalLangs)/2))*25
}

// CalculatePieLayoutHeight calculates height for the pie layout.
func CalculatePieLayoutHeight(totalLangs int) int {
	return 300 + int(math.Round(float64(totalLangs)/2))*25
}

// DonutCenterTranslation calculates the center translation needed to keep the
// donut chart centred.
func DonutCenterTranslation(totalLangs int) int {
	extra := totalLangs - 5
	if extra < 0 {
		extra = 0
	}
	return -45 + extra*16
}

// TrimTopLanguages trims top languages to langsCount while also hiding
// certain languages. It returns the trimmed languages and their total size.
func TrimTopLanguages(topLangs fetchers.TopLangData, langsCount int, hide []string) ([]*fetchers.Lang, float64) {
	count := int(common.ClampValue(float64(langsCount), 1, maximumLangsCount))

	langsToHide := map[string]bool{}
	for _, h := range hide {
		langsToHide[common.LowercaseTrim(h)] = true
	}

	langs := make([]*fetchers.Lang, 0, len(topLangs))
	for _, lang := range topLangs {
		if langsToHide[common.LowercaseTrim(lang.Name)] {
			continue
		}
		langs = append(langs, lang)
	}
	sort.Slice(langs, func(i, j int) bool {
		return float64(langs[i].Size) > float64(langs[j].Size)
	})
	if len(langs) > count {
		langs = langs[:count]
	}

	var totalLanguageSize float64
	for _, l := range langs {
		totalLanguageSize += float64(l.Size)
	}

	return langs, totalLanguageSize
}

// GetDefaultLanguagesCountByLayout returns the default languages count for
// the provided card layout.
func GetDefaultLanguagesCountByLayout(layout string, hideProgress bool) int {
	if layout == "compact" || hideProgress {
		return compactDefaultLangs
	} else if layout == "donut" {
		return donutDefaultLangs
	} else if layout == "donut-vertical" {
		return donutVerticalDefaults
	} else if layout == "pie" {
		return pieDefaultLangs
	}
	return normalDefaultLangs
}

// langColor returns the display color of a language, defaulting to gray.
func langColor(color string) string {
	if color == "" {
		return topLangDefaultColor
	}
	return color
}

// langProgressTextNode creates a progress bar text item for a language.
func langProgressTextNode(width float64, color, name string, progress float64, index int) string {
	staggerDelay := (index + 3) * 150
	paddingRight := 95.0
	progressTextX := width - paddingRight + 10
	progressWidth := width - paddingRight

	return `
    <g class="stagger" style="animation-delay: ` + fmtNum(float64(staggerDelay)) + `ms">
      <text data-testid="lang-name" x="2" y="15" class="lang-name">` + name + `</text>
      <text x="` + fmtNum(progressTextX) + `" y="34" class="lang-name">` + fmtNum(progress) + `%</text>
      ` + common.CreateProgressNode(0, 25, progressWidth, color, progress, "#ddd", staggerDelay+300) + `
    </g>
  `
}

// compactLangNode creates a compact text item for a language.
func compactLangNode(lang *fetchers.Lang, totalSize float64, hideProgress bool, index int) string {
	percentage := fmt.Sprintf("%.2f", float64(lang.Size)/totalSize*100)
	staggerDelay := (index + 3) * 150
	color := langColor(lang.Color)

	progress := " " + percentage + "%"
	if hideProgress {
		progress = ""
	}

	return `
    <g class="stagger" style="animation-delay: ` + fmtNum(float64(staggerDelay)) + `ms">
      <circle cx="5" cy="6" r="5" fill="` + color + `" />
      <text data-testid="lang-name" x="15" y="10" class='lang-name'>
        ` + lang.Name + progress + `
      </text>
    </g>
  `
}

// chunkLangs splits languages into columns, replicating the JS chunkArray
// behavior including fractional per-chunk sizes.
func chunkLangs(langs []*fetchers.Lang, perChunk float64) [][]*fetchers.Lang {
	var result [][]*fetchers.Lang
	for i, lang := range langs {
		chunkIndex := 0
		if perChunk != 0 {
			chunkIndex = int(math.Floor(float64(i) / perChunk))
		}
		for len(result) <= chunkIndex {
			result = append(result, []*fetchers.Lang{})
		}
		result[chunkIndex] = append(result[chunkIndex], lang)
	}
	return result
}

// languageTextNode creates compact language text items for all languages.
func languageTextNode(langs []*fetchers.Lang, totalSize float64, hideProgress bool) string {
	longestLang := GetLongestLang(langs)
	chunked := chunkLangs(langs, float64(len(langs))/2)
	layouts := make([]string, 0, len(chunked))
	for _, array := range chunked {
		items := make([]string, 0, len(array))
		for i, lang := range array {
			items = append(items, compactLangNode(lang, totalSize, hideProgress, i))
		}
		layouts = append(layouts, strings.Join(common.FlexLayout(items, 25, "column", nil), ""))
	}

	percent := fmt.Sprintf("%.2f", float64(longestLang.Size)/totalSize*100)
	minGap := 150.0
	maxGap := 20 + common.MeasureText(longestLang.Name+" "+percent+"%", 11)
	gap := maxGap
	if gap < minGap {
		gap = minGap
	}
	return strings.Join(common.FlexLayout(layouts, int(gap), "", nil), "")
}

// donutLanguagesNode creates donut language text items for all languages.
func donutLanguagesNode(langs []*fetchers.Lang, totalSize float64) string {
	items := make([]string, 0, len(langs))
	for i, lang := range langs {
		items = append(items, compactLangNode(lang, totalSize, false, i))
	}
	return strings.Join(common.FlexLayout(items, 32, "column", nil), "")
}

// RenderNormalLayout renders the default language card layout.
func RenderNormalLayout(langs []*fetchers.Lang, width, totalLanguageSize float64) string {
	items := make([]string, 0, len(langs))
	for i, lang := range langs {
		items = append(items, langProgressTextNode(
			width,
			langColor(lang.Color),
			lang.Name,
			round2(float64(lang.Size)/totalLanguageSize*100),
			i,
		))
	}
	return strings.Join(common.FlexLayout(items, 40, "column", nil), "")
}

// RenderCompactLayout renders the compact language card layout.
func RenderCompactLayout(langs []*fetchers.Lang, width, totalLanguageSize float64, hideProgress bool) string {
	paddingRight := 50.0
	offsetWidth := width - paddingRight
	// progressOffset holds the previous language's width and is used to offset
	// the next language so that we can stack them one after another.
	progressOffset := 0.0
	parts := make([]string, 0, len(langs))
	for _, lang := range langs {
		percentage := round2(float64(lang.Size) / totalLanguageSize * offsetWidth)

		progress := percentage
		if percentage < 10 {
			progress = percentage + 10
		}

		parts = append(parts, `
        <rect
          mask="url(#rect-mask)"
          data-testid="lang-progress"
          x="`+fmtNum(progressOffset)+`"
          y="0"
          width="`+fmtNum(progress)+`"
          height="8"
          fill="`+langColor(lang.Color)+`"
        />
      `)
		progressOffset += percentage
	}
	compactProgressBar := strings.Join(parts, "")

	mask := ""
	if !hideProgress {
		mask = `
      <mask id="rect-mask">
          <rect x="0" y="0" width="` + fmtNum(offsetWidth) + `" height="8" fill="white" rx="5"/>
        </mask>
        ` + compactProgressBar + `
      `
	}
	translate := "25"
	if hideProgress {
		translate = "0"
	}

	return `
  ` + mask + `
    <g transform="translate(0, ` + translate + `)">
      ` + languageTextNode(langs, totalLanguageSize, hideProgress) + `
    </g>
  `
}

// RenderDonutVerticalLayout renders the donut vertical layout.
func RenderDonutVerticalLayout(langs []*fetchers.Lang, totalLanguageSize float64) string {
	radius := 80.0
	totalCircleLength := GetCircleLength(radius)

	circles := make([]string, 0, len(langs))
	indent := 0.0
	startDelayCoefficient := 1
	for _, lang := range langs {
		percentage := float64(lang.Size) / totalLanguageSize * 100
		circleLength := totalCircleLength * (percentage / 100)
		delay := startDelayCoefficient * 100

		circles = append(circles, `
      <g class="stagger" style="animation-delay: `+fmtNum(float64(delay))+`ms">
        <circle
          cx="150"
          cy="100"
          r="`+fmtNum(radius)+`"
          fill="transparent"
          stroke="`+lang.Color+`"
          stroke-width="25"
          stroke-dasharray="`+fmtNum(totalCircleLength)+`"
          stroke-dashoffset="`+fmtNum(indent)+`"
          size="`+fmtNum(percentage)+`"
          data-testid="lang-donut"
        />
      </g>
    `)

		indent += circleLength
		startDelayCoefficient++
	}

	return `
    <svg data-testid="lang-items">
      <g transform="translate(0, 0)">
        <svg data-testid="donut">
          ` + strings.Join(circles, "") + `
        </svg>
      </g>
      <g transform="translate(0, 220)">
        <svg data-testid="lang-names" x="` + fmtNum(topLangCardPadding) + `">
          ` + languageTextNode(langs, totalLanguageSize, false) + `
        </svg>
      </g>
    </svg>
  `
}

// RenderPieLayout renders the pie layout.
func RenderPieLayout(langs []*fetchers.Lang, totalLanguageSize float64) string {
	radius := 90.0
	centerX := 150.0
	centerY := 100.0

	startAngle := 0.0
	startDelayCoefficient := 1

	paths := make([]string, 0, len(langs))
	for _, lang := range langs {
		if len(langs) == 1 {
			paths = append(paths, `
        <circle
          cx="`+fmtNum(centerX)+`"
          cy="`+fmtNum(centerY)+`"
          r="`+fmtNum(radius)+`"
          stroke="none"
          fill="`+lang.Color+`"
          data-testid="lang-pie"
          size="100"
        />
      `)
			break
		}

		langSizePart := float64(lang.Size) / totalLanguageSize
		percentage := langSizePart * 100
		angle := langSizePart * 360
		endAngle := startAngle + angle

		startPoint := PolarToCartesian(centerX, centerY, radius, startAngle)
		endPoint := PolarToCartesian(centerX, centerY, radius, endAngle)

		largeArcFlag := 0
		if angle > 180 {
			largeArcFlag = 1
		}
		delay := startDelayCoefficient * 100

		paths = append(paths, `
      <g class="stagger" style="animation-delay: `+fmtNum(float64(delay))+`ms">
        <path
          data-testid="lang-pie"
          size="`+fmtNum(percentage)+`"
          d="M `+fmtNum(centerX)+` `+fmtNum(centerY)+` L `+fmtNum(startPoint.X)+` `+fmtNum(startPoint.Y)+` A `+fmtNum(radius)+` `+fmtNum(radius)+` 0 `+fmtNum(float64(largeArcFlag))+` 1 `+fmtNum(endPoint.X)+` `+fmtNum(endPoint.Y)+` Z"
          fill="`+lang.Color+`"
        />
      </g>
    `)

		startAngle = endAngle
		startDelayCoefficient++
	}

	return `
    <svg data-testid="lang-items">
      <g transform="translate(0, 0)">
        <svg data-testid="pie">
          ` + strings.Join(paths, "") + `
        </svg>
      </g>
      <g transform="translate(0, 220)">
        <svg data-testid="lang-names" x="` + fmtNum(topLangCardPadding) + `">
          ` + languageTextNode(langs, totalLanguageSize, false) + `
        </svg>
      </g>
    </svg>
  `
}

// donutSection is a single arc of the donut chart.
type donutSection struct {
	d       string
	percent float64
}

// createDonutPaths creates the SVG paths for the language donut chart.
func createDonutPaths(cx, cy, radius float64, percentages []float64) []donutSection {
	paths := make([]donutSection, 0, len(percentages))
	startAngle := 0.0
	endAngle := 0.0

	var totalPercent float64
	for _, p := range percentages {
		totalPercent += p
	}
	for _, p := range percentages {
		percent := round2(p / totalPercent * 100)

		endAngle = 3.6*percent + startAngle
		startPoint := PolarToCartesian(cx, cy, radius, endAngle-90)
		endPoint := PolarToCartesian(cx, cy, radius, startAngle-90)
		largeArc := 0
		if endAngle-startAngle > 180 {
			largeArc = 1
		}

		paths = append(paths, donutSection{
			percent: percent,
			d:       `M ` + fmtNum(startPoint.X) + ` ` + fmtNum(startPoint.Y) + ` A ` + fmtNum(radius) + ` ` + fmtNum(radius) + ` 0 ` + fmtNum(float64(largeArc)) + ` 0 ` + fmtNum(endPoint.X) + ` ` + fmtNum(endPoint.Y),
		})
		startAngle = endAngle
	}

	return paths
}

// RenderDonutLayout renders the donut language card layout.
func RenderDonutLayout(langs []*fetchers.Lang, width, totalLanguageSize float64) string {
	centerX := width / 3
	centerY := width / 3
	radius := centerX - 60
	strokeWidth := 12

	colors := make([]string, 0, len(langs))
	langsPercents := make([]float64, 0, len(langs))
	for _, lang := range langs {
		colors = append(colors, lang.Color)
		langsPercents = append(langsPercents, round2(float64(lang.Size)/totalLanguageSize*100))
	}

	langPaths := createDonutPaths(centerX, centerY, radius, langsPercents)

	var donutPaths string
	if len(langs) == 1 {
		donutPaths = `<circle cx="` + fmtNum(centerX) + `" cy="` + fmtNum(centerY) + `" r="` + fmtNum(radius) + `" stroke="` + colors[0] + `" fill="none" stroke-width="` + fmtNum(float64(strokeWidth)) + `" data-testid="lang-donut" size="100"/>`
	} else {
		parts := make([]string, 0, len(langPaths))
		for index, section := range langPaths {
			staggerDelay := (index + 3) * 100
			delay := staggerDelay + 300

			parts = append(parts, `
       <g class="stagger" style="animation-delay: `+fmtNum(float64(delay))+`ms">
        <path
          data-testid="lang-donut"
          size="`+fmtNum(section.percent)+`"
          d="`+section.d+`"
          stroke="`+colors[index]+`"
          fill="none"
          stroke-width="`+fmtNum(float64(strokeWidth))+`">
        </path>
      </g>
      `)
		}
		donutPaths = strings.Join(parts, "")
	}

	donut := `<svg width="` + fmtNum(width) + `" height="` + fmtNum(width) + `">` + donutPaths + `</svg>`

	return `
    <g transform="translate(0, 0)">
      <g transform="translate(0, 0)">
        ` + donutLanguagesNode(langs, totalLanguageSize) + `
      </g>

      <g transform="translate(125, ` + fmtNum(float64(DonutCenterTranslation(len(langs)))) + `)">
        ` + donut + `
      </g>
    </g>
  `
}

// NoLanguagesDataNode creates the no languages data SVG node.
func NoLanguagesDataNode(color, text, layout string) string {
	x := "0"
	if layout == "pie" || layout == "donut-vertical" {
		x = fmtNum(topLangCardPadding)
	}
	return `
    <text x="` + x + `" y="11" class="stat bold" fill="` + color + `">` + text + `</text>
  `
}

// RenderTopLanguages renders the top languages card SVG.
func RenderTopLanguages(data fetchers.TopLangData, o TopLangOptions) string {
	hideTitle := boolOpt(o.HideTitle)
	hideBorder := boolOpt(o.HideBorder)
	hideProgress := boolOpt(o.HideProgress)
	disableAnimations := boolOpt(o.DisableAnimations)

	theme := o.Theme
	if theme == "" {
		theme = "default"
	}

	langsCount := GetDefaultLanguagesCountByLayout(o.Layout, hideProgress)
	if o.LangsCount != nil {
		langsCount = *o.LangsCount
	}

	i18n := common.NewI18n(o.Locale, translations.LangCardLocales)

	langs, totalLanguageSize := TrimTopLanguages(data, langsCount, o.Hide)

	width := float64(topLangDefaultCardWidth)
	if o.CardWidth != nil {
		if math.IsNaN(*o.CardWidth) {
			width = topLangDefaultCardWidth
		} else if *o.CardWidth < MinCardWidth {
			width = MinCardWidth
		} else {
			width = *o.CardWidth
		}
	}
	height := CalculateNormalLayoutHeight(len(langs))

	colors := common.GetCardColors(o.TitleColor, o.TextColor, "", o.BgColor, o.BorderColor, "", theme, "default")

	var finalLayout string
	if len(langs) == 0 {
		height = compactLayoutBaseH
		finalLayout = NoLanguagesDataNode(colors.TextColor, i18n.T("langcard.nodata"), o.Layout)
	} else if o.Layout == "pie" {
		height = CalculatePieLayoutHeight(len(langs))
		finalLayout = RenderPieLayout(langs, totalLanguageSize)
	} else if o.Layout == "donut-vertical" {
		height = CalculateDonutVerticalLayoutHeight(len(langs))
		finalLayout = RenderDonutVerticalLayout(langs, totalLanguageSize)
	} else if o.Layout == "compact" || hideProgress {
		height = CalculateCompactLayoutHeight(len(langs))
		if hideProgress {
			height -= 25
		}
		finalLayout = RenderCompactLayout(langs, width, totalLanguageSize, hideProgress)
	} else if o.Layout == "donut" {
		height = CalculateDonutLayoutHeight(len(langs))
		width += 50 // padding
		finalLayout = RenderDonutLayout(langs, width, totalLanguageSize)
	} else {
		finalLayout = RenderNormalLayout(langs, width, totalLanguageSize)
	}

	card := common.NewCard(width, float64(height), cardBorderRadius(o.BorderRadius), colors, o.CustomTitle, i18n.T("langcard.title"), "")

	if disableAnimations {
		card.DisableAnimations()
	}

	card.SetHideBorder(hideBorder)
	card.SetHideTitle(hideTitle)
	card.SetCSS(
		`
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
    .stat {
      font: 600 14px 'Segoe UI', Ubuntu, "Helvetica Neue", Sans-Serif; fill: ` + colors.TextColor + `;
    }
    @supports(-moz-appearance: auto) {
      /* Selector detects Firefox */
      .stat { font-size:12px; }
    }
    .bold { font-weight: 700 }
    .lang-name {
      font: 400 11px "Segoe UI", Ubuntu, Sans-Serif;
      fill: ` + colors.TextColor + `;
    }
    .stagger {
      opacity: 0;
      animation: fadeInAnimation 0.3s ease-in-out forwards;
    }
    #rect-mask rect{
      animation: slideInAnimation 1s ease-in-out forwards;
    }
    .lang-progress{
      animation: growWidthAnimation 0.6s ease-in-out forwards;
    }
    `,
	)

	if o.Layout == "pie" || o.Layout == "donut-vertical" {
		return card.Render(finalLayout)
	}

	return card.Render(`
    <svg data-testid="lang-items" x="` + fmtNum(topLangCardPadding) + `">
      ` + finalLayout + `
    </svg>
  `)
}
