package rancher

import (
	"github.com/mark3labs/mcp-go/server"
	"github.com/mrostamii/rancher-mcp-server/internal/security"
	"github.com/mrostamii/rancher-mcp-server/pkg/client/rancher"
	"github.com/mrostamii/rancher-mcp-server/pkg/formatter"
)

const localCluster = "local" // Rancher management resources live on "local"

// Toolset implements the Rancher management toolset (clusters, projects, overview, Norman /v3).
type Toolset struct {
	client    *rancher.SteveClient
	norman    *rancher.NormanClient
	policy    *security.Policy
	formatter formatter.Formatter
}

// NewToolset creates a Rancher toolset. norman may be nil to skip Norman API tools (tests only).
func NewToolset(client *rancher.SteveClient, norman *rancher.NormanClient, policy *security.Policy) *Toolset {
	return &Toolset{
		client:    client,
		norman:    norman,
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

	if t.norman == nil {
		return
	}

	s.AddTool(t.normanSchemaListTool(), t.normanSchemaListHandler)
	s.AddTool(t.normanSchemaGetTool(), t.normanSchemaGetHandler)
	s.AddTool(t.normanTokenListTool(), t.normanTokenListHandler)
	s.AddTool(t.normanTokenGetTool(), t.normanTokenGetHandler)
	s.AddTool(t.normanAuthConfigListTool(), t.normanAuthConfigListHandler)
	s.AddTool(t.normanAuthConfigGetTool(), t.normanAuthConfigGetHandler)
	s.AddTool(t.normanUserListTool(), t.normanUserListHandler)
	s.AddTool(t.normanUserGetTool(), t.normanUserGetHandler)
	s.AddTool(t.normanGlobalRoleBindingListTool(), t.normanGlobalRoleBindingListHandler)
	s.AddTool(t.normanGlobalRoleBindingGetTool(), t.normanGlobalRoleBindingGetHandler)
	s.AddTool(t.normanClusterRegistrationTokenListTool(), t.normanClusterRegistrationTokenListHandler)
	s.AddTool(t.normanClusterRegistrationTokenGetTool(), t.normanClusterRegistrationTokenGetHandler)
	s.AddTool(t.normanNodeDriverListTool(), t.normanNodeDriverListHandler)
	s.AddTool(t.normanCloudCredentialListTool(), t.normanCloudCredentialListHandler)
	s.AddTool(t.normanCloudCredentialGetTool(), t.normanCloudCredentialGetHandler)
	s.AddTool(t.normanCatalogListTool(), t.normanCatalogListHandler)
	s.AddTool(t.normanCatalogGetTool(), t.normanCatalogGetHandler)
	s.AddTool(t.normanClusterRepoListTool(), t.normanClusterRepoListHandler)
	s.AddTool(t.normanClusterRepoGetTool(), t.normanClusterRepoGetHandler)
	s.AddTool(t.normanFeatureListTool(), t.normanFeatureListHandler)
	s.AddTool(t.normanFeatureGetTool(), t.normanFeatureGetHandler)
	s.AddTool(t.normanSettingListTool(), t.normanSettingListHandler)
	s.AddTool(t.normanSettingGetTool(), t.normanSettingGetHandler)
	s.AddTool(t.normanAuditLogListTool(), t.normanAuditLogListHandler)

	if t.policy.CanWrite() {
		s.AddTool(t.normanActionTool(), t.normanActionHandler)
		s.AddTool(t.normanTokenCreateTool(), t.normanTokenCreateHandler)
		s.AddTool(t.normanAuthConfigUpdateTool(), t.normanAuthConfigUpdateHandler)
		s.AddTool(t.normanUserCreateTool(), t.normanUserCreateHandler)
		s.AddTool(t.normanUserDisableTool(), t.normanUserDisableHandler)
		s.AddTool(t.normanUserEnableTool(), t.normanUserEnableHandler)
		s.AddTool(t.normanGlobalRoleBindingCreateTool(), t.normanGlobalRoleBindingCreateHandler)
		s.AddTool(t.normanClusterRegistrationTokenCreateTool(), t.normanClusterRegistrationTokenCreateHandler)
		s.AddTool(t.normanCloudCredentialCreateTool(), t.normanCloudCredentialCreateHandler)
		s.AddTool(t.normanCatalogRefreshTool(), t.normanCatalogRefreshHandler)
		s.AddTool(t.normanFeatureSetTool(), t.normanFeatureSetHandler)
		s.AddTool(t.normanSettingUpdateTool(), t.normanSettingUpdateHandler)
		s.AddTool(t.normanSupportBundleGenerateTool(), t.normanSupportBundleGenerateHandler)
	}
	if t.policy.CanDelete() {
		s.AddTool(t.normanTokenDeleteTool(), t.normanTokenDeleteHandler)
		s.AddTool(t.normanGlobalRoleBindingDeleteTool(), t.normanGlobalRoleBindingDeleteHandler)
		s.AddTool(t.normanClusterRegistrationTokenDeleteTool(), t.normanClusterRegistrationTokenDeleteHandler)
		s.AddTool(t.normanCloudCredentialDeleteTool(), t.normanCloudCredentialDeleteHandler)
	}
}
