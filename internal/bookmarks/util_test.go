package bookmarks

import (
	"testing"
	"time"

	"github.com/bradleyjkemp/cupaloy/v2"
	"github.com/malleatus/tamjaweb/internal/browser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFilterBookmarksByTerm(t *testing.T) {
	fixedTime := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)

	bookmarks := map[string][]browser.Bookmark{
		"TestBrowser1": {
			{
				Title:      "GitHub Homepage",
				URL:        "https://github.com",
				DateAdded:  fixedTime,
				FolderPath: "Dev/Resources",
			},
			{
				Title:      "Stack Overflow",
				URL:        "https://stackoverflow.com",
				DateAdded:  fixedTime,
				FolderPath: "Dev/Resources",
			},
		},
		"TestBrowser2": {
			{
				Title:      "Google",
				URL:        "https://google.com",
				DateAdded:  fixedTime,
				FolderPath: "Search Engines",
			},
			{
				Title:      "GitHub Projects",
				URL:        "https://github.com/projects",
				DateAdded:  fixedTime,
				FolderPath: "Dev/Projects",
			},
		},
	}

	testCases := []struct {
		name            string
		searchTerm      string
		expectedResults map[string][]struct {
			title string
			url   string
			path  string
		}
	}{
		{
			name:       "Filter by title exact match",
			searchTerm: "GitHub Homepage",
			expectedResults: map[string][]struct {
				title string
				url   string
				path  string
			}{
				"TestBrowser1": {
					{
						title: "GitHub Homepage",
						url:   "https://github.com",
						path:  "Dev/Resources",
					},
				},
			},
		},
		{
			name:       "Filter by title partial match",
			searchTerm: "GitHub",
			expectedResults: map[string][]struct {
				title string
				url   string
				path  string
			}{
				"TestBrowser1": {
					{
						title: "GitHub Homepage",
						url:   "https://github.com",
						path:  "Dev/Resources",
					},
				},
				"TestBrowser2": {
					{
						title: "GitHub Projects",
						url:   "https://github.com/projects",
						path:  "Dev/Projects",
					},
				},
			},
		},
		{
			name:       "Filter by title partial match -- lower-case",
			searchTerm: "github",
			expectedResults: map[string][]struct {
				title string
				url   string
				path  string
			}{
				"TestBrowser1": {
					{
						title: "GitHub Homepage",
						url:   "https://github.com",
						path:  "Dev/Resources",
					},
				},
				"TestBrowser2": {
					{
						title: "GitHub Projects",
						url:   "https://github.com/projects",
						path:  "Dev/Projects",
					},
				},
			},
		},
		{
			name:       "Filter by fzf style match",
			searchTerm: "ghub",
			expectedResults: map[string][]struct {
				title string
				url   string
				path  string
			}{
				"TestBrowser1": {
					{
						title: "GitHub Homepage",
						url:   "https://github.com",
						path:  "Dev/Resources",
					},
				},
				"TestBrowser2": {
					{
						title: "GitHub Projects",
						url:   "https://github.com/projects",
						path:  "Dev/Projects",
					},
				},
			},
		},
		{
			name:       "Filter by URL",
			searchTerm: "stackoverflow",
			expectedResults: map[string][]struct {
				title string
				url   string
				path  string
			}{
				"TestBrowser1": {
					{
						title: "Stack Overflow",
						url:   "https://stackoverflow.com",
						path:  "Dev/Resources",
					},
				},
			},
		},
		{
			name:       "Filter by folder path",
			searchTerm: "Projects",
			expectedResults: map[string][]struct {
				title string
				url   string
				path  string
			}{
				"TestBrowser2": {
					{
						title: "GitHub Projects",
						url:   "https://github.com/projects",
						path:  "Dev/Projects",
					},
				},
			},
		},
		{
			name:       "Case insensitive search",
			searchTerm: "github",
			expectedResults: map[string][]struct {
				title string
				url   string
				path  string
			}{
				"TestBrowser1": {
					{
						title: "GitHub Homepage",
						url:   "https://github.com",
						path:  "Dev/Resources",
					},
				},
				"TestBrowser2": {
					{
						title: "GitHub Projects",
						url:   "https://github.com/projects",
						path:  "Dev/Projects",
					},
				},
			},
		},
		{
			name:       "No matches",
			searchTerm: "nonexistent",
			expectedResults: map[string][]struct {
				title string
				url   string
				path  string
			}{},
		},
		{
			name:       "Empty search term matches all",
			searchTerm: "",
			expectedResults: map[string][]struct {
				title string
				url   string
				path  string
			}{
				"TestBrowser1": {
					{
						title: "GitHub Homepage",
						url:   "https://github.com",
						path:  "Dev/Resources",
					},
					{
						title: "Stack Overflow",
						url:   "https://stackoverflow.com",
						path:  "Dev/Resources",
					},
				},
				"TestBrowser2": {
					{
						title: "Google",
						url:   "https://google.com",
						path:  "Search Engines",
					},
					{
						title: "GitHub Projects",
						url:   "https://github.com/projects",
						path:  "Dev/Projects",
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := FilterBookmarksByTerm(bookmarks, tc.searchTerm)

			assert.Equal(t, len(tc.expectedResults), len(result), "Number of browsers with matches")

			if len(tc.expectedResults) == 0 {
				assert.Empty(t, result, "Result should be empty when no matches")
				return
			}

			for browserName, expectedBookmarks := range tc.expectedResults {
				actualBookmarks, ok := result[browserName]
				assert.True(t, ok, "Browser %s should exist in filtered results", browserName)
				assert.Equal(t, len(expectedBookmarks), len(actualBookmarks), "Number of bookmarks for browser: %s", browserName)

				for i, expectedBookmark := range expectedBookmarks {
					assert.Equal(t, expectedBookmark.title, actualBookmarks[i].Title, "Bookmark title for browser %s at index %d", browserName, i)
					assert.Equal(t, expectedBookmark.url, actualBookmarks[i].URL, "Bookmark URL for browser %s at index %d", browserName, i)
					assert.Equal(t, expectedBookmark.path, actualBookmarks[i].FolderPath, "Bookmark folder path for browser %s at index %d", browserName, i)
				}
			}
		})
	}
}

func TestPrintBookmarks(t *testing.T) {
	fixedTime := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)
	bookmarks := map[string][]browser.Bookmark{
		"TestBrowser": {
			{
				Title:      "Test Bookmark",
				URL:        "https://example.com",
				DateAdded:  fixedTime,
				FolderPath: "Test Folder",
			},
		},
	}

	output, err := PrintBookmarks(bookmarks)
	require.NoError(t, err)

	// Snapshot test replaces multiple assert statements
	cupaloy.SnapshotT(t, output)
}

func TestPrintBookmarksEmpty(t *testing.T) {
	bookmarks := map[string][]browser.Bookmark{}

	output, err := PrintBookmarks(bookmarks)
	require.NoError(t, err)

	cupaloy.SnapshotT(t, output)
}
