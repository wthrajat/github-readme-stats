package cards

import (
	"strings"

	"github.com/wthrajat/github-readme-stats/internal/common"
	"github.com/wthrajat/github-readme-stats/internal/fetchers"
	"github.com/wthrajat/github-readme-stats/internal/translations"
)

// RepoCardOptions mirrors the JS RepoCardOptions for the repo card.
type RepoCardOptions struct {
	HideBorder                                                            *bool
	TitleColor, IconColor, TextColor, BgColor, Theme, BorderColor, Locale string
	BorderRadius                                                          *float64
	ShowOwner                                                             *bool
	DescriptionLinesCount                                                 *int
}

const (
	repoIconSize             = 16
	repoDescriptionLineWidth = 59
	repoDescriptionMaxLines  = 3
)

// getBadgeSVG retrieves the repository badge (template/archived) SVG.
func getBadgeSVG(label, textColor string) string {
	return `
  <g data-testid="badge" class="badge" transform="translate(320, -18)">
    <rect stroke="` + textColor + `" stroke-width="1" width="70" height="20" x="-12" y="-14" ry="10" rx="10"></rect>
    <text
      x="23" y="-5"
      alignment-baseline="central"
      dominant-baseline="central"
      text-anchor="middle"
      fill="` + textColor + `"
    >
      ` + label + `
    </text>
  </g>
`
}

// RenderRepoCard renders the repository card SVG.
func RenderRepoCard(repo *fetchers.RepositoryData, o RepoCardOptions) string {
	theme := o.Theme
	if theme == "" {
		theme = "default_repocard"
	}
	hideBorder := boolOpt(o.HideBorder)
	showOwner := boolOpt(o.ShowOwner)

	header := repo.Name
	if showOwner {
		header = repo.NameWithOwner
	}
	langName := "Unspecified"
	langColor := "#333"
	hasLang := false
	if repo.PrimaryLanguage != nil {
		hasLang = true
		if repo.PrimaryLanguage.Name != "" {
			langName = repo.PrimaryLanguage.Name
		}
		if repo.PrimaryLanguage.Color != "" {
			langColor = repo.PrimaryLanguage.Color
		}
	}

	descriptionMaxLines := repoDescriptionMaxLines
	if o.DescriptionLinesCount != nil {
		descriptionMaxLines = int(common.ClampValue(float64(*o.DescriptionLinesCount), 1, repoDescriptionMaxLines))
	}

	desc := stringVal(repo.Description)
	if desc == "" {
		desc = "No description provided"
	}
	desc = common.ParseEmojis(desc)
	multiLineDescription := common.WrapTextMultiline(desc, repoDescriptionLineWidth, descriptionMaxLines)
	descriptionLinesCount := len(multiLineDescription)
	if o.DescriptionLinesCount != nil {
		descriptionLinesCount = int(common.ClampValue(float64(*o.DescriptionLinesCount), 1, repoDescriptionMaxLines))
	}

	parts := make([]string, 0, len(multiLineDescription))
	for _, line := range multiLineDescription {
		parts = append(parts, `<tspan dy="1.2em" x="25">`+common.EncodeHTML(line)+`</tspan>`)
	}
	descriptionSvg := strings.Join(parts, "")

	height := 110
	if descriptionLinesCount > 1 {
		height = 120
	}
	height += descriptionLinesCount * 10

	i18n := common.NewI18n(o.Locale, translations.RepoCardLocales)

	colors := common.GetCardColors(o.TitleColor, o.TextColor, o.IconColor, o.BgColor, o.BorderColor, "", theme, "default")

	svgLanguage := ""
	if hasLang {
		svgLanguage = common.CreateLanguageNode(langName, langColor)
	}

	totalStars := common.KFormatter(float64(repo.StarCount))
	totalForks := common.KFormatter(float64(repo.ForkCount))
	svgStars := common.IconWithLabel(common.Icons["star"], totalStars, "stargazers", repoIconSize)
	svgForks := common.IconWithLabel(common.Icons["fork"], totalForks, "forkcount", repoIconSize)

	starAndForkCount := strings.Join(common.FlexLayout(
		[]string{svgLanguage, svgStars, svgForks},
		25,
		"",
		[]int{
			int(common.MeasureText(langName, 12)),
			repoIconSize + int(common.MeasureText(totalStars, 12)),
			repoIconSize + int(common.MeasureText(totalForks, 12)),
		},
	), "")

	title := header
	if len([]rune(header)) > 35 {
		title = string([]rune(header)[:35]) + "..."
	}

	card := common.NewCard(400, float64(height), cardBorderRadius(o.BorderRadius), colors, "", title, common.Icons["contribs"])

	card.DisableAnimations()
	card.SetHideBorder(hideBorder)
	card.SetHideTitle(false)
	card.SetCSS(`
    .description { font: 400 13px 'Segoe UI', Ubuntu, Sans-Serif; fill: ` + colors.TextColor + ` }
    .gray { font: 400 12px 'Segoe UI', Ubuntu, Sans-Serif; fill: ` + colors.TextColor + ` }
    .icon { fill: ` + colors.IconColor + ` }
    .badge { font: 600 11px 'Segoe UI', Ubuntu, Sans-Serif; }
    .badge rect { opacity: 0.2 }
  `)

	badge := ""
	if repo.IsTemplate {
		badge = getBadgeSVG(i18n.T("repocard.template"), colors.TextColor)
	} else if repo.IsArchived {
		badge = getBadgeSVG(i18n.T("repocard.archived"), colors.TextColor)
	}

	return card.Render(`
    ` + badge + `

    <text class="description" x="25" y="-5">
      ` + descriptionSvg + `
    </text>

    <g transform="translate(30, ` + fmtNum(float64(height-75)) + `)">
      ` + starAndForkCount + `
    </g>
  `)
}
