// cmd/crewsbase/main.go
package main

import (
	"errors"
	"os"

	"github.com/crewsbase/crewsbase-cli/internal/api"
	"github.com/crewsbase/crewsbase-cli/internal/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		var apiErr *api.APIError
		if errors.As(err, &apiErr) {
			os.Exit(apiErr.Code)
		}
		os.Exit(1)
	}
}
