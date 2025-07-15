package main

import (
	"context"
	"encoding/xml"
	"html"
	"io"
	"net/http"
)

type RSSFeed struct {
	Channel struct {
		Title       string    `xml:"title"`
		Link        string    `xml:"link"`
		Description string    `xml:"description"`
		Item        []RSSItem `xml:"item"`
	} `xml:"channel"`
}

type RSSItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
}

func fetchFeed(ctx context.Context, feedURL string) (*RSSFeed, error) {
	rss := RSSFeed{}

	// make the request string
	req, err := http.NewRequestWithContext(context.Background(), "GET", feedURL, nil)
	if err != nil {
		return &rss, err
	}
	// set the headers
	req.Header.Set("User-Agent", "gator")

	// make the request
	client := http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return &rss, err
	}
	defer res.Body.Close()

	// read the response
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return &rss, err
	}

	// unmarshal the response into the RSSFeed
	if err := xml.Unmarshal(body, &rss); err != nil {
		return &rss, err
	}

	rss.Channel.Title = html.UnescapeString(rss.Channel.Title)
	rss.Channel.Description = html.UnescapeString(rss.Channel.Description)
	for i, _ := range rss.Channel.Item {
		rss.Channel.Item[i].Title = html.UnescapeString(rss.Channel.Item[i].Title)
		rss.Channel.Item[i].Description = html.UnescapeString(rss.Channel.Item[i].Description)
	}
	return &rss, nil
}
