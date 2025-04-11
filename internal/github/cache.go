package github

import (
	"github.com/malleatus/tamjaweb/internal/cache"
)

// getStarsCache returns the cache for stars
func getStarsCache() (*cache.CacheStore[Star], error) {
	return cache.New[Star]("stars.json")
}

// GetCachedStars retrieves all stars from the cache
func GetCachedStars() ([]Star, error) {
	starsCache, err := getStarsCache()
	if err != nil {
		return nil, err
	}
	return starsCache.Read()
}

// WriteCachedStars updates the cache with stars for a specific stargazer
func WriteCachedStars(stargazer string, stars []Star) error {
	starsCache, err := getStarsCache()
	if err != nil {
		return err
	}

	// Filter to remove existing stars for this stargazer
	stargazerFilter := func(star Star) bool {
		return star.Stargazer == stargazer
	}

	return starsCache.UpdateWithFilter(stargazerFilter, stars)
}
