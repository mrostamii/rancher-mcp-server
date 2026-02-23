package kubernetes

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mrostamii/rancher-mcp-server/pkg/client/rancher"
)

func (t *Toolset) listTool() mcp.Tool {
	return mcp.NewTool(
		"kubernetes_list",
		mcp.WithDescription("List Kubernetes resources by apiVersion and kind (e.g. v1 Pod, apps/v1 Deployment)"),
		mcp.WithString("cluster", mcp.Required(), mcp.Description("Cluster ID")),
		mcp.WithString("api_version", mcp.Required(), mcp.Description("apiVersion (e.g. v1, apps/v1)")),
		mcp.WithString("kind", mcp.Required(), mcp.Description("Kind (e.g. Pod, Deployment)")),
		mcp.WithString("namespace", mcp.Description("Namespace (empty = all for namespaced; omit for cluster-scoped)")),
		mcp.WithString("label_selector", mcp.Description("Label selector")),
		mcp.WithString("format", mcp.Description("Output format: json, table (default: json)")),
		mcp.WithNumber("limit", mcp.Description("Max items (default: 100)")),
	)
}

func (t *Toolset) listHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	cluster, err := req.RequireString("cluster")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	apiVersion, err := req.RequireString("api_version")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	kind, err := req.RequireString("kind")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	resourceType := rancher.SteveType(apiVersion, kind)
	namespace := req.GetString("namespace", "")
	labelSelector := req.GetString("label_selector", "")
	format := req.GetString("format", "json")
	limit := req.GetInt("limit", 100)
	if limit <= 0 {
		limit = 100
	}

	opts := rancher.ListOpts{Limit: limit, LabelSelector: labelSelector}
	if namespace != "" {
		opts.Namespace = namespace
	}
	col, err := t.client.List(ctx, cluster, resourceType, opts)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to list %s: %v", kind, err)), nil
	}
	items := make([]map[string]interface{}, 0, len(col.Data))
	for _, r := range col.Data {
		items = append(items, map[string]interface{}{
			"name":      r.ObjectMeta.Name,
			"namespace": r.ObjectMeta.Namespace,
			"metadata":  r.ObjectMeta,
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
