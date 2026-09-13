package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/emredurukan/asactl/pkg/appleads"
	"github.com/emredurukan/asactl/pkg/config"
	"github.com/spf13/cobra"
)

var (
	loginName           string
	loginKeyID          string
	loginTeamID         string
	loginClientID       string
	loginOrgID          string
	loginPrivateKeyPath string
	loginBypassKeychain bool
	loginValidate       bool
)

var authLoginCmd = &cobra.Command{
	Use:   "login",
	Short: "Configure credentials for Apple Ads API",
	Long: `Configure and store Apple Search Ads API credentials for a given profile.
Credentials can be saved in macOS Keychain or in the local config file (~/.asactl/config.yaml).`,
	Example: `  # Log in with full credentials
  asactl auth login \
    --name default \
    --key-id "SEARCHADS.1234..." \
    --team-id "SEARCHADS.1234..." \
    --client-id "SEARCHADS.1234..." \
    --org-id "1234567" \
    --private-key ~/.keys/AuthKey.p8 \
    --validate`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if loginKeyID == "" {
			return fmt.Errorf("--key-id is required")
		}
		if loginTeamID == "" {
			return fmt.Errorf("--team-id is required")
		}
		if loginClientID == "" {
			return fmt.Errorf("--client-id is required")
		}
		if loginPrivateKeyPath == "" {
			return fmt.Errorf("--private-key is required")
		}

		// Verify private key file exists
		if _, err := os.Stat(loginPrivateKeyPath); err != nil {
			return fmt.Errorf("cannot access private key file at %s: %w", loginPrivateKeyPath, err)
		}

		profile := config.Profile{
			Name:           loginName,
			KeyID:          loginKeyID,
			TeamID:         loginTeamID,
			ClientID:       loginClientID,
			OrgID:          loginOrgID,
			PrivateKeyPath: loginPrivateKeyPath,
			BypassKeychain: loginBypassKeychain,
		}

		if !loginBypassKeychain {
			keyBytes, err := os.ReadFile(loginPrivateKeyPath)
			if err == nil {
				if kerr := config.StorePrivateKeyInKeyring(loginName, string(keyBytes)); kerr != nil {
					fmt.Printf("Warning: Could not save key to Keychain (%v). Falling back to file path.\n", kerr)
					profile.BypassKeychain = true
				}
			}
		}

		// Optionally validate credentials right now
		if loginValidate {
			fmt.Println("Validating credentials with Apple...")
			client := appleads.NewClient(&profile)
			ctx := context.Background()

			token, err := client.TokenManager.GetAccessToken(ctx, true)
			if err != nil {
				return fmt.Errorf("validation failed: %w", err)
			}
			fmt.Println("✓ Successfully authenticated with Apple ID OAuth endpoint")

			acls, err := client.GetACLs(ctx)
			if err != nil {
				fmt.Printf("Warning: Token generated but failed to fetch ACLs: %v\n", err)
			} else {
				fmt.Printf("✓ Retrieved %d organization(s) accessible with these credentials\n", len(acls))
				if profile.OrgID == "" && len(acls) > 0 {
					profile.OrgID = fmt.Sprintf("%d", acls[0].OrgID)
					fmt.Printf("✓ Automatically selected default orgId: %s (%s)\n", profile.OrgID, acls[0].OrgName)
				}
			}
			_ = token
		}

		if err := config.SaveProfile(profile, true, GetConfigFile()); err != nil {
			return fmt.Errorf("failed to save profile: %w", err)
		}

		fmt.Printf("✓ Successfully saved profile '%s'\n", profile.Name)
		return nil
	},
}

func init() {
	authLoginCmd.Flags().StringVar(&loginName, "name", "default", "profile name")
	authLoginCmd.Flags().StringVar(&loginKeyID, "key-id", "", "Apple Ads Key ID")
	authLoginCmd.Flags().StringVar(&loginTeamID, "team-id", "", "Apple Ads Team ID")
	authLoginCmd.Flags().StringVar(&loginClientID, "client-id", "", "Apple Ads Client ID")
	authLoginCmd.Flags().StringVar(&loginOrgID, "org-id", "", "Apple Ads Organization ID")
	authLoginCmd.Flags().StringVar(&loginPrivateKeyPath, "private-key", "", "Path to private key (.p8 or .key)")
	authLoginCmd.Flags().BoolVar(&loginBypassKeychain, "bypass-keychain", false, "Do not store private key in OS Keychain")
	authLoginCmd.Flags().BoolVar(&loginValidate, "validate", false, "Validate credentials with Apple immediately")

	authCmd.AddCommand(authLoginCmd)
}
