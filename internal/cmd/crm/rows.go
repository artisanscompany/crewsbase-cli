// internal/cmd/crm/rows.go
package crm

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/artisanscompany/crewsbase-cli/internal/api"
	"github.com/artisanscompany/crewsbase-cli/internal/output"
	"github.com/spf13/cobra"
)

// Note: flags are read per-command via cmd.Flags().GetString() etc.
// Do NOT use shared package-level variables for flags bound to multiple commands.

var rowsCmd = &cobra.Command{
	Use:   "rows",
	Short: "Manage CRM rows",
}

var rowsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List rows in a CRM table",
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
		page, _ := cmd.Flags().GetInt("page")
		perPage, _ := cmd.Flags().GetInt("per-page")

		client, err := api.NewClient(accountFlag, tokenFlag, debugFlag)
		if err != nil {
			return err
		}

		params := map[string]string{
			"page":     fmt.Sprintf("%d", page),
			"per_page": fmt.Sprintf("%d", perPage),
		}

		body, headers, err := client.Get(fmt.Sprintf("/crm/tables/%s/rows", tableID), params)
		if err != nil {
			return err
		}

		var resp struct {
			Data []struct {
				ID        string                 `json:"id"`
				DisplayID int                    `json:"display_id"`
				Values    map[string]interface{} `json:"values"`
				CreatedAt string                 `json:"created_at"`
			} `json:"data"`
			Meta struct {
				Page       int `json:"page"`
				PerPage    int `json:"per_page"`
				Total      int `json:"total"`
				TotalPages int `json:"total_pages"`
			} `json:"meta"`
		}
		if err := json.Unmarshal(body, &resp); err != nil {
			return fmt.Errorf("failed to parse response: %w", err)
		}

		if outputFlag == "json" {
			return output.PrintRaw(outputFlag, resp)
		}

		// Build dynamic headers from values
		columnSet := make(map[string]bool)
		var columns []string
		for _, row := range resp.Data {
			for k := range row.Values {
				if !columnSet[k] {
					columnSet[k] = true
					columns = append(columns, k)
				}
			}
		}

		allHeaders := append([]string{"ID"}, columns...)
		var rows [][]string
		for _, row := range resp.Data {
			r := []string{row.ID}
			for _, col := range columns {
				val := row.Values[col]
				r = append(r, fmt.Sprintf("%v", val))
			}
			rows = append(rows, r)
		}

		if err := output.Print(outputFlag, allHeaders, rows, quietFlag); err != nil {
			return err
		}

		_ = headers // pagination info available in headers too
		if !quietFlag && outputFlag == "table" {
			fmt.Printf("\nPage %d of %d (%d total)\n", resp.Meta.Page, resp.Meta.TotalPages, resp.Meta.Total)
		}
		return nil
	},
}

var rowsShowCmd = &cobra.Command{
	Use:   "show <id>",
	Short: "Show a CRM row",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		tableID, _ := cmd.Flags().GetString("table")
		if tableID == "" {
			return fmt.Errorf("--table flag is required")
		}

		accountFlag, _ := cmd.Root().Flags().GetString("account")
		tokenFlag, _ := cmd.Root().Flags().GetString("token")
		outputFlag, _ := cmd.Root().Flags().GetString("output")
		debugFlag, _ := cmd.Root().Flags().GetBool("debug")

		client, err := api.NewClient(accountFlag, tokenFlag, debugFlag)
		if err != nil {
			return err
		}

		body, _, err := client.Get(fmt.Sprintf("/crm/tables/%s/rows/%s", tableID, args[0]), nil)
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

var rowsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a CRM row",
	RunE: func(cmd *cobra.Command, args []string) error {
		tableID, _ := cmd.Flags().GetString("table")
		if tableID == "" {
			return fmt.Errorf("--table flag is required")
		}
		fieldVals, _ := cmd.Flags().GetStringArray("field")
		if len(fieldVals) == 0 {
			return fmt.Errorf("at least one --field is required")
		}

		accountFlag, _ := cmd.Root().Flags().GetString("account")
		tokenFlag, _ := cmd.Root().Flags().GetString("token")
		outputFlag, _ := cmd.Root().Flags().GetString("output")
		debugFlag, _ := cmd.Root().Flags().GetBool("debug")

		client, err := api.NewClient(accountFlag, tokenFlag, debugFlag)
		if err != nil {
			return err
		}

		fields := parseFields(fieldVals)
		body, _, err := client.Post(fmt.Sprintf("/crm/tables/%s/rows", tableID), map[string]interface{}{
			"fields": fields,
		})
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

var rowsUpdateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update a CRM row",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		tableID, _ := cmd.Flags().GetString("table")
		if tableID == "" {
			return fmt.Errorf("--table flag is required")
		}
		fieldVals, _ := cmd.Flags().GetStringArray("field")
		if len(fieldVals) == 0 {
			return fmt.Errorf("at least one --field is required")
		}

		accountFlag, _ := cmd.Root().Flags().GetString("account")
		tokenFlag, _ := cmd.Root().Flags().GetString("token")
		outputFlag, _ := cmd.Root().Flags().GetString("output")
		debugFlag, _ := cmd.Root().Flags().GetBool("debug")

		client, err := api.NewClient(accountFlag, tokenFlag, debugFlag)
		if err != nil {
			return err
		}

		fields := parseFields(fieldVals)
		body, _, err := client.Patch(fmt.Sprintf("/crm/tables/%s/rows/%s", tableID, args[0]), map[string]interface{}{
			"fields": fields,
		})
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

var rowsDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a CRM row",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		tableID, _ := cmd.Flags().GetString("table")
		if tableID == "" {
			return fmt.Errorf("--table flag is required")
		}

		accountFlag, _ := cmd.Root().Flags().GetString("account")
		tokenFlag, _ := cmd.Root().Flags().GetString("token")
		debugFlag, _ := cmd.Root().Flags().GetBool("debug")
		quietFlag, _ := cmd.Root().Flags().GetBool("quiet")

		client, err := api.NewClient(accountFlag, tokenFlag, debugFlag)
		if err != nil {
			return err
		}

		_, _, err = client.Delete(fmt.Sprintf("/crm/tables/%s/rows/%s", tableID, args[0]))
		if err != nil {
			return err
		}

		if !quietFlag {
			fmt.Println("Row deleted.")
		}
		return nil
	},
}

func parseFields(flags []string) map[string]string {
	fields := make(map[string]string)
	for _, f := range flags {
		parts := strings.SplitN(f, "=", 2)
		if len(parts) == 2 {
			fields[parts[0]] = parts[1]
		}
	}
	return fields
}

func init() {
	// Add --table flag to all row commands
	for _, c := range []*cobra.Command{rowsListCmd, rowsShowCmd, rowsCreateCmd, rowsUpdateCmd, rowsDeleteCmd} {
		c.Flags().String("table", "", "CRM table ID (required)")
	}

	rowsListCmd.Flags().Int("page", 1, "page number")
	rowsListCmd.Flags().Int("per-page", 25, "items per page (max 100)")

	rowsCreateCmd.Flags().StringArray("field", nil, "field value as key=value")
	rowsUpdateCmd.Flags().StringArray("field", nil, "field value as key=value")

	rowsCmd.AddCommand(rowsListCmd)
	rowsCmd.AddCommand(rowsShowCmd)
	rowsCmd.AddCommand(rowsCreateCmd)
	rowsCmd.AddCommand(rowsUpdateCmd)
	rowsCmd.AddCommand(rowsDeleteCmd)
}
