package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "nizam",
	Short: "Local structured service manager for dev environments",
	Long: `nizam is a powerful CLI tool to manage, monitor, and interact with 
local development services (Postgres, Redis, Meilisearch, etc.) using Docker.

It helps you spin up, shut down, and interact with common services without 
manually writing docker run or service-specific commands.`,
	Version: "0.1.0",
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	// Global flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is .nizam.yaml)")
	rootCmd.PersistentFlags().StringP("profile", "p", "dev", "configuration profile to use")
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "enable verbose logging")

	// Bind flags to viper
	viper.BindPFlag("profile", rootCmd.PersistentFlags().Lookup("profile"))
	viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Search for config in current directory
		viper.AddConfigPath(".")
		viper.SetConfigName(".nizam")
		viper.SetConfigType("yaml")
	}

	// Read in environment variables that match
	viper.SetEnvPrefix("NIZAM")
	viper.AutomaticEnv()

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		if viper.GetBool("verbose") {
			fmt.Fprintf(os.Stderr, "Using config file: %s\n", viper.ConfigFileUsed())
		}
	}
}
