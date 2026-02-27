package rancher

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mrostamii/rancher-mcp-server/pkg/client/rancher"
	"github.com/mrostamii/rancher-mcp-server/pkg/formatter"
)

func (t *Toolset) clusterListTool() mcp.Tool {
	return mcp.NewTool(
		"rancher_cluster_list",
		mcp.WithDescription("List all Rancher clusters with health, K8s version, provider, node count"),
		mcp.WithString("format", mcp.Description("Output format: json, table (default: json)")),
		mcp.WithNumber("limit", mcp.Description("Max items (default: 100)")),
		mcp.WithString("continue", mcp.Description("Pagination token from previous response (for next page)")),
	)
}

func (t *Toolset) clusterListHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	format := req.GetString("format", "json")
	limit := req.GetInt("limit", 100)
	if limit <= 0 {
		limit = 100
	}
	continueToken := req.GetString("continue", "")

	opts := rancher.ListOpts{Limit: limit, Continue: continueToken}
	col, err := t.client.List(ctx, localCluster, rancher.TypeManagementClusters, opts)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to list clusters: %v", err)), nil
	}
	items := make([]map[string]interface{}, 0, len(col.Data))
	for _, r := range col.Data {
		items = append(items, map[string]interface{}{
			"name":      r.ObjectMeta.Name,
			"metadata": r.ObjectMeta,
			"spec":     r.Spec,
			"status":   r.Status,
		})
	}
	out, err := formatter.FormatListWithContinue(t.formatter, items, col.Continue, format)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("format: %v", err)), nil
	}
	return mcp.NewToolResultText(out), nil
}
