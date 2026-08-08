package main

import (
	"fmt"
	"strings"
)

// drawHeading renders a section heading in the mono face with a hairline
// running right. GitHub strips <style> from markdown, so rendering the label
// as SVG is the only way to put the page's own typeface on it. The rule starts
// past the widest plausible advance so a narrower font widens the gap rather
// than colliding.
func drawHeading(word string) string {
	const (
		fs = 16
		H  = 26
	)
	textEnd := advance(word, fs) + 18

	var p strings.Builder
	p.WriteString(head(Width, H, fontHead()))
	p.WriteString(label(I(0), I(18), word, fs,
		opts{cls: "e-f", extra: ` font-weight="600"`}))
	p.WriteString(fmt.Sprintf(`<line x1="%.0f" y1="12.5" x2="%d" y2="12.5" `+
		`class="u-s" stroke-width="1"/>`, textEnd, Width))
	p.WriteString(`</svg>`)
	return p.String()
}
