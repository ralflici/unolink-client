package cmd

import (
	"fmt"
	"os"

	"unolink-client/connection"
	"unolink-client/display"

	"github.com/spf13/cobra"
)

var (
	// Used for flags.
	ulAddress  string
	streamPort uint16
	restPort   uint16

	rootCmd = &cobra.Command{
		Use:   "unolink-client",
		Short: "Client for unolink for packet statistics",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Starting the client...")
			run(ulAddress, streamPort, restPort)
		},
	}
)

func init() {
	rootCmd.PersistentFlags().StringVarP(&ulAddress, "address", "a", "127.0.0.1", "unolink ip address")
    rootCmd.PersistentFlags().Uint16VarP(&restPort, "rest-port", "r", 2280, "port of REST API")
	rootCmd.PersistentFlags().Uint16VarP(&streamPort, "stream-port", "s", 2281, "port of stream TCP connection")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(address string, streamPort, restPort uint16) {
	go connection.Handle(address, streamPort, restPort)
	display.RenderTable()
}
