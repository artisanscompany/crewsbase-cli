// internal/cmd/crm/tables.go
package crm

import (
	"encoding/json"
	"fmt"

	"github.com/artisanscompany/crewsbase-cli/internal/api"
	"github.com/artisanscompany/crewsbase-cli/internal/output"
	"github.com/spf13/cobra"
)

var tablesCmd = &cobra.Command{
	Use:   "tables",
	Short: "Manage CRM tables",
}

var tablesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all CRM tables",
	RunE: func(cmd *cobra.Command, args []string) error {
		accountFlag, _ := cmd.Root().Flags().GetString("account")
		tokenFlag, _ := cmd.Root().Flags().GetString("token")
		outputFlag, _ := cmd.Root().Flags().GetString("output")
		quietFlag, _ := cmd.Root().Flags().GetBool("quiet")
		debugFlag, _ := cmd.Root().Flags().GetBool("debug")

		client, err := api.NewClient(accountFlag, tokenFlag, debugFlag)
		if err != nil {
			return err
		}

		body, _, err := client.Get("/crm/tables", nil)
		if err != nil {
			return err
		}

		var resp struct {
			Data []struct {
				ID        string `json:"id"`
				Name      string `json:"name"`
				RowCount  int    `json:"row_count"`
				CreatedAt string `json:"created_at"`
			} `json:"data"`
		}
		if err := json.Unmarshal(body, &resp); err != nil {
			return fmt.Errorf("failed to parse response: %w", err)
		}

		if outputFlag == "json" {
			return output.PrintRaw(outputFlag, resp.Data)
		}

		headers := []string{"ID", "NAME", "ROWS", "CREATED"}
		var rows [][]string
		for _, t := range resp.Data {
			rows = append(rows, []string{t.ID, t.Name, fmt.Sprintf("%d", t.RowCount), t.CreatedAt[:10]})
		}
		return output.Print(outputFlag, headers, rows, quietFlag)
	},
}

var tablesShowCmd = &cobra.Command{
	Use:   "show <id>",
	Short: "Show a CRM table",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		accountFlag, _ := cmd.Root().Flags().GetString("account")
		tokenFlag, _ := cmd.Root().Flags().GetString("token")
		outputFlag, _ := cmd.Root().Flags().GetString("output")
		debugFlag, _ := cmd.Root().Flags().GetBool("debug")

		client, err := api.NewClient(accountFlag, tokenFlag, debugFlag)
		if err != nil {
			return err
		}

		body, _, err := client.Get("/crm/tables/"+args[0], nil)
		if err != nil {
			return err
		}

		var resp struct {
			Data json.RawMessage `json:"data"`
		}
		if err := json.Unmarshal(body, &resp); err != nil {
			return fmt.Errorf("failed to parse response: %w", err)
		}

		return output.PrintRaw(outputFlag, json.RawMessage(resp.Data))
	},
}

func init() {
	tablesCmd.AddCommand(tablesListCmd)
	tablesCmd.AddCommand(tablesShowCmd)
}
