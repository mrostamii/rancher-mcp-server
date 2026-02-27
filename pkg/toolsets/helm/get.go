package helm

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"helm.sh/helm/v3/pkg/action"
)

func (t *Toolset) getTool() mcp.Tool {
	return mcp.NewTool(
		"helm_get",
		mcp.WithDescription("Get detailed info for a Helm release (manifest, values, notes)"),
		mcp.WithString("cluster", mcp.Required(), mcp.Description("Rancher cluster ID")),
		mcp.WithString("release", mcp.Required(), mcp.Description("Release name")),
		mcp.WithString("namespace", mcp.Description("Namespace (default: default)")),
		mcp.WithString("format", mcp.Description("Output format: json, table (default: json)")),
		mcp.WithNumber("revision", mcp.Description("Revision number (0 = current)")),
	)
}

func (t *Toolset) getHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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
	revision := req.GetInt("revision", 0)

	cfg, err := t.actionConfigFor(cluster, namespace)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("helm action config: %v", err)), nil
	}
	getAction := action.NewGet(cfg)
	getAction.Version = revision

	rel, err := getAction.Run(releaseName)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("helm get: %v", err)), nil
	}
	data := map[string]interface{}{
		"name":       rel.Name,
		"namespace":  rel.Namespace,
		"revision":   rel.Version,
		"status":     rel.Info.Status.String(),
		"chart":      rel.Chart.Metadata.Name + "-" + rel.Chart.Metadata.Version,
		"updated":    rel.Info.LastDeployed.Format("2006-01-02 15:04:05"),
		"manifest":   rel.Manifest,
		"notes":      rel.Info.Notes,
		"values":     rel.Config,
	}
	out, err := t.formatter.Format(data, format)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("format: %v", err)), nil
	}
	return mcp.NewToolResultText(out), nil
}
