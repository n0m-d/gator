package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
	"github.com/n0m-d/gator/internal/cli"
	"github.com/n0m-d/gator/internal/config"
	"github.com/n0m-d/gator/internal/database"
	"github.com/n0m-d/gator/internal/tui"
)

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
	app := &cli.App{Cfg: &cfg, DB: dbQueries}

	if len(os.Args) >= 2 {
		if os.Args[1] == "help" || os.Args[1] == "--help" || os.Args[1] == "-h" {
			app.PrintUsage()
			return
		}
		if err := app.Run(os.Args[1], os.Args[2:]); err != nil {
			log.Fatal(err)
		}
		return
	}

	if err := tui.Run(dbQueries, &cfg); err != nil {
		fmt.Fprintf(os.Stderr, "tui error: %v\n", err)
		os.Exit(1)
	}
}
