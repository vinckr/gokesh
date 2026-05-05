package build

import (
	"encoding/xml"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type rssItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate,omitempty"`
}

type rssChannel struct {
	Title       string    `xml:"title"`
	Link        string    `xml:"link"`
	Description string    `xml:"description"`
	Items       []rssItem `xml:"item"`
}

type rssFeed struct {
	XMLName xml.Name   `xml:"rss"`
	Version string     `xml:"version,attr"`
	Channel rssChannel `xml:"channel"`
}

// GenerateRSS writes a valid RSS 2.0 feed.xml to outpath listing all dated non-draft pages.
// Skips generation if baseURL is empty.
func GenerateRSS(pages []PageSummary, outpath, siteTitle, baseURL string) error {
	if baseURL == "" {
		slog.Warn("skipping RSS feed: base_url not set in gokesh.toml")
		return nil
	}
	baseURL = strings.TrimRight(baseURL, "/")

	var items []rssItem
	for _, p := range pages {
		if p.Date.IsZero() {
			continue
		}
		items = append(items, rssItem{
			Title:       p.Title,
			Link:        p.URL,
			Description: p.Title,
			PubDate:     p.Date.Format(time.RFC1123Z),
		})
	}

	feed := rssFeed{
		Version: "2.0",
		Channel: rssChannel{
			Title:       siteTitle,
			Link:        baseURL + "/",
			Description: siteTitle,
			Items:       items,
		},
	}

	out, err := xml.MarshalIndent(feed, "", "  ")
	if err != nil {
		return fmt.Errorf("encoding RSS feed: %w", err)
	}

	dest := filepath.Join(outpath, "feed.xml")
	content := append([]byte(xml.Header), out...)
	if err := os.WriteFile(dest, content, 0644); err != nil {
		return fmt.Errorf("writing feed.xml: %w", err)
	}
	slog.Info("RSS feed written", "file", dest)
	return nil
}
