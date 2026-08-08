package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

// write only touches path when the bytes differ, so an unchanged graphic
// never produces an empty commit.
func write(path, svg string) (bool, error) {
	if old, err := os.ReadFile(path); err == nil && string(old) == svg {
		return false, nil
	}
	return true, os.WriteFile(path, []byte(svg), 0o644)
}

func load() (*User, error) {
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		return nil, fmt.Errorf("GITHUB_TOKEN is not set")
	}
	return fetch(login(), token)
}

func login() string {
	if v := os.Getenv("GH_LOGIN"); v != "" {
		return v
	}
	return "0xmthan"
}

func run() error {
	user, err := load()
	if err != nil {
		return err
	}
	s := summarise(user)

	outDir := os.Getenv("OUT_DIR")
	if outDir == "" {
		outDir = "assets"
	}
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return err
	}

	files := map[string]string{
		"stats.svg":    drawStats(s),
		"about.svg":    drawAbout(aboutCard()),
		"social.svg":   drawSocial(s),
		"streak.svg":   drawStreak(s),
		"langs.svg":    drawLangs(s),
		"weekday.svg":  drawWeekday(s),
		"year.svg":     drawYear(s),
		"repos.svg":    drawRepos(s),
		"activity.svg": drawActivity(s),
	}
	for _, word := range []string{"about", "projects", "social", "activity", "stats"} {
		files["hd-"+word+".svg"] = drawHeading(word)
	}

	var changed []string
	for name, svg := range files {
		didWrite, err := write(filepath.Join(outDir, name), svg)
		if err != nil {
			return err
		}
		if didWrite {
			changed = append(changed, name)
		}
	}
	sort.Strings(changed)

	fmt.Printf("%d contributions, %d active days, best week %d\n",
		s.Total, s.Active, s.BestWeek)
	if len(changed) == 0 {
		fmt.Println("updated: nothing")
	} else {
		fmt.Printf("updated: %s\n", joinComma(changed))
	}
	return nil
}

func joinComma(v []string) string {
	out := ""
	for i, s := range v {
		if i > 0 {
			out += ", "
		}
		out += s
	}
	return out
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
