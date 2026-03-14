package main

import (
	"bufio"
	"encoding/json"
	"math"
	"os"
	"path/filepath"
	"sort"
	"time"
)

type PastedContent struct {
	Content string `json:"content"`
}

type Entry struct {
	Display        string                    `json:"display"`
	PastedContents map[string]PastedContent `json:"pastedContents"`
	Timestamp      int64                     `json:"timestamp"`
	Project        string                    `json:"project"`
	SessionID      string                    `json:"sessionId"`
}

func (e *Entry) EstimateChars() int {
	n := len(e.Display)
	for _, p := range e.PastedContents {
		n += len(p.Content)
	}
	return n
}

type ProjectStats struct {
	Name       string  `json:"name"`
	Prompts    int     `json:"prompts"`
	Sessions   int     `json:"sessions"`
	First      string  `json:"first_date"`
	Last       string  `json:"last_date"`
	ActiveDays int     `json:"active_days"`
	Intensity  float64 `json:"intensity"`
}

type collector struct {
	name     string
	prompts  int
	first    time.Time
	last     time.Time
	sessions map[string]struct{}
	days     map[string]struct{}
}

func ParseHistory() ([]ProjectStats, error) {
	historyPath := os.Getenv("HOME") + "/.claude/history.jsonl"

	file, err := os.Open(historyPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	collectors := make(map[string]*collector)

	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)

	for scanner.Scan() {
		var entry Entry
		if err := json.Unmarshal(scanner.Bytes(), &entry); err != nil {
			continue
		}

		name := filepath.Base(entry.Project)
		ts := time.UnixMilli(entry.Timestamp)

		c, ok := collectors[name]
		if !ok {
			c = &collector{
				name:     name,
				first:    ts,
				last:     ts,
				sessions: make(map[string]struct{}),
				days:     make(map[string]struct{}),
			}
			collectors[name] = c
		}

		c.prompts++
		c.sessions[entry.SessionID] = struct{}{}
		c.days[ts.Format("2006-01-02")] = struct{}{}

		if ts.Before(c.first) {
			c.first = ts
		}
		if ts.After(c.last) {
			c.last = ts
		}
	}

	var results []ProjectStats
	for _, c := range collectors {
		totalDays := int(c.last.Sub(c.first).Hours()/24) + 1
		if totalDays < 1 {
			totalDays = 1
		}
		intensity := math.Round(float64(c.prompts)/float64(totalDays)*100) / 100

		results = append(results, ProjectStats{
			Name:       c.name,
			Prompts:    c.prompts,
			Sessions:   len(c.sessions),
			First:      c.first.Format("2006-01-02"),
			Last:       c.last.Format("2006-01-02"),
			ActiveDays: len(c.days),
			Intensity:  intensity,
		})
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Prompts > results[j].Prompts
	})

	return results, nil
}

// ── Timeline analytics ──

type WeekStats struct {
	Week     string            `json:"week"`      // "2026-W07"
	Label    string            `json:"label"`      // "Feb 10 – Feb 16"
	Prompts  int               `json:"prompts"`
	Tokens   int               `json:"tokens"`
	Sessions int               `json:"sessions"`
	Projects map[string]int    `json:"-"`
	TopProjects []ProjectCount `json:"top_projects"`
}

type ProjectCount struct {
	Name    string `json:"name"`
	Prompts int    `json:"prompts"`
}

type MonthStats struct {
	Month   string `json:"month"`   // "2026-02"
	Label   string `json:"label"`   // "Feb 2026"
	Prompts int    `json:"prompts"`
	Tokens  int    `json:"tokens"`
	Days    int    `json:"active_days"`
}

type Timeline struct {
	TotalTokens    int          `json:"total_tokens"`
	TotalWeeks     int          `json:"total_weeks"`
	ActiveWeeks    int          `json:"active_weeks"`
	AvgPerWeek     float64      `json:"avg_prompts_per_week"`
	AvgTokensWeek  float64      `json:"avg_tokens_per_week"`
	Weeks          []WeekStats  `json:"weeks"`
	Months         []MonthStats `json:"months"`
}

