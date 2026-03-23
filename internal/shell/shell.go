// internal/shell/shell.go
package shell

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/artisanscompany/crewsbase-cli/internal/api"
	"github.com/artisanscompany/crewsbase-cli/internal/types"
	"github.com/chzyer/readline"
)

// Shell is the interactive REPL
type Shell struct {
	client    *api.Client
	domains   []types.ToolDomain
	account   string
	debug     bool
	toolIndex map[string]types.ToolInfo // human command -> tool info
	commands  []string                  // sorted list of human commands
}

// New creates a new shell instance
func New(client *api.Client, domains []types.ToolDomain, account string, debug bool) *Shell {
	s := &Shell{
		client:    client,
		domains:   domains,
		account:   account,
		debug:     debug,
		toolIndex: make(map[string]types.ToolInfo),
	}
	s.buildIndex()
	return s
}

func (s *Shell) buildIndex() {
	for _, d := range s.domains {
		for _, t := range d.Tools {
			cmd := toolNameToCommand(t.Name)
			s.toolIndex[cmd] = t
			s.commands = append(s.commands, cmd)
		}
	}
	sort.Strings(s.commands)
}

// Run starts the interactive shell
func (s *Shell) Run() error {
	completer := readline.NewPrefixCompleter(s.buildCompleterItems()...)

	rl, err := readline.NewEx(&readline.Config{
		Prompt:          s.prompt(),
		HistoryFile:     historyFile(),
		AutoComplete:    completer,
		InterruptPrompt: "^C",
		EOFPrompt:       "exit",
	})
	if err != nil {
		return err
	}
	defer rl.Close()

	for {
		line, err := rl.Readline()
		if err == readline.ErrInterrupt {
			continue
		}
		if err == io.EOF {
			fmt.Println("Bye!")
			return nil
		}

		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Meta commands start with /
		if strings.HasPrefix(line, "/") {
			if quit := s.handleMeta(line, rl); quit {
				return nil
			}
			continue
		}

		s.executeCommand(line)
	}
}

func (s *Shell) prompt() string {
	return fmt.Sprintf("\033[1;36mcrewsbase\033[0m (\033[33m%s\033[0m)> ", s.account)
}

func (s *Shell) handleMeta(line string, rl *readline.Instance) bool {
	parts := strings.Fields(line)
	cmd := parts[0]

	switch cmd {
	case "/exit", "/quit", "/q":
		fmt.Println("Bye!")
		return true

	case "/help", "/h":
		if len(parts) > 1 {
			s.showDomainHelp(parts[1])
		} else {
			s.showHelp()
		}

	case "/tools", "/t":
		if len(parts) > 1 {
			s.showDomainHelp(parts[1])
		} else {
			s.showDomainSummary()
		}

	case "/account":
		if len(parts) < 2 {
			fmt.Printf("Current account: %s\n", s.account)
		} else {
			s.account = parts[1]
			rl.SetPrompt(s.prompt())
			fmt.Printf("Switched to %s\n", s.account)
		}

	case "/debug":
		s.debug = !s.debug
		fmt.Printf("Debug mode: %v\n", s.debug)

	default:
		fmt.Printf("Unknown command: %s (try /help)\n", cmd)
	}

	return false
}

func (s *Shell) showHelp() {
	fmt.Println(`
Commands:
  <domain> <action>          Execute a tool (e.g. "crm list-tables")
  <domain> <action> --help   Show tool parameters

Meta commands:
  /tools                     List all domains and tool counts
  /tools <domain>            List tools in a domain
  /help                      Show this help
  /account [slug]            Show or switch account
  /debug                     Toggle debug mode
  /exit                      Exit shell

Tab completion is available for all commands.`)
}

func (s *Shell) showDomainSummary() {
	fmt.Println()
	for _, d := range s.domains {
		fmt.Printf("  %-20s %d tools\n", d.Domain, len(d.Tools))
	}
	fmt.Println()
	fmt.Println("  Use /tools <domain> to list tools in a domain")
}

func (s *Shell) showDomainHelp(domain string) {
	for _, d := range s.domains {
		if d.Domain == domain {
			fmt.Println()
			for _, t := range d.Tools {
				cmd := toolNameToCommand(t.Name)
				fmt.Printf("  %-40s %s\n", cmd, t.Description)
			}
			fmt.Println()
			return
		}
	}
	fmt.Printf("Unknown domain: %s\n", domain)
}

