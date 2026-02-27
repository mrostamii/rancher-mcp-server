package helm

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"helm.sh/helm/v3/pkg/action"
)

func (t *Toolset) historyTool() mcp.Tool {
	return mcp.NewTool(
		"helm_history",
		mcp.WithDescription("Get revision history for a Helm release"),
		mcp.WithString("cluster", mcp.Required(), mcp.Description("Rancher cluster ID")),
		mcp.WithString("release", mcp.Required(), mcp.Description("Release name")),
		mcp.WithString("namespace", mcp.Description("Namespace (default: default)")),
		mcp.WithString("format", mcp.Description("Output format: json, table (default: json)")),
		mcp.WithNumber("max", mcp.Description("Max revisions to return (default: 256)")),
	)
}

func (t *Toolset) historyHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	cluster, err := req.RequireString("cluster")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	releaseName, err := req.RequireString("release")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	namespace := req.GetString("namespace", "default")
	format := req.GetString("format", "json")
	max := req.GetInt("max", 256)

	cfg, err := t.actionConfigFor(cluster, namespace)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("helm action config: %v", err)), nil
	}
	histAction := action.NewHistory(cfg)
	histAction.Max = max

	releases, err := histAction.Run(releaseName)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("helm history: %v", err)), nil
	}
	items := make([]map[string]interface{}, 0, len(releases))
	for _, r := range releases {
		items = append(items, map[string]interface{}{
			"revision": r.Version,
			"updated":  r.Info.LastDeployed.Format("2006-01-02 15:04:05"),
			"status":   r.Info.Status.String(),
			"chart":    r.Chart.Metadata.Name + "-" + r.Chart.Metadata.Version,
			"description": r.Info.Description,
		})
	}
	out, err := t.formatter.Format(items, format)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("format: %v", err)), nil
	}
	return mcp.NewToolResultText(out), nil
}
