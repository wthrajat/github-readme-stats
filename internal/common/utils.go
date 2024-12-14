package common

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/wthrajat/github-readme-stats/internal/themes"
)

// ErrorCardLength is the width of the error card SVG.
const ErrorCardLength = 576.5

// Error type identifiers, mirroring CustomError statics in utils.js.
const (
	ErrMaxRetry        = "MAX_RETRY"
	ErrNoTokens        = "NO_TOKENS"
	ErrUserNotFound    = "USER_NOT_FOUND"
	ErrGraphQLError    = "GRAPHQL_ERROR"
	ErrGithubRestError = "GITHUB_REST_API_ERROR"
	ErrWakatime        = "WAKATIME_ERROR"
)

// TryAgainLater is the shared secondary message for retryable errors.
const TryAgainLater = "Please try again later"

var secondaryErrorMessages = map[string]string{
	ErrMaxRetry:               "You can deploy own instance or wait until public will be no longer limited",
	ErrNoTokens:               "Please add an env variable called PAT_1 with your GitHub API token in vercel",
	ErrUserNotFound:           "Make sure the provided username is not an organization",
	ErrGraphQLError:           TryAgainLater,
	ErrGithubRestError:        TryAgainLater,
	"WAKATIME_USER_NOT_FOUND": "Make sure you have a public WakaTime profile",
}

// CustomError is a typed error with a secondary message for the error card.
type CustomError struct {
	Message          string
	Type             string
	SecondaryMessage string
}

// NewCustomError creates a CustomError, mapping the type to its secondary
// message and defaulting to the type itself when unknown.
func NewCustomError(msg, typ string) *CustomError {
	secondary, ok := secondaryErrorMessages[typ]
	if !ok {
		secondary = typ
	}
	return &CustomError{Message: msg, Type: typ, SecondaryMessage: secondary}
}

func (e *CustomError) Error() string { return e.Message }

// MissingParamError is raised when required query parameters are absent.
type MissingParamError struct {
	MissedParams []string
	Secondary    string
}

// NewMissingParamError creates a MissingParamError.
func NewMissingParamError(missedParams []string, secondary string) *MissingParamError {
	return &MissingParamError{MissedParams: missedParams, Secondary: secondary}
}

func (e *MissingParamError) Error() string {
	quoted := make([]string, len(e.MissedParams))
	for i, p := range e.MissedParams {
		quoted[i] = `"` + p + `"`
	}
	return fmt.Sprintf("Missing params %s make sure you pass the parameters in URL",
		strings.Join(quoted, ", "))
}

// FlexLayout lays out items horizontally or vertically with gaps, wrapping
// each non-empty item in a translated <g> element.
// Ported from the flexLayout helper in utils.js.
func FlexLayout(items []string, gap int, direction string, sizes []int) []string {
	var out []string
	lastSize := 0
	idx := 0
	for _, item := range items {
		if item == "" {
			continue
		}
		size := 0
		if idx < len(sizes) {
			size = sizes[idx]
		}
		transform := fmt.Sprintf("translate(%d, 0)", lastSize)
		if direction == "column" {
			transform = fmt.Sprintf("translate(0, %d)", lastSize)
		}
		lastSize += size + gap
		out = append(out, `<g transform="`+transform+`">`+item+`</g>`)
		idx++
	}
	return out
}

// CreateLanguageNode renders the primary-language badge node.
func CreateLanguageNode(langName, langColor string) string {
	return `
    <g data-testid="primary-lang">
      <circle data-testid="lang-color" cx="0" cy="-5" r="6" fill="` + langColor + `" />
      <text data-testid="lang-name" class="gray" x="15">` + langName + `</text>
    </g>
    `
}

// IconWithLabel renders an icon followed by a label. Numeric labels <= 0
// render as the empty string, mirroring iconWithLabel in utils.js.
func IconWithLabel(icon string, label any, testid string, iconSize int) string {
	switch v := label.(type) {
	case int:
		if v <= 0 {
			return ""
		}
	case int8:
		if v <= 0 {
			return ""
		}
	case int16:
		if v <= 0 {
			return ""
		}
	case int32:
		if v <= 0 {
			return ""
		}
	case int64:
		if v <= 0 {
			return ""
		}
	case uint:
		if v == 0 {
			return ""
		}
	case uint8:
		if v == 0 {
			return ""
		}
	case uint16:
		if v == 0 {
			return ""
		}
	case uint32:
		if v == 0 {
			return ""
		}
	case uint64:
		if v == 0 {
			return ""
		}
	case float32:
		if v <= 0 {
			return ""
		}
	case float64:
		if v <= 0 {
			return ""
		}
	}
	iconSVG := `
      <svg
        class="icon"
        y="-12"
        viewBox="0 0 16 16"
        version="1.1"
        width="` + strconv.Itoa(iconSize) + `"
        height="` + strconv.Itoa(iconSize) + `"
      >
        ` + icon + `
      </svg>
    `
	text := `<text data-testid="` + testid + `" class="gray">` + fmt.Sprint(label) + `</text>`
	return strings.Join(FlexLayout([]string{iconSVG, text}, 20, "", nil), "")
}

