package rancher

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mrostamii/rancher-mcp-server/pkg/client/rancher"
)

func (t *Toolset) clusterGetTool() mcp.Tool {
	return mcp.NewTool(
		"rancher_cluster_get",
		mcp.WithDescription("Get detailed info for one Rancher cluster (health, version, node count)"),
		mcp.WithString("name", mcp.Required(), mcp.Description("Cluster name or ID")),
		mcp.WithString("format", mcp.Description("Output format: json, table (default: json)")),
	)
}

func (t *Toolset) clusterGetHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	name, err := req.RequireString("name")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	format := req.GetString("format", "json")

	res, err := t.client.Get(ctx, localCluster, rancher.TypeManagementClusters, "", name)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("cluster %q not found: %v", name, err)), nil
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
