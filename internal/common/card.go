package common

import (
	"os"
	"strconv"
	"strings"
)

// Card renders an SVG stats card.
// Ported from src/common/Card.js.
type Card struct {
	Width           float64
	Height          float64
	BorderRadius    float64
	HideBorder      bool
	HideTitle       bool
	Colors          CardColors
	Title           string
	CSS             string
	TitlePrefixIcon string
	PaddingX        float64
	PaddingY        float64
	Animations      bool
	A11yTitle       string
	A11yDesc        string
}

// NewCard creates a Card. An empty customTitle selects the defaultTitle.
func NewCard(width, height, borderRadius float64, colors CardColors, customTitle, defaultTitle, titlePrefixIcon string) *Card {
	title := customTitle
	if title == "" {
		title = defaultTitle
	}
	return &Card{
		Width:           width,
		Height:          height,
		BorderRadius:    borderRadius,
		HideBorder:      false,
		HideTitle:       false,
		Colors:          colors,
		Title:           EncodeHTML(title),
		CSS:             "",
		PaddingX:        25,
		PaddingY:        35,
		TitlePrefixIcon: titlePrefixIcon,
		Animations:      true,
		A11yTitle:       "",
		A11yDesc:        "",
	}
}

// DisableAnimations turns off card animations.
func (c *Card) DisableAnimations() {
	c.Animations = false
}

// SetAccessibilityLabel sets the a11y title and description.
func (c *Card) SetAccessibilityLabel(title, desc string) {
	c.A11yTitle = title
	c.A11yDesc = desc
}

// SetCSS sets additional card CSS.
func (c *Card) SetCSS(v string) {
	c.CSS = v
}

// SetHideBorder toggles the card border.
func (c *Card) SetHideBorder(v bool) {
	c.HideBorder = v
}

// SetHideTitle toggles the title, shrinking the card height by 30.
func (c *Card) SetHideTitle(v bool) {
	c.HideTitle = v
	if v {
		c.Height -= 30
	}
}

// SetTitle sets the card title (stored raw, like the JS implementation).
func (c *Card) SetTitle(t string) {
	c.Title = t
}

func fmtNum(v float64) string {
	return strconv.FormatFloat(v, 'f', -1, 64)
}

// RenderTitle renders the card title group.
func (c *Card) RenderTitle() string {
	titleText := `
      <text
        x="0"
        y="0"
        class="header"
        data-testid="header"
      >` + c.Title + `</text>
    `
	prefixIcon := `
      <svg
        class="icon"
        x="0"
        y="-13"
        viewBox="0 0 16 16"
        version="1.1"
        width="16"
        height="16"
      >
        ` + c.TitlePrefixIcon + `
      </svg>
    `
	items := []string{"", titleText}
	if c.TitlePrefixIcon != "" {
		items[0] = prefixIcon
	}
	return `
      <g
        data-testid="card-title"
        transform="translate(` + fmtNum(c.PaddingX) + `, ` + fmtNum(c.PaddingY) + `)"
      >
        ` + strings.Join(FlexLayout(items, 25, "", nil), "") + `
      </g>
    `
}

// RenderGradient renders the background gradient defs, or "" for plain colors.
func (c *Card) RenderGradient() string {
	if !c.Colors.IsGradient {
		return ""
	}
	gradients := c.Colors.BgGradient[1:]
	stops := make([]string, 0, len(gradients))
	for i, grad := range gradients {
		var offset float64
		if len(gradients) > 1 {
			offset = float64(i*100) / float64(len(gradients)-1)
		}
		stops = append(stops, `<stop offset="`+fmtNum(offset)+`%" stop-color="#`+grad+`" />`)
	}
	return `
        <defs>
          <linearGradient
            id="gradient"
            gradientTransform="rotate(` + c.Colors.BgGradient[0] + `)"
            gradientUnits="userSpaceOnUse"
          >
            ` + strings.Join(stops, "") + `
          </linearGradient>
        </defs>
        `
}

// RenderEffectDefs renders SVG defs for the glass effects.
func (c *Card) RenderEffectDefs() string {
	return `
      <defs>
        <clipPath id="card-clip">
          <rect
            x="0.5"
            y="0.5"
            rx="` + fmtNum(c.BorderRadius) + `"
            height="99%"
            width="` + fmtNum(c.Width-1) + `"
          />
        </clipPath>
        <linearGradient id="card-depth" x1="0" y1="0" x2="1" y2="1">
          <stop offset="0%" stop-color="#1d4ed8" stop-opacity="0.16" />
          <stop offset="45%" stop-color="#1d4ed8" stop-opacity="0.05" />
          <stop offset="75%" stop-color="#1d4ed8" stop-opacity="0" />
        </linearGradient>
        <linearGradient id="card-reflection" x1="0" y1="0" x2="0.7" y2="1">
          <stop offset="0%" stop-color="#ffffff" stop-opacity="0.22" />
          <stop offset="30%" stop-color="#ffffff" stop-opacity="0.09" />
          <stop offset="52%" stop-color="#ffffff" stop-opacity="0" />
        </linearGradient>
        <linearGradient id="card-streak" x1="0" y1="0" x2="1" y2="0">
          <stop offset="0%" stop-color="#ffffff" stop-opacity="0" />
          <stop offset="50%" stop-color="#dbeafe" stop-opacity="0.09" />
          <stop offset="100%" stop-color="#ffffff" stop-opacity="0" />
        </linearGradient>
        <radialGradient id="card-ambient" cx="0.5" cy="0.42" r="0.85">
          <stop offset="0%" stop-color="#3b82f6" stop-opacity="0" />
          <stop offset="72%" stop-color="#3b82f6" stop-opacity="0" />
          <stop offset="100%" stop-color="#3b82f6" stop-opacity="0.12" />
        </radialGradient>
        <filter id="card-shadow" x="-20%" y="-20%" width="140%" height="140%">
          <feDropShadow
            dx="0"
            dy="0"
            stdDeviation="9"
            flood-color="#3b82f6"
            flood-opacity="0.16"
          />
          <feDropShadow
            dx="0"
            dy="3"
            stdDeviation="4"
            flood-color="#000000"
            flood-opacity="0.16"
          />
        </filter>
      </defs>
    `
}

