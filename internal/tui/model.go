package tui

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
	"github.com/n0m-d/gator/internal/config"
	"github.com/n0m-d/gator/internal/database"
)

const (
	tabPosts            = 0
	tabFollowing        = 1
	defaultPostsPerPage = 10
	pagerHeight         = 1
	aggProgressHeight   = 1
	toastDuration       = 2 * time.Second
	defaultAggInterval  = time.Minute
	progressTick        = 200 * time.Millisecond
)

type tab int

type model struct {
	db       *database.Queries
	cfg      *config.Config
	user     database.User
	username string
	styles   Styles

	authenticated bool
	authMode      authMode
	usernameInput textinput.Model

	activeTab    tab
	cursor       int
	listOffset   int
	postPage     int
	postsPerPage int
	totalPosts   int64
	width        int
	height       int

	posts []database.Post
	feeds []database.GetFeedFollowsForUserRow

	toast      string
	toastError bool

	addingFeed   bool
	addFeedField addFeedField
	nameInput    textinput.Model
	urlInput     textinput.Model

	aggregating  bool
	aggInterval  time.Duration
	scraping     bool
	totalFeeds   int64
	feedsFetched int
	lastScrapeAt time.Time
	aggProgress  progress.Model

	err     error
	loading bool
}

type dataLoadedMsg struct {
	posts      []database.Post
	feeds      []database.GetFeedFollowsForUserRow
	totalPosts int64
	err        error
}

type clearToastMsg struct{}

func NewModel(db *database.Queries, cfg *config.Config) model {
	aggBar := progress.New(
		progress.WithGradient("#00A95C", "#73F59F"),
		progress.WithWidth(40),
		progress.WithoutPercentage(),
	)

	m := model{
		db:          db,
		cfg:         cfg,
		styles:      NewStyles(),
		aggInterval: defaultAggInterval,
		aggProgress: aggBar,
	}

	if user, username, ok := tryLoadSession(db, cfg); ok {
		m.authenticated = true
		m.user = user
		m.username = username
		m.loading = true
	} else {
		m.initAuthForm()
	}

	return m
}

func (m model) Init() tea.Cmd {
	if !m.authenticated {
		return textinput.Blink
	}
	return m.loadData
}

func (m model) calcPostsPerPage() int {
	if m.height == 0 {
		return defaultPostsPerPage
	}
	header := m.renderHeader()
	footer := m.renderStatusBar()
	used := lipgloss.Height(header) + lipgloss.Height(footer) + pagerHeight + 3
	if m.aggregating {
		used += aggProgressHeight
	}
	remaining := m.height - used
	detailReserve := min(10, remaining/2)
	rows := remaining - detailReserve - 2
	if rows < 3 {
		return 3
	}
	return rows
}

func (m model) pageSize() int {
	if m.postsPerPage > 0 {
		return m.postsPerPage
	}
	return m.calcPostsPerPage()
}

func (m model) totalPages() int {
	perPage := m.pageSize()
	if perPage <= 0 || m.totalPosts == 0 {
		return 1
	}
	return int((m.totalPosts + int64(perPage) - 1) / int64(perPage))
}

func (m model) loadData() tea.Msg {
	perPage := m.pageSize()

	total, err := m.db.CountPostsForUser(context.Background(), m.user.ID)
	if err != nil {
		return dataLoadedMsg{err: fmt.Errorf("couldn't count posts: %w", err)}
	}

	posts, err := m.db.GetPostsForUserPaginated(context.Background(), database.GetPostsForUserPaginatedParams{
		UserID: m.user.ID,
		Limit:  int32(perPage),
		Offset: int32(m.postPage * perPage),
	})
	if err != nil {
		return dataLoadedMsg{err: fmt.Errorf("couldn't load posts: %w", err)}
	}

	feeds, err := m.db.GetFeedFollowsForUser(context.Background(), m.user.ID)
	if err != nil {
		return dataLoadedMsg{err: fmt.Errorf("couldn't load feeds: %w", err)}
	}

	return dataLoadedMsg{posts: posts, feeds: feeds, totalPosts: total}
}

