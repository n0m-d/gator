package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/n0m-d/gator/internal/database"
)

func handlerFollowFeed(state *state, cmd command, user database.User) error {
	if len(cmd.Args) != 1 {
		return errors.New("follow requires one argument")
	}

	url := cmd.Args[0]

	feed, err := state.db.GetFeedByURL(context.Background(), url)
	if err != nil {
		return fmt.Errorf("couldn't get feed: %w", err)
	}

	fmt.Printf("User ID: %v\n", user.ID)
	fmt.Printf("Feed ID: %v\n", feed.ID)

	data, err := insertFeedFollow(state, user.ID, feed.ID)
	if err != nil {
		return fmt.Errorf("couldn't follow feed: %w", err)
	}

	fmt.Printf("Follow created: %v\n", data)

	return nil
}

func insertFeedFollow(state *state, userID uuid.UUID, feedID uuid.UUID) (database.CreateFeedFollowRow, error) {
	return state.db.CreateFeedFollow(context.Background(), database.CreateFeedFollowParams{
		ID:     uuid.New(),
		UserID: userID,
		FeedID: feedID,
	})
}

func handlerFollowingFeed(state *state, cmd command, user database.User) error {

	feedFollows, err := state.db.GetFeedFollowsForUser(context.Background(), user.ID)
	if err != nil {
		return fmt.Errorf("couldn't get feed follows: %w", err)
	}

	fmt.Printf("Following %d feeds:\n", len(feedFollows))
	for _, follow := range feedFollows {
		fmt.Printf("%s\n", follow.FeedName)
	}

	return nil
}

func handlerUnfollowFeed(state *state, cmd command, user database.User) error {

	if len(cmd.Args) != 1 {
		return errors.New("follow requires one argument")
	}

	url := cmd.Args[0]

	feed, err := state.db.GetFeedByURL(context.Background(), url)
	if err != nil {
		return fmt.Errorf("couldn't get feed: %w", err)
	}

	fmt.Printf("Unfollowing feed: %s\n", feed)

	err = state.db.DeleteFeedFollowByFeedIdAndUserId(context.Background(), database.DeleteFeedFollowByFeedIdAndUserIdParams{
		FeedID: feed.ID,
		UserID: user.ID,
	})
	if err != nil {
		return fmt.Errorf("couldn't unfollow feed: %w", err)
	}

	fmt.Printf("Unfollowed feed successfully\n")
	return nil
}
