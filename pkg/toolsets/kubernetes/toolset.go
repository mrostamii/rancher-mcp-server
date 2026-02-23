package kubernetes

import (
	"github.com/mark3labs/mcp-go/server"
	"github.com/mrostamii/rancher-mcp-server/internal/security"
	"github.com/mrostamii/rancher-mcp-server/pkg/client/rancher"
	"github.com/mrostamii/rancher-mcp-server/pkg/formatter"
)

// Toolset implements the Kubernetes MCP toolset (generic resources, describe, events, capacity).
type Toolset struct {
	client    *rancher.SteveClient
	policy    *security.Policy
	formatter formatter.Formatter
}

// NewToolset creates a Kubernetes toolset.
func NewToolset(client *rancher.SteveClient, policy *security.Policy) *Toolset {
	return &Toolset{
		client:    client,
		policy:    policy,
		formatter: formatter.JSONFormatter{},
	}
}

// Register adds all Kubernetes tools to the MCP server.
func (t *Toolset) Register(s *server.MCPServer) {
	s.AddTool(t.listTool(), t.listHandler)
	s.AddTool(t.getTool(), t.getHandler)
	s.AddTool(t.describeTool(), t.describeHandler)
	s.AddTool(t.eventsTool(), t.eventsHandler)
	s.AddTool(t.capacityTool(), t.capacityHandler)
	if t.policy.CanWrite() {
		s.AddTool(t.createTool(), t.createHandler)
		s.AddTool(t.patchTool(), t.patchHandler)
	}
	if t.policy.CanDelete() {
		s.AddTool(t.deleteTool(), t.deleteHandler)
	}
}
