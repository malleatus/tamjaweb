package browser

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetBraveBookmarksPathForPlatform(t *testing.T) {
	testCases := []struct {
		name         string
		goos         string
		homeDir      string
		localAppData string
		profile      string
		expected     string
		expectError  bool
	}{
		{
			name:         "Windows path",
			goos:         "windows",
			homeDir:      `C:\Users\testuser`,
			localAppData: `C:\Users\testuser\AppData\Local`,
			profile:      "Default",
			expected:     filepath.Join(`C:\Users\testuser\AppData\Local`, "BraveSoftware", "Brave-Browser", "User Data", "Default", "Bookmarks"),
		},
		{
			name:     "macOS path",
			goos:     "darwin",
			homeDir:  "/Users/testuser",
			profile:  "Default",
			expected: filepath.Join("/Users/testuser", "Library", "Application Support", "BraveSoftware", "Brave-Browser", "Default", "Bookmarks"),
		},
		{
			name:     "Linux path",
			goos:     "linux",
			homeDir:  "/home/testuser",
			profile:  "Default",
			expected: filepath.Join("/home/testuser", ".config", "BraveSoftware", "Brave-Browser", "Default", "Bookmarks"),
		},
		{
			name:        "Unsupported OS",
			goos:        "solaris",
			homeDir:     "/home/testuser",
			profile:     "Default",
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			path, err := getBraveBookmarksPathForPlatform(tc.goos, tc.homeDir, tc.localAppData, tc.profile)

			if tc.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "unsupported operating system")
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expected, path)
			}
		})
	}
}

func TestGetBookmarksWithMockFile(t *testing.T) {
	// Create a temporary directory that will be cleaned up after the test
	tmpDir, err := os.MkdirTemp("", "brave-test")
	require.NoError(t, err)
	defer func() {
		err := os.RemoveAll(tmpDir)
		if err != nil {
			t.Errorf("Failed to remove temporary directory: %v", err)
		}
	}()

	// Create nested directories
	bookmarksDir := filepath.Join(tmpDir, "Default")
	err = os.MkdirAll(bookmarksDir, 0755)
	require.NoError(t, err)

	// Create a sample bookmarks file
	bookmarksFile := filepath.Join(bookmarksDir, "Bookmarks")
	sampleJSON := createSampleBookmarksJSON()

	err = os.WriteFile(bookmarksFile, []byte(sampleJSON), 0644)
	require.NoError(t, err)

	// Create a custom Brave instance that uses our mock location
	brave := &Brave{
		getBookmarksPath: func(profile string) (string, error) {
			return bookmarksFile, nil
		},
	}

	// Call GetBookmarks
	bookmarks, err := brave.GetBookmarks("Default")
	require.NoError(t, err)

	// Validate the results
	assert.Equal(t, 3, len(bookmarks), "Should have 3 bookmarks total")

	// Check bookmark bar entry
	assert.Equal(t, "Example Site", bookmarks[0].Title)
	assert.Equal(t, "https://example.com", bookmarks[0].URL)
	assert.Equal(t, "Bookmark Bar", bookmarks[0].FolderPath)

	// Check nested folder entry
	assert.Equal(t, "GitHub", bookmarks[1].Title)
	assert.Equal(t, "https://github.com", bookmarks[1].URL)
	assert.Equal(t, filepath.Join("Bookmark Bar", "Work"), bookmarks[1].FolderPath)

	// Check "Other Bookmarks" entry
	assert.Equal(t, "Other Site", bookmarks[2].Title)
	assert.Equal(t, "https://othersite.com", bookmarks[2].URL)
	assert.Equal(t, "Other Bookmarks", bookmarks[2].FolderPath)
}

func TestProcessBookmarkNodes(t *testing.T) {
	testCases := []struct {
		name           string
		nodes          []ChromiumBookmarkNode
		folderPath     string
		expectedCount  int
		expectedTitles []string
		expectedURLs   []string
		expectedPaths  []string
	}{
		{
			name: "Simple URL bookmarks",
			nodes: []ChromiumBookmarkNode{
				{
					Type:      "url",
					Name:      "Example Site",
					URL:       "https://example.com",
					DateAdded: json.Number("13214422057039153"),
				},
			},
			folderPath:     "Bookmark Bar",
			expectedCount:  1,
			expectedTitles: []string{"Example Site"},
			expectedURLs:   []string{"https://example.com"},
			expectedPaths:  []string{"Bookmark Bar"},
		},
		{
			name: "Nested folder structure",
			nodes: []ChromiumBookmarkNode{
				{
					Type: "folder",
					Name: "Work",
					Children: []ChromiumBookmarkNode{
						{
							Type:      "url",
							Name:      "GitHub",
							URL:       "https://github.com",
							DateAdded: json.Number("13214422057039154"),
						},
					},
				},
			},
			folderPath:     "Bookmark Bar",
			expectedCount:  1,
			expectedTitles: []string{"GitHub"},
			expectedURLs:   []string{"https://github.com"},
			expectedPaths:  []string{filepath.Join("Bookmark Bar", "Work")},
		},
		{
			name: "Invalid DateAdded",
			nodes: []ChromiumBookmarkNode{
				{
					Type:      "url",
					Name:      "Bad Date",
					URL:       "https://example.com",
					DateAdded: json.Number("invalid"),
				},
			},
			folderPath:    "Bookmark Bar",
			expectedCount: 0,
		},
		{
			name: "Empty folder",
			nodes: []ChromiumBookmarkNode{
				{
					Type:     "folder",
					Name:     "Empty Folder",
					Children: []ChromiumBookmarkNode{},
				},
			},
			folderPath:    "Bookmark Bar",
			expectedCount: 0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var bookmarks []Bookmark
			processBookmarkNodes(&bookmarks, tc.nodes, tc.folderPath)

			assert.Equal(t, tc.expectedCount, len(bookmarks))

			for i := range tc.expectedCount {
				assert.Equal(t, tc.expectedTitles[i], bookmarks[i].Title)
				assert.Equal(t, tc.expectedURLs[i], bookmarks[i].URL)
				assert.Equal(t, tc.expectedPaths[i], bookmarks[i].FolderPath)
			}
		})
	}
}

// Helper function to create sample bookmarks JSON
func createSampleBookmarksJSON() string {
	return `{
		"checksum": "test-checksum",
		"roots": {
			"bookmark_bar": {
				"children": [
					{
						"date_added": "13214422057039153",
						"date_last_used": "13214422057039153",
						"guid": "guid1",
						"id": "1",
						"name": "Example Site",
						"type": "url",
						"url": "https://example.com"
					},
					{
						"children": [
							{
								"date_added": "13214422057039154",
								"date_last_used": "13214422057039154",
								"guid": "guid2",
								"id": "2",
								"name": "GitHub",
								"type": "url",
								"url": "https://github.com"
							}
						],
						"date_added": "13214422057039155",
						"date_last_used": "13214422057039155",
						"guid": "guid3",
						"id": "3",
						"name": "Work",
						"type": "folder"
					}
				],
				"name": "Bookmark Bar",
				"type": "folder"
			},
			"other": {
				"children": [
					{
						"date_added": "13214422057039156",
						"date_last_used": "13214422057039156",
						"guid": "guid4",
						"id": "4",
						"name": "Other Site",
						"type": "url",
						"url": "https://othersite.com"
					}
				],
				"name": "Other Bookmarks",
				"type": "folder"
			}
		},
		"version": 1
	}`
}
