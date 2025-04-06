package bookmarks

import (
	"bytes"
	"fmt"
	"text/tabwriter"

	"github.com/charmbracelet/log"
	"github.com/malleatus/tamjaweb/internal/browser"
	"github.com/malleatus/tamjaweb/internal/fzf"
)

type Options struct {
	Profile string
}

// FilterBookmarksByTerm filters bookmarks using fzf's filter functionality
// Returns a map of browser names to matching bookmarks
func FilterBookmarksByTerm(bookmarks map[string][]browser.Bookmark, term string) map[string][]browser.Bookmark {
	// If term is empty, return all bookmarks
	if term == "" {
		return bookmarks
	}

	// Create a slice to hold all bookmarks with their browser names
	type bookmarkEntry struct {
		browserName string
		bookmark    browser.Bookmark
	}

	var entries []bookmarkEntry

	// Populate entries
	for browserName, bookmarkList := range bookmarks {
		for _, bookmark := range bookmarkList {
			entries = append(entries, bookmarkEntry{
				browserName: browserName,
				bookmark:    bookmark,
			})
		}
	}

	// Create inputs for fzf
	inputs := make([]string, len(entries))
	for i, entry := range entries {
		// Create searchable text combining all fields (same as original)
		inputs[i] = fmt.Sprintf("%d\t%s\t%s",
			i,
			entry.bookmark.Title,
			entry.bookmark.URL,
		)
	}

	// Use the FZF utility to filter
	matchedIndices, err := fzf.FilterStrings(inputs, term)
	if err != nil {
		log.Error("Failed to filter bookmarks", "error", err)
		return make(map[string][]browser.Bookmark)
	}

	// Build filtered bookmarks map
	filteredBookmarks := make(map[string][]browser.Bookmark)

	for _, idx := range matchedIndices {
		entry := entries[idx]
		filteredBookmarks[entry.browserName] = append(
			filteredBookmarks[entry.browserName],
			entry.bookmark,
		)
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
