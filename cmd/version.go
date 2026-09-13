package cmd

import (
	"os"

	"github.com/emredurukan/asactl/internal/version"
	"github.com/emredurukan/asactl/pkg/output"
	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Display version and build information",
	RunE: func(cmd *cobra.Command, args []string) error {
		info := version.Get()
		fmtType := output.ResolveFormat(GetRequestedOutput())

		if fmtType == output.FormatJSON {
			return output.RenderJSON(os.Stdout, info, true)
		}

		headers := []string{"FIELD", "VALUE"}
		rows := [][]string{
			{"Version", info.Version},
			{"Git Commit", info.GitCommit},
			{"Build Date", info.BuildDate},
			{"Go Version", info.GoVersion},
			{"Platform", info.Platform},
		}

		return output.Render(os.Stdout, fmtType, info, headers, rows)
	},
}

func init() {
	RootCmd.AddCommand(versionCmd)
}
