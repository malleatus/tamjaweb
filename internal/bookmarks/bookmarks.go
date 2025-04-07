package bookmarks

import (
	"bytes"
	"fmt"

	"github.com/charmbracelet/log"
	"github.com/malleatus/tamjaweb/internal/browser"
	"github.com/malleatus/tamjaweb/internal/fzf"
	"github.com/olekukonko/tablewriter"
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
	if len(bookmarks) == 0 {
		return "No bookmarks found", nil
	}

	var buf bytes.Buffer
	table := tablewriter.NewWriter(&buf)

	table.SetHeader([]string{"Browser", "Title", "URL", "Folder", "Date Added"})
	table.SetAutoWrapText(true)
	table.SetColumnAlignment([]int{
		tablewriter.ALIGN_LEFT,
		tablewriter.ALIGN_LEFT,
		tablewriter.ALIGN_LEFT,
		tablewriter.ALIGN_LEFT,
		tablewriter.ALIGN_LEFT,
	})
	table.SetColWidth(50)

	for browserName, bookmarkList := range bookmarks {
		for _, bookmark := range bookmarkList {
			table.Append([]string{
				browserName,
				bookmark.Title,
				bookmark.URL,
				bookmark.FolderPath,
				bookmark.DateAdded.Format("2006-01-02 15:04:05"),
			})
		}
	}

	table.Render()

	return buf.String(), nil
}
