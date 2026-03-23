// internal/cmd/shell.go
package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"

	"github.com/artisanscompany/crewsbase-cli/internal/api"
	"github.com/artisanscompany/crewsbase-cli/internal/config"
	"github.com/artisanscompany/crewsbase-cli/internal/shell"
	"github.com/artisanscompany/crewsbase-cli/internal/types"
	"github.com/spf13/cobra"
)

var shellCmd = &cobra.Command{
	Use:   "shell",
	Short: "Interactive shell session",
	Long:  "Start an interactive Crewsbase shell with tab completion and tool discovery.",
	RunE: func(cmd *cobra.Command, args []string) error {
		accountFlag, _ := cmd.Root().Flags().GetString("account")
		tokenFlag, _ := cmd.Root().Flags().GetString("token")
		debugFlag, _ := cmd.Root().Flags().GetBool("debug")

		client, err := api.NewClient(accountFlag, tokenFlag, debugFlag)
		if err != nil {
			return err
		}

		// Fetch available tools
		fmt.Fprintf(os.Stderr, "Loading tools...")
		tools, err := fetchTools(client)
		if err != nil {
			fmt.Fprintf(os.Stderr, " failed: %v\n", err)
			return err
		}
		fmt.Fprintf(os.Stderr, " %d tools across %d domains\n", countTools(tools), len(tools))

		accountName := config.GetAccount(accountFlag)
		s := shell.New(client, tools, accountName, debugFlag)
		return s.Run()
	},
}

func fetchTools(client *api.Client) ([]types.ToolDomain, error) {
	body, _, err := client.Get("/tools", nil)
	if err != nil {
		return nil, err
	}

	var resp struct {
		Data []types.ToolDomain `json:"data"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse tools: %w", err)
	}

	sort.Slice(resp.Data, func(i, j int) bool {
		return resp.Data[i].Domain < resp.Data[j].Domain
	})

	return resp.Data, nil
}

func countTools(domains []types.ToolDomain) int {
	n := 0
	for _, d := range domains {
		n += len(d.Tools)
	}
	return n
}

func init() {
	rootCmd.AddCommand(shellCmd)

	rootCmd.Long = `A command-line interface for Crewsbase. Manage your workspace from the terminal.

Run "crewsbase shell" for an interactive session with tab completion.`
}
