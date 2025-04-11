package cache

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type CacheTestSuite struct {
	suite.Suite

	originalHomeDir string
	tempHomeDir     string
}

// SetupTest runs before each test in the suite.
func (s *CacheTestSuite) SetupTest() {
	homeDir := os.Getenv("HOME")

	s.originalHomeDir = homeDir
	tempHomeDir, err := os.MkdirTemp("", "tamjaweb-test-cache")
	s.NoError(err)

	err = os.Setenv("HOME", tempHomeDir)
	s.NoError(err)
	s.tempHomeDir = tempHomeDir
}

// TearDownTest runs after each test in the suite.
func (s *CacheTestSuite) TearDownTest() {
	err := os.Setenv("HOME", s.originalHomeDir)
	s.NoError(err)
	defer func() {
		err := os.RemoveAll(s.tempHomeDir)
		s.NoError(err)
	}()
}

// TestItem is a simple struct for testing cache operations
type TestItem struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func (s *CacheTestSuite) Test_GetCacheDir() {
	cacheDir, err := GetCacheDir()
	s.NoError(err)

	s.Equal(filepath.Join(s.tempHomeDir, ".cache/tamjaweb"), cacheDir)
}

func (s *CacheTestSuite) Test_New() {
	t := s.T()
	// Create a temporary directory to use as the home directory
	tempHome := t.TempDir()

	// Save original home directory setting and restore after test
	t.Setenv("HOME", tempHome)

	cache, err := New[TestItem]("test-cache.json")
	assert.NoError(t, err)
	assert.Contains(t, cache.filePath, "test-cache.json")

	// Verify the cache directory was created
	cacheDir, err := GetCacheDir()
	assert.NoError(t, err)
	_, err = os.Stat(cacheDir)
	assert.NoError(t, err, "Cache directory should exist")
}

func (s *CacheTestSuite) Test_CacheStore_Read() {
	t := s.T()

	t.Run("empty cache", func(t *testing.T) {
		cache, err := New[TestItem]("empty-cache.json")
		s.NoError(err)

		items, err := cache.Read()
		s.NoError(err)
		s.Empty(items)
	})

	t.Run("populated cache", func(t *testing.T) {
		initialItems := []TestItem{
			{ID: 1, Name: "Item 1"},
			{ID: 2, Name: "Item 2"},
		}

		cache, err := New[TestItem]("populated-cache.json")
		s.NoError(err)
		err = cache.Write(initialItems)
		s.NoError(err)

		items, err := cache.Read()
		assert.NoError(t, err)
		assert.Len(t, items, 2)
		assert.Equal(t, initialItems, items)
	})
}

func (s *CacheTestSuite) Test_CacheStore_Write() {
	cache, err := New[TestItem]("write-test.json")
	s.NoError(err)

	items := []TestItem{
		{ID: 1, Name: "Item 1"},
		{ID: 2, Name: "Item 2"},
	}

	err = cache.Write(items)
	s.NoError(err)

	// Verify the file exists and contains the expected content
	data, err := os.ReadFile(cache.filePath)
	s.NoError(err)
	s.Contains(string(data), `"id": 1`)
	s.Contains(string(data), `"name": "Item 1"`)

	// Verify we can read the items back
	readItems, err := cache.Read()
	s.NoError(err)
	s.Equal(items, readItems)
}

func (s *CacheTestSuite) Test_CacheStore_UpdateWithFilter() {
	initialItems := []TestItem{
		{ID: 1, Name: "Item 1"},
		{ID: 2, Name: "Item 2"},
		{ID: 3, Name: "Item 3"},
	}

	cache, err := New[TestItem]("update-test.json")
	s.NoError(err)
	err = cache.Write(initialItems)
	s.NoError(err)

	// Filter out items with ID 2
	filter := func(item TestItem) bool {
		return item.ID == 2
	}

	// Add new items
	newItems := []TestItem{
		{ID: 4, Name: "Item 4"},
	}

	err = cache.UpdateWithFilter(filter, newItems)
	s.NoError(err)

	// Verify the result
	items, err := cache.Read()
	s.NoError(err)
	s.Len(items, 3) // 3 initial - 1 filtered + 1 new = 3

	// The filtered item (ID=2) should be gone
	for _, item := range items {
		s.NotEqual(2, item.ID)
	}

	// Check for the new item
	found := false
	for _, item := range items {
		if item.ID == 4 {
			found = true
			s.Equal("Item 4", item.Name)
			break
		}
	}
	s.True(found, "New item should be in the cache")
}

func TestCacheSuite(t *testing.T) {
	suite.Run(t, new(CacheTestSuite))
}
