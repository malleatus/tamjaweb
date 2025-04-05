package browser

import (
	"time"
)

// Bookmark represents a browser bookmark
type Bookmark struct {
	Title      string
	URL        string
	DateAdded  time.Time
	FolderPath string
}

// Browser defines methods that all browser implementations must provide
type Browser interface {
	Name() string
	GetBookmarks(profile string) ([]Bookmark, error)
}

// RegisteredBrowsers is a slice of all available browser implementations
var RegisteredBrowsers []Browser

// RegisterBrowser adds a browser to the list of registered browsers
func RegisterBrowser(b Browser) {
	RegisteredBrowsers = append(RegisteredBrowsers, b)
}

// GetAllBookmarks returns bookmarks from all registered browsers
func GetAllBookmarks(profile string) (map[string][]Bookmark, error) {
	result := make(map[string][]Bookmark)

	for _, browser := range RegisteredBrowsers {
		bookmarks, err := browser.GetBookmarks(profile)
		if err != nil {
			continue // Skip this browser if there's an error
		}
		result[browser.Name()] = bookmarks
	}

	return result, nil
}
