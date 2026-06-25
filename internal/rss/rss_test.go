package rss

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

const sampleRSS = `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
  <channel>
    <title>Test &amp; Feed</title>
    <link>https://example.com</link>
    <description>A test feed</description>
    <item>
      <title>First Post</title>
      <link>https://example.com/1</link>
      <description>Hello &lt;world&gt;</description>
      <pubDate>Mon, 02 Jan 2006 15:04:05 MST</pubDate>
    </item>
  </channel>
</rss>`

func TestFetchParsesFeed(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("User-Agent") != "gator" {
			t.Errorf("User-Agent = %q, want gator", r.Header.Get("User-Agent"))
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(sampleRSS))
	}))
	defer srv.Close()

	feed, err := Fetch(context.Background(), srv.URL)
	if err != nil {
		t.Fatalf("Fetch() error: %v", err)
	}

	if feed.Channel.Title != "Test & Feed" {
		t.Errorf("title = %q, want %q", feed.Channel.Title, "Test & Feed")
	}
	if len(feed.Channel.Item) != 1 {
		t.Fatalf("items = %d, want 1", len(feed.Channel.Item))
	}
	item := feed.Channel.Item[0]
	if item.Title != "First Post" {
		t.Errorf("item title = %q", item.Title)
	}
	if item.Description != "Hello <world>" {
		t.Errorf("item description = %q", item.Description)
	}
}

func TestFetchRejectsNonOKStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	_, err := Fetch(context.Background(), srv.URL)
	if err == nil {
		t.Fatal("expected error for 404 response")
	}
}

func TestParsePublishedAt(t *testing.T) {
	tests := []struct {
		input string
		valid bool
	}{
		{"", false},
		{"not a date", false},
		{"Mon, 02 Jan 2006 15:04:05 MST", true},
		{"2006-01-02T15:04:05Z", true},
	}

	for _, tt := range tests {
		got := ParsePublishedAt(tt.input)
		if got.Valid != tt.valid {
			t.Errorf("ParsePublishedAt(%q).Valid = %v, want %v", tt.input, got.Valid, tt.valid)
		}
	}

	parsed := ParsePublishedAt("Mon, 02 Jan 2006 15:04:05 MST")
	if !parsed.Valid {
		t.Fatal("expected valid time")
	}
	if parsed.Time.Year() != 2006 || parsed.Time.Month() != time.January || parsed.Time.Day() != 2 {
		t.Errorf("unexpected date: %v", parsed.Time)
	}
	if parsed.Time.Hour() != 15 || parsed.Time.Minute() != 4 || parsed.Time.Second() != 5 {
		t.Errorf("unexpected time of day: %v", parsed.Time)
	}
}
