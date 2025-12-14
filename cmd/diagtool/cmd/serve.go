package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/mark/dsl-diagram-tool/pkg/server"
)

var serveCmd = &cobra.Command{
	Use:   "serve [file.d2]",
	Short: "Start the diagram editor web server",
	Long: `Start a local web server that provides a browser-based diagram editor.

The editor provides:
  - Split-pane interface with code editor and live preview
  - Real-time SVG rendering as you type
  - File save functionality (Ctrl+S)
  - External file change detection

Examples:
  # Start server with a D2 file
  diagtool serve diagram.d2

  # Start on a specific port
  diagtool serve diagram.d2 --port 3000

  # Start without a file (empty editor)
  diagtool serve`,
	Args: cobra.MaximumNArgs(1),
	RunE: runServe,
}

var (
	servePort int
)

func init() {
	serveCmd.Flags().IntVarP(&servePort, "port", "p", 8080, "port to listen on")
	rootCmd.AddCommand(serveCmd)
}

func runServe(cmd *cobra.Command, args []string) error {
	var filePath string
	if len(args) > 0 {
		filePath = args[0]
		// Check file exists
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			return fmt.Errorf("file not found: %s", filePath)
		}
	}

	srv, err := server.New(server.Options{
		Port:     servePort,
		FilePath: filePath,
	})
	if err != nil {
		return err
	}

	// Handle shutdown signals
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigCh
		fmt.Println("\nShutting down...")
		cancel()
	}()

	// Print startup message
	url := fmt.Sprintf("http://localhost:%d", servePort)
	fmt.Printf("Starting diagram editor server...\n")
	fmt.Printf("  URL: %s\n", url)
	if filePath != "" {
		fmt.Printf("  File: %s\n", filePath)
	}
	fmt.Printf("\nPress Ctrl+C to stop\n\n")

	return srv.Start(ctx)
}
