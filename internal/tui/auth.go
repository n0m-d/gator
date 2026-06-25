package tui

import (
	"context"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
	"github.com/google/uuid"
	"github.com/n0m-d/gator/internal/config"
	"github.com/n0m-d/gator/internal/database"
)

type authMode int

const (
	authLogin authMode = iota
	authRegister
)

type authDoneMsg struct {
	user     database.User
	username string
	err      error
}

func (m *model) initAuthForm() {
	input := textinput.New()
	input.Placeholder = "username"
	input.CharLimit = 64
	input.Width = 40
	input.Prompt = "User: "

	m.usernameInput = input
	m.authMode = authLogin
	m.usernameInput.Focus()
}

func (m model) updateAuth(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.toast != "" {
		m.toast = ""
		m.toastError = false
	}

	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit
	case "tab":
		if m.authMode == authLogin {
			m.authMode = authRegister
		} else {
			m.authMode = authLogin
		}
		return m, nil
	case "enter":
		name := strings.TrimSpace(m.usernameInput.Value())
		if name == "" {
			m.toast = "Username is required"
			m.toastError = true
			return m, m.dismissToastCmd()
		}
		return m, m.authCmd(name)
	}

	var cmd tea.Cmd
	m.usernameInput, cmd = m.usernameInput.Update(msg)
	return m, cmd
}

func (m model) authCmd(name string) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()

		switch m.authMode {
		case authRegister:
			user, err := m.db.CreateUser(ctx, database.CreateUserParams{
				ID:   uuid.New(),
				Name: name,
			})
			if err != nil {
				return authDoneMsg{err: fmt.Errorf("couldn't create user: %w", err)}
			}
			if err := m.cfg.SetUser(name); err != nil {
				return authDoneMsg{err: fmt.Errorf("couldn't save session: %w", err)}
			}
			return authDoneMsg{user: user, username: name}

		default:
			user, err := m.db.GetUser(ctx, name)
			if err != nil {
				return authDoneMsg{err: fmt.Errorf("user not found: %w", err)}
			}
			if err := m.cfg.SetUser(name); err != nil {
				return authDoneMsg{err: fmt.Errorf("couldn't save session: %w", err)}
			}
			return authDoneMsg{user: user, username: name}
		}
	}
}

func (m model) renderAuthScreen() string {
	contentWidth := m.width - 4
	if contentWidth < 40 {
		contentWidth = 40
	}

	m.usernameInput.Width = contentWidth - lipgloss.Width(m.usernameInput.Prompt)

	mode := "Login"
	if m.authMode == authRegister {
		mode = "Register"
	}

	body := lipgloss.JoinVertical(lipgloss.Center,
		m.styles.DetailLabel.Render(mode),
		"",
		m.usernameInput.View(),
		"",
		m.styles.Help.Render("enter: submit  tab: switch login/register  q: quit"),
	)

	dialog := m.styles.DetailBox.Width(min(contentWidth, 50)).Render(body)
	return lipgloss.NewStyle().
		Width(contentWidth).
		Align(lipgloss.Center).
		Render(dialog)
}

func tryLoadSession(db *database.Queries, cfg *config.Config) (database.User, string, bool) {
	if db == nil || cfg == nil {
		return database.User{}, "", false
	}
	username, err := cfg.GetUser()
	if err != nil {
		return database.User{}, "", false
	}

	user, err := db.GetUser(context.Background(), username)
	if err != nil {
		return database.User{}, "", false
	}

	return user, username, true
}
