package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/emredurukan/asactl/pkg/appleads"
	"github.com/emredurukan/asactl/pkg/config"
	"github.com/emredurukan/asactl/pkg/output"
	"github.com/spf13/cobra"
)

var statusValidate bool

var authStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show status of current credentials and profiles",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadConfig(GetConfigFile())
		if err != nil {
			return err
		}

		profile, err := config.GetActiveProfile(GetRequestedProfile(), GetConfigFile())
		if err != nil {
			fmt.Println("No active profile configured. Run 'asactl auth login' to get started.")
			return nil
		}

		fmtType := output.ResolveFormat(GetRequestedOutput())

		statusData := map[string]any{
			"active_profile": cfg.ActiveProfile,
			"profile":        profile,
			"profiles_count": len(cfg.Profiles),
		}

		if statusValidate {
			client := appleads.NewClient(profile)
			ctx := context.Background()
			token, terr := client.TokenManager.GetAccessToken(ctx, false)
			if terr != nil {
				statusData["validation"] = map[string]any{
					"success": false,
					"error":   terr.Error(),
				}
			} else {
				statusData["validation"] = map[string]any{
					"success":    true,
					"token_type": "Bearer",
					"token_len":  len(token),
				}
			}
		}

		if fmtType == output.FormatJSON {
			return output.RenderJSON(os.Stdout, statusData, true)
		}

		headers := []string{"FIELD", "VALUE"}
		mask := func(s string) string {
			if len(s) <= 8 {
				return "****"
			}
			return s[:4] + "..." + s[len(s)-4:]
		}

		rows := [][]string{
			{"Active Profile", profile.Name},
			{"Key ID", mask(profile.KeyID)},
			{"Client ID", profile.ClientID},
			{"Team ID", profile.TeamID},
			{"Org ID", profile.OrgID},
			{"Private Key Path", profile.PrivateKeyPath},
			{"Bypass Keychain", fmt.Sprintf("%t", profile.BypassKeychain)},
			{"Total Configured Profiles", fmt.Sprintf("%d", len(cfg.Profiles))},
		}

		if statusValidate {
			if v, ok := statusData["validation"].(map[string]any); ok {
				if v["success"] == true {
					rows = append(rows, []string{"Token Validation", "✓ Valid / Connected"})
				} else {
					rows = append(rows, []string{"Token Validation", fmt.Sprintf("✗ Failed: %v", v["error"])})
				}
			}
		}

		return output.Render(os.Stdout, fmtType, statusData, headers, rows)
	},
}

func init() {
	authStatusCmd.Flags().BoolVar(&statusValidate, "validate", false, "Validate token generation against Apple ID")
	authCmd.AddCommand(authStatusCmd)
}
