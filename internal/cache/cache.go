package cache

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// CacheStore manages generic caching for any type
type CacheStore[T any] struct {
	filePath string
}

func GetCacheDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	cacheDir := filepath.Join(homeDir, ".cache", "tamjaweb")
	return cacheDir, nil
}

// New creates a new cache instance for type T
func New[T any](fileName string) (*CacheStore[T], error) {
	cacheDir, err := GetCacheDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get cache directory: %w", err)
	}

	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create cache directory: %w", err)
	}

	return &CacheStore[T]{
		filePath: filepath.Join(cacheDir, fileName),
	}, nil
}

// Read gets all items from the cache
func (c *CacheStore[T]) Read() ([]T, error) {
	data, err := os.ReadFile(c.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return []T{}, nil // Return empty slice if no cache exists
		}
		return nil, fmt.Errorf("failed to read cache: %w", err)
	}

	var items []T
	if err := json.Unmarshal(data, &items); err != nil {
		return nil, fmt.Errorf("failed to parse cache: %w", err)
	}

	return items, nil
}

// Write stores items to the cache
func (c *CacheStore[T]) Write(items []T) error {
	data, err := json.MarshalIndent(items, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to encode cache: %w", err)
	}

	return os.WriteFile(c.filePath, data, 0644)
}

// UpdateWithFilter updates cache by removing items that match the filter and adding new ones
func (c *CacheStore[T]) UpdateWithFilter(filter func(T) bool, newItems []T) error {
	currentItems, err := c.Read()
	if err != nil {
		return err
	}

	// Keep items that don't match the filter
	filteredItems := make([]T, 0)
	for _, item := range currentItems {
		if !filter(item) {
			filteredItems = append(filteredItems, item)
		}
	}

	// Add new items and save
	updatedItems := append(filteredItems, newItems...)
	return c.Write(updatedItems)
}

// IsOutdated checks if the cache is older than the specified duration
func (c *CacheStore[T]) IsOutdated(maxAge time.Duration) (bool, error) {
	info, err := os.Stat(c.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return true, nil
		}
		return false, err
	}
	return time.Since(info.ModTime()) > maxAge, nil
}
