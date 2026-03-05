package fleet

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mrostamii/rancher-mcp-server/pkg/client/rancher"
)

func (t *Toolset) gitrepoDeleteTool() mcp.Tool {
	return mcp.NewTool(
		"fleet_gitrepo_delete",
		mcp.WithDescription("Delete a Fleet GitRepo (requires read_only=false, destructive allowed)"),
		mcp.WithString("name", mcp.Required(), mcp.Description("GitRepo name")),
		mcp.WithString("namespace", mcp.Description("Namespace (default: fleet-default)")),
	)
}

func (t *Toolset) gitrepoDeleteHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if err := t.policy.CheckDestructive(); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	name, err := req.RequireString("name")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	namespace := req.GetString("namespace", "fleet-default")

	if err := t.client.Delete(ctx, localCluster, rancher.TypeFleetGitRepos, namespace, name); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("fleet_gitrepo_delete: %v", err)), nil
	}
	return mcp.NewToolResultText(fmt.Sprintf("Deleted GitRepo %q from namespace %s", name, namespace)), nil
}
