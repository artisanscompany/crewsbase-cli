// internal/output/formatter.go
package output

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
)

func Print(format string, headers []string, rows [][]string, quiet bool) error {
	switch format {
	case "json":
		return printJSON(headers, rows)
	case "csv":
		return printCSV(headers, rows)
	default:
		if quiet {
			return printQuiet(rows)
		}
		return printTable(headers, rows)
	}
}

func PrintRaw(format string, data interface{}) error {
	if format == "json" {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(data)
	}
	return nil
}

func printTable(headers []string, rows [][]string) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, strings.Join(headers, "\t"))
	for _, row := range rows {
		fmt.Fprintln(w, strings.Join(row, "\t"))
	}
	return w.Flush()
}

func printJSON(headers []string, rows [][]string) error {
	var result []map[string]string
	for _, row := range rows {
		item := make(map[string]string)
		for i, h := range headers {
			if i < len(row) {
				item[h] = row[i]
			}
		}
		result = append(result, item)
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(result)
}

func printCSV(headers []string, rows [][]string) error {
	w := csv.NewWriter(os.Stdout)
	if err := w.Write(headers); err != nil {
		return err
	}
	for _, row := range rows {
		if err := w.Write(row); err != nil {
			return err
		}
	}
	w.Flush()
	return w.Error()
}

func printQuiet(rows [][]string) error {
	for _, row := range rows {
		if len(row) > 0 {
			fmt.Println(row[0]) // Print first column only (typically ID)
		}
	}
	return nil
}
