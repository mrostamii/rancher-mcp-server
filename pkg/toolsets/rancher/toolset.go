package rancher

import (
	"github.com/mark3labs/mcp-go/server"
	"github.com/mrostamii/rancher-mcp-server/internal/security"
	"github.com/mrostamii/rancher-mcp-server/pkg/client/rancher"
	"github.com/mrostamii/rancher-mcp-server/pkg/formatter"
)

const localCluster = "local" // Rancher management resources live on "local"

// Toolset implements the Rancher management toolset (clusters, projects, overview).
type Toolset struct {
	client    *rancher.SteveClient
	policy    *security.Policy
	formatter formatter.Formatter
}

// NewToolset creates a Rancher toolset.
func NewToolset(client *rancher.SteveClient, policy *security.Policy) *Toolset {
	return &Toolset{
		client:    client,
		policy:    policy,
		formatter: formatter.JSONFormatter{},
	}
}

// Register adds all Rancher tools to the MCP server.
func (t *Toolset) Register(s *server.MCPServer) {
	s.AddTool(t.clusterListTool(), t.clusterListHandler)
	s.AddTool(t.clusterGetTool(), t.clusterGetHandler)
	s.AddTool(t.projectListTool(), t.projectListHandler)
	s.AddTool(t.overviewTool(), t.overviewHandler)
}
