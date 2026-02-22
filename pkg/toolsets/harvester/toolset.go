package harvester

import (
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/mrostamii/rancher-mcp-server/internal/security"
	"github.com/mrostamii/rancher-mcp-server/pkg/client/rancher"
	"github.com/mrostamii/rancher-mcp-server/pkg/formatter"
)

// cluster returns the cluster ID from the request, or the default if empty.
func (t *Toolset) cluster(req mcp.CallToolRequest) string {
	c := req.GetString("cluster", "")
	if c != "" {
		return c
	}
	return t.defaultCluster
}

// Toolset implements the Harvester MCP toolset (VMs, images, volumes, networks, hosts).
type Toolset struct {
	client             *rancher.SteveClient
	policy             *security.Policy
	formatter          formatter.Formatter
	defaultCluster     string // optional; used when cluster param is empty
}

// NewToolset creates a Harvester toolset. defaultCluster is optional (e.g. from config); used when a tool omits cluster.
func NewToolset(client *rancher.SteveClient, policy *security.Policy, defaultCluster string) *Toolset {
	return &Toolset{
		client:         client,
		policy:         policy,
		formatter:      formatter.JSONFormatter{},
		defaultCluster: defaultCluster,
	}
}

// Register adds all Harvester tools to the MCP server.
func (t *Toolset) Register(s *server.MCPServer) {
	s.AddTool(t.vmListTool(), t.vmListHandler)
	s.AddTool(t.vmGetTool(), t.vmGetHandler)
	s.AddTool(t.imageListTool(), t.imageListHandler)
	s.AddTool(t.volumeListTool(), t.volumeListHandler)
	s.AddTool(t.networkListTool(), t.networkListHandler)
	s.AddTool(t.hostListTool(), t.hostListHandler)
	if t.policy.CanWrite() {
		s.AddTool(t.vmActionTool(), t.vmActionHandler)
	}
}
