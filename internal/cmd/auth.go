// internal/cmd/auth.go
package cmd

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"time"

	"github.com/crewsbase/crewsbase-cli/internal/config"
	"github.com/spf13/cobra"
)

var tokenFlag string

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Manage authentication",
}

var authLoginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate with Crewsbase",
	RunE: func(cmd *cobra.Command, args []string) error {
		if tokenFlag != "" {
			if err := config.SetToken(tokenFlag); err != nil {
				return fmt.Errorf("failed to save token: %w", err)
			}
			fmt.Println("Token saved successfully.")
			return nil
		}

		// Browser flow
		listener, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			return fmt.Errorf("failed to start local server: %w", err)
		}
		port := listener.Addr().(*net.TCPAddr).Port

		tokenCh := make(chan string, 1)
		errCh := make(chan error, 1)

		mux := http.NewServeMux()
		mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
			token := r.URL.Query().Get("token")
			if token == "" {
				http.Error(w, "No token received", http.StatusBadRequest)
				errCh <- fmt.Errorf("no token received in callback")
				return
			}
			w.Header().Set("Content-Type", "text/html")
			fmt.Fprint(w, "<html><body><h2>Authentication successful!</h2><p>You can close this tab.</p></body></html>")
			tokenCh <- token
		})

		srv := &http.Server{Handler: mux}
		go func() {
			if err := srv.Serve(listener); err != http.ErrServerClosed {
				errCh <- err
			}
		}()

		url := fmt.Sprintf("%s/cli/auth?port=%d", config.GetAPIURL(), port)
		fmt.Printf("Opening browser to authenticate...\n")
		fmt.Printf("If the browser doesn't open, visit: %s\n", url)
		openBrowser(url)

		select {
		case token := <-tokenCh:
			if err := config.SetToken(token); err != nil {
				return fmt.Errorf("failed to save token: %w", err)
			}
			fmt.Println("Authentication successful!")
		case err := <-errCh:
			return err
		case <-time.After(5 * time.Minute):
			return fmt.Errorf("authentication timed out after 5 minutes")
		}

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		srv.Shutdown(ctx)
		return nil
	},
}

var authLogoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Remove stored credentials",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := config.RemoveToken(); err != nil {
			return fmt.Errorf("failed to remove token: %w", err)
		}
		fmt.Println("Logged out successfully.")
		if os.Getenv("CREWSBASE_TOKEN") != "" {
			fmt.Fprintln(os.Stderr, "Warning: CREWSBASE_TOKEN environment variable is still set.")
		}
		return nil
	},
}

var authStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show current authentication status",
	Run: func(cmd *cobra.Command, args []string) {
		token := config.GetToken("")
		if token == "" {
			fmt.Println("Not authenticated.")
			return
		}
		fmt.Printf("Authenticated (token: %s...)\n", token[:6])
		acc := config.GetAccount("")
		if acc != "" {
			fmt.Printf("Default account: %s\n", acc)
		}
	},
}

func openBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	}
	if cmd != nil {
		cmd.Start()
	}
}

func init() {
	authLoginCmd.Flags().StringVar(&tokenFlag, "token", "", "API token (for non-interactive auth)")
	authCmd.AddCommand(authLoginCmd)
	authCmd.AddCommand(authLogoutCmd)
	authCmd.AddCommand(authStatusCmd)
	rootCmd.AddCommand(authCmd)
}
