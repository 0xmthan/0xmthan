package main

import (
	"fmt"
	"strconv"
	"strings"
)

// drawLangs: share of bytes, and count of repos by main language.
func drawLangs(s Summary) string {
	rows := max(len(s.BySize), len(s.ByRepo), 1)
	H := 26 + rows*22 + 6
	colw := float64(Width-Left-30) / 2
	const nameW = 82
	barMax := colw - 82 - 44

	var p strings.Builder
	p.WriteString(head(Width, H, ""))

	groups := []struct {
		gx    num
		title string
		data  []Count
		asPct bool
	}{
		{I(Left), "by bytes", s.BySize, true},
		{F(float64(Left) + colw + 30), "by repos", s.ByRepo, false},
	}

	for gi, g := range groups {
		p.WriteString(`<g opacity="0">`)
		p.WriteString(fade(delay(0.10, 0.10, gi)))
		p.WriteString(label(g.gx, I(12), strings.ToUpper(g.title), 9,
			opts{cls: "m-f", extra: ` letter-spacing="1.3"`}))
		p.WriteString(`</g>`)
		if len(g.data) == 0 {
			continue
		}

		top, total := 0, 0
		for _, c := range g.data {
			if c.Value > top {
				top = c.Value
			}
			total += c.Value
		}
		if top == 0 {
			top = 1
		}
		if total == 0 {
			total = 1
		}

		cid := fmt.Sprintf("rl%d", gi)
		clip, cursor := wipe(cid, g.gx.Add(I(nameW)), I(20), F(barMax),
			I(rows*22), delay(0.34, 0.12, gi), 0.95)
		p.WriteString(clip)

		for ri, c := range g.data {
			y := 26 + ri*22
			shown := strconv.Itoa(c.Value)
			if g.asPct {
				shown = fmt.Sprintf("%.0f%%", float64(c.Value)/float64(total)*100)
			}
			name := strings.ToLower(c.Name)
			if len(name) > 11 {
				name = name[:11]
			}
			p.WriteString(`<g opacity="0">`)
			p.WriteString(fade(delay(delay(0.24, 0.10, gi), 0.05, ri)))
			p.WriteString(label(g.gx, I(y+8), name, 11, opts{cls: "e-f"}))
			p.WriteString(label(F(g.gx.v+colw-6), I(y+8), shown, 11,
				opts{cls: "m-f", anchor: "end"}))
			p.WriteString(`</g>`)
			fmt.Fprintf(&p, `<g clip-path="url(#%s)">`, cid)
			p.WriteString(hbar(g.gx.v+nameW, float64(y), barMax, 7, "w"))
			p.WriteString(hbar(g.gx.v+nameW, float64(y), barMax*float64(c.Value)/float64(top), 7, ""))
			p.WriteString(`</g>`)
		}
		p.WriteString(cursor)
	}

	p.WriteString(`</svg>`)
	return p.String()
}

// drawWeekday: contributions by weekday, Monday first.
func drawWeekday(s Summary) string {
	labels := []string{"mon", "tue", "wed", "thu", "fri", "sat", "sun"}
	totals := s.ByWeekday
	if len(totals) == 0 {
		totals = make([]int, 7)
	}

	const (
		H      = 20 + 7*20 + 6
		nameW  = 40
		valW   = 34
		barMax = Width - Left - nameW - valW - 10
	)

	var p strings.Builder
	p.WriteString(head(Width, H, ""))
	p.WriteString(`<g opacity="0">`)
	p.WriteString(fade(0.10))
	p.WriteString(label(I(Left), I(12), "BY WEEKDAY", 9,
		opts{cls: "m-f", extra: ` letter-spacing="1.3"`}))
	p.WriteString(`</g>`)

	top := 0
	for _, v := range totals {
		if v > top {
			top = v
		}
	}
	if top == 0 {
		top = 1
	}

	clip, cursor := wipe("rw", I(Left+nameW), I(20), I(barMax), I(7*20), 0.34, 1.05)
	p.WriteString(clip)

	for i, lab := range labels {
		val := totals[i]
		y := 22 + i*20
		cls := "m-f"
		if val == top && val != 0 {
			cls = "e-f"
		}
		p.WriteString(`<g opacity="0">`)
		p.WriteString(fade(delay(0.24, 0.05, i)))
		p.WriteString(label(I(Left), I(y+8), lab, 11, opts{cls: cls}))
		p.WriteString(label(I(Width), I(y+8), strconv.Itoa(val), 11,
			opts{cls: "m-f", anchor: "end"}))
		p.WriteString(`</g>`)
		p.WriteString(`<g clip-path="url(#rw)">`)
		p.WriteString(hbar(Left+nameW, float64(y), barMax, 7, "w"))
		p.WriteString(hbar(Left+nameW, float64(y), barMax*float64(val)/float64(top), 7, ""))
		p.WriteString(`</g>`)
	}

	p.WriteString(cursor)
	p.WriteString(`</svg>`)
	return p.String()
}
