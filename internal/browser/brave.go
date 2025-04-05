package browser

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/malleatus/tamjaweb/internal/logger"
)

var braveLogger = logger.New("browser:brave")

// BookmarksPathProvider is a function type that returns the path to bookmarks
type BookmarksPathProvider func(profile string) (string, error)

type Brave struct {
	getBookmarksPath BookmarksPathProvider
}

func init() {
	braveLogger.Info("Registering Brave browser")
	RegisterBrowser(NewBrave())
}

// NewBrave creates a new Brave browser instance with the default path provider
func NewBrave() *Brave {
	return &Brave{
		getBookmarksPath: getBraveBookmarksPath,
	}
}

func (b *Brave) Name() string {
	return "Brave"
}

// ChromiumBookmarks represents the structure of the Bookmarks JSON file
type ChromiumBookmarks struct {
	Checksum string `json:"checksum"`
	Roots    struct {
		BookmarkBar struct {
			Children []ChromiumBookmarkNode `json:"children"`
			Name     string                 `json:"name"`
			Type     string                 `json:"type"`
		} `json:"bookmark_bar"`
		Other struct {
			Children []ChromiumBookmarkNode `json:"children"`
			Name     string                 `json:"name"`
			Type     string                 `json:"type"`
		} `json:"other"`
	} `json:"roots"`
	Version int `json:"version"`
}

// ChromiumBookmarkNode represents a node in the bookmarks tree
type ChromiumBookmarkNode struct {
	DateAdded    json.Number            `json:"date_added"`
	DateLastUsed json.Number            `json:"date_last_used"`
	GUID         string                 `json:"guid"`
	ID           string                 `json:"id"`
	Name         string                 `json:"name"`
	Type         string                 `json:"type"`
	URL          string                 `json:"url,omitempty"`
	Children     []ChromiumBookmarkNode `json:"children,omitempty"`
}

// GetBookmarksPath is exposed for testing
func (b *Brave) GetBookmarksPath(profile string) (string, error) {
	return b.getBookmarksPath(profile)
}

func (b *Brave) GetBookmarks(profile string) ([]Bookmark, error) {
	bookmarksPath, err := b.getBookmarksPath(profile)
	if err != nil {
		braveLogger.Error("Failed to get Brave bookmarks path", "error", err)
		return nil, err
	}

	data, err := os.ReadFile(bookmarksPath)
	if err != nil {
		braveLogger.Error("Failed to read Brave bookmarks file", "path", bookmarksPath, "error", err)
		return nil, err
	}

	var chromeBookmarks ChromiumBookmarks
	if err := json.Unmarshal(data, &chromeBookmarks); err != nil {
		braveLogger.Error("Failed to unmarshal Brave bookmarks JSON", "path", bookmarksPath, "error", err)
		return nil, err
	}

	var bookmarks []Bookmark

	// Process bookmark bar
	processBookmarkNodes(&bookmarks, chromeBookmarks.Roots.BookmarkBar.Children, "Bookmark Bar")

	// Process other bookmarks
	processBookmarkNodes(&bookmarks, chromeBookmarks.Roots.Other.Children, "Other Bookmarks")

	return bookmarks, nil
}

// processBookmarkNodes recursively processes the bookmark nodes
func processBookmarkNodes(bookmarks *[]Bookmark, nodes []ChromiumBookmarkNode, folderPath string) {
	for _, node := range nodes {
		if node.Type == "url" {
			// Convert timestamp (microseconds since epoch) to time.Time
			dateAddedInt64, err := node.DateAdded.Int64()
			if err != nil {
				braveLogger.Error("Failed to convert date_added to int64", "value", node.DateAdded, "error", err)
				continue
			}

			// Windows epoch adjustment (difference between 1601 and 1970 in microseconds)
			windowsToUnixEpochDiff := int64(11644473600 * 1000000)
			unixMicroseconds := dateAddedInt64 - windowsToUnixEpochDiff
			dateAdded := time.Unix(0, unixMicroseconds*1000) // convert to nanoseconds

			*bookmarks = append(*bookmarks, Bookmark{
				Title:      node.Name,
				URL:        node.URL,
				DateAdded:  dateAdded,
				FolderPath: folderPath,
			})
		} else if node.Type == "folder" && len(node.Children) > 0 {
			// Recurse into folder
			newPath := filepath.Join(folderPath, node.Name)
			processBookmarkNodes(bookmarks, node.Children, newPath)
		}
	}
}

// getBraveBookmarksPathForPlatform returns the path to Brave bookmarks for the specified platform
func getBraveBookmarksPathForPlatform(goos, homeDir, localAppData, profile string) (string, error) {
	var path string

	switch goos {
	case "windows":
		path = filepath.Join(localAppData, "BraveSoftware", "Brave-Browser", "User Data", profile, "Bookmarks")
	case "darwin":
		path = filepath.Join(homeDir, "Library", "Application Support", "BraveSoftware", "Brave-Browser", profile, "Bookmarks")
	case "linux":
		path = filepath.Join(homeDir, ".config", "BraveSoftware", "Brave-Browser", profile, "Bookmarks")
	default:
		return "", fmt.Errorf("unsupported operating system: %s", goos)
	}

	return path, nil
}

// getBraveBookmarksPath returns the path to Brave bookmarks file based on OS
func getBraveBookmarksPath(profile string) (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	localAppData := ""
	if runtime.GOOS == "windows" {
		localAppData = os.Getenv("LOCALAPPDATA")
	}

	path, err := getBraveBookmarksPathForPlatform(runtime.GOOS, homeDir, localAppData, profile)
	if err != nil {
		return "", err
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return "", fmt.Errorf("Brave browser bookmarks not found at %s", path)
	}

	return path, nil
}
