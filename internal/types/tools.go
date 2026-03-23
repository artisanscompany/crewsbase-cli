// internal/types/tools.go
package types

// ToolDomain represents a group of tools in a domain
type ToolDomain struct {
	Domain string     `json:"domain"`
	Tools  []ToolInfo `json:"tools"`
}

// ToolInfo represents a single tool
type ToolInfo struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Category    string                 `json:"category"`
	Parameters  map[string]interface{} `json:"parameters"`
}