// KFormatter formats numbers over 999 with a "k" suffix precise to one
// decimal, mirroring kFormatter in utils.js.
func KFormatter(num float64) string {
	abs := math.Abs(num)
	if abs > 999 {
		v := abs / 1000
		s := strconv.FormatFloat(math.Round(v*10)/10, 'f', -1, 64)
		if num < 0 {
			return "-" + s + "k"
		}
		return s + "k"
	}
	if num == math.Trunc(num) {
		return strconv.FormatInt(int64(num), 10)
	}
	return strconv.FormatFloat(num, 'f', -1, 64)
}

var hexColorRegex = regexp.MustCompile(`^([A-Fa-f0-9]{8}|[A-Fa-f0-9]{6}|[A-Fa-f0-9]{3}|[A-Fa-f0-9]{4})$`)

// IsValidHexColor reports whether s is a valid hex color (without '#').
func IsValidHexColor(s string) bool {
	return hexColorRegex.MatchString(s)
}

// ParseBoolean parses "true"/"false" (case-insensitive) into a bool pointer.
// Empty or unrecognized strings yield nil.
func ParseBoolean(s string) *bool {
	if s == "" {
		return nil
	}
	if strings.EqualFold(s, "true") {
		v := true
		return &v
	}
	if strings.EqualFold(s, "false") {
		v := false
		return &v
	}
	return nil
}

// ParseArray splits a comma-separated string, returning [] for empty input.
func ParseArray(s string) []string {
	if s == "" {
		return []string{}
	}
	return strings.Split(s, ",")
}

// ClampValue clamps n into [min, max], returning min for NaN.
func ClampValue(n, min, max float64) float64 {
	if math.IsNaN(n) {
		return min
	}
	return math.Max(min, math.Min(n, max))
}

// IsValidGradient reports whether colors describe a valid gradient
// (a rotation plus at least two valid hex colors).
func IsValidGradient(colors []string) bool {
	if len(colors) <= 2 {
		return false
	}
	for _, c := range colors[1:] {
		if !IsValidHexColor(c) {
			return false
		}
	}
	return true
}

// FallbackColor resolves a color string to a gradient ([]string), a "#hex"
// string, or the fallback when invalid. Mirrors fallbackColor in utils.js.
func FallbackColor(color string, fallback any) any {
	var colors []string
	if color != "" {
		colors = strings.Split(color, ",")
	}
	if len(colors) > 1 && IsValidGradient(colors) {
		return colors
	}
	if IsValidHexColor(color) {
		return "#" + color
	}
	return fallback
}

// CardColors holds resolved card colors. BgColor is a plain color string;
// gradient backgrounds are stored in BgGradient with IsGradient set.
type CardColors struct {
	TitleColor  string
	IconColor   string
	TextColor   string
	BgColor     string
	BgGradient  []string
	IsGradient  bool
	BorderColor string
	RingColor   string
}

func fallbackColorStr(color, fallback string) string {
	switch v := FallbackColor(color, fallback).(type) {
	case string:
		return v
	case []string:
		return fallback
	default:
		return fallback
	}
}

// bgFill returns the SVG fill value for the resolved background.
func (c CardColors) bgFill() string {
	if c.IsGradient {
		return "url(#gradient)"
	}
	return c.BgColor
}

// BgFillValue returns the raw background value (gradient joined by comma
// like JS Array.toString, or the plain color). Used by RenderError.
func (c CardColors) BgFillValue() string {
	if c.IsGradient {
		return strings.Join(c.BgGradient, ",")
	}
	return c.BgColor
}

