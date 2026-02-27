package harvester

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mrostamii/rancher-mcp-server/pkg/client/rancher"
	"github.com/mrostamii/rancher-mcp-server/pkg/formatter"
)

func (t *Toolset) settingsTool() mcp.Tool {
	return mcp.NewTool(
		"harvester_settings",
		mcp.WithDescription("List or get Harvester cluster settings (backup-target, auto-disk-provision-paths, etc.)"),
		mcp.WithString("cluster", mcp.Required(), mcp.Description("Harvester cluster ID")),
		mcp.WithString("name", mcp.Description("Setting name (empty = list all)")),
		mcp.WithString("format", mcp.Description("Output format: json, table (default: json)")),
		mcp.WithNumber("limit", mcp.Description("Max items when listing (default: 100)")),
		mcp.WithString("continue", mcp.Description("Pagination token from previous response (for next page; list mode only)")),
	)
}

func (t *Toolset) settingsHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	cluster, err := req.RequireString("cluster")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	name := req.GetString("name", "")
	format := req.GetString("format", "json")
	limit := req.GetInt("limit", 100)
	if limit <= 0 {
		limit = 100
	}
	continueToken := req.GetString("continue", "")

	if name != "" {
		// Get single setting (cluster-scoped, empty namespace)
		setting, err := t.client.Get(ctx, cluster, rancher.TypeSettings, "", name)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("harvester_settings get: %v", err)), nil
		}
		var value string
		if spec, ok := setting.Spec.(map[string]interface{}); ok {
			if v, ok := spec["value"].(string); ok {
				value = v
			}
		}
		data := map[string]interface{}{
			"name":  setting.ObjectMeta.Name,
			"value": value,
			"spec":  setting.Spec,
		}
		out, err := t.formatter.Format(data, format)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("format: %v", err)), nil
		}
		return mcp.NewToolResultText(out), nil
	}

	// List all settings
	opts := rancher.ListOpts{Limit: limit, Continue: continueToken}
	col, err := t.client.List(ctx, cluster, rancher.TypeSettings, opts)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("harvester_settings list: %v", err)), nil
	}
	items := make([]map[string]interface{}, 0, len(col.Data))
	for _, r := range col.Data {
		var value string
		if spec, ok := r.Spec.(map[string]interface{}); ok {
			if v, ok := spec["value"].(string); ok {
				value = v
			}
		}
		items = append(items, map[string]interface{}{
			"name":  r.ObjectMeta.Name,
			"value": value,
		})
	}
	out, err := formatter.FormatListWithContinue(t.formatter, items, col.Continue, format)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("format: %v", err)), nil
	}
	return mcp.NewToolResultText(out), nil
}
