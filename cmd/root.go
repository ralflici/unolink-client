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
	ulAddress string

	rootCmd = &cobra.Command{
		Use:   "unolink-client",
		Short: "Client for unolink for packet statistics",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Starting the client...")
			run(ulAddress)
		},
	}
)

func init() {
	rootCmd.PersistentFlags().StringVarP(&ulAddress, "address", "a", "127.0.0.1:2281", "unolink address [ip:port]")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(address string) {
	go connection.Handle(address)
	display.Show()
}
