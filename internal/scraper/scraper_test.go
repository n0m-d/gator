package scraper

import (
	"context"
	"database/sql"
	"os"
	"testing"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/n0m-d/gator/internal/database"
)

func TestScrapeNextFeedNoFollows(t *testing.T) {
	dbURL := os.Getenv("TEST_DATABASE_URL")
	if dbURL == "" {
		t.Skip("TEST_DATABASE_URL not set")
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()

	q := database.New(db)
	name, err := ScrapeNextFeed(context.Background(), q, uuid.New())
	if err != nil {
		t.Fatalf("ScrapeNextFeed() error: %v", err)
	}
	if name != "" {
		t.Errorf("name = %q, want empty when user follows no feeds", name)
	}
}
