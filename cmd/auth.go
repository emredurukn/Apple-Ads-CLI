package cmd

import "github.com/spf13/cobra"

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Manage Apple Ads authentication credentials and profiles",
}

func init() {
	RootCmd.AddCommand(authCmd)
}
