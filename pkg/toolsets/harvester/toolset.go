package harvester

import (
	"github.com/mark3labs/mcp-go/server"
	"github.com/mrostamii/rancher-mcp-server/internal/security"
	"github.com/mrostamii/rancher-mcp-server/pkg/client/rancher"
	"github.com/mrostamii/rancher-mcp-server/pkg/formatter"
)

// Toolset implements the Harvester MCP toolset (VMs, images, volumes, networks, hosts).
type Toolset struct {
	client    *rancher.SteveClient
	policy    *security.Policy
	formatter formatter.Formatter
}

// NewToolset creates a Harvester toolset.
func NewToolset(client *rancher.SteveClient, policy *security.Policy) *Toolset {
	return &Toolset{
		client:    client,
		policy:    policy,
		formatter: formatter.JSONFormatter{},
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
		s.AddTool(t.vmCreateTool(), t.vmCreateHandler)
		s.AddTool(t.vmSnapshotTool(), t.vmSnapshotHandler)
		s.AddTool(t.vmBackupTool(), t.vmBackupHandler)
		s.AddTool(t.imageCreateTool(), t.imageCreateHandler)
		s.AddTool(t.volumeCreateTool(), t.volumeCreateHandler)
	}
}
