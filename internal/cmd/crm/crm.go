// internal/cmd/crm/crm.go
package crm

import (
	"github.com/spf13/cobra"
)

var CrmCmd = &cobra.Command{
	Use:   "crm",
	Short: "Manage CRM tables and data",
}

func init() {
	CrmCmd.AddCommand(tablesCmd)
	CrmCmd.AddCommand(rowsCmd)
	CrmCmd.AddCommand(fieldsCmd)
}
