/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"

	"github.com/spf13/cobra"
)

var (
	accountBalanceCmd_AccountUUID string
)

// accountBalanceCmd represents the accountBalance command
var accountBalanceCmd = &cobra.Command{
	Use:   "accountBalance",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		cli_service, err := client(ctx)
		if err != nil {
			return
		}
		cli_service.ShowCurrentBalance(ctx, accountBalanceCmd_AccountUUID)
	},
}

func init() {
	clientCmd.AddCommand(accountBalanceCmd)
	accountBalanceCmd.Flags().StringVar(&accountBalanceCmd_AccountUUID, "account_uuid", "", "uuid of the account to get the balance")
}
