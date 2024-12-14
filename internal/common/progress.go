package common

import "strconv"

// CreateProgressNode renders a horizontal progress bar node.
// Progress is clamped to [2, 100], mirroring createProgressNode in
// src/common/createProgressNode.js.
func CreateProgressNode(x, y, width float64, color string, progress float64, progressBarBg string, delay int) string {
	progressPercentage := ClampValue(progress, 2, 100)

	f := func(v float64) string { return strconv.FormatFloat(v, 'f', -1, 64) }

	return `
    <svg width="` + f(width) + `" x="` + f(x) + `" y="` + f(y) + `">
      <rect rx="5" ry="5" x="0" y="0" width="` + f(width) + `" height="8" fill="` + progressBarBg + `"></rect>
      <svg data-testid="lang-progress" width="` + f(progressPercentage) + `%">
        <rect
            height="8"
            fill="` + color + `"
            rx="5" ry="5" x="0" y="0"
            class="lang-progress"
            style="animation-delay: ` + strconv.Itoa(delay) + `ms;"
        />
      </svg>
    </svg>
  `
}
