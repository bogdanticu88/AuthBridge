package cmd

import (
	"github.com/spf13/cobra"
)

var (
	cfgFile string
)

var rootCmd = &cobra.Command{
	Use:   "authbridge",
	Short: "AuthBridge centralizes authentication for pentesting tools",
	Long: `AuthBridge is a lightweight daemon that manages JWT, OAuth2, 
Basic Auth, and more for tools like Burp Suite and Nuclei.`,
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.authbridge/config.yaml)")
}
