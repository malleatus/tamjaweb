package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/charmbracelet/log"
	"github.com/malleatus/tamjaweb/internal/cache"
	"github.com/spf13/cobra"
)

func newCacheCommand() *cobra.Command {
	var clearCache bool

	cmd := &cobra.Command{
		Use:   "cache",
		Short: "Manage tamjaweb cache",
		Run: func(cmd *cobra.Command, args []string) {
			out := cmd.OutOrStdout()
			cacheDir, err := cache.GetCacheDir()
			if err != nil {
				log.Error("Failed to get cache directory", "error", err)
				return
			}

			if clearCache {
				entries, err := os.ReadDir(cacheDir)
				if err != nil {
					if os.IsNotExist(err) {
						_, err = fmt.Fprintln(out, "Cache directory does not exist, nothing to clear")
						if err != nil {
							log.Error("[INTERNAL] failed to print to stdout")
						}
						return
					}
					log.Error("Failed to read cache directory", "error", err)
					return
				}

				for _, entry := range entries {
					if !entry.IsDir() {
						filePath := filepath.Join(cacheDir, entry.Name())
						err := os.Remove(filePath)
						if err != nil {
							log.Error("Failed to remove cache file", "file", entry.Name(), "error", err)
						}
					}
				}
				_, err = fmt.Fprintln(out, "Cache cleared successfully")
				if err != nil {
					log.Error("[INTERNAL] failed to print to stdout")
				}
			} else {
				// Just print the cache location
				_, err = fmt.Fprintln(out, "Cache location:", cacheDir)
				if err != nil {
					log.Error("[INTERNAL] failed to print to stdout")
				}
			}
		},
	}

	cmd.Flags().BoolVar(&clearCache, "clear", false, "Clear all cached files")

	return cmd
}

func init() {
	rootCmd.AddCommand(newCacheCommand())
}
