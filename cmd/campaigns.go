package cmd

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/emredurukn/asactl/pkg/appleads"
	"github.com/emredurukn/asactl/pkg/config"
	"github.com/emredurukn/asactl/pkg/output"
	"github.com/spf13/cobra"
)

var (
	campaignsLimit  int
	campaignsOffset int
)

var campaignsCmd = &cobra.Command{
	Use:     "campaigns",
	Aliases: []string{"campaign"},
	Short:   "Manage Apple Search Ads campaigns",
}

var campaignsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List advertising campaigns",
	Example: `  # List campaigns with default table view
  asactl campaigns list

  # List campaigns as JSON
  asactl campaigns list --output json

  # Paginate campaigns
  asactl campaigns list --limit 50 --offset 10`,
	RunE: func(cmd *cobra.Command, args []string) error {
		profile, err := config.GetActiveProfile(GetRequestedProfile(), GetConfigFile())
		if err != nil {
			return err
		}

		if profile.OrgID == "" {
			return fmt.Errorf("profile '%s' has no org_id set. Please update your profile or specify --org-id", profile.Name)
		}

		client := appleads.NewClient(profile)
		ctx := context.Background()

		campaigns, _, err := client.ListCampaigns(ctx, campaignsLimit, campaignsOffset)
		if err != nil {
			return err
		}

		fmtType := output.ResolveFormat(GetRequestedOutput())

		if fmtType == output.FormatJSON {
			return output.RenderJSON(os.Stdout, campaigns, true)
		}

		headers := []string{"ID", "NAME", "STATUS", "SERVING", "DAILY BUDGET", "CURRENCY", "REGIONS"}
		rows := make([][]string, 0, len(campaigns))

		for _, c := range campaigns {
			dailyBudget := "-"
			currency := c.BillingCurrency
			if c.DailyBudgetAmount != nil {
				dailyBudget = c.DailyBudgetAmount.Amount
				if c.DailyBudgetAmount.Currency != "" {
					currency = c.DailyBudgetAmount.Currency
				}
			}

			regions := strings.Join(c.CountriesOrRegions, ",")
			if len(regions) > 20 {
				regions = fmt.Sprintf("%d regions", len(c.CountriesOrRegions))
			}

			rows = append(rows, []string{
				strconv.FormatInt(c.ID, 10),
				c.Name,
				c.Status,
				c.ServingStatus,
				dailyBudget,
				currency,
				regions,
			})
		}

		return output.Render(os.Stdout, fmtType, campaigns, headers, rows)
	},
}

var campaignsGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get details for a specific campaign",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid campaign id '%s': must be numeric", args[0])
		}

		profile, err := config.GetActiveProfile(GetRequestedProfile(), GetConfigFile())
		if err != nil {
			return err
		}

		if profile.OrgID == "" {
			return fmt.Errorf("profile '%s' has no org_id set. Please update your profile or specify --org-id", profile.Name)
		}

		client := appleads.NewClient(profile)
		ctx := context.Background()

		campaign, err := client.GetCampaign(ctx, id)
		if err != nil {
			return err
		}

		fmtType := output.ResolveFormat(GetRequestedOutput())

		if fmtType == output.FormatJSON {
			return output.RenderJSON(os.Stdout, campaign, true)
		}

		headers := []string{"FIELD", "VALUE"}
		rows := [][]string{
			{"ID", strconv.FormatInt(campaign.ID, 10)},
			{"Org ID", strconv.FormatInt(campaign.OrgID, 10)},
			{"Name", campaign.Name},
			{"Status", campaign.Status},
			{"Serving Status", campaign.ServingStatus},
			{"Adam ID (App ID)", strconv.FormatInt(campaign.AdamID, 10)},
			{"Countries / Regions", strings.Join(campaign.CountriesOrRegions, ", ")},
		}

		if campaign.BudgetAmount != nil {
			rows = append(rows, []string{"Total Budget", fmt.Sprintf("%s %s", campaign.BudgetAmount.Amount, campaign.BudgetAmount.Currency)})
		}
		if campaign.DailyBudgetAmount != nil {
			rows = append(rows, []string{"Daily Budget", fmt.Sprintf("%s %s", campaign.DailyBudgetAmount.Amount, campaign.DailyBudgetAmount.Currency)})
		}
		if campaign.ModificationTime != nil {
			rows = append(rows, []string{"Last Modified", campaign.ModificationTime.Format("2006-01-02 15:04:05 MST")})
		}

		return output.Render(os.Stdout, fmtType, campaign, headers, rows)
	},
}

func init() {
	campaignsListCmd.Flags().IntVar(&campaignsLimit, "limit", 20, "Number of campaigns to fetch (max 1000)")
	campaignsListCmd.Flags().IntVar(&campaignsOffset, "offset", 0, "Pagination offset")

	campaignsCmd.AddCommand(campaignsListCmd)
	campaignsCmd.AddCommand(campaignsGetCmd)
	RootCmd.AddCommand(campaignsCmd)
}
