package helm

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/release"
)

func (t *Toolset) listTool() mcp.Tool {
	return mcp.NewTool(
		"helm_list",
		mcp.WithDescription("List Helm releases in a cluster (optionally filter by namespace)"),
		mcp.WithString("cluster", mcp.Required(), mcp.Description("Rancher cluster ID")),
		mcp.WithString("namespace", mcp.Description("Namespace (empty = all namespaces)")),
		mcp.WithString("format", mcp.Description("Output format: json, table (default: json)")),
		mcp.WithBoolean("deployed", mcp.Description("If true, only show deployed releases (default: false)")),
		mcp.WithBoolean("failed", mcp.Description("If true, only show failed releases (default: false)")),
		mcp.WithBoolean("pending", mcp.Description("If true, only show pending releases (default: false)")),
	)
}

func (t *Toolset) listHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	cluster, err := req.RequireString("cluster")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	namespace := req.GetString("namespace", "")
	format := req.GetString("format", "json")
	deployed := req.GetBool("deployed", false)
	failed := req.GetBool("failed", false)
	pending := req.GetBool("pending", false)

	// When listing all namespaces, we must iterate or use a fixed namespace for the action config
	// Helm List with AllNamespaces uses the config's namespace only for storage driver; it lists from all
	nsForConfig := namespace
	if nsForConfig == "" {
		nsForConfig = "default"
	}
	cfg, err := t.actionConfigFor(cluster, nsForConfig)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("helm action config: %v", err)), nil
	}
	listAction := action.NewList(cfg)
	listAction.AllNamespaces = (namespace == "")
	listAction.Deployed = deployed
	listAction.Failed = failed
	listAction.Pending = pending

	releases, err := listAction.Run()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("helm list: %v", err)), nil
	}
	if releases == nil {
		releases = []*release.Release{}
	}
	items := make([]map[string]interface{}, 0, len(releases))
	for _, r := range releases {
		items = append(items, map[string]interface{}{
			"name":       r.Name,
			"namespace":  r.Namespace,
			"revision":   r.Version,
			"status":     r.Info.Status.String(),
			"chart":      r.Chart.Metadata.Name + "-" + r.Chart.Metadata.Version,
			"updated":    r.Info.LastDeployed.Format("2006-01-02 15:04:05"),
			"app_version": r.Chart.Metadata.AppVersion,
		})
	}
	out, err := t.formatter.Format(items, format)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("format: %v", err)), nil
	}
	return mcp.NewToolResultText(out), nil
}
