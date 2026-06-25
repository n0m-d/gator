package tui

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/n0m-d/gator/internal/database"
)

func TestViewPostsDoesNotPanic(t *testing.T) {
	posts := make([]database.Post, 5)
	for i := range posts {
		posts[i] = database.Post{
			ID:    uuid.New(),
			Title: "Test post",
			Url:   "https://example.com",
		}
	}

	m := model{
		username:     "tester",
		styles:       NewStyles(),
		activeTab:    tabPosts,
		width:        80,
		height:       24,
		posts:        posts,
		postsPerPage: 5,
		totalPosts:   50,
		postPage:     1,
		cursor:       2,
	}

	_ = m.View()
}

func TestViewFollowingDoesNotPanic(t *testing.T) {
	m := model{
		username:  "tester",
		styles:    NewStyles(),
		activeTab: tabFollowing,
		width:     80,
		height:    24,
		feeds: []database.GetFeedFollowsForUserRow{
			{FeedName: "HN", CreatedAt: time.Now()},
		},
		cursor: 0,
	}

	_ = m.View()
}

func TestViewEmptyPostsDoesNotPanic(t *testing.T) {
	m := model{
		username:   "tester",
		styles:     NewStyles(),
		activeTab:  tabPosts,
		width:      80,
		height:     24,
		totalPosts: 10,
		posts:      nil,
	}

	_ = m.View()
}
