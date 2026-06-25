package main

import (
	"context"
	"fmt"

	"github.com/n0m-d/gator/internal/scraper"
)

func scrapeFeeds(s *state) error {
	name, err := scraper.ScrapeNextFeed(context.Background(), s.db)
	if err != nil {
		return err
	}
	if name != "" {
		fmt.Printf("Feed: %s\n", name)
	}
	return nil
}
