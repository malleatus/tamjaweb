package bookmarks

import (
	"bytes"
	"fmt"
	"strings"
	"text/tabwriter"

	"github.com/malleatus/tamjaweb/internal/browser"
)

type Options struct {
	Profile string
}

// filterBookmarksByTerm filters bookmarks that match the provided search term
// Returns a map of browser names to matching bookmarks
func FilterBookmarksByTerm(bookmarks map[string][]browser.Bookmark, term string) map[string][]browser.Bookmark {
	filteredBookmarks := make(map[string][]browser.Bookmark)
	for browserName, browserBookmarks := range bookmarks {
		var matches []browser.Bookmark
		for _, bookmark := range browserBookmarks {
			if strings.Contains(strings.ToLower(bookmark.Title), strings.ToLower(term)) ||
				strings.Contains(strings.ToLower(bookmark.URL), strings.ToLower(term)) ||
				strings.Contains(strings.ToLower(bookmark.FolderPath), strings.ToLower(term)) {
				matches = append(matches, bookmark)
			}
		}
		if len(matches) > 0 {
			filteredBookmarks[browserName] = matches
		}
	}
	return filteredBookmarks
}

// prints the bookmarks in a tabular format
func PrintBookmarks(bookmarks map[string][]browser.Bookmark) (string, error) {
	var buf bytes.Buffer
	w := tabwriter.NewWriter(&buf, 0, 0, 2, ' ', 0)

	if _, err := fmt.Fprintln(w, "Browser\tTitle\tURL\tFolder\tDate Added"); err != nil {
		return "", fmt.Errorf("failed to write header: %w", err)
	}
	if _, err := fmt.Fprintln(w, "-------\t-----\t---\t------\t----------"); err != nil {
		return "", fmt.Errorf("failed to write separator: %w", err)
	}

	// Format bookmarks from each browser
	for browserName, bookmarkList := range bookmarks {
		for _, bookmark := range bookmarkList {
			if _, err := fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
				browserName,
				bookmark.Title,
				bookmark.URL,
				bookmark.FolderPath,
				bookmark.DateAdded.Format("2006-01-02 15:04:05"),
			); err != nil {
				return "", fmt.Errorf("failed to write bookmark: %w", err)
			}
		}
	}

	if err := w.Flush(); err != nil {
		return "", fmt.Errorf("failed to flush output: %w", err)
	}

	return buf.String(), nil
}
