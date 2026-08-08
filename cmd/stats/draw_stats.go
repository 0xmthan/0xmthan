package main

import (
	"fmt"
	"strconv"
	"strings"
)

// drawStats: hero number, two secondary counts, weekly sparkline.
func drawStats(s Summary) string {
	const H = 148

	weekly := s.Weekly
	if len(weekly) == 0 {
		weekly = []int{0}
	}
	peak := 0
	for _, v := range weekly {
		if v > peak {
			peak = v
		}
	}
	if peak == 0 {
		peak = 1
	}

	var p strings.Builder
	p.WriteString(head(Width, H, ""))

	p.WriteString(`<g opacity="0">` + fade(0.10) +
		label(I(0), I(50), strconv.Itoa(s.Total), 52,
			opts{cls: "e-f", extra: ` font-weight="600"`}) +
		label(I(0), I(72), "contributions in the last year", 12, opts{}) +
		`</g>`)

	secondary := []struct {
		val int
		lab string
	}{{s.Active, "active days"}, {s.BestWeek, "best week"}}
	for i, sec := range secondary {
		p.WriteString(`<g opacity="0">` + fade(delay(0.30, 0.12, i)) +
			label(I(Width), I(30+i*40), strconv.Itoa(sec.val), 19,
				opts{cls: "e-f", anchor: "end", extra: ` font-weight="600"`}) +
			label(I(Width), I(47+i*40), sec.lab, 11,
				opts{cls: "m-f", anchor: "end"}) + `</g>`)
	}

	base, top := H-10, H-58
	span := base - top
	step := float64(Width) / float64(max(len(weekly)-1, 1))

	type pt struct{ x, y float64 }
	pts := make([]pt, len(weekly))
	for i, v := range weekly {
		pts[i] = pt{
			x: float64(i) * step,
			y: float64(base) - float64((float64(v)/float64(peak))*float64(span)),
		}
	}

	for _, y := range []float64{float64(top), float64(top) + float64(span)/2, float64(base)} {
		p.WriteString(fmt.Sprintf(`<line x1="0" y1="%.1f" x2="%d" y2="%.1f" `+
			`class="u-s" stroke-width="1" stroke-dasharray="2 4" opacity="0">`+
			`%s</line>`, y, Width, y, fade(0.15)))
	}

	clip, cursor := wipe("rs", I(0), I(top-6), I(Width), I(span+8), 0.50)
	p.WriteString(clip)
	p.WriteString(`<g clip-path="url(#rs)">`)

	var area strings.Builder
	fmt.Fprintf(&area, `<path d="M%.1f %.1f`, pts[0].x, float64(base))
	for _, q := range pts {
		fmt.Fprintf(&area, "L%.1f %.1f", q.x, q.y)
	}
	fmt.Fprintf(&area, `L%.1f %.1fZ" class="w"/>`, pts[len(pts)-1].x, float64(base))
	p.WriteString(area.String())

	var line strings.Builder
	fmt.Fprintf(&line, `<path d="M%.1f %.1f`, pts[0].x, pts[0].y)
	for _, q := range pts[1:] {
		fmt.Fprintf(&line, "L%.1f %.1f", q.x, q.y)
	}
	line.WriteString(`" class="d-s" stroke-width="2" stroke-linejoin="round" ` +
		`stroke-linecap="round"/>`)
	p.WriteString(line.String())

	p.WriteString(`</g>`)
	p.WriteString(cursor)

	end := pts[len(pts)-1]
	p.WriteString(fmt.Sprintf(`<circle cx="%.1f" cy="%.1f" r="4.5" `+
		`class="e-f r" stroke-width="2" opacity="0">%s</circle>`,
		end.x-2, end.y, fade(0.50+Reveal, 0.35)))

	p.WriteString(`</svg>`)
	return p.String()
}
