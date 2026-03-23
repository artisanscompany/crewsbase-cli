// cmd/crewsbase/main.go
package main

import (
	"os"

	"github.com/crewsbase/crewsbase-cli/internal/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
