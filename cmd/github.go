package cmd

import (
	"github.com/charmbracelet/log"
	githubcmd "github.com/malleatus/tamjaweb/cmd/github"
	github "github.com/malleatus/tamjaweb/internal/github"
	"github.com/spf13/cobra"
)

func init() {
	opts := &github.Options{}

	cmd := &cobra.Command{
		Use:   "github",
		Short: "GitHub Utilities",
	}

	cmd.PersistentFlags().StringVar(&opts.User, "user", "", "GitHub user to use")
	err := cmd.MarkPersistentFlagRequired("user")
	// TODO handle the error properly, log.Error and exit non-zero this should not happen in normal circumstances
	if err != nil {
		log.Error("Failed to mark user flag as required", "error", err)
		return
	}

	cmd.AddCommand(githubcmd.NewStarsCommand(opts))

	rootCmd.AddCommand(cmd)
}
