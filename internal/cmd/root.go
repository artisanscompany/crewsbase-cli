// internal/cmd/root.go
package cmd

import (
	"fmt"
	"os"

	"github.com/crewsbase/crewsbase-cli/internal/api"
	"github.com/crewsbase/crewsbase-cli/internal/cmd/crm"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile   string
	account   string
	outputFmt string
	quiet     bool
	noColor   bool
	debug     bool
)

var rootCmd = &cobra.Command{
	Use:   "crewsbase",
	Short: "Crewsbase CLI — manage your workspace from the terminal",
	Long:  "A command-line interface for Crewsbase. Manage CRM tables, rows, and more.",
}

func Execute() error {
	api.Version = Version // Propagate ldflags version to API client
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default $HOME/.crewsbase/config.yaml)")
	rootCmd.PersistentFlags().StringVar(&account, "account", "", "account slug (overrides default)")
	rootCmd.PersistentFlags().StringVar(&outputFmt, "output", "table", "output format: table, json, csv")
	rootCmd.PersistentFlags().BoolVar(&quiet, "quiet", false, "suppress non-essential output")
	rootCmd.PersistentFlags().BoolVar(&noColor, "no-color", false, "disable color output")
	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "print HTTP request/response details")
	rootCmd.PersistentFlags().String("token", "", "API token (overrides config and env var)")

	rootCmd.AddCommand(crm.CrmCmd)
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		viper.AddConfigPath(home + "/.crewsbase")
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
	}

	viper.SetEnvPrefix("CREWSBASE")
	viper.AutomaticEnv()

	_ = viper.ReadInConfig()
}
