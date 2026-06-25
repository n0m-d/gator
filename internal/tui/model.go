package tui

import (
	"context"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/n0m-d/gator/internal/database"
)

const (
	tabPosts            = 0
	tabFollowing        = 1
	defaultPostsPerPage = 10
	pagerHeight         = 1
)

type tab int

type model struct {
	db       *database.Queries
	user     database.User
	username string
	styles   Styles

	activeTab  tab
	cursor     int
	listOffset int
	postPage     int
	postsPerPage int
	totalPosts   int64
	width      int
	height     int

	posts []database.Post
	feeds []database.GetFeedFollowsForUserRow

	err     error
	loading bool
}

type dataLoadedMsg struct {
	posts      []database.Post
	feeds      []database.GetFeedFollowsForUserRow
	totalPosts int64
	err        error
}

func NewModel(db *database.Queries, user database.User, username string) model {
	return model{
		db:       db,
		user:     user,
		username: username,
		styles:   NewStyles(),
		loading:  true,
	}
}

func (m model) Init() tea.Cmd {
	return m.loadData
}

func (m model) calcPostsPerPage() int {
	if m.height == 0 {
		return defaultPostsPerPage
	}
	header := m.renderHeader()
	footer := m.renderStatusBar()
	used := lipgloss.Height(header) + lipgloss.Height(footer) + pagerHeight + 3
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

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.postsPerPage = m.calcPostsPerPage()
		m.loading = true
		return m, m.loadData

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

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
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
			m.loading = true
			m.err = nil
			return m, m.loadData
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

	return lipgloss.Place(m.width, m.height, lipgloss.Left, lipgloss.Top,
		lipgloss.JoinVertical(lipgloss.Left, header, "", content, "", pager, "", footer),
	)
}

func (m model) renderHeader() string {
	return lipgloss.JoinVertical(lipgloss.Left,
		m.styles.Banner.Render(banner),
		m.renderTabs(),
	)
}

func Run(db *database.Queries, user database.User, username string) error {
	m := NewModel(db, user, username)
	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err := p.Run()
	return err
}