// GetCardColors resolves theme colors with user overrides and defaults.
// Mirrors getCardColors in utils.js.
func GetCardColors(titleColor, textColor, iconColor, bgColor, borderColor, ringColor, theme, fallbackTheme string) CardColors {
	if fallbackTheme == "" {
		fallbackTheme = "default"
	}
	defaultTheme, ok := themes.Themes[fallbackTheme]
	if !ok {
		defaultTheme = themes.Themes["default"]
	}
	selectedTheme, ok := themes.Themes[theme]
	if theme == "" || !ok {
		selectedTheme = defaultTheme
	}
	defaultBorderColor := selectedTheme.BorderColor
	if defaultBorderColor == "" {
		defaultBorderColor = defaultTheme.BorderColor
	}

	orSelected := func(given, selected string) string {
		if given != "" {
			return given
		}
		return selected
	}

	title := fallbackColorStr(orSelected(titleColor, selectedTheme.TitleColor), "#"+defaultTheme.TitleColor)
	ring := fallbackColorStr(orSelected(ringColor, selectedTheme.RingColor), title)
	icon := fallbackColorStr(orSelected(iconColor, selectedTheme.IconColor), "#"+defaultTheme.IconColor)
	text := fallbackColorStr(orSelected(textColor, selectedTheme.TextColor), "#"+defaultTheme.TextColor)
	border := fallbackColorStr(orSelected(borderColor, defaultBorderColor), "#"+defaultBorderColor)

	out := CardColors{
		TitleColor:  title,
		IconColor:   icon,
		TextColor:   text,
		BorderColor: border,
		RingColor:   ring,
	}
	switch v := FallbackColor(orSelected(bgColor, selectedTheme.BgColor), "#"+defaultTheme.BgColor).(type) {
	case string:
		out.BgColor = v
	case []string:
		out.BgGradient = v
		out.IsGradient = true
	default:
		out.BgColor = "#" + defaultTheme.BgColor
	}
	return out
}

