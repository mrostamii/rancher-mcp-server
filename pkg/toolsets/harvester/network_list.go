package harvester

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mrostamii/rancher-mcp-server/pkg/client/rancher"
)

func (t *Toolset) networkListTool() mcp.Tool {
	return mcp.NewTool(
		"harvester_network_list",
		mcp.WithDescription("List Harvester VLAN networks (NetworkAttachmentDefinition)"),
		mcp.WithString("cluster", mcp.Description("Harvester cluster ID (optional if default_harvester_cluster is set in config)")),
		mcp.WithString("namespace", mcp.Description("Namespace (empty = all)")),
		mcp.WithString("format", mcp.Description("Output format: json, table (default: json)")),
		mcp.WithNumber("limit", mcp.Description("Max items (default: 100)")),
	)
}

func (t *Toolset) networkListHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	cluster := t.cluster(req)
	if cluster == "" {
		return mcp.NewToolResultError("cluster is required (or set default_harvester_cluster in config)"), nil
	}
	namespace := req.GetString("namespace", "")
	format := req.GetString("format", "json")
	limit := req.GetInt("limit", 100)
	if limit <= 0 {
		limit = 100
	}

	opts := rancher.ListOpts{Limit: limit}
	if namespace != "" {
		opts.Namespace = namespace
	}
	col, err := t.client.List(ctx, cluster, rancher.TypeNetworkAttachmentDefinition, opts)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to list networks: %v", err)), nil
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
	out, err := t.formatter.Format(items, format)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("format: %v", err)), nil
	}
	return mcp.NewToolResultText(out), nil
}
