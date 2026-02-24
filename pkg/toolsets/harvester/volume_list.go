package harvester

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mrostamii/rancher-mcp-server/pkg/client/rancher"
)

func (t *Toolset) volumeListTool() mcp.Tool {
	return mcp.NewTool(
		"harvester_volume_list",
		mcp.WithDescription("List PersistentVolumeClaims (Longhorn-backed volumes) with size and health"),
		mcp.WithString("cluster", mcp.Required(), mcp.Description("Harvester cluster ID")),
		mcp.WithString("namespace", mcp.Description("Namespace (empty = all)")),
		mcp.WithString("format", mcp.Description("Output format: json, table (default: json)")),
		mcp.WithNumber("limit", mcp.Description("Max items (default: 100)")),
	)
}

func (t *Toolset) volumeListHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	cluster := req.GetString("cluster", "")
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
	col, err := t.client.List(ctx, cluster, rancher.TypePersistentVolumeClaims, opts)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to list volumes: %v", err)), nil
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