// RenderEffectOverlays renders the glass overlay shapes clipped to the card.
func (c *Card) RenderEffectOverlays() string {
	w := c.Width
	h := c.Height
	return `
      <g clip-path="url(#card-clip)">
        <rect x="0" y="0" width="` + fmtNum(w) + `" height="` + fmtNum(h) + `" fill="url(#card-depth)" />
        <rect
          x="0"
          y="0"
          width="` + fmtNum(w) + `"
          height="` + fmtNum(h) + `"
          fill="url(#card-reflection)"
        />
        <g opacity="0.55">
          <polygon
            points="` + strconv.FormatFloat(w*0.18, 'f', 1, 64) + `,0 ` + strconv.FormatFloat(w*0.34, 'f', 1, 64) + `,0 ` + strconv.FormatFloat(w*0.1, 'f', 1, 64) + `,` + fmtNum(h) + ` ` + strconv.FormatFloat(w*-0.06, 'f', 1, 64) + `,` + fmtNum(h) + `"
            fill="url(#card-streak)"
          />
          <polygon
            points="` + strconv.FormatFloat(w*0.42, 'f', 1, 64) + `,0 ` + strconv.FormatFloat(w*0.5, 'f', 1, 64) + `,0 ` + strconv.FormatFloat(w*0.26, 'f', 1, 64) + `,` + fmtNum(h) + ` ` + strconv.FormatFloat(w*0.18, 'f', 1, 64) + `,` + fmtNum(h) + `"
            fill="url(#card-streak)"
            opacity="0.7"
          />
        </g>
        <rect x="0" y="0" width="` + fmtNum(w) + `" height="` + fmtNum(h) + `" fill="url(#card-ambient)" />
      </g>
    `
}

// GetAnimations returns the card CSS animations.
func (c *Card) GetAnimations() string {
	return `
      /* Animations */
      @keyframes scaleInAnimation {
        from {
          transform: translate(-5px, 5px) scale(0);
        }
        to {
          transform: translate(-5px, 5px) scale(1);
        }
      }
      @keyframes fadeInAnimation {
        from {
          opacity: 0;
        }
        to {
          opacity: 1;
        }
      }
    `
}

// Render renders the full card SVG around body.
func (c *Card) Render(body string) string {
	animations := ""
	if os.Getenv("GO_ENV") != "test" {
		animations = c.GetAnimations()
	}
	disableCSS := ""
	if !c.Animations {
		disableCSS = `* { animation-duration: 0s !important; animation-delay: 0s !important; }`
	}
	strokeOpacity := "1"
	if c.HideBorder {
		strokeOpacity = "0"
	}
	title := ""
	if !c.HideTitle {
		title = c.RenderTitle()
	}
	bodyY := c.PaddingX
	if !c.HideTitle {
		bodyY = c.PaddingY + 20
	}
	return `
      <svg
        width="` + fmtNum(c.Width) + `"
        height="` + fmtNum(c.Height) + `"
        viewBox="0 0 ` + fmtNum(c.Width) + ` ` + fmtNum(c.Height) + `"
        fill="none"
        xmlns="http://www.w3.org/2000/svg"
        role="img"
        aria-labelledby="descId"
      >
        <title id="titleId">` + c.A11yTitle + `</title>
        <desc id="descId">` + c.A11yDesc + `</desc>
        <style>
          .header {
            font: 600 17px 'Segoe UI', Ubuntu, Sans-Serif;
            fill: ` + c.Colors.TitleColor + `;
            animation: fadeInAnimation 0.8s ease-in-out forwards;
          }
          @supports(-moz-appearance: auto) {
            /* Selector detects Firefox */
            .header { font-size: 14.5px; }
          }
          ` + c.CSS + `

          ` + animations + `
          ` + disableCSS + `
        </style>

        ` + c.RenderGradient() + `

        ` + c.RenderEffectDefs() + `

        <rect
          data-testid="card-bg"
          x="0.5"
          y="0.5"
          rx="` + fmtNum(c.BorderRadius) + `"
          height="99%"
          stroke="` + c.Colors.BorderColor + `"
          width="` + fmtNum(c.Width-1) + `"
          fill="` + c.Colors.bgFill() + `"
          stroke-opacity="` + strokeOpacity + `"
          filter="url(#card-shadow)"
        />

        <rect
          data-testid="card-edge-highlight"
          x="1.5"
          y="1.5"
          rx="` + fmtNum(c.BorderRadius) + `"
          height="98%"
          width="` + fmtNum(c.Width-3) + `"
          fill="none"
          stroke="#ffffff"
          stroke-opacity="0.06"
        />

        ` + c.RenderEffectOverlays() + `

        ` + title + `

        <g
          data-testid="main-card-body"
          transform="translate(0, ` + fmtNum(bodyY) + `)"
        >
          ` + body + `
        </g>
      </svg>
    `
}
