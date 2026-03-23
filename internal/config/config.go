// internal/config/config.go
package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

const (
	DefaultAPIURL = "https://crewsbase.app"
)

func GetToken(flagOverride string) string {
	if flagOverride != "" {
		return flagOverride
	}
	if token := os.Getenv("CREWSBASE_TOKEN"); token != "" {
		return token
	}
	return viper.GetString("auth.token")
}

func GetAccount(flagOverride string) string {
	if flagOverride != "" {
		return flagOverride
	}
	return viper.GetString("default_account")
}

func GetAPIURL() string {
	url := viper.GetString("api_url")
	if url == "" {
		return DefaultAPIURL
	}
	return url
}

func SetToken(token string) error {
	viper.Set("auth.token", token)
	return writeConfig()
}

func RemoveToken() error {
	viper.Set("auth.token", "")
	return writeConfig()
}

func Set(key, value string) error {
	validKeys := map[string]bool{
		"default_account": true,
		"api_url":         true,
	}
	if !validKeys[key] {
		return fmt.Errorf("invalid config key %q (valid: default_account, api_url)", key)
	}
	viper.Set(key, value)
	return writeConfig()
}

func writeConfig() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	dir := filepath.Join(home, ".crewsbase")
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	return viper.WriteConfigAs(filepath.Join(dir, "config.yaml"))
}
