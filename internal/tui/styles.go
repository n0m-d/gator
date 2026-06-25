package tui

import (
	"github.com/charmbracelet/lipgloss"
)

const gatorGreen = "#00A95C"

type Styles struct {
	Subtle       lipgloss.AdaptiveColor
	Highlight    lipgloss.AdaptiveColor
	Special      lipgloss.AdaptiveColor
	Tab          lipgloss.Style
	ActiveTab    lipgloss.Style
	TabGap       lipgloss.Style
	Title        lipgloss.Style
	List         lipgloss.Style
	ListHeader   lipgloss.Style
	ListItem     lipgloss.Style
	SelectedItem lipgloss.Style
	DetailBox    lipgloss.Style
	DetailLabel  lipgloss.Style
	DetailValue  lipgloss.Style
	StatusBar    lipgloss.Style
	StatusKey    lipgloss.Style
	StatusValue  lipgloss.Style
	Help         lipgloss.Style
	Error        lipgloss.Style
	Empty        lipgloss.Style
	Banner       lipgloss.Style
	URL          lipgloss.Style
	Toast        lipgloss.Style
	ToastError   lipgloss.Style
}

func NewStyles() Styles {
	subtle := lipgloss.AdaptiveColor{Light: "#D9DCCF", Dark: "#383838"}
	highlight := lipgloss.AdaptiveColor{Light: "#874BFD", Dark: "#7D56F4"}
	special := lipgloss.AdaptiveColor{Light: "#43BF6D", Dark: "#73F59F"}

	activeTabBorder := lipgloss.Border{
		Top: "─", Bottom: " ", Left: "│", Right: "│",
		TopLeft: "╭", TopRight: "╮", BottomLeft: "┘", BottomRight: "└",
	}
	tabBorder := lipgloss.Border{
		Top: "─", Bottom: "─", Left: "│", Right: "│",
		TopLeft: "╭", TopRight: "╮", BottomLeft: "┴", BottomRight: "┴",
	}

	tab := lipgloss.NewStyle().
		Border(tabBorder, true).
		BorderForeground(highlight).
		Padding(0, 1)

	return Styles{
		Subtle:    subtle,
		Highlight: highlight,
		Special:   special,
		Tab:       tab,
		ActiveTab: tab.Border(activeTabBorder, true),
		TabGap: tab.Border(lipgloss.Border{
			Top: "─", Bottom: "─", Left: "│", Right: "│",
			TopLeft: "╭", TopRight: "╮", BottomLeft: "┴", BottomRight: "┴",
		}, true).BorderTop(false).BorderLeft(false).BorderRight(false),
		Title: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(gatorGreen)),
		List: lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), false, true, false, false).
			BorderForeground(subtle).
			Padding(0, 1),
		ListHeader: lipgloss.NewStyle().
			Bold(true).
			BorderStyle(lipgloss.NormalBorder()).
			BorderBottom(true).
			BorderForeground(subtle).
			MarginBottom(1),
		ListItem: lipgloss.NewStyle().PaddingLeft(1),
		SelectedItem: lipgloss.NewStyle().
			PaddingLeft(1).
			Bold(true).
			Foreground(lipgloss.Color("#FFF7DB")).
			Background(highlight),
		DetailBox: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(highlight).
			Padding(1, 2).
			MarginTop(1),
		DetailLabel: lipgloss.NewStyle().Foreground(subtle).Bold(true),
		DetailValue: lipgloss.NewStyle(),
		StatusBar: lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#343433", Dark: "#C1C6B2"}).
			Background(lipgloss.AdaptiveColor{Light: "#D9DCCF", Dark: "#353533"}),
		StatusKey: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFDF5")).
			Background(lipgloss.Color(gatorGreen)).
			Padding(0, 1).
			MarginRight(1),
		StatusValue: lipgloss.NewStyle(),
		Help: lipgloss.NewStyle().
			Foreground(subtle).
			Italic(true),
		Error: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF5F87")).
			Bold(true),
		Empty: lipgloss.NewStyle().
			Foreground(subtle).
			Italic(true).
			Padding(1, 2),
		Banner: lipgloss.NewStyle().
			Foreground(lipgloss.Color(gatorGreen)),
		URL: lipgloss.NewStyle().Foreground(special),
		Toast: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFF7DB")).
			Background(lipgloss.Color(gatorGreen)).
			Padding(0, 2).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(gatorGreen)),
		ToastError: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFF7DB")).
			Background(lipgloss.Color("#FF5F87")).
			Padding(0, 2).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#FF5F87")),
	}
}
