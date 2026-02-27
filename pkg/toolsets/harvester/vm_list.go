package harvester

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mrostamii/rancher-mcp-server/pkg/client/rancher"
	"github.com/mrostamii/rancher-mcp-server/pkg/formatter"
)

func (t *Toolset) vmListTool() mcp.Tool {
	return mcp.NewTool(
		"harvester_vm_list",
		mcp.WithDescription("List virtual machines across Harvester clusters with status, IP, node, CPU/memory"),
		mcp.WithString("cluster", mcp.Required(), mcp.Description("Harvester cluster ID")),
		mcp.WithString("namespace", mcp.Description("Namespace (empty = all namespaces)")),
		mcp.WithString("format", mcp.Description("Output format: json, table (default: json)")),
		mcp.WithNumber("limit", mcp.Description("Max items to return (default: 100)")),
		mcp.WithString("continue", mcp.Description("Pagination token from previous response (for next page)")),
	)
}

func (t *Toolset) vmListHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	cluster := req.GetString("cluster", "")
	namespace := req.GetString("namespace", "")
	format := req.GetString("format", "json")
	limit := req.GetInt("limit", 100)
	if limit <= 0 {
		limit = 100
	}
	continueToken := req.GetString("continue", "")

	opts := rancher.ListOpts{Limit: limit, Continue: continueToken}
	if namespace != "" {
		opts.Namespace = namespace
	}
	col, err := t.client.List(ctx, cluster, rancher.TypeVirtualMachines, opts)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to list VMs: %v", err)), nil
	}
	items := make([]map[string]interface{}, 0, len(col.Data))
	for _, r := range col.Data {
		items = append(items, map[string]interface{}{
			"name":      r.ObjectMeta.Name,
			"namespace": r.ObjectMeta.Namespace,
			"spec":      r.Spec,
			"status":    r.Status,
		})
	}
	out, err := formatter.FormatListWithContinue(t.formatter, items, col.Continue, format)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("format: %v", err)), nil
	}
	return mcp.NewToolResultText(out), nil
}