func (m model) dismissToastCmd() tea.Cmd {
	return tea.Tick(toastDuration, func(time.Time) tea.Msg {
		return clearToastMsg{}
	})
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		barWidth := msg.Width - 24
		if barWidth < 20 {
			barWidth = 20
		}
		m.aggProgress.Width = barWidth
		if !m.authenticated {
			return m, nil
		}
		m.postsPerPage = m.calcPostsPerPage()
		m.loading = true
		return m, m.loadData

	case clearToastMsg:
		m.toast = ""
		m.toastError = false
		return m, nil

	case dataLoadedMsg:
		m.loading = false
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		}
		m.posts = msg.posts
		m.feeds = msg.feeds
		m.totalPosts = msg.totalPosts
		m.err = nil

		if m.postPage >= m.totalPages() {
			m.postPage = max(0, m.totalPages()-1)
		}
		m.clampCursor()
		return m, nil

	case feedCreatedMsg:
		m.loading = false
		if msg.err != nil {
			m.toast = msg.err.Error()
			m.toastError = true
			return m, tea.Batch(m.loadData, m.dismissToastCmd())
		}
		m.toast = fmt.Sprintf("Added and followed %s", msg.name)
		m.toastError = false
		return m, tea.Batch(m.loadData, m.dismissToastCmd())

	case authDoneMsg:
		if msg.err != nil {
			m.toast = msg.err.Error()
			m.toastError = true
			return m, m.dismissToastCmd()
		}
		m.authenticated = true
		m.user = msg.user
		m.username = msg.username
		m.loading = true
		m.toast = fmt.Sprintf("Welcome, %s", msg.username)
		m.toastError = false
		return m, tea.Batch(m.loadData, m.dismissToastCmd())

	case aggTickMsg:
		return m.handleAggTick()

	case scrapeDoneMsg:
		return m.handleScrapeDone(msg)

	case feedCountMsg:
		return m.handleFeedCount(msg)

	case progressTickMsg:
		if !m.aggregating {
			return m, nil
		}
		return m, m.scheduleProgressTick()

	case tea.KeyMsg:
		if !m.authenticated {
			return m.updateAuth(msg)
		}

		if m.addingFeed {
			return m.updateAddFeed(msg)
		}

		if m.toast != "" {
			m.toast = ""
			m.toastError = false
		}

		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "q":
			if m.addingFeed {
				m.addingFeed = false
				return m, nil
			}
			return m, tea.Quit
		case "tab":
			m.activeTab = (m.activeTab + 1) % 2
			m.cursor = 0
			m.listOffset = 0
			return m, nil
		case "1":
			m.activeTab = tabPosts
			m.cursor = 0
			m.listOffset = 0
			return m, nil
		case "2":
			m.activeTab = tabFollowing
			m.cursor = 0
			m.listOffset = 0
			return m, nil
		case "r":
			return m.toggleAgg()
		case "left", "h", "[":
			if m.activeTab == tabPosts && m.postPage > 0 {
				m.postPage--
				m.cursor = 0
				m.loading = true
				return m, m.loadData
			}
			return m, nil
		case "right", "l", "]":
			if m.activeTab == tabPosts && m.postPage < m.totalPages()-1 {
				m.postPage++
				m.cursor = 0
				m.loading = true
				return m, m.loadData
			}
			return m, nil
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
				if m.activeTab == tabFollowing {
					m.ensureFollowingVisible()
				}
			}
			return m, nil
		case "down", "j":
			if m.cursor < m.listLen()-1 {
				m.cursor++
				if m.activeTab == tabFollowing {
					m.ensureFollowingVisible()
				}
			}
			return m, nil
		case "v":
			if m.activeTab != tabPosts || len(m.posts) == 0 || m.cursor >= len(m.posts) {
				return m, nil
			}

			post := m.posts[m.cursor]
			u, err := url.Parse(post.Url)
			if err != nil {
				m.toast = "Couldn't parse URL"
				m.toastError = true
				return m, m.dismissToastCmd()
			}
			if u.Scheme != "http" && u.Scheme != "https" {
				m.toast = "Unsupported URL scheme"
				m.toastError = true
				return m, m.dismissToastCmd()
			}

			if err := clipboard.WriteAll(post.Url); err != nil {
				m.toast = "Couldn't copy to clipboard"
				m.toastError = true
				return m, m.dismissToastCmd()
			}

			m.toast = "URL copied to clipboard"
			m.toastError = false
			return m, m.dismissToastCmd()

		case "u":
			if m.activeTab != tabFollowing || len(m.feeds) == 0 || m.cursor >= len(m.feeds) {
				return m, nil
			}

			feed := m.feeds[m.cursor]
			err := m.db.DeleteFeedFollowByFeedIdAndUserId(context.Background(), database.DeleteFeedFollowByFeedIdAndUserIdParams{
				FeedID: feed.FeedID,
				UserID: m.user.ID,
			})
			if err != nil {
				m.toast = "Couldn't unfollow feed"
				m.toastError = true
				return m, m.dismissToastCmd()
			}

			m.toast = fmt.Sprintf("Unfollowed %s", feed.FeedName)
			m.toastError = false
			m.loading = true
			return m, tea.Batch(m.loadData, m.dismissToastCmd())

		case "a":
			if m.activeTab != tabFollowing || m.addingFeed {
				return m, nil
			}
			m.initAddFeedForm()
			return m, textinput.Blink
		}
	}

	return m, nil
}

