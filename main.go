package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
	"github.com/n0m-d/gator/internal/config"
	"github.com/n0m-d/gator/internal/database"
	"github.com/n0m-d/gator/internal/tui"
)

type state struct {
	cfg *config.Config
	db  *database.Queries
}

func main() {
	cfg, err := config.Read()
	if err != nil {
		log.Fatalf("error reading config: %v", err)
	}

	db, err := sql.Open("postgres", cfg.DBURL)
	if err != nil {
		log.Fatalf("error opening database: %v", err)
	}

	dbQueries := database.New(db)
	programState := &state{
		cfg: &cfg,
		db:  dbQueries,
	}

	if len(os.Args) >= 2 {
		runCLI(programState, os.Args[1], os.Args[2:])
		return
	}

	runTUI(programState)
}

func runCLI(s *state, cmdName string, cmdArgs []string) {
	cmds := commands{
		registeredCommands: make(map[string]func(*state, command) error),
	}
	cmds.register("login", handlerLogin)
	cmds.register("register", handlerRegister)
	cmds.register("reset", handlerReSet)
	cmds.register("users", handlerListUsers)
	cmds.register("agg", handlerAgg)
	cmds.register("addfeed", middlewareLoggedIn(handlerAdd))
	cmds.register("feeds", handlerListFeeds)
	cmds.register("follow", middlewareLoggedIn(handlerFollowFeed))
	cmds.register("following", middlewareLoggedIn(handlerFollowingFeed))
	cmds.register("unfollow", middlewareLoggedIn(handlerUnfollowFeed))
	cmds.register("browse", middlewareLoggedIn(handlerBrowse))

	if err := cmds.run(s, command{Name: cmdName, Args: cmdArgs}); err != nil {
		log.Fatal(err)
	}
}

func runTUI(s *state) {
	username, err := s.cfg.GetUser()
	if err != nil {
		log.Fatalf("not logged in: %v\nUse: ./gator login <name>", err)
	}

	user, err := s.db.GetUser(context.Background(), username)
	if err != nil {
		log.Fatalf("couldn't get user %q: %v", username, err)
	}

	if err := tui.Run(s.db, user, username); err != nil {
		fmt.Fprintf(os.Stderr, "tui error: %v\n", err)
		os.Exit(1)
	}
}
