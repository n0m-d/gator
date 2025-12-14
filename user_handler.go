package main

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/n0m-d/gator/internal/database"
)

func handlerLogin(s *state, cmd command) error {
	if len(cmd.Args) != 1 {
		return fmt.Errorf("usage: %s <name>", cmd.Name)
	}
	name := cmd.Args[0]

	user, err := s.db.GetUser(context.Background(), name)
	if err != nil {
		return fmt.Errorf("couldn't get user: %w", err)
	}

	err = s.cfg.SetUser(user.Name)
	if err != nil {
		return fmt.Errorf("couldn't set current user: %w", err)
	}

	fmt.Println("User switched successfully!")
	return nil
}

func handlerRegister(s *state, cmd command) error {
	if len(cmd.Args) != 1 {
		return fmt.Errorf("usage: %s <name>", cmd.Name)
	}
	name := cmd.Args[0]

	data, err := s.db.CreateUser(context.Background(), database.CreateUserParams{
		ID:   uuid.New(),
		Name: name,
	})
	if err != nil {
		return fmt.Errorf("couldn't create user: %w", err)
	}

	fmt.Printf("User created successfully! %v\n", data)

	err = s.cfg.SetUser(name)
	if err != nil {
		return fmt.Errorf("couldn't set current user: %w", err)
	}
	return nil
}
