package main

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

var ramp = []string{" ", ":", "+", "#", "@"}

// drawYear: seven rows by fifty-three weeks, intensity as a character.
func drawYear(s Summary) string {
	const (
		fs   = 9.2
		lh   = 11.0
		colW = 2
		padL = Left
		padT = 44
	)
	cw := fs * 0.6
	weeks := s.Weeks
	H := int(padT + 7*lh + 26)

	level := func(v int) int {
		for i, cut := range []int{0, 2, 5, 9} {
			if v <= cut {
				return i
			}
		}
		return 4
	}

	totalDays := 0
	for _, w := range weeks {
		totalDays += len(w)
	}

	var p strings.Builder
	p.WriteString(head(Width, H, ""))
	p.WriteString(`<g opacity="0">` + fade(0.10) +
		label(I(padL), I(16), "THE YEAR", 9,
			opts{cls: "m-f", extra: ` letter-spacing="1.3"`}) +
		label(I(padL), I(32), strconv.Itoa(s.Active)+" of "+
			strconv.Itoa(totalDays)+" days had a contribution", 11, opts{}) +
		`</g>`)

	lx := Width - 6
	p.WriteString(`<g opacity="0">` + fade(1.30) +
		label(I(lx-78), I(32), "less", 9, opts{cls: "m-f", anchor: "end"}) +
		fmt.Sprintf(`<text xml:space="preserve" x="%d" y="32" class="d-f" `+
			`font-size="%s">%s</text>`, lx-72, f2s(fs),
			strings.Join(ramp[1:], " ")) +
		label(I(lx), I(32), "more", 9, opts{cls: "m-f", anchor: "end"}) +
		`</g>`)

	for r := 0; r < 7; r++ {
		var b strings.Builder
		for _, w := range weeks {
			v := 0
			for _, d := range w {
				if d.Weekday == r {
					v = d.ContributionCount
					break
				}
			}
			b.WriteString(strings.Repeat(ramp[level(v)], colW))
		}
		line := strings.TrimRightFunc(b.String(), unicode.IsSpace)
		if line == "" {
			continue
		}

		y := float64(padT) + float64(float64(r)*lh)
		wpx := float64(max(len(line), 1)) * cw
		cid := fmt.Sprintf("ry%d", r)
		d := delay(0.30, 0.07, r)

		p.WriteString(fmt.Sprintf(`<clipPath id="%s"><rect x="%d" y="%s" `+
			`height="%s" width="0"><animate attributeName="width" `+
			`from="0" to="%.1f" begin="%.2fs" dur="0.40s" `+
			`fill="freeze"/></rect></clipPath>`,
			cid, padL, f2s(y), f2s(lh), wpx, d))

		safe := strings.ReplaceAll(line, "&", "&amp;")
		safe = strings.ReplaceAll(safe, "<", "&lt;")
		p.WriteString(fmt.Sprintf(`<g clip-path="url(#%s)"><text `+
			`xml:space="preserve" x="%d" y="%.1f" class="d-f" `+
			`font-size="%s">%s</text></g>`,
			cid, padL, y+fs-0.6, f2s(fs), safe))
	}

	for _, wd := range []struct {
		r   int
		lab string
	}{{1, "mon"}, {3, "wed"}, {5, "fri"}} {
		p.WriteString(label(I(padL-7),
			F(float64(padT)+float64(float64(wd.r)*lh)+fs-0.6), wd.lab, 9,
			opts{cls: "m-f", anchor: "end"}))
	}

	// Month ticks, skipped when two would collide.
	lastM, lastX := -1, -999.0
	baseY := float64(padT) + 7*lh + 13
	for i, w := range weeks {
		if len(w) == 0 {
			continue
		}
		m, err := strconv.Atoi(w[0].Date[5:7])
		if err != nil {
			continue
		}
		x := float64(padL) + float64(float64(i*colW)*cw)
		if m != lastM && i < len(weeks)-1 && x-lastX >= 34 {
			p.WriteString(label(F(x), F(baseY), mon[m-1], 9, opts{cls: "m-f"}))
			lastX = x
		}
		lastM = m
	}

	p.WriteString(`</svg>`)
	return p.String()
}
