/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/spf13/cobra"
)

var (
	clientCmdGrpcHost string
	clientCmdGrpcPort string
)

// clientCmd represents the client command
var clientCmd = &cobra.Command{
	Use:   "client",
	Short: "gRPC client command for the current gRPC server",
	Long: `This cli will be used to get connected to the gRPC server. 
	Ivoking the gRPC services returning the response in json format`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func init() {
	rootCmd.AddCommand(clientCmd)
	clientCmd.PersistentFlags().StringVar(&clientCmdGrpcHost, "grpc-host", "localhost", "grpc server address")
	clientCmd.PersistentFlags().StringVar(&clientCmdGrpcPort, "grpc-port", "9090", "grpc server port")
}
