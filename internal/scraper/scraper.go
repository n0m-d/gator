package scraper

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/n0m-d/gator/internal/database"
	"github.com/n0m-d/gator/internal/rss"
)

// ScrapeNextFeed fetches the next due feed the user follows and saves its posts.
// Returns the feed name when one was scraped, or "" when there are no followed feeds.
func ScrapeNextFeed(ctx context.Context, db *database.Queries, userID uuid.UUID) (string, error) {
	feed, err := db.GetNextFeedToFetchForUser(ctx, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", nil
		}
		return "", fmt.Errorf("couldn't get next feed to fetch: %w", err)
	}

	rssFeed, err := rss.Fetch(ctx, feed.Url)
	if err != nil {
		return feed.Name, fmt.Errorf("couldn't fetch feed %s: %w", feed.Name, err)
	}

	if err := db.MarkFeedFetched(ctx, feed.ID); err != nil {
		return feed.Name, fmt.Errorf("couldn't mark feed as fetched: %w", err)
	}

	for _, item := range rssFeed.Channel.Item {
		if err := savePost(ctx, db, feed.ID, item); err != nil {
			log.Printf("couldn't create post: %v", err)
		}
	}

	return feed.Name, nil
}

func savePost(ctx context.Context, db *database.Queries, feedID uuid.UUID, item rss.Item) error {
	var description sql.NullString
	if item.Description != "" {
		description = sql.NullString{String: item.Description, Valid: true}
	}

	_, err := db.CreatePost(ctx, database.CreatePostParams{
		ID:          uuid.New(),
		Title:       item.Title,
		Url:         item.Link,
		Description: description,
		PublishedAt: rss.ParsePublishedAt(item.PubDate),
		FeedID:      feedID,
	})
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			return nil
		}
		return err
	}

	return nil
}
