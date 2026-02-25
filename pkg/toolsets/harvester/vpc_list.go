package harvester

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mrostamii/rancher-mcp-server/pkg/client/rancher"
)

func (t *Toolset) vpcListTool() mcp.Tool {
	return mcp.NewTool(
		"harvester_vpc_list",
		mcp.WithDescription("List KubeOVN VPCs (requires kubeovn-operator addon enabled)"),
		mcp.WithString("cluster", mcp.Required(), mcp.Description("Harvester cluster ID (where KubeOVN runs)")),
		mcp.WithString("format", mcp.Description("Output format: json, table (default: json)")),
		mcp.WithNumber("limit", mcp.Description("Max items (default: 100)")),
	)
}

func (t *Toolset) vpcListHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	cluster := req.GetString("cluster", "")
	format := req.GetString("format", "json")
	limit := req.GetInt("limit", 100)
	if limit <= 0 {
		limit = 100
	}

	opts := rancher.ListOpts{Limit: limit}
	col, err := t.client.List(ctx, cluster, rancher.TypeVpcs, opts)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to list VPCs: %v", err)), nil
	}

	items := make([]map[string]interface{}, 0, len(col.Data))
	for _, r := range col.Data {
		items = append(items, map[string]interface{}{
			"name":      r.ObjectMeta.Name,
			"spec":      r.Spec,
			"status":    r.Status,
		})
	}
	out, err := t.formatter.Format(items, format)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("format: %v", err)), nil
	}
	return mcp.NewToolResultText(out), nil
}
