package main

import (
	"fmt"
	"strings"
)

// About is the profile card's copy. Nothing here comes from the API. Every
// block sizes to its contents, so adding a fact or a tag just grows the card.
type About struct {
	Name  string
	Title string
	Bio   []string
	Facts [][2]string
	Tags  []string
}

// aboutCard is the copy for the profile card. Two rules when editing: write
// ampersands as &amp;, since the text is emitted into XML as-is, and keep to
// ASCII, since the embedded font is subset to basic latin and anything outside
// it falls back to the viewer's own monospace.
func aboutCard() About {
	return About{
		Name:  login(),
		Title: "Muhammed T.",
		Bio: []string{
			"Computer engineering student and 42 cadet.",
			"I like gathering data and self-hosting stuff.",
			"You can stalk me more by running " + hi("ssh mthan.dev") +
				" in your terminal :)",
		},
		Facts: [][2]string{
			{"based in", "Istanbul, TR"},
			{"focus", "data &amp; self-hosting"},
			{"working on", "42 common core"},
			{"reachable at", "me@mthan.dev"},
		},
		Tags: []string{"c", "go", "python", "typescript",
			"kotlin", "dart", "shell", "linux"},
	}
}

// hi marks a run of bio text as emphasised ink: near-white in dark mode,
// near-black in light. A literal white would disappear on light backgrounds.
func hi(s string) string {
	return `<tspan class="e-f">` + s + `</tspan>`
}

// drawAbout: name, bio, quick facts, tag cloud.
func drawAbout(cfg About) string {
	const (
		nameFS  = 26
		bioLH   = 18
		factLH  = 20
		tagH    = 20
		tagGap  = 10
		yName   = 32
		tagFS   = 10
		bodyPad = 14
	)

	y := yName + 14
	bioStart := y + 12
	if len(cfg.Bio) > 0 {
		y = bioStart + len(cfg.Bio)*bioLH
	}

	factsStart := y + 6
	if len(cfg.Bio) > 0 {
		factsStart = y + bodyPad
	}
	factRows := (len(cfg.Facts) + 1) / 2
	if len(cfg.Facts) > 0 {
		y = factsStart + factRows*factLH
	}

	tagsStart := y + 6
	if len(cfg.Facts) > 0 {
		tagsStart = y + bodyPad
	}

	var tagRows []placedTag
	tagsBottom := float64(tagsStart)
	if len(cfg.Tags) > 0 {
		tagRows, tagsBottom = flow(cfg.Tags, Left, float64(tagsStart),
			Width-Left-6, tagH, tagGap)
	}

	bottom := float64(y)
	if len(cfg.Tags) > 0 {
		bottom = tagsBottom
	}
	H := int(bottom + bodyPad)

	dName, dTitle := 0.10, 0.24
	dBio := dTitle + 0.14
	dFacts := dBio
	if len(cfg.Bio) > 0 {
		dFacts = dBio + (float64(float64(len(cfg.Bio))*0.08) + 0.14)
	}
	dTags := dFacts
	if len(cfg.Facts) > 0 {
		dTags = dFacts + (float64(float64(factRows)*0.06) + 0.14)
	}

	var p strings.Builder
	p.WriteString(head(Width, H, ""))

	p.WriteString(`<g opacity="0">`)
	p.WriteString(fade(dName))
	p.WriteString(label(I(0), I(yName), cfg.Name, nameFS,
		opts{cls: "e-f", extra: ` font-weight="600"`}))
	p.WriteString(`</g>`)

	if cfg.Title != "" {
		tx := advance(cfg.Name, nameFS) + 14
		p.WriteString(`<g opacity="0">`)
		p.WriteString(fade(dTitle))
		p.WriteString(label(F(tx), I(yName), cfg.Title, 13, opts{}))
		p.WriteString(`</g>`)
	}

	for i, line := range cfg.Bio {
		p.WriteString(`<g opacity="0">`)
		p.WriteString(fade(delay(dBio, 0.08, i)))
		p.WriteString(label(I(0), I(bioStart+i*bioLH), line, 12, opts{}))
		p.WriteString(`</g>`)
	}

	if len(cfg.Facts) > 0 {
		colw := float64(Width-Left) / 2
		for i, f := range cfg.Facts {
			col, row := i%2, i/2
			x := I(Left)
			if col != 0 {
				x = F(float64(Left) + colw)
			}
			fy := factsStart + row*factLH
			p.WriteString(`<g opacity="0">`)
			p.WriteString(fade(delay(dFacts, 0.06, row)))
			p.WriteString(kv(x, I(fy), f[0], f[1], 11))
			p.WriteString(`</g>`)
		}
	}

	for i, t := range tagRows {
		fmt.Fprintf(&p, `<g opacity="0">%s`+
			`<rect x="%.1f" y="%.1f" width="%.1f" height="%d" rx="%.0f" `+
			`class="w u-s" stroke-width="1"/>`+
			`<text x="%.1f" y="%.1f" font-size="%d" class="d-f" `+
			`text-anchor="middle">%s</text></g>`,
			fade(delay(dTags, 0.045, i)),
			t.x, t.y, t.w, tagH, float64(tagH)/2,
			t.x+t.w/2, t.y+float64(tagH*0.67), tagFS, t.text)
	}

	p.WriteString(`</svg>`)
	return p.String()
}
