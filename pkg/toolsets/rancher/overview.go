package rancher

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mrostamii/rancher-mcp-server/pkg/client/rancher"
)

func (t *Toolset) overviewTool() mcp.Tool {
	return mcp.NewTool(
		"rancher_overview",
		mcp.WithDescription("Cross-cluster summary: total cluster count and total project count"),
		mcp.WithString("format", mcp.Description("Output format: json, table (default: json)")),
	)
}

func (t *Toolset) overviewHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	format := req.GetString("format", "json")

	col, err := t.client.List(ctx, localCluster, rancher.TypeManagementClusters, rancher.ListOpts{Limit: 500})
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to list clusters for overview: %v", err)), nil
	}
	total := len(col.Data)
	// Status.conditions often carry "Ready" or similar; we don't parse deeply here, just count
	summary := map[string]interface{}{
		"total_clusters": total,
		"clusters":       col.Data,
	}
	out, err := t.formatter.Format(summary, format)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("format: %v", err)), nil
	}
	return mcp.NewToolResultText(out), nil
}
