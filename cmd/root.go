package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile        string
	profileFlag    string
	outputFlag     string
	verboseFlag    bool
)

// RootCmd is the base command when called without any subcommands.
var RootCmd = &cobra.Command{
	Use:     "asactl",
	Aliases: []string{"apple-ads"},
	Short:   "asactl - Fast, scriptable CLI for the Apple Search Ads API",
	Long: `asactl is a fast and lightweight command-line interface for the Apple Search Ads API.
Automate campaigns, ad groups, keywords, reports, budgets, and authentication workflows.

Complete documentation: https://github.com/emredurukan/asactl`,
}

func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file path (default is $HOME/.asactl/config.yaml)")
	RootCmd.PersistentFlags().StringVarP(&profileFlag, "profile", "p", "", "profile name to use")
	RootCmd.PersistentFlags().StringVarP(&outputFlag, "output", "o", "", "output format (table, json, csv)")
	RootCmd.PersistentFlags().BoolVarP(&verboseFlag, "verbose", "v", false, "enable verbose output")

	_ = viper.BindPFlag("profile", RootCmd.PersistentFlags().Lookup("profile"))
	_ = viper.BindPFlag("output", RootCmd.PersistentFlags().Lookup("output"))
	_ = viper.BindPFlag("verbose", RootCmd.PersistentFlags().Lookup("verbose"))
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		if err == nil {
			viper.AddConfigPath(home + "/.asactl")
			viper.AddConfigPath(home + "/.apple-ads")
			viper.SetConfigName("config")
			viper.SetConfigType("yaml")
		}
	}

	viper.SetEnvPrefix("ASACTL")
	viper.AutomaticEnv()
	_ = viper.ReadInConfig()
}

// GetRequestedProfile returns the active profile name from flags or config.
func GetRequestedProfile() string {
	return profileFlag
}

// GetRequestedOutput returns the format flag or default.
func GetRequestedOutput() string {
	return outputFlag
}

// GetConfigFile returns custom config file if specified.
func GetConfigFile() string {
	return cfgFile
}
