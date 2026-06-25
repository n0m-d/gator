package cli

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/n0m-d/gator/internal/database"
)

func handleLogin(a *App, cmd command) error {
	if len(cmd.args) != 1 {
		return fmt.Errorf("usage: login <name>")
	}
	name := cmd.args[0]

	user, err := a.DB.GetUser(context.Background(), name)
	if err != nil {
		return fmt.Errorf("couldn't get user: %w", err)
	}

	if err := a.Cfg.SetUser(user.Name); err != nil {
		return fmt.Errorf("couldn't set current user: %w", err)
	}

	fmt.Println("Logged in successfully!")
	return nil
}

func handleRegister(a *App, cmd command) error {
	if len(cmd.args) != 1 {
		return fmt.Errorf("usage: register <name>")
	}
	name := cmd.args[0]

	user, err := a.DB.CreateUser(context.Background(), database.CreateUserParams{
		ID:   uuid.New(),
		Name: name,
	})
	if err != nil {
		return fmt.Errorf("couldn't create user: %w", err)
	}

	if err := a.Cfg.SetUser(name); err != nil {
		return fmt.Errorf("couldn't set current user: %w", err)
	}

	fmt.Printf("User created: %s\n", user.Name)
	return nil
}

func handleListUsers(a *App, cmd command) error {
	users, err := a.DB.GetUsers(context.Background())
	if err != nil {
		return fmt.Errorf("couldn't get users: %w", err)
	}

	for _, user := range users {
		if user.Name == a.Cfg.CurrentUserName {
			fmt.Printf("* %s (current)\n", user.Name)
		} else {
			fmt.Printf("  %s\n", user.Name)
		}
	}
	return nil
}

func handleReset(a *App, cmd command) error {
	if err := a.DB.DeleteAllUsers(context.Background()); err != nil {
		return fmt.Errorf("couldn't delete users: %w", err)
	}
	fmt.Println("All users deleted.")
	return nil
}
