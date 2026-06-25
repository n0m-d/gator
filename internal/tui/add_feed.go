package tui

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
	"github.com/google/uuid"
	"github.com/n0m-d/gator/internal/database"
)

type addFeedField int

const (
	addFieldName addFeedField = iota
	addFieldURL
)

type feedCreatedMsg struct {
	name string
	err  error
}

func (m *model) initAddFeedForm() {
	name := textinput.New()
	name.Placeholder = "Feed name"
	name.CharLimit = 255
	name.Width = 50
	name.Prompt = "Name: "

	feedURL := textinput.New()
	feedURL.Placeholder = "https://example.com/feed.xml"
	feedURL.CharLimit = 255
	feedURL.Width = 50
	feedURL.Prompt = "URL:  "

	m.nameInput = name
	m.urlInput = feedURL
	m.addingFeed = true
	m.addFeedField = addFieldName
	m.focusAddFeedField()
}

func (m *model) focusAddFeedField() {
	if m.addFeedField == addFieldName {
		m.nameInput.Focus()
		m.urlInput.Blur()
	} else {
		m.urlInput.Focus()
		m.nameInput.Blur()
	}
}

func (m model) updateAddFeed(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.addingFeed = false
		return m, nil
	case "tab", "down":
		m.addFeedField = addFieldURL
		m.focusAddFeedField()
		return m, textinput.Blink
	case "shift+tab", "up":
		m.addFeedField = addFieldName
		m.focusAddFeedField()
		return m, textinput.Blink
	case "enter":
		if m.addFeedField == addFieldName {
			m.addFeedField = addFieldURL
			m.focusAddFeedField()
			return m, textinput.Blink
		}
		return m.submitAddFeed()
	}

	var cmd tea.Cmd
	if m.addFeedField == addFieldName {
		m.nameInput, cmd = m.nameInput.Update(msg)
	} else {
		m.urlInput, cmd = m.urlInput.Update(msg)
	}
	return m, cmd
}

func (m model) submitAddFeed() (tea.Model, tea.Cmd) {
	name := strings.TrimSpace(m.nameInput.Value())
	feedURL := strings.TrimSpace(m.urlInput.Value())

	if name == "" || feedURL == "" {
		m.toast = "Name and URL are required"
		m.toastError = true
		return m, m.dismissToastCmd()
	}

	parsed, err := url.Parse(feedURL)
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		m.toast = "Enter a valid http(s) URL"
		m.toastError = true
		return m, m.dismissToastCmd()
	}

	m.addingFeed = false
	m.loading = true
	return m, tea.Batch(m.createFeedCmd(name, feedURL), m.dismissToastCmd())
}

func (m model) createFeedCmd(name, feedURL string) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()

		feed, err := m.db.CreateFeed(ctx, database.CreateFeedParams{
			ID:     uuid.New(),
			Name:   name,
			Url:    feedURL,
			UserID: m.user.ID,
		})
		if err != nil {
			return feedCreatedMsg{err: fmt.Errorf("couldn't create feed: %w", err)}
		}

		_, err = m.db.CreateFeedFollow(ctx, database.CreateFeedFollowParams{
			ID:     uuid.New(),
			UserID: m.user.ID,
			FeedID: feed.ID,
		})
		if err != nil {
			return feedCreatedMsg{err: fmt.Errorf("couldn't follow feed: %w", err)}
		}

		return feedCreatedMsg{name: name}
	}
}

func (m model) renderAddFeedDialog() string {
	contentWidth := m.width - 8
	if contentWidth < 40 {
		contentWidth = 40
	}

	m.nameInput.Width = contentWidth - lipgloss.Width(m.nameInput.Prompt)
	m.urlInput.Width = contentWidth - lipgloss.Width(m.urlInput.Prompt)

	nameLabel := m.styles.DetailLabel.Render("Name")
	urlLabel := m.styles.DetailLabel.Render("URL")
	if m.addFeedField == addFieldName {
		nameLabel = m.styles.SelectedItem.Render("› Name")
	} else {
		urlLabel = m.styles.SelectedItem.Render("› URL")
	}

	body := lipgloss.JoinVertical(lipgloss.Left,
		m.styles.DetailLabel.Render("Add Feed"),
		"",
		nameLabel,
		m.nameInput.View(),
		"",
		urlLabel,
		m.urlInput.View(),
		"",
		m.styles.Help.Render("enter: next/submit  tab: fields  esc: cancel"),
	)

	return m.styles.DetailBox.Width(contentWidth).Render(body)
}
