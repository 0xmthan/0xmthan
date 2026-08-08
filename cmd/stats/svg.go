package main

import (
	"encoding/base64"
	"fmt"
	"math"
	"strconv"
	"strings"
)

const (
	Width  = 620
	Left   = 34
	Reveal = 1.30
)

var mon = []string{"jan", "feb", "mar", "apr", "may", "jun",
	"jul", "aug", "sep", "oct", "nov", "dec"}

type theme struct{ data, emph, dim, rule, surface string }

var (
	light = theme{data: "#6e7681", emph: "#424a53", dim: "#8c959f",
		rule: "#d8dee4", surface: "#ffffff"}
	dark = theme{data: "#c9d1d9", emph: "#f0f6fc", dim: "#8b949e",
		rule: "#30363d", surface: "#0d1117"}
)

const mono = "JBMono,ui-monospace,SFMono-Regular,Menlo,Consolas," +
	"&apos;Liberation Mono&apos;,monospace"

// num carries whether a coordinate came from integer arithmetic, because the
// SVG text differs: an int renders as "620", a float as "620.0".
type num struct {
	v     float64
	isInt bool
}

func I(v int) num     { return num{float64(v), true} }
func F(v float64) num { return num{v, false} }

// Add promotes to float if either side is float.
func (n num) Add(m num) num { return num{n.v + m.v, n.isInt && m.isInt} }

func (n num) String() string {
	if n.isInt {
		return strconv.Itoa(int(n.v))
	}
	return f2s(n.v)
}

// f2s formats shortest round-trip but always keeps a decimal point, so 152.0
// does not print as "152".
func f2s(v float64) string {
	s := strconv.FormatFloat(v, 'g', -1, 64)
	if !strings.ContainsAny(s, ".eE") {
		s += ".0"
	}
	return s
}

// delay is the animation pattern base+step*i, with the product rounded before
// the add.
//
// The conversion is load-bearing. On arm64 the compiler fuses x*y+z into a
// single FMA, keeping the product at extra precision and landing one ULP off.
// An unguarded a+b*c renders "1.07" where it should say "1.08". Every product
// feeding an addition in this package is wrapped the same way.
func delay(base, step float64, i int) float64 {
	return base + float64(step*float64(i))
}

// advance is a rough monospace advance width, for layout maths only. The
// result is rounded before returning so callers cannot fuse it into their add.
func advance(text string, fs int) float64 {
	return float64(float64(len(text)*fs) * 0.6)
}

// face inlines a subset as a data URI. An external font URL cannot work: these
// SVGs load through <img>, and browsers refuse subresources for an image
// document. Inlining also pins the 0.600em advance the year grid assumes.
func face(blob []byte, weight int) string {
	return fmt.Sprintf("@font-face{font-family:JBMono;font-style:normal;"+
		"font-weight:%d;font-display:block;"+
		"src:url(data:font/woff2;base64,%s) format('woff2')}",
		weight, base64.StdEncoding.EncodeToString(blob))
}

func block(t theme) string {
	return fmt.Sprintf(".d-f{fill:%s}.d-s{stroke:%s}.e-f{fill:%s}"+
		".m-f{fill:%s}.u-s{stroke:%s}.r{stroke:%s}",
		t.data, t.data, t.emph, t.dim, t.rule, t.surface)
}

// style emits both palettes; the dark one is selected by the viewer's OS
// preference, since an <img>-loaded SVG cannot see GitHub's theme setting.
func style(extra, font string) string {
	if font == "" {
		font = fontText()
	}
	return "<style>" + font + block(light) +
		fmt.Sprintf(".w{fill:%s;opacity:.13}", light.data) + extra +
		"@media(prefers-color-scheme:dark){" + block(dark) +
		fmt.Sprintf(".w{fill:%s;opacity:.16}}", dark.data) + "</style>"
}

func head(w, h int, font string) string {
	return fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" width="%d" `+
		`height="%d" viewBox="0 0 %d %d" fill="none" font-family="%s">`,
		w, h, w, h, mono) + style("", font)
}

