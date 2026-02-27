package harvester

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mrostamii/rancher-mcp-server/pkg/client/rancher"
	"github.com/mrostamii/rancher-mcp-server/pkg/formatter"
)

func (t *Toolset) subnetListTool() mcp.Tool {
	return mcp.NewTool(
		"harvester_subnet_list",
		mcp.WithDescription("List KubeOVN Subnets (requires kubeovn-operator addon). Subnets define VM network CIDRs and are linked to Networks via provider."),
		mcp.WithString("cluster", mcp.Required(), mcp.Description("Harvester cluster ID")),
		mcp.WithString("format", mcp.Description("Output format: json, table (default: json)")),
		mcp.WithNumber("limit", mcp.Description("Max items (default: 100)")),
		mcp.WithString("continue", mcp.Description("Pagination token from previous response (for next page)")),
	)
}

func (t *Toolset) subnetListHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	cluster := req.GetString("cluster", "")
	format := req.GetString("format", "json")
	limit := req.GetInt("limit", 100)
	if limit <= 0 {
		limit = 100
	}
	continueToken := req.GetString("continue", "")

	opts := rancher.ListOpts{Limit: limit, Continue: continueToken}
	col, err := t.client.List(ctx, cluster, rancher.TypeSubnets, opts)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to list subnets: %v", err)), nil
	}

	items := make([]map[string]interface{}, 0, len(col.Data))
	for _, r := range col.Data {
		items = append(items, map[string]interface{}{
			"name":   r.ObjectMeta.Name,
			"spec":   r.Spec,
			"status": r.Status,
		})
	}
	out, err := formatter.FormatListWithContinue(t.formatter, items, col.Continue, format)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("format: %v", err)), nil
	}
	return mcp.NewToolResultText(out), nil
}
