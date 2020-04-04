package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/cameronstanley/go-reddit"
	"github.com/gorilla/feeds"
)

func handler(w http.ResponseWriter, r *http.Request) {
	url := fmt.Sprintf("https://reddit.com%s", r.URL)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		http.Error(w, err.Error(), 500)
	}

	req.Header.Add("User-Agent", "reddit-rss 1.0")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		http.Error(w, err.Error(), 500)
	}
	defer resp.Body.Close()

	var result linkListing
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		http.Error(w, err.Error(), 500)
	}

	feed := &feeds.Feed{
		Title:       fmt.Sprintf("reddit-rss %s", r.URL),
		Link:        &feeds.Link{Href: "https://github.com/trashhalo/reddit-rss"},
		Description: "Reddit RSS feed that links directly to the content",
		Author:      &feeds.Author{Name: "Stephen Solka", Email: "s@0qz.fun"},
		Created:     time.Now(),
	}

	for _, link := range result.Data.Children {
		feed.Items = append(feed.Items, linkToFeed(&link.Data))
	}

	rss, err := feed.ToRss()
	if err != nil {
		http.Error(w, err.Error(), 500)
	}

	w.Header().Set("Content-Type", "application/rss+xml")
	io.WriteString(w, rss)
}

func linkToFeed(link *reddit.Link) *feeds.Item {
	t := time.Unix(int64(link.CreatedUtc), 0)
	return &feeds.Item{
		Title:   link.Title,
		Link:    &feeds.Link{Href: link.URL},
		Author:  &feeds.Author{Name: link.Author},
		Created: t,
	}
}

func main() {
	log.Println("starting reddit-rss")

	http.HandleFunc("/", handler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), nil))
}