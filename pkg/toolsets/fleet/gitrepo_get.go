package fleet

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mrostamii/rancher-mcp-server/pkg/client/rancher"
)

func (t *Toolset) gitrepoGetTool() mcp.Tool {
	return mcp.NewTool(
		"fleet_gitrepo_get",
		mcp.WithDescription("Get detailed info for one Fleet GitRepo"),
		mcp.WithString("name", mcp.Required(), mcp.Description("GitRepo name")),
		mcp.WithString("namespace", mcp.Description("Namespace (default: fleet-default)")),
		mcp.WithString("format", mcp.Description("Output format: json, table (default: json)")),
	)
}

func (t *Toolset) gitrepoGetHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	name, err := req.RequireString("name")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	namespace := req.GetString("namespace", "fleet-default")
	format := req.GetString("format", "json")

	res, err := t.client.Get(ctx, localCluster, rancher.TypeFleetGitRepos, namespace, name)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("GitRepo %q not found: %v", name, err)), nil
	}
	data := map[string]interface{}{
		"metadata": res.ObjectMeta,
		"spec":     res.Spec,
		"status":   res.Status,
	}
	out, err := t.formatter.Format(data, format)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("format: %v", err)), nil
	}
	return mcp.NewToolResultText(out), nil
}