// EncodeHTML escapes HTML special and non-ASCII BMP characters as numeric
// entities and strips backspace characters. Mirrors encodeHTML in utils.js.
func EncodeHTML(s string) string {
	runes := []rune(s)
	var b strings.Builder
	b.Grow(len(s))
	for i, r := range runes {
		if r == 0x08 {
			continue
		}
		if r == '<' || r == '>' || r == '&' || (r >= 0xA0 && r <= 0x9999) {
			if i+1 < len(runes) && runes[i+1] == '#' {
				b.WriteRune(r)
				continue
			}
			b.WriteString("&#")
			b.WriteString(strconv.Itoa(int(r)))
			b.WriteString(";")
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}

// RenderError renders the error card SVG. opts keys: title_color,
// text_color, bg_color, border_color, theme.
func RenderError(message, secondary string, opts map[string]string) string {
	var titleColorOpt, textColorOpt, bgColorOpt, borderColorOpt, theme string
	if opts != nil {
		titleColorOpt = opts["title_color"]
		textColorOpt = opts["text_color"]
		bgColorOpt = opts["bg_color"]
		borderColorOpt = opts["border_color"]
		theme = opts["theme"]
	}
	if theme == "" {
		theme = "default"
	}
	colors := GetCardColors(titleColorOpt, textColorOpt, "", bgColorOpt, borderColorOpt, "", theme, "default")

	suffix := " file an issue at https://github.com/wthrajat/github-readme-stats/issues"
	if secondary == TryAgainLater || secondary == secondaryErrorMessages[ErrMaxRetry] {
		suffix = ""
	}

	bg := colors.BgFillValue()
	length := strconv.FormatFloat(ErrorCardLength, 'f', -1, 64)
	inner := strconv.FormatFloat(ErrorCardLength-1, 'f', -1, 64)

	return `
    <svg width="` + length + `"  height="120" viewBox="0 0 ` + length + ` 120" fill="` + bg + `" xmlns="http://www.w3.org/2000/svg">
    <style>
    .text { font: 600 16px 'Segoe UI', Ubuntu, Sans-Serif; fill: ` + colors.TitleColor + ` }
    .small { font: 600 12px 'Segoe UI', Ubuntu, Sans-Serif; fill: ` + colors.TextColor + ` }
    .gray { fill: #858585 }
    </style>
    <rect x="0.5" y="0.5" width="` + inner + `" height="99%" rx="4.5" fill="` + bg + `" stroke="` + colors.BorderColor + `"/>
    <text x="25" y="45" class="text">Something went wrong!` + suffix + `</text>
    <text data-testid="message" x="25" y="55" class="text small">
      <tspan x="25" dy="18">` + EncodeHTML(message) + `</tspan>
      <tspan x="25" dy="18" class="gray">` + secondary + `</tspan>
    </text>
    </svg>
  `
}

// WrapTextMultiline splits text into at most maxLines lines of the given
// width, mirroring wrapTextMultiline in utils.js.
func WrapTextMultiline(text string, width, maxLines int) []string {
	const fullWidthComma = "，"
	encoded := EncodeHTML(text)

	var wrapped []string
	if strings.Contains(encoded, fullWidthComma) {
		wrapped = strings.Split(encoded, fullWidthComma)
	} else {
		wrapped = wordWrap(encoded, width)
	}

	end := maxLines
	if len(wrapped) < end {
		end = len(wrapped)
	}
	lines := make([]string, 0, end)
	for _, line := range wrapped[:end] {
		lines = append(lines, strings.TrimSpace(line))
	}
	if len(wrapped) > maxLines && len(lines) > 0 {
		lines[maxLines-1] += "..."
	}
	out := lines[:0]
	for _, line := range lines {
		if line != "" {
			out = append(out, line)
		}
	}
	return out
}

// wordWrap greedily wraps s at word boundaries to the given width
// (measured in runes), without breaking long words.
func wordWrap(s string, width int) []string {
	if width <= 0 {
		return []string{s}
	}
	words := strings.Fields(s)
	if len(words) == 0 {
		return []string{""}
	}
	var lines []string
	cur := words[0]
	curLen := len([]rune(words[0]))
	for _, w := range words[1:] {
		wLen := len([]rune(w))
		if curLen+1+wLen <= width {
			cur += " " + w
			curLen += 1 + wLen
		} else {
			lines = append(lines, cur)
			cur = w
			curLen = wLen
		}
	}
	lines = append(lines, cur)
	return lines
}

// Cache durations in seconds, mirroring CONSTANTS in utils.js.
const (
	OneMinute            = 60
	FiveMinutes          = 300
	TenMinutes           = 600
	FifteenMinutes       = 900
	ThirtyMinutes        = 1800
	TwoHours             = 7200
	FourHours            = 14400
	SixHours             = 21600
	EightHours           = 28800
	TwelveHours          = 43200
	OneDay               = 86400
	TwoDay               = 172800
	SixDay               = 518400
	TenDay               = 864000
	CardCacheSeconds     = OneDay
	TopLangsCacheSeconds = SixDay
	PinCardCacheSeconds  = TenDay
	ErrorCacheSeconds    = TenMinutes
)

// TextWidths maps char codes to their relative width at font size 1.
// Copied from the TEXT_WIDTHS table in utils.js.
var TextWidths = [128]float64{
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0.2796875, 0.2765625,
	0.3546875, 0.5546875, 0.5546875, 0.8890625, 0.665625, 0.190625,
	0.3328125, 0.3328125, 0.3890625, 0.5828125, 0.2765625, 0.3328125,
	0.2765625, 0.3015625, 0.5546875, 0.5546875, 0.5546875, 0.5546875,
	0.5546875, 0.5546875, 0.5546875, 0.5546875, 0.5546875, 0.5546875,
	0.2765625, 0.2765625, 0.584375, 0.5828125, 0.584375, 0.5546875,
	1.0140625, 0.665625, 0.665625, 0.721875, 0.721875, 0.665625,
	0.609375, 0.7765625, 0.721875, 0.2765625, 0.5, 0.665625,
	0.5546875, 0.8328125, 0.721875, 0.7765625, 0.665625, 0.7765625,
	0.721875, 0.665625, 0.609375, 0.721875, 0.665625, 0.94375,
	0.665625, 0.665625, 0.609375, 0.2765625, 0.3546875, 0.2765625,
	0.4765625, 0.5546875, 0.3328125, 0.5546875, 0.5546875, 0.5,
	0.5546875, 0.5546875, 0.2765625, 0.5546875, 0.5546875, 0.221875,
	0.240625, 0.5, 0.221875, 0.8328125, 0.5546875, 0.5546875,
	0.5546875, 0.5546875, 0.3328125, 0.5, 0.2765625, 0.5546875,
	0.5, 0.721875, 0.5, 0.5, 0.5, 0.3546875, 0.259375, 0.353125, 0.5890625,
}

// AvgCharWidth is the fallback width for chars outside TextWidths.
const AvgCharWidth = 0.5279276315789471

// MeasureText estimates the rendered width of s at the given font size.
func MeasureText(s string, fontSize float64) float64 {
	width := 0.0
	for _, r := range s {
		if r >= 0 && int(r) < len(TextWidths) {
			width += TextWidths[int(r)]
		} else {
			width += AvgCharWidth
		}
	}
	return width * fontSize
}

// LowercaseTrim lowercases and trims surrounding whitespace.
func LowercaseTrim(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}

// ChunkArray splits arr into sequential chunks of perChunk items.
func ChunkArray[T any](arr []T, perChunk int) [][]T {
	if perChunk <= 0 {
		return [][]T{arr}
	}
	var out [][]T
	for i := 0; i < len(arr); i += perChunk {
		end := i + perChunk
		if end > len(arr) {
			end = len(arr)
		}
		out = append(out, arr[i:end])
	}
	return out
}

var emojiPattern = regexp.MustCompile(`:([A-Za-z0-9_]+):`)

var emojiMap = map[string]string{
	"heart": "❤️", "rocket": "🚀", "star": "⭐", "fire": "🔥",
	"tada": "🎉", "sparkles": "✨", "eyes": "👀", "smile": "😄",
	"laughing": "😆", "wink": "😉", "thumbsup": "👍", "+1": "👍",
	"thumbsdown": "👎", "-1": "👎", "clap": "👏", "pray": "🙏",
	"muscle": "💪", "bug": "🐛", "book": "📚", "books": "📚",
	"bulb": "💡", "white_check_mark": "✅", "x": "❌",
	"warning": "⚠️", "information_source": "ℹ️", "question": "❓",
	"zap": "⚡", "art": "🎨", "construction": "🚧",
	"recycle": "♻️", "package": "📦", "wrench": "🔧",
	"hammer": "🔨", "gear": "⚙️", "link": "🔗", "lock": "🔒",
	"key": "🔑", "globe_with_meridians": "🌐", "computer": "💻",
	"iphone": "📱", "email": "📧", "calendar": "📅", "alarm_clock": "⏰",
	"moneybag": "💰", "gift": "🎁", "trophy": "🏆", "medal": "🏅",
	"triangular_flag_on_post": "🚩", "balloon": "🎈", "cake": "🎂",
	"coffee": "☕", "beer": "🍺", "pizza": "🍕", "apple": "🍎",
	"evergreen_tree": "🌲", "sunny": "☀️", "moon": "🌙",
	"cloud": "☁️", "umbrella": "☂️", "snowflake": "❄️",
	"ocean": "🌊", "earth_americas": "🌍", "wave": "👋",
	"cry": "😢", "joy": "😂", "sunglasses": "😎", "thinking": "🤔",
	"sleeping": "😴", "mask": "😷", "ghost": "👻", "skull": "💀",
	"poop": "💩", "see_no_evil": "🙈", "hear_no_evil": "🙉",
	"speak_no_evil": "🙊", "100": "💯", "partying_face": "🥳",
	"confetti_ball": "🎊", "gem": "💎", "crown": "👑",
	"checkered_flag": "🏁", "dart": "🎯", "game_die": "🎲",
	"headphones": "🎧", "microphone": "🎤", "camera": "📷",
	"tv": "📺", "bulb2": "💡", "memo": "📝", "pencil": "✏️",
	"mag": "🔍", "chart_with_upwards_trend": "📈",
	"chart_with_downwards_trend": "📉", "bar_chart": "📊",
	"green_heart": "💚", "blue_heart": "💙", "yellow_heart": "💛",
	"purple_heart": "💜", "black_heart": "🖤", "white_heart": "🤍",
	"broken_heart": "💔", "kiss": "💋", "hugs": "🤗",
	"wave2": "🌊", "penguin": "🐧", "whale": "🐳", "dolphin": "🐬",
	"snake": "🐍", "turtle": "🐢", "octocat": "🐙", "cat": "🐱",
	"dog": "🐶",
}

// ParseEmojis replaces :shortcode: occurrences with their emoji,
// using "" for unknown shortcodes like emoji-name-map does.
func ParseEmojis(s string) string {
	if s == "" {
		return ""
	}
	return emojiPattern.ReplaceAllStringFunc(s, func(m string) string {
		name := m[1 : len(m)-1]
		if e, ok := emojiMap[name]; ok {
			return e
		}
		return ""
	})
}

// DateDiff returns the difference between d1 and d2 in rounded minutes.
func DateDiff(d1, d2 time.Time) int {
	return int(math.Round(d1.Sub(d2).Minutes()))
}

// GraphQLRequest POSTs data to the GitHub GraphQL API with a 10s timeout,
// returning the raw body, status code, and any error.
func GraphQLRequest(data any, headers map[string]string) ([]byte, int, error) {
	payload, err := json.Marshal(data)
	if err != nil {
		return nil, 0, err
	}
	req, err := http.NewRequest(http.MethodPost, "https://api.github.com/graphql", bytes.NewReader(payload))
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("Content-Type", "application/json")
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()
	var buf bytes.Buffer
	if _, err := buf.ReadFrom(resp.Body); err != nil {
		return nil, resp.StatusCode, err
	}
	return buf.Bytes(), resp.StatusCode, nil
}
