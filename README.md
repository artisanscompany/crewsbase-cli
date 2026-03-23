# Crewsbase CLI

Command-line interface for [Crewsbase](https://crewsbase.app).

## Install

### Homebrew (macOS/Linux)

    brew install crewsbase/tap/crewsbase

### Binary

Download from [GitHub Releases](https://github.com/crewsbase/crewsbase-cli/releases).

### Go

    go install github.com/crewsbase/crewsbase-cli/cmd/crewsbase@latest

## Authentication

Interactive (opens browser):

    crewsbase auth login

Token (for CI/agents):

    crewsbase auth login --token cb_your_token
    # or
    export CREWSBASE_TOKEN=cb_your_token

## Usage

    crewsbase config set default_account myteam

    crewsbase crm tables list
    crewsbase crm rows list --table TABLE_ID
    crewsbase crm rows create --table TABLE_ID --field name="Jane" --field email="jane@co.com"
    crewsbase crm rows update ROW_ID --table TABLE_ID --field name="Jane Smith"
    crewsbase crm rows delete ROW_ID --table TABLE_ID

    # Output formats
    crewsbase crm tables list --output json
    crewsbase crm tables list --output csv

## Shell Completions

    crewsbase completion bash >> ~/.bashrc
    crewsbase completion zsh >> ~/.zshrc
    crewsbase completion fish > ~/.config/fish/completions/crewsbase.fish
