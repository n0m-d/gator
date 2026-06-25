package tui

import (
	"fmt"
	"html"
	"regexp"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/n0m-d/gator/internal/database"
)

const banner = `
 ██████╗  █████╗ ████████╗ ██████╗ ██████╗ 
██╔════╝ ██╔══██╗╚══██╔══╝██╔═══██╗██╔══██╗             
██║  ███╗███████║   ██║   ██║   ██║██████╔╝            
██║   ██║██╔══██║   ██║   ██║   ██║██╔══██╗    
╚██████╔╝██║  ██║   ██║   ╚██████╔╝██║  ██║ 
 ╚═════╝ ╚═╝  ╚═╝   ╚═╝    ╚═════╝ ╚═╝  ╚═╝ 🐊 
`

var htmlTagRe = regexp.MustCompile(`<[^>]*>`)

func (m model) renderTabs() string {
	tabs := []string{"Posts", "Following"}
	var rendered []string
	for i, name := range tabs {
		if tab(i) == m.activeTab {
			rendered = append(rendered, m.styles.ActiveTab.Render(name))
		} else {
			rendered = append(rendered, m.styles.Tab.Render(name))
		}
	}
	row := lipgloss.JoinHorizontal(lipgloss.Top, rendered...)
	contentWidth := m.width - 4
	if contentWidth < 20 {
		contentWidth = 20
	}
	gap := m.styles.TabGap.Render(strings.Repeat(" ", max(0, contentWidth-lipgloss.Width(row)-2)))
	return lipgloss.JoinHorizontal(lipgloss.Bottom, row, gap)
}

func (m model) renderContent() string {
	contentWidth := m.width - 4
	if contentWidth < 40 {
		contentWidth = 40
	}
	listHeight := m.visibleListRows()

	switch m.activeTab {
	case tabPosts:
		return m.renderPosts(contentWidth, listHeight)
	case tabFollowing:
		return m.renderFollowing(contentWidth, listHeight)
	default:
		return ""
	}
}

func (m model) renderList(items []string, header string, contentWidth, listHeight int) string {
	listStyle := m.styles.List.Width(contentWidth).Height(listHeight)
	body := append([]string{m.styles.ListHeader.Render(header)}, items...)
	return listStyle.Render(lipgloss.JoinVertical(lipgloss.Left, body...))
}

func (m model) renderPosts(contentWidth, listHeight int) string {
	if m.totalPosts == 0 {
		return m.styles.Empty.Render("No posts yet. Follow some feeds and run the aggregator (./gator agg 1m).")
	}
	if len(m.posts) == 0 {
		return m.styles.Empty.Render("Loading page...")
	}

	var items []string
	for i, post := range m.posts {
		label := truncate(post.Title, contentWidth-4)
		if i == m.cursor {
			items = append(items, m.styles.SelectedItem.Render("› "+label))
		} else {
			items = append(items, m.styles.ListItem.Render("  "+label))
		}
	}

	perPage := m.pageSize()
	start := m.postPage*perPage + 1
	end := min(m.postPage*perPage+len(m.posts), int(m.totalPosts))
	header := fmt.Sprintf("Recent Posts (%d–%d of %d)", start, end, m.totalPosts)

	list := m.renderList(items, header, contentWidth, listHeight)
	detail := m.renderPostDetail(m.posts[m.cursor], contentWidth)
	return lipgloss.JoinVertical(lipgloss.Left, list, detail)
}

func (m model) renderPager() string {
	if m.activeTab != tabPosts || m.totalPosts == 0 {
		return ""
	}

	contentWidth := m.width - 4
	if contentWidth < 20 {
		contentWidth = 20
	}

	page := m.postPage + 1
	pages := m.totalPages()

	prev := "← prev"
	next := "next →"
	if m.postPage == 0 {
		prev = m.styles.Help.Render("← prev")
	}
	if m.postPage >= pages-1 {
		next = m.styles.Help.Render("next →")
	}

	label := fmt.Sprintf("Page %d / %d", page, pages)
	return lipgloss.NewStyle().
		Width(contentWidth).
		Align(lipgloss.Center).
		Render(fmt.Sprintf("%s   %s   %s", prev, label, next))
}

