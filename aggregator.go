package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/n0m-d/gator/internal/database"
)

func scrapeFeeds(s *state) error {
	ctx := context.Background()

	feed, err := s.db.GetNextFeedToFetch(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		return fmt.Errorf("couldn't get next feed to fetch: %w", err)
	}

	err = s.db.MarkFeedFetched(ctx, feed.ID)
	if err != nil {
		return fmt.Errorf("couldn't mark feed as fetched: %w", err)
	}

	rssFeed, err := fetchFeed(ctx, feed.Url)
	if err != nil {
		return fmt.Errorf("couldn't fetch feed: %w", err)
	}

	fmt.Printf("Feed: %s\n", feed.Name)
	for _, item := range rssFeed.Channel.Item {
		if err := savePost(ctx, s.db, feed.ID, item); err != nil {
			log.Printf("couldn't create post: %v", err)
		}
	}

	return nil
}

func savePost(ctx context.Context, db *database.Queries, feedID uuid.UUID, item RSSItem) error {
	var description sql.NullString
	if item.Description != "" {
		description = sql.NullString{String: item.Description, Valid: true}
	}

	_, err := db.CreatePost(ctx, database.CreatePostParams{
		ID:          uuid.New(),
		Title:       item.Title,
		Url:         item.Link,
		Description: description,
		PublishedAt: parsePublishedAt(item.PubDate),
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
