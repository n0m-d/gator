package tui

import (
	"context"
	"fmt"
	"math"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/n0m-d/gator/internal/scraper"
)

type aggTickMsg struct{}

type progressTickMsg struct{}

type scrapeDoneMsg struct {
	feedName string
	err      error
}

func (m model) scrapeCmd() tea.Cmd {
	return func() tea.Msg {
		name, err := scraper.ScrapeNextFeed(context.Background(), m.db, m.user.ID)
		return scrapeDoneMsg{feedName: name, err: err}
	}
}

func (m model) scheduleAggTick() tea.Cmd {
	return tea.Tick(m.aggInterval, func(time.Time) tea.Msg {
		return aggTickMsg{}
	})
}

func (m model) scheduleProgressTick() tea.Cmd {
	return tea.Tick(progressTick, func(time.Time) tea.Msg {
		return progressTickMsg{}
	})
}

func (m model) loadFeedCount() tea.Cmd {
	return func() tea.Msg {
		count, err := m.db.CountFeedFollowsForUser(context.Background(), m.user.ID)
		if err != nil {
			return feedCountMsg{err: err}
		}
		return feedCountMsg{total: count}
	}
}

type feedCountMsg struct {
	total int64
	err   error
}

func (m model) toggleAgg() (tea.Model, tea.Cmd) {
	if m.aggregating {
		m.aggregating = false
		m.scraping = false
		m.toast = "Aggregation stopped"
		m.toastError = false
		return m, m.dismissToastCmd()
	}

	m.aggregating = true
	m.scraping = true
	m.feedsFetched = 0
	m.lastScrapeAt = time.Now()
	m.toast = fmt.Sprintf("Collecting feeds every %s", m.aggInterval)
	m.toastError = false
	return m, tea.Batch(
		m.loadFeedCount(),
		m.scrapeCmd(),
		m.scheduleAggTick(),
		m.scheduleProgressTick(),
		m.dismissToastCmd(),
	)
}

func (m model) handleAggTick() (tea.Model, tea.Cmd) {
	if !m.aggregating || m.scraping {
		return m, nil
	}
	m.scraping = true
	return m, tea.Batch(m.scrapeCmd(), m.scheduleAggTick())
}

func (m model) handleScrapeDone(msg scrapeDoneMsg) (tea.Model, tea.Cmd) {
	m.scraping = false
	m.lastScrapeAt = time.Now()

	if msg.err != nil {
		m.toast = msg.err.Error()
		m.toastError = true
		return m, tea.Batch(m.loadData, m.dismissToastCmd())
	}

	if msg.feedName != "" {
		m.feedsFetched++
		m.toast = fmt.Sprintf("Fetched %s", msg.feedName)
		m.toastError = false
		return m, tea.Batch(m.loadData, m.dismissToastCmd())
	}

	if m.totalFeeds == 0 {
		m.toast = "No followed feeds to collect — add one on the Following tab"
	} else {
		m.toast = "No feeds due right now"
	}
	m.toastError = false
	return m, tea.Batch(m.loadData, m.dismissToastCmd())
}

func (m model) handleFeedCount(msg feedCountMsg) (tea.Model, tea.Cmd) {
	if msg.err != nil {
		m.toast = msg.err.Error()
		m.toastError = true
		return m, m.dismissToastCmd()
	}
	m.totalFeeds = msg.total
	return m, nil
}

func (m model) aggProgressPercent() float64 {
	if m.scraping {
		wave := math.Sin(float64(time.Now().UnixMilli()) / 180)
		return 0.35 + 0.3*wave
	}

	if m.aggInterval <= 0 {
		return 0
	}

	elapsed := time.Since(m.lastScrapeAt)
	p := float64(elapsed) / float64(m.aggInterval)
	if p > 1 {
		return 1
	}
	if p < 0 {
		return 0
	}
	return p
}

func (m model) aggProgressLabel() string {
	if m.scraping {
		return "Fetching feed..."
	}

	if m.totalFeeds == 0 {
		return "No followed feeds — press a to add one"
	}

	remaining := m.aggInterval - time.Since(m.lastScrapeAt)
	if remaining < 0 {
		remaining = 0
	}

	return fmt.Sprintf("%d/%d feeds · next in %s",
		m.cycleFeedPosition(),
		m.totalFeeds,
		formatDuration(remaining),
	)
}

func (m model) cycleFeedPosition() int {
	if m.totalFeeds == 0 {
		return 0
	}
	pos := m.feedsFetched % int(m.totalFeeds)
	if pos == 0 && m.feedsFetched > 0 {
		return int(m.totalFeeds)
	}
	return pos
}

func formatDuration(d time.Duration) string {
	d = d.Round(time.Second)
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
	m := int(d.Minutes())
	s := int(d.Seconds()) % 60
	if s == 0 {
		return fmt.Sprintf("%dm", m)
	}
	return fmt.Sprintf("%dm%ds", m, s)
}
