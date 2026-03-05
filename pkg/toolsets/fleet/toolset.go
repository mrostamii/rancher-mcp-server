package fleet

import (
	"github.com/mark3labs/mcp-go/server"
	"github.com/mrostamii/rancher-mcp-server/internal/security"
	"github.com/mrostamii/rancher-mcp-server/pkg/client/rancher"
	"github.com/mrostamii/rancher-mcp-server/pkg/formatter"
)

const localCluster = "local" // Fleet runs in the Rancher management cluster

// Toolset implements the Fleet GitOps toolset (gitrepos, bundles, clusters, drift_detect).
type Toolset struct {
	client    *rancher.SteveClient
	policy    *security.Policy
	formatter formatter.Formatter
}

// NewToolset creates a Fleet toolset.
func NewToolset(client *rancher.SteveClient, policy *security.Policy) *Toolset {
	return &Toolset{
		client:    client,
		policy:    policy,
		formatter: formatter.JSONFormatter{},
	}
}

// Register adds all Fleet tools to the MCP server.
func (t *Toolset) Register(s *server.MCPServer) {
	s.AddTool(t.gitrepoListTool(), t.gitrepoListHandler)
	s.AddTool(t.gitrepoGetTool(), t.gitrepoGetHandler)
	s.AddTool(t.bundleListTool(), t.bundleListHandler)
	s.AddTool(t.clusterListTool(), t.clusterListHandler)
	s.AddTool(t.driftDetectTool(), t.driftDetectHandler)
	if t.policy.CanWrite() {
		s.AddTool(t.gitrepoCreateTool(), t.gitrepoCreateHandler)
		s.AddTool(t.gitrepoActionTool(), t.gitrepoActionHandler)
		s.AddTool(t.gitrepoCloneTool(), t.gitrepoCloneHandler)
	}
	if t.policy.CanDelete() {
		s.AddTool(t.gitrepoDeleteTool(), t.gitrepoDeleteHandler)
	}
}
