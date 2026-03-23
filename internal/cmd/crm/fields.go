// internal/cmd/crm/fields.go
package crm

import (
	"encoding/json"
	"fmt"

	"github.com/crewsbase/crewsbase-cli/internal/api"
	"github.com/crewsbase/crewsbase-cli/internal/output"
	"github.com/spf13/cobra"
)

var fieldsCmd = &cobra.Command{
	Use:   "fields",
	Short: "Manage CRM fields",
}

var fieldsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List fields in a CRM table",
	RunE: func(cmd *cobra.Command, args []string) error {
		tableID, _ := cmd.Flags().GetString("table")
		if tableID == "" {
			return fmt.Errorf("--table flag is required")
		}

		accountFlag, _ := cmd.Root().Flags().GetString("account")
		tokenFlag, _ := cmd.Root().Flags().GetString("token")
		outputFlag, _ := cmd.Root().Flags().GetString("output")
		quietFlag, _ := cmd.Root().Flags().GetBool("quiet")
		debugFlag, _ := cmd.Root().Flags().GetBool("debug")

		client, err := api.NewClient(accountFlag, tokenFlag, debugFlag)
		if err != nil {
			return err
		}

		body, _, err := client.Get(fmt.Sprintf("/crm/tables/%s/fields", tableID), nil)
		if err != nil {
			return err
		}

		var resp struct {
			Data []struct {
				ID   string `json:"id"`
				Name string `json:"name"`
				Type string `json:"type"`
			} `json:"data"`
		}
		if err := json.Unmarshal(body, &resp); err != nil {
			return fmt.Errorf("failed to parse response: %w", err)
		}

		if outputFlag == "json" {
			return output.PrintRaw(outputFlag, resp.Data)
		}

		headers := []string{"ID", "NAME", "TYPE"}
		var rows [][]string
		for _, f := range resp.Data {
			rows = append(rows, []string{f.ID, f.Name, f.Type})
		}
		return output.Print(outputFlag, headers, rows, quietFlag)
	},
}

func init() {
	fieldsListCmd.Flags().String("table", "", "CRM table ID (required)")
	fieldsCmd.AddCommand(fieldsListCmd)
}