func ParseTimeline() (*Timeline, error) {
	historyPath := os.Getenv("HOME") + "/.claude/history.jsonl"

	file, err := os.Open(historyPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	weeks := make(map[string]*WeekStats)
	months := make(map[string]*MonthStats)
	monthDays := make(map[string]map[string]struct{})
	allSessions := make(map[string]map[string]struct{})
	totalTokens := 0
	var earliest, latest time.Time

	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)

	for scanner.Scan() {
		var entry Entry
		if err := json.Unmarshal(scanner.Bytes(), &entry); err != nil {
			continue
		}

		ts := time.UnixMilli(entry.Timestamp)
		chars := entry.EstimateChars()
		tokens := chars / 4 // rough estimate: ~4 chars per token

		if earliest.IsZero() || ts.Before(earliest) {
			earliest = ts
		}
		if ts.After(latest) {
			latest = ts
		}

		totalTokens += tokens
		project := filepath.Base(entry.Project)

		// weekly
		y, w := ts.ISOWeek()
		wk := time.Date(y, 1, 1, 0, 0, 0, 0, ts.Location())
		// find monday of that ISO week
		for wk.Weekday() != time.Monday {
			wk = wk.AddDate(0, 0, 1)
		}
		wk = wk.AddDate(0, 0, (w-1)*7)
		wkEnd := wk.AddDate(0, 0, 6)

		weekKey := wk.Format("2006-W") + padWeek(w)
		ws, ok := weeks[weekKey]
		if !ok {
			ws = &WeekStats{
				Week:     weekKey,
				Label:    wk.Format("Jan 02") + " – " + wkEnd.Format("Jan 02"),
				Projects: make(map[string]int),
			}
			weeks[weekKey] = ws
		}
		ws.Prompts++
		ws.Tokens += tokens
		ws.Projects[project]++

		// session tracking per week
		if _, ok := allSessions[weekKey]; !ok {
			allSessions[weekKey] = make(map[string]struct{})
		}
		allSessions[weekKey][entry.SessionID] = struct{}{}

		// monthly
		mk := ts.Format("2006-01")
		ms, ok := months[mk]
		if !ok {
			ms = &MonthStats{
				Month: mk,
				Label: ts.Format("Jan 2006"),
			}
			months[mk] = ms
		}
		ms.Prompts++
		ms.Tokens += tokens

		if _, ok := monthDays[mk]; !ok {
			monthDays[mk] = make(map[string]struct{})
		}
		monthDays[mk][ts.Format("2006-01-02")] = struct{}{}
	}

	// build sorted week list + top projects per week
	var weekList []WeekStats
	for k, ws := range weeks {
		ws.Sessions = len(allSessions[k])
		// top 3 projects
		var pcs []ProjectCount
		for name, count := range ws.Projects {
			pcs = append(pcs, ProjectCount{name, count})
		}
		sort.Slice(pcs, func(i, j int) bool { return pcs[i].Prompts > pcs[j].Prompts })
		if len(pcs) > 3 {
			pcs = pcs[:3]
		}
		ws.TopProjects = pcs
		weekList = append(weekList, *ws)
	}
	sort.Slice(weekList, func(i, j int) bool { return weekList[i].Week < weekList[j].Week })

	// build sorted month list
	var monthList []MonthStats
	for mk, ms := range months {
		ms.Days = len(monthDays[mk])
		monthList = append(monthList, *ms)
	}
	sort.Slice(monthList, func(i, j int) bool { return monthList[i].Month < monthList[j].Month })

	// total weeks span
	totalWeeks := 1
	if !earliest.IsZero() && !latest.IsZero() {
		totalWeeks = int(latest.Sub(earliest).Hours()/(24*7)) + 1
	}

	avgPerWeek := 0.0
	avgTokensWeek := 0.0
	if len(weekList) > 0 {
		total := 0
		totalTk := 0
		for _, w := range weekList {
			total += w.Prompts
			totalTk += w.Tokens
		}
		avgPerWeek = math.Round(float64(total)/float64(len(weekList))*10) / 10
		avgTokensWeek = math.Round(float64(totalTk) / float64(len(weekList)))
	}

	return &Timeline{
		TotalTokens:   totalTokens,
		TotalWeeks:    totalWeeks,
		ActiveWeeks:   len(weekList),
		AvgPerWeek:    avgPerWeek,
		AvgTokensWeek: avgTokensWeek,
		Weeks:         weekList,
		Months:        monthList,
	}, nil
}

func padWeek(w int) string {
	if w < 10 {
		return "0" + string(rune('0'+w))
	}
	s := ""
	s += string(rune('0' + w/10))
	s += string(rune('0' + w%10))
	return s
}