func (m model) renderPostDetail(post database.Post, width int) string {
	lines := []string{
		m.detailLine("Title", post.Title, width),
		m.detailLine("URL", post.Url, width),
	}
	if post.PublishedAt.Valid {
		lines = append(lines, m.detailLine("Published", post.PublishedAt.Time.Format(time.RFC1123), width))
	}
	if post.Description.Valid {
		desc := stripHTML(post.Description.String)
		lines = append(lines, m.detailLine("Description", desc, width))
	}
	content := strings.Join(lines, "\n")
	return m.styles.DetailBox.Width(width).Render(content)
}

func (m model) renderFollowing(contentWidth, listHeight int) string {
	if len(m.feeds) == 0 {
		return m.styles.Empty.Render("You're not following any feeds. Use ./gator follow <url> from the CLI.")
	}

	var items []string
	for i, feed := range m.feeds {
		if i < m.listOffset {
			continue
		}
		if i >= m.listOffset+m.visibleListRows() {
			break
		}
		label := truncate(feed.FeedName, contentWidth-4)
		if i == m.cursor {
			items = append(items, m.styles.SelectedItem.Render("› "+label))
		} else {
			items = append(items, m.styles.ListItem.Render("  "+label))
		}
	}

	header := fmt.Sprintf("Following (%d)", len(m.feeds))
	if len(m.feeds) > m.visibleListRows() {
		header = fmt.Sprintf("Following (%d–%d of %d)", m.listOffset+1, min(m.listOffset+m.visibleListRows(), len(m.feeds)), len(m.feeds))
	}

	list := m.renderList(items, header, contentWidth, listHeight)
	feed := m.feeds[m.cursor]
	detail := m.styles.DetailBox.Width(contentWidth).Render(
		strings.Join([]string{
			m.detailLine("Feed", feed.FeedName, contentWidth),
			m.detailLine("URL", feed.FeedURL, contentWidth),
			m.detailLine("Followed", feed.CreatedAt.Format(time.RFC1123), contentWidth),
		}, "\n"),
	)
	return lipgloss.JoinVertical(lipgloss.Left, list, detail)
}

func (m model) detailLine(label, value string, width int) string {
	maxVal := width - len(label) - 4
	if maxVal < 10 {
		maxVal = 10
	}
	return m.styles.DetailLabel.Render(label+": ") +
		m.styles.DetailValue.Render(truncate(value, maxVal))
}

func (m model) renderToast() string {
	style := m.styles.Toast
	if m.toastError {
		style = m.styles.ToastError
	}
	return style.Render(m.toast)
}

func (m model) renderStatusBar() string {
	contentWidth := m.width - 4
	if contentWidth < 20 {
		contentWidth = 20
	}

	userKey := m.styles.StatusKey.Render("User")
	userVal := m.styles.StatusValue.
		Width(max(10, contentWidth/3-lipgloss.Width(userKey))).
		Render(m.username)

	help := m.styles.Help.Render("tab: switch  j/k: move  h/l: page  v: copy url  r: refresh  q: quit")

	bar := lipgloss.JoinHorizontal(lipgloss.Top, userKey, userVal, "  ", help)
	return m.styles.StatusBar.Width(contentWidth).Render(bar)
}

func truncate(s string, max int) string {
	s = strings.TrimSpace(s)
	if max <= 0 {
		return ""
	}
	if lipgloss.Width(s) <= max {
		return s
	}
	runes := []rune(s)
	for len(runes) > 0 && lipgloss.Width(string(runes)) > max-1 {
		runes = runes[:len(runes)-1]
	}
	return string(runes) + "…"
}

func stripHTML(s string) string {
	s = htmlTagRe.ReplaceAllString(s, "")
	return html.UnescapeString(strings.TrimSpace(s))
}
