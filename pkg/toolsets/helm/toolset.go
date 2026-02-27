package helm

import (
	"github.com/mark3labs/mcp-go/server"
	"github.com/mrostamii/rancher-mcp-server/internal/security"
	"github.com/mrostamii/rancher-mcp-server/pkg/formatter"
)

// Toolset implements the Helm MCP toolset (list, get, history, install, upgrade, uninstall, rollback, repo_list).
type Toolset struct {
	baseURL  string
	token    string
	insecure bool
	policy   *security.Policy
	formatter formatter.Formatter
}

// NewToolset creates a Helm toolset. baseURL and token are the Rancher server connection; cluster is passed per-call.
func NewToolset(baseURL, token string, insecure bool, policy *security.Policy) *Toolset {
	return &Toolset{
		baseURL:   baseURL,
		token:     token,
		insecure:  insecure,
		policy:    policy,
		formatter: formatter.JSONFormatter{},
	}
}

// Register adds all Helm tools to the MCP server.
func (t *Toolset) Register(s *server.MCPServer) {
	s.AddTool(t.listTool(), t.listHandler)
	s.AddTool(t.getTool(), t.getHandler)
	s.AddTool(t.historyTool(), t.historyHandler)
	s.AddTool(t.repoListTool(), t.repoListHandler)
	if t.policy.CanWrite() {
		s.AddTool(t.installTool(), t.installHandler)
		s.AddTool(t.upgradeTool(), t.upgradeHandler)
		s.AddTool(t.rollbackTool(), t.rollbackHandler)
	}
	if t.policy.CanDelete() {
		s.AddTool(t.uninstallTool(), t.uninstallHandler)
	}
}
