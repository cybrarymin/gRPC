/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"

	"github.com/spf13/cobra"
)

// showExchangeRateCmd represents the showExchangeRate command
var showExchangeRateCmd = &cobra.Command{
	Use:   "showExchangeRate",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		cli_service, err := client()
		if err != nil {
			return
		}
		cli_service.ShowExchangeRate(ctx, "USD", "CAD", 10)
	},
}

func init() {
	clientCmd.AddCommand(showExchangeRateCmd)

}
