package main

import (
	"fmt"
	"strconv"
	"strings"
)

// drawActivity: commits, PRs, reviews, issues.
func drawActivity(s Summary) string {
	rows := s.Activity
	H := 20 + max(len(rows), 1)*24 + 6
	const (
		nameW  = 130
		valW   = 44
		barMax = Width - Left - nameW - valW - 10
	)

	var p strings.Builder
	p.WriteString(head(Width, H, ""))
	p.WriteString(`<g opacity="0">` + fade(0.10) +
		label(I(Left), I(12), "CONTRIBUTION MIX", 9,
			opts{cls: "m-f", extra: ` letter-spacing="1.3"`}) + `</g>`)

	top := 0
	for _, r := range rows {
		if r.Value > top {
			top = r.Value
		}
	}
	if top == 0 {
		top = 1
	}

	clip, cursor := wipe("ra", I(Left+nameW), I(20), I(barMax),
		I(len(rows)*24), 0.34, 1.05)
	p.WriteString(clip)

	for i, r := range rows {
		y := 24 + i*24
		p.WriteString(`<g opacity="0">` + fade(delay(0.24, 0.06, i)) +
			label(I(Left), I(y+8), r.Name, 11, opts{cls: "e-f"}) +
			label(I(Width), I(y+8), strconv.Itoa(r.Value), 11,
				opts{cls: "m-f", anchor: "end"}) + `</g>`)
		p.WriteString(`<g clip-path="url(#ra)">` +
			hbar(Left+nameW, float64(y), barMax, 7, "w") +
			hbar(Left+nameW, float64(y), barMax*float64(r.Value)/float64(top), 7, "") +
			`</g>`)
	}

	p.WriteString(cursor)
	p.WriteString(`</svg>`)
	return p.String()
}

// drawRepos: top starred repos, language in dim ink.
func drawRepos(s Summary) string {
	rows := s.TopRepos
	H := 20 + max(len(rows), 1)*24 + 6
	const (
		nameW  = 210
		valW   = 34
		barMax = Width - Left - nameW - valW - 10
	)

	var p strings.Builder
	p.WriteString(head(Width, H, ""))
	p.WriteString(`<g opacity="0">` + fade(0.10) +
		label(I(Left), I(12), "TOP REPOS", 9,
			opts{cls: "m-f", extra: ` letter-spacing="1.3"`}) + `</g>`)
	if len(rows) == 0 {
		p.WriteString(`</svg>`)
		return p.String()
	}

	top := 0
	for _, r := range rows {
		if r.Stars > top {
			top = r.Stars
		}
	}
	if top == 0 {
		top = 1
	}

	clip, cursor := wipe("rr", I(Left+nameW), I(20), I(barMax),
		I(len(rows)*24), 0.34, 1.05)
	p.WriteString(clip)

	for i, r := range rows {
		y := 24 + i*24
		name := r.Name
		if len(name) > 22 {
			name = name[:22]
		}
		if r.Lang != "" {
			name += fmt.Sprintf(`  <tspan class="m-f" font-size="10">%s</tspan>`,
				strings.ToLower(r.Lang))
		}
		p.WriteString(`<g opacity="0">` + fade(delay(0.24, 0.06, i)) +
			label(I(Left), I(y+8), name, 11, opts{cls: "e-f"}) +
			label(I(Width), I(y+8), "&#9733; "+strconv.Itoa(r.Stars), 11,
				opts{cls: "m-f", anchor: "end"}) + `</g>`)
		p.WriteString(`<g clip-path="url(#rr)">` +
			hbar(Left+nameW, float64(y), barMax, 7, "w") +
			hbar(Left+nameW, float64(y), barMax*float64(r.Stars)/float64(top), 7, "") +
			`</g>`)
	}

	p.WriteString(cursor)
	p.WriteString(`</svg>`)
	return p.String()
}
