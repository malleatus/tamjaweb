package cmd

import (
	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
)

// bookmarksCmd represents the bookmarks command
var bookmarksCmd = &cobra.Command{
	Use:   "bookmarks",
	Short: "list all bookmarks",
	Run: func(cmd *cobra.Command, args []string) {
		log.Info("bookmarks called")
	},
}

func init() {
	rootCmd.AddCommand(bookmarksCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// bookmarksCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// bookmarksCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