func (m *model) clampCursor() {
	max := m.listLen() - 1
	if max < 0 {
		m.cursor = 0
		m.listOffset = 0
		return
	}
	if m.cursor > max {
		m.cursor = max
	}
	if m.activeTab == tabFollowing {
		m.ensureFollowingVisible()
	}
}

func (m model) visibleListRows() int {
	return m.pageSize()
}

func (m *model) ensureFollowingVisible() {
	if m.height == 0 {
		return
	}
	visible := m.visibleListRows()
	if m.cursor < m.listOffset {
		m.listOffset = m.cursor
	}
	if m.cursor >= m.listOffset+visible {
		m.listOffset = m.cursor - visible + 1
	}
	maxOffset := m.listLen() - visible
	if maxOffset < 0 {
		maxOffset = 0
	}
	if m.listOffset > maxOffset {
		m.listOffset = maxOffset
	}
}

func (m model) listLen() int {
	if m.activeTab == tabPosts {
		return len(m.posts)
	}
	return len(m.feeds)
}

func (m model) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	if !m.authenticated {
		header := m.styles.Banner.Render(banner)
		auth := m.renderAuthScreen()
		parts := []string{header, "", auth}
		if m.toast != "" {
			contentWidth := m.width - 4
			if contentWidth < 20 {
				contentWidth = 20
			}
			toast := lipgloss.NewStyle().
				Width(contentWidth).
				Align(lipgloss.Center).
				Render(m.renderToast())
			parts = append(parts, "", toast)
		}
		return lipgloss.Place(m.width, m.height, lipgloss.Left, lipgloss.Top,
			lipgloss.JoinVertical(lipgloss.Left, parts...),
		)
	}

	header := m.renderHeader()
	footer := m.renderStatusBar()

	var content string
	if m.loading {
		content = m.styles.Empty.Render("Loading...")
	} else if m.err != nil {
		content = m.styles.Error.Render(m.err.Error())
	} else {
		content = m.renderContent()
	}

	pager := ""
	if m.activeTab == tabPosts && m.totalPosts > 0 {
		pager = m.renderPager()
	}

	aggBar := m.renderAggProgress()

	parts := []string{header, ""}
	if aggBar != "" {
		parts = append(parts, aggBar, "")
	}
	parts = append(parts, content, "", pager)
	if m.addingFeed {
		contentWidth := m.width - 4
		if contentWidth < 20 {
			contentWidth = 20
		}
		dialog := lipgloss.NewStyle().
			Width(contentWidth).
			Align(lipgloss.Center).
			Render(m.renderAddFeedDialog())
		parts = append(parts, "", dialog)
	}
	if m.toast != "" {
		contentWidth := m.width - 4
		if contentWidth < 20 {
			contentWidth = 20
		}
		toast := lipgloss.NewStyle().
			Width(contentWidth).
			Align(lipgloss.Center).
			Render(m.renderToast())
		parts = append(parts, "", toast)
	}
	parts = append(parts, "", footer)

	return lipgloss.Place(m.width, m.height, lipgloss.Left, lipgloss.Top,
		lipgloss.JoinVertical(lipgloss.Left, parts...),
	)
}

func (m model) renderHeader() string {
	return lipgloss.JoinVertical(lipgloss.Center,
		m.styles.Banner.Render(banner),
		m.renderTabs(),
	)
}

func Run(db *database.Queries, cfg *config.Config) error {
	defer restoreTerminal()

	m := NewModel(db, cfg)
	p := tea.NewProgram(
		m,
		tea.WithAltScreen(),
		tea.WithInput(os.Stdin),
		tea.WithOutput(os.Stdout),
	)
	_, err := p.Run()
	return err
}
