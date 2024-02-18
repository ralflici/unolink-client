package cmd

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

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
			// fmt.Println("Starting the client...")
			run(ulAddress, streamPort, restPort)
		},
	}
)

func init() {
	rootCmd.PersistentFlags().StringVarP(&ulAddress, "host", "H", "127.0.0.1", "unolink ip address")
	rootCmd.PersistentFlags().Uint16VarP(&restPort, "rest-port", "r", 2280, "port of REST API")
	rootCmd.PersistentFlags().Uint16VarP(&streamPort, "stream-port", "s", 2281, "port of stream TCP connection")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}

func run(address string, streamPort, restPort uint16) {
	ctx, cancel := context.WithCancel(context.Background())
	quitCh := make(chan struct{})
	defer cancel()

	var wg sync.WaitGroup
	errCh := make(chan error)

	wg.Add(2)
	go connection.Handle(ctx, &wg, errCh, quitCh, address, streamPort, restPort)
	go display.RenderTable(ctx, &wg, errCh, quitCh)

	go func() {
		for {
			select {
			case <-errCh:
				return
			case <-quitCh:
				fmt.Println("Terminating...")
				return
			case <-ctx.Done():
				fmt.Println("Terminating...")
				return
			default:
				time.Sleep(100 * time.Microsecond)
			}
		}
	}()

	// Wait for an error or context cancellation
	// select {
	// case <-errCh:
	// 	// cancel() // Cancel the context to signal termination
	// 	ctx.Done()
	// case <-quitCh:
	// 	cancel()
	// }
	wg.Wait()
}
