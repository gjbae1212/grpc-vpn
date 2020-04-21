package cmd

import (
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use:   "vpn-server",
		Short: color.GreenString(`vpn-server is the VPN server communicating through GRPC.`),
		Long:  color.GreenString(`vpn-server is the VPN server communicating through GRPC.`),
	}
)

// Execute runs main command.
func Execute() {
	rootCmd.Execute()
}

func initConfig() {}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.SetHelpCommand(&cobra.Command{
		Use:    "no-help",
		Hidden: true,
	})
}
