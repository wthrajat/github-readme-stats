package cards

import (
	"strings"

	"github.com/wthrajat/github-readme-stats/internal/common"
	"github.com/wthrajat/github-readme-stats/internal/fetchers"
)

// GistCardOptions mirrors the JS GistCardOptions for the gist card.
type GistCardOptions struct {
	TitleColor, IconColor, TextColor, BgColor, Theme, BorderColor string
	BorderRadius                                                  *float64
	ShowOwner, HideBorder                                         *bool
}

const (
	gistIconSize  = 16
	gistCardWidth = 400
	// HeaderMaxLength mirrors HEADER_MAX_LENGTH in gist-card.js.
	HeaderMaxLength = 35
	gistLineWidth   = 59
	gistLinesLimit  = 10
)

// RenderGistCard renders the gist card SVG.
func RenderGistCard(g *fetchers.GistData, o GistCardOptions) string {
	theme := o.Theme
	if theme == "" {
		theme = "default"
	}
	showOwner := boolOpt(o.ShowOwner)
	hideBorder := boolOpt(o.HideBorder)

	colors := common.GetCardColors(o.TitleColor, o.TextColor, o.IconColor, o.BgColor, o.BorderColor, "", theme, "default")

	desc := stringVal(g.Description)
	if desc == "" {
		desc = "No description provided"
	}
	desc = common.ParseEmojis(desc)
	multiLineDescription := common.WrapTextMultiline(desc, gistLineWidth, gistLinesLimit)
	descriptionLines := len(multiLineDescription)
	parts := make([]string, 0, len(multiLineDescription))
	for _, line := range multiLineDescription {
		parts = append(parts, `<tspan dy="1.2em" x="25">`+common.EncodeHTML(line)+`</tspan>`)
	}
	descriptionSvg := strings.Join(parts, "")

	lineHeight := 10
	if descriptionLines > 3 {
		lineHeight = 12
	}
	height := 110
	if descriptionLines > 1 {
		height = 120
	}
	height += descriptionLines * lineHeight

	totalStars := common.KFormatter(float64(g.StarsCount))
	totalForks := common.KFormatter(float64(g.ForksCount))
	svgStars := common.IconWithLabel(common.Icons["star"], totalStars, "starsCount", gistIconSize)
	svgForks := common.IconWithLabel(common.Icons["fork"], totalForks, "forksCount", gistIconSize)

	languageName := stringVal(g.Language)
	if languageName == "" {
		languageName = "Unspecified"
	}
	languageColor := languageColorFor(languageName)

	svgLanguage := common.CreateLanguageNode(languageName, languageColor)

	starAndForkCount := strings.Join(common.FlexLayout(
		[]string{svgLanguage, svgStars, svgForks},
		25,
		"",
		[]int{
			int(common.MeasureText(languageName, 12)),
			gistIconSize + int(common.MeasureText(totalStars, 12)),
			gistIconSize + int(common.MeasureText(totalForks, 12)),
		},
	), "")

	header := g.Name
	if showOwner {
		header = g.NameWithOwner
	}

	title := header
	if len([]rune(header)) > HeaderMaxLength {
		title = string([]rune(header)[:HeaderMaxLength]) + "..."
	}

	card := common.NewCard(gistCardWidth, float64(height), cardBorderRadius(o.BorderRadius), colors, "", title, common.Icons["gist"])

	card.SetCSS(`
    .description { font: 400 13px 'Segoe UI', Ubuntu, Sans-Serif; fill: ` + colors.TextColor + ` }
    .gray { font: 400 12px 'Segoe UI', Ubuntu, Sans-Serif; fill: ` + colors.TextColor + ` }
    .icon { fill: ` + colors.IconColor + ` }
  `)
	card.SetHideBorder(hideBorder)

	return card.Render(`
    <text class="description" x="25" y="-5">
        ` + descriptionSvg + `
    </text>

    <g transform="translate(30, ` + fmtNum(float64(height-75)) + `)">
        ` + starAndForkCount + `
    </g>
  `)
}
