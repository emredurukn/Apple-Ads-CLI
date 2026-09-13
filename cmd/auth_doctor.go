package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/emredurukan/asactl/pkg/appleads"
	"github.com/emredurukan/asactl/pkg/config"
	"github.com/spf13/cobra"
)

var authDoctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Run diagnostics on Apple Ads credentials and connectivity",
	Long: `Diagnose authentication configuration, private key validity,
Apple ID token exchange, and Apple Ads API access permissions.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("asactl Authentication Doctor")
		fmt.Println("==============================")

		// 1. Profile check
		profile, err := config.GetActiveProfile(GetRequestedProfile(), GetConfigFile())
		if err != nil {
			fmt.Println("✗ Active profile: NOT FOUND")
			fmt.Printf("  Hint: Run 'asactl auth login' to configure a profile.\n")
			return nil
		}
		fmt.Printf("✓ Profile '%s': Loaded\n", profile.Name)

		// 2. Field completeness
		missing := []string{}
		if profile.KeyID == "" {
			missing = append(missing, "key-id")
		}
		if profile.ClientID == "" {
			missing = append(missing, "client-id")
		}
		if profile.TeamID == "" {
			missing = append(missing, "team-id")
		}
		if len(missing) > 0 {
			fmt.Printf("✗ Profile fields: Missing required fields: %s\n", strings.Join(missing, ", "))
			return nil
		}
		fmt.Println("✓ Profile fields: Complete")

		// 3. Private Key File
		var keyBytes []byte
		if !profile.BypassKeychain {
			savedKey, _ := config.GetPrivateKeyFromKeyring(profile.Name)
			if savedKey != "" {
				keyBytes = []byte(savedKey)
				fmt.Println("✓ Private key: Retrieved from OS Keychain")
			}
		}

		if len(keyBytes) == 0 {
			path := profile.PrivateKeyPath
			if strings.HasPrefix(path, "~/") {
				home, _ := os.UserHomeDir()
				path = filepath.Join(home, path[2:])
			}
			data, rerr := os.ReadFile(path)
			if rerr != nil {
				fmt.Printf("✗ Private key file: Cannot read '%s': %v\n", path, rerr)
				return nil
			}
			keyBytes = data
			fmt.Printf("✓ Private key file: Found and accessible at '%s'\n", path)
		}

		// 4. Private Key Parsing
		ecdsaKey, err := appleads.ParsePrivateKey(keyBytes)
		if err != nil {
			fmt.Printf("✗ Private key parsing: %v\n", err)
			return nil
		}
		_ = ecdsaKey
		fmt.Println("✓ Private key parsing: Valid ECDSA private key")

		// 5. JWT Generation
		secretJWT, err := appleads.GenerateClientSecret(profile)
		if err != nil {
			fmt.Printf("✗ JWT generation: %v\n", err)
			return nil
		}
		_ = secretJWT
		fmt.Println("✓ Client Secret JWT: Successfully created (ES256)")

		// 6. Token Exchange
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()

		client := appleads.NewClient(profile)
		fmt.Println("• Contacting Apple ID OAuth endpoint...")
		token, err := client.TokenManager.GetAccessToken(ctx, true)
		if err != nil {
			fmt.Printf("✗ Apple ID OAuth: Failed to exchange token: %v\n", err)
			fmt.Println("  Hint: Verify your Key ID, Client ID, and Team ID in Apple Search Ads console.")
			return nil
		}
		_ = token
		fmt.Println("✓ Apple ID OAuth: Successfully obtained access token")

		// 7. ACL / Permission Check
		fmt.Println("• Verifying Apple Search Ads API permissions...")
		acls, err := client.GetACLs(ctx)
		if err != nil {
			fmt.Printf("✗ Apple Ads API: Failed to fetch ACLs: %v\n", err)
			return nil
		}

		fmt.Printf("✓ Apple Ads API: Organization access verified (%d orgs accessible)\n", len(acls))
		for _, acl := range acls {
			activeMark := " "
			if fmt.Sprintf("%d", acl.OrgID) == profile.OrgID {
				activeMark = "*"
			}
			fmt.Printf("  [%s] Org ID: %d | Name: %s | Roles: %s\n",
				activeMark, acl.OrgID, acl.OrgName, strings.Join(acl.RoleNames, ", "))
		}

		fmt.Println("\nAll checks passed! asactl is ready for use.")
		return nil
	},
}

func init() {
	authCmd.AddCommand(authDoctorCmd)
}
