package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"time"
)

const api = "https://api.github.com/graphql"

const query = `
query($login: String!, $from: DateTime!, $to: DateTime!) {
  user(login: $login) {
    followers { totalCount }
    following { totalCount }
    contributionsCollection(from: $from, to: $to) {
      totalCommitContributions
      totalIssueContributions
      totalPullRequestContributions
      totalPullRequestReviewContributions
      contributionCalendar {
        totalContributions
        weeks { contributionDays { contributionCount date weekday } }
      }
    }
    repositories(first: 100, ownerAffiliations: OWNER, isFork: false,
                 privacy: PUBLIC) {
      nodes {
        name
        stargazerCount
        languages(first: 12, orderBy: {field: SIZE, direction: DESC}) {
          edges { size node { name } }
        }
      }
    }
  }
}
`

type Day struct {
	ContributionCount int    `json:"contributionCount"`
	Date              string `json:"date"`
	Weekday           int    `json:"weekday"`
}

type Repo struct {
	Name           string `json:"name"`
	StargazerCount int    `json:"stargazerCount"`
	Languages      struct {
		Edges []struct {
			Size int `json:"size"`
			Node struct {
				Name string `json:"name"`
			} `json:"node"`
		} `json:"edges"`
	} `json:"languages"`
}

type User struct {
	Followers struct {
		TotalCount int `json:"totalCount"`
	} `json:"followers"`
	Following struct {
		TotalCount int `json:"totalCount"`
	} `json:"following"`
	ContributionsCollection struct {
		TotalCommitContributions            int `json:"totalCommitContributions"`
		TotalIssueContributions             int `json:"totalIssueContributions"`
		TotalPullRequestContributions       int `json:"totalPullRequestContributions"`
		TotalPullRequestReviewContributions int `json:"totalPullRequestReviewContributions"`
		ContributionCalendar                struct {
			TotalContributions int `json:"totalContributions"`
			Weeks              []struct {
				ContributionDays []Day `json:"contributionDays"`
			} `json:"weeks"`
		} `json:"contributionCalendar"`
	} `json:"contributionsCollection"`
	Repositories struct {
		Nodes []Repo `json:"nodes"`
	} `json:"repositories"`
}

type response struct {
	Data struct {
		User *User `json:"user"`
	} `json:"data"`
	Errors json.RawMessage `json:"errors"`
}

// window is the trailing 365 days.
func window() (string, string) {
	today := time.Now().UTC()
	start := today.AddDate(0, 0, -364)
	return start.Format("2006-01-02") + "T00:00:00Z",
		today.Format("2006-01-02") + "T23:59:59Z"
}

func fetch(login, token string) (*User, error) {
	since, until := window()
	body, err := json.Marshal(map[string]any{
		"query": query,
		"variables": map[string]string{
			"login": login, "from": since, "to": until,
		},
	})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, api, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", login+"-profile-stats")

	client := &http.Client{Timeout: 30 * time.Second}
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("github returned %s", res.Status)
	}
	var payload response
	if err := json.NewDecoder(res.Body).Decode(&payload); err != nil {
		return nil, err
	}
	if len(payload.Errors) > 0 {
		return nil, fmt.Errorf("GraphQL errors: %s", payload.Errors)
	}
	if payload.Data.User == nil {
		return nil, fmt.Errorf("no such user: %s", login)
	}
	return payload.Data.User, nil
}

// Streak is a run of consecutive days with at least one contribution.
type Streak struct {
	Length     int
	Start, End string
}

type Count struct {
	Name  string
	Value int
}

type RepoStat struct {
	Name  string
	Stars int
	Lang  string
}

// Summary is everything the graphics draw, derived from one API response.
type Summary struct {
	Total    int
	Active   int
	BestWeek int
	Weekly   []int
	Weeks    [][]Day

	Current Streak
	Longest Streak

	BySize   []Count
	ByRepo   []Count
	TopRepos []RepoStat
	Activity []Count

	ByWeekday []int

	Followers int
	Following int
	Stars     int
}

// streaks finds the current and longest runs. A zero on the final day does
// not break the current streak — the day isn't over yet. Any earlier zero does.
func streaks(days []Day) (current, longest Streak) {
	run, runStart := 0, ""
	for _, d := range days {
		if d.ContributionCount > 0 {
			run++
			if runStart == "" {
				runStart = d.Date
			}
			if run > longest.Length {
				longest = Streak{Length: run, Start: runStart, End: d.Date}
			}
		} else {
			run, runStart = 0, ""
		}
	}

	tail := days
	if n := len(days); n > 0 && days[n-1].ContributionCount == 0 {
		tail = days[:n-1]
	}
	for i := len(tail) - 1; i >= 0; i-- {
		if tail[i].ContributionCount == 0 {
			break
		}
		current.Length++
		current.Start = tail[i].Date
		if current.End == "" {
			current.End = tail[i].Date
		}
	}
	return current, longest
}