func (s *Shell) executeCommand(line string) {
	parts := strings.Fields(line)
	if len(parts) < 2 {
		fmt.Println("Usage: <domain> <action> [--param value ...]")
		fmt.Println("Try /tools to see available commands")
		return
	}

	// Build the command key (first two parts)
	cmdKey := parts[0] + " " + parts[1]
	remainingParts := parts[2:]

	tool, ok := s.toolIndex[cmdKey]
	if !ok {
		// Try multi-word domains: "social_media list-posts" stored as "social_media list-posts"
		// Input might be "social_media list-posts" already matching
		found := false
		for i := 2; i < len(parts) && i <= 3; i++ {
			cmdKey = strings.Join(parts[:i], " ")
			if tool, ok = s.toolIndex[cmdKey]; ok {
				remainingParts = parts[i:]
				found = true
				break
			}
		}
		if !found {
			fmt.Printf("Unknown command: %s (try /tools)\n", parts[0]+" "+parts[1])
			return
		}
	}

	// Check for --help
	for _, p := range remainingParts {
		if p == "--help" || p == "-h" {
			s.showToolHelp(tool)
			return
		}
	}

	// Parse remaining args as --key value pairs
	arguments := parseArgs(remainingParts)

	// Execute via API
	result, err := s.client.ExecuteTool(tool.Name, arguments)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return
	}

	// Pretty print the result
	prettyJSON, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		fmt.Println(result)
		return
	}
	fmt.Println(string(prettyJSON))
}

func (s *Shell) showToolHelp(tool types.ToolInfo) {
	fmt.Printf("\n  %s\n", tool.Description)
	fmt.Printf("  Category: %s\n\n", tool.Category)

	props, ok := tool.Parameters["properties"].(map[string]interface{})
	if !ok || len(props) == 0 {
		fmt.Println("  No parameters")
		return
	}

	required := map[string]bool{}
	if reqList, ok := tool.Parameters["required"].([]interface{}); ok {
		for _, r := range reqList {
			if s, ok := r.(string); ok {
				required[s] = true
			}
		}
	}

	fmt.Println("  Parameters:")
	for name, spec := range props {
		specMap, ok := spec.(map[string]interface{})
		if !ok {
			continue
		}
		typ := specMap["type"]
		desc := specMap["description"]
		req := ""
		if required[name] {
			req = " (required)"
		}
		fmt.Printf("    --%-25s %-10v %s%s\n", name, typ, desc, req)
	}
	fmt.Println()
}

func (s *Shell) buildCompleterItems() []readline.PrefixCompleterInterface {
	domainCmds := map[string][]string{}
	for _, cmd := range s.commands {
		parts := strings.SplitN(cmd, " ", 2)
		if len(parts) == 2 {
			domainCmds[parts[0]] = append(domainCmds[parts[0]], parts[1])
		}
	}

	var items []readline.PrefixCompleterInterface

	domains := make([]string, 0, len(domainCmds))
	for d := range domainCmds {
		domains = append(domains, d)
	}
	sort.Strings(domains)

	for _, domain := range domains {
		subCmds := domainCmds[domain]
		sort.Strings(subCmds)
		var subItems []readline.PrefixCompleterInterface
		for _, sub := range subCmds {
			subItems = append(subItems, readline.PcItem(sub))
		}
		items = append(items, readline.PcItem(domain, subItems...))
	}

	// Meta commands
	items = append(items,
		readline.PcItem("/help"),
		readline.PcItem("/tools"),
		readline.PcItem("/account"),
		readline.PcItem("/debug"),
		readline.PcItem("/exit"),
	)

	return items
}

// parseArgs converts ["--search", "foo", "--page", "2"] to {"search": "foo", "page": "2"}
func parseArgs(parts []string) map[string]interface{} {
	args := map[string]interface{}{}
	for i := 0; i < len(parts); i++ {
		if strings.HasPrefix(parts[i], "--") {
			key := strings.TrimPrefix(parts[i], "--")
			key = strings.ReplaceAll(key, "-", "_")
			if i+1 < len(parts) && !strings.HasPrefix(parts[i+1], "--") {
				args[key] = parts[i+1]
				i++
			} else {
				args[key] = true
			}
		}
	}
	return args
}

// toolNameToCommand converts "crm_list_tables" to "crm list-tables"
func toolNameToCommand(name string) string {
	parts := strings.SplitN(name, "_", 2)
	if len(parts) < 2 {
		return name
	}
	rest := strings.ReplaceAll(parts[1], "_", "-")
	return parts[0] + " " + rest
}

func historyFile() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return home + "/.crewsbase/shell_history"
}
