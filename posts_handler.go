package main

import (
	"context"
	"fmt"
	"strconv"

	"github.com/n0m-d/gator/internal/database"
)

func handlerBrowse(s *state, cmd command, user database.User) error {
	limit := int32(2)
	if len(cmd.Args) > 0 {
		parsed, err := strconv.Atoi(cmd.Args[0])
		if err != nil {
			return fmt.Errorf("invalid limit: %w", err)
		}
		limit = int32(parsed)
	}

	posts, err := s.db.GetPostsForUser(context.Background(), database.GetPostsForUserParams{
		UserID: user.ID,
		Limit:  limit,
	})
	if err != nil {
		return fmt.Errorf("couldn't get posts: %w", err)
	}

	for _, post := range posts {
		printPost(post)
	}

	return nil
}

func printPost(post database.Post) {
	fmt.Printf("* ID:            %s\n", post.ID)
	fmt.Printf("* Created:       %v\n", post.CreatedAt)
	fmt.Printf("* Updated:       %v\n", post.UpdatedAt)
	fmt.Printf("* Title:         %s\n", post.Title)
	fmt.Printf("* URL:           %s\n", post.Url)
	if post.Description.Valid {
		fmt.Printf("* Description:   %s\n", post.Description.String)
	}
	if post.PublishedAt.Valid {
		fmt.Printf("* Published:     %v\n", post.PublishedAt.Time)
	}
	fmt.Printf("* Feed ID:       %s\n", post.FeedID)
	fmt.Println()
}
