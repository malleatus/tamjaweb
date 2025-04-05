package cmd

import (
	"github.com/malleatus/tamjaweb/cmd/bookmarks"
	internalBookmarks "github.com/malleatus/tamjaweb/internal/bookmarks"
	"github.com/spf13/cobra"
)

func init() {
	opts := &internalBookmarks.Options{}

	cmd := &cobra.Command{
		Use:   "bookmarks",
		Short: "Manage browser bookmarks",
	}

	cmd.PersistentFlags().StringVar(&opts.Profile, "profile", "Default", "Browser profile to use")

	cmd.AddCommand(bookmarks.NewSearchCommand(opts))
	cmd.AddCommand(bookmarks.NewListCommand(opts))

	rootCmd.AddCommand(cmd)
}
