package main

import (
	"fmt"
	"strconv"
	"strings"
)

const (
	iconFlame = "M12 2C11.5 3 10.3 5.3 10.3 7C10.3 8.7 11.2 10 12 10C12.8 10 13.7 8.7 13.7 7C13.7 5.3" +
		" 12.5 3 12 2ZM17 12C15.7 9.8 14 8.6 14 8.6C14 8.6 13 9.6 12.5 10.7C12 11.8 12.8 13.3" +
		" 11.5 14.6C10.2 13.3 11 11.8 10.5 10.7C10 9.6 9 8.6 9 8.6C9 8.6 7.3 9.8 6 12C4.7" +
		" 14.4 5.6 17.7 8.2 19.5C9.9 20.6 12 20.6 13.7 19.5C16.3 17.7 17.3 14.4 17 12Z"
	iconTrophy = "M19 5h-2V3H7v2H5c-1.1 0-2 .9-2 2v3c0 2.44 1.72 4.44 4 4.9V19h-3v2h16v-2h-3v-4.1" +
		"c2.28-.46 4-2.46 4-4.9V7c0-1.1-.9-2-2-2zM5 10V7h2v3H5zm14 0h-2V7h2v3z"
	iconUser = "M12 12c2.21 0 4-1.79 4-4s-1.79-4-4-4-4 1.79-4 4 1.79 4 4 4zm0 2c-2.67 0-8 1.34-8 4v2h16v-2c0-2.66-5.33-4-8-4z"
	iconAdd  = "M9 12c2.21 0 4-1.79 4-4s-1.79-4-4-4-4 1.79-4 4 1.79 4 4 4zm0 2c-2.67 0-8 1.34-8 4v2h16v-2c0-2.66-5.33-4-8-4zm11-4V7h-2v3h-3v2h3v3h2v-3h3v-2h-3z"
	iconStar = "M12 17.27L18.18 21l-1.64-7.03L22 9.24l-7.19-.61L12 2 9.19 8.63 2 9.24l5.46 4.73L5.82 21z"
)

func icon(x num, path string) string {
	return fmt.Sprintf(`<svg x="%.0f" y="32" width="24" height="24" `+
		`viewBox="0 0 24 24" fill="none"><path d="%s" class="d-f"/></svg>`,
		x.v, path)
}

// drawStreak: current and longest streak, split by a hairline.
func drawStreak(s Summary) string {
	const H = 96

	type cell struct {
		val  int
		lab  string
		span string
	}
	cells := make([]cell, 0, 2)
	for i, r := range []Streak{s.Current, s.Longest} {
		lab := "current streak"
		if i == 1 {
			lab = "longest streak"
		}
		span := "&#8212;"
		if r.Length > 0 {
			span = pretty(r.Start) + " &#8211; " + pretty(r.End)
		}
		cells = append(cells, cell{r.Length, lab, span})
	}

	var p strings.Builder
	p.WriteString(head(Width, H, ""))

	mid := float64(Width) / 2
	p.WriteString(fmt.Sprintf(`<line x1="%.0f" y1="16" x2="%.0f" y2="80" `+
		`class="u-s" stroke-width="1" opacity="0">%s</line>`,
		mid, mid, fade(0.20)))

	for i, c := range cells {
		// First column is a plain int offset; the second lands on a float.
		x := I(Left)
		glyph := iconFlame
		if i == 1 {
			x = F(mid + Left)
			glyph = iconTrophy
		}
		p.WriteString(`<g opacity="0">` + fade(delay(0.12, 0.14, i)) +
			icon(x, glyph) +
			label(x.Add(I(36)), I(44), strconv.Itoa(c.val), 34,
				opts{cls: "e-f", extra: ` font-weight="600"`}) +
			label(x.Add(I(36)), I(64), c.lab, 11, opts{}) +
			label(x.Add(I(36)), I(80), c.span, 10, opts{}) + `</g>`)
	}

	p.WriteString(`</svg>`)
	return p.String()
}

// drawSocial: followers, following, stars earned. Three columns instead of
// drawStreak's two.
func drawSocial(s Summary) string {
	const H = 96

	cells := []struct {
		val  int
		lab  string
		icon string
	}{
		{s.Followers, "followers", iconUser},
		{s.Following, "following", iconAdd},
		{s.Stars, "stars earned", iconStar},
	}
	n := len(cells)
	colw := float64(Width) / float64(n)

	var p strings.Builder
	p.WriteString(head(Width, H, ""))

	for i := 1; i < n; i++ {
		x := colw * float64(i)
		p.WriteString(fmt.Sprintf(`<line x1="%.0f" y1="16" x2="%.0f" y2="80" `+
			`class="u-s" stroke-width="1" opacity="0">%s</line>`,
			x, x, fade(0.20)))
	}

	for i, c := range cells {
		x := F(float64(colw*float64(i)) + Left)
		p.WriteString(`<g opacity="0">` + fade(delay(0.12, 0.14, i)) +
			icon(x, c.icon) +
			label(x.Add(I(36)), I(44), strconv.Itoa(c.val), 30,
				opts{cls: "e-f", extra: ` font-weight="600"`}) +
			label(x.Add(I(36)), I(64), c.lab, 11, opts{}) + `</g>`)
	}

	p.WriteString(`</svg>`)
	return p.String()
}
