package cli

import (
	"errors"
	"fmt"

	"github.com/n0m-d/gator/internal/config"
	"github.com/n0m-d/gator/internal/database"
)

type App struct {
	Cfg *config.Config
	DB  *database.Queries
}

type command struct {
	name string
	args []string
}

type commands struct {
	handlers map[string]func(*App, command) error
}

func (c *commands) register(name string, f func(*App, command) error) {
	c.handlers[name] = f
}

func (c *commands) run(app *App, cmd command) error {
	f, ok := c.handlers[cmd.name]
	if !ok {
		return errors.New("command not found")
	}
	return f(app, cmd)
}

func (a *App) Run(cmdName string, args []string) error {
	cmds := commands{handlers: make(map[string]func(*App, command) error)}
	cmds.register("login", handleLogin)
	cmds.register("register", handleRegister)
	cmds.register("users", handleListUsers)
	cmds.register("reset", handleReset)

	if err := cmds.run(a, command{name: cmdName, args: args}); err != nil {
		return err
	}
	return nil
}

func usage() string {
	return `Gator auth commands (the app runs as a TUI by default):

  login <name>       Log in as an existing user
  register <name>    Create a new user and log in
  users              List all users
  reset              Delete all users (development)`
}

func (a *App) PrintUsage() {
	fmt.Println(usage())
}