func fade(delay float64, dur ...float64) string {
	d := 0.45
	if len(dur) > 0 {
		d = dur[0]
	}
	return fmt.Sprintf(`<animate attributeName="opacity" from="0" to="1" `+
		`begin="%.2fs" dur="%ss" fill="freeze"/>`, delay, f2s(d))
}

// wipe returns a clipPath reveal plus the cursor block riding its edge.
func wipe(cid string, x, y, w, h num, delay float64, dur ...float64) (string, string) {
	d := Reveal
	if len(dur) > 0 {
		d = dur[0]
	}
	clip := fmt.Sprintf(`<clipPath id="%s"><rect x="%s" y="%s" height="%s" `+
		`width="0"><animate attributeName="width" from="0" to="%s" `+
		`begin="%.2fs" dur="%ss" fill="freeze"/></rect></clipPath>`,
		cid, x, y, h, w, delay, f2s(d))
	cursor := fmt.Sprintf(`<rect y="%s" width="2" height="%s" class="d-f" `+
		`opacity="0"><animate attributeName="x" from="%s" to="%s" `+
		`begin="%.2fs" dur="%ss" fill="freeze"/>`+
		`<set attributeName="opacity" to="0.55" begin="%.2fs"/>`+
		`<set attributeName="opacity" to="0" begin="%.2fs"/></rect>`,
		y, h, x, x.Add(w), delay, f2s(d), delay, delay+d)
	return clip, cursor
}

// opts stands in for label's Python keyword arguments.
type opts struct {
	cls    string // default "m-f"
	anchor string // default "start"
	extra  string
}

func label(x, y num, text string, size int, o opts) string {
	cls := o.cls
	if cls == "" {
		cls = "m-f"
	}
	anchor := ""
	if o.anchor != "" && o.anchor != "start" {
		anchor = fmt.Sprintf(` text-anchor="%s"`, o.anchor)
	}
	return fmt.Sprintf(`<text x="%s" y="%s" class="%s" font-size="%d"%s%s>%s</text>`,
		x, y, cls, size, anchor, o.extra, text)
}

// hbar is a horizontal bar: rounded data-end, square baseline. Bars too short
// to round are dropped rather than drawn as a nub.
func hbar(x, y, w, h float64, cls string) string {
	if cls == "" {
		cls = "d-f"
	}
	if w <= 0.6 {
		return ""
	}
	r := math.Min(math.Min(3.0, h/2.0), w)
	return fmt.Sprintf(`<path d="M%.1f %.1fH%.1f`+
		`Q%.1f %.1f %.1f %.1f`+
		`V%.1fQ%.1f %.1f %.1f %.1f`+
		`H%.1fZ" class="%s"/>`,
		x, y, x+w-r, x+w, y, x+w, y+r,
		y+h-r, x+w, y+h, x+w-r, y+h, x, cls)
}

// kv is one "label value" line kept in a single text node, so the two can
// never drift out of baseline alignment.
func kv(x, y num, k, v string, size int) string {
	return fmt.Sprintf(`<text x="%s" y="%s" font-size="%d">`+
		`<tspan class="m-f">%s</tspan>`+
		`<tspan class="e-f" dx="6">%s</tspan></text>`, x, y, size, k, v)
}

type placedTag struct {
	text    string
	x, y, w float64
}

// flow lays items left to right, wrapping past maxW, and reports where the
// last row ends so callers can size their canvas to it.
func flow(items []string, x0, y0, maxW, h, gapY float64) ([]placedTag, float64) {
	const (
		gapX = 8.0
		padX = 10.0
		fs   = 10
	)
	var placed []placedTag
	x, y, started := x0, y0, false
	for _, text := range items {
		w := advance(text, fs) + padX*2
		if started && x+w > x0+maxW {
			x, y = x0, y+h+gapY
		}
		placed = append(placed, placedTag{text, x, y, w})
		x += w + gapX
		started = true
	}
	if len(placed) == 0 {
		return placed, y0
	}
	return placed, y + h
}