// rank sorts by count, ties by name so runs never reshuffle. Top five.
func rank(m map[string]int) []Count {
	out := make([]Count, 0, len(m))
	for name, v := range m {
		out = append(out, Count{name, v})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Value != out[j].Value {
			return out[i].Value > out[j].Value
		}
		return out[i].Name < out[j].Name
	})
	if len(out) > 5 {
		out = out[:5]
	}
	return out
}

func languages(repos []Repo) (bySize, byRepo []Count) {
	sizes, counts := map[string]int{}, map[string]int{}
	for _, r := range repos {
		edges := r.Languages.Edges
		for _, e := range edges {
			sizes[e.Node.Name] += e.Size
		}
		if len(edges) > 0 {
			counts[edges[0].Node.Name]++
		}
	}
	return rank(sizes), rank(counts)
}

func topRepos(repos []Repo, n int) []RepoStat {
	rows := make([]RepoStat, 0, len(repos))
	for _, r := range repos {
		lang := ""
		if len(r.Languages.Edges) > 0 {
			lang = r.Languages.Edges[0].Node.Name
		}
		rows = append(rows, RepoStat{r.Name, r.StargazerCount, lang})
	}
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].Stars != rows[j].Stars {
			return rows[i].Stars > rows[j].Stars
		}
		return rows[i].Name < rows[j].Name
	})
	if len(rows) > n {
		rows = rows[:n]
	}
	return rows
}

// weekdayTotals sums by weekday, Monday first. GitHub's weekday is 0=Sunday,
// and the extra +7 is required: Go's % keeps the dividend's sign, so a bare
// (0-1)%7 is -1 and panics on index where Python's floored modulo gives 6.
func weekdayTotals(days []Day) []int {
	totals := make([]int, 7)
	for _, d := range days {
		totals[((d.Weekday-1)%7+7)%7] += d.ContributionCount
	}
	return totals
}

// activityBreakdown orders the four kinds high to low. The sort must be
// stable: equal counts keep the declared order, as Python's does.
func activityBreakdown(u *User) []Count {
	cc := u.ContributionsCollection
	rows := []Count{
		{"commits", cc.TotalCommitContributions},
		{"pull requests", cc.TotalPullRequestContributions},
		{"reviews", cc.TotalPullRequestReviewContributions},
		{"issues", cc.TotalIssueContributions},
	}
	sort.SliceStable(rows, func(i, j int) bool { return rows[i].Value > rows[j].Value })
	return rows
}

func summarise(u *User) Summary {
	cal := u.ContributionsCollection.ContributionCalendar

	weeks := make([][]Day, 0, len(cal.Weeks))
	weekly := make([]int, 0, len(cal.Weeks))
	var days []Day
	active, best := 0, 0
	for _, w := range cal.Weeks {
		weeks = append(weeks, w.ContributionDays)
		days = append(days, w.ContributionDays...)
		sum := 0
		for _, d := range w.ContributionDays {
			sum += d.ContributionCount
			if d.ContributionCount > 0 {
				active++
			}
		}
		weekly = append(weekly, sum)
		if sum > best {
			best = sum
		}
	}

	repos := u.Repositories.Nodes
	bySize, byRepo := languages(repos)
	current, longest := streaks(days)

	stars := 0
	for _, r := range repos {
		stars += r.StargazerCount
	}

	return Summary{
		Total:     cal.TotalContributions,
		Active:    active,
		BestWeek:  best,
		Weekly:    weekly,
		Weeks:     weeks,
		Current:   current,
		Longest:   longest,
		BySize:    bySize,
		ByRepo:    byRepo,
		TopRepos:  topRepos(repos, 5),
		Activity:  activityBreakdown(u),
		ByWeekday: weekdayTotals(days),
		Followers: u.Followers.TotalCount,
		Following: u.Following.TotalCount,
		Stars:     stars,
	}
}

// pretty renders an ISO date as the streak card prints it: "may 10".
func pretty(iso string) string {
	t, err := time.Parse("2006-01-02", iso)
	if err != nil {
		return iso
	}
	return fmt.Sprintf("%s %d", mon[int(t.Month())-1], t.Day())
}
