package kubernetes

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mrostamii/rancher-mcp-server/pkg/client/rancher"
)

func (t *Toolset) getTool() mcp.Tool {
	return mcp.NewTool(
		"kubernetes_get",
		mcp.WithDescription("Get a single Kubernetes resource by apiVersion, kind, namespace, name"),
		mcp.WithString("cluster", mcp.Required(), mcp.Description("Cluster ID")),
		mcp.WithString("api_version", mcp.Required(), mcp.Description("apiVersion (e.g. v1, apps/v1)")),
		mcp.WithString("kind", mcp.Required(), mcp.Description("Kind (e.g. Pod, Deployment)")),
		mcp.WithString("namespace", mcp.Description("Namespace (required for namespaced resources)")),
		mcp.WithString("name", mcp.Required(), mcp.Description("Resource name")),
		mcp.WithString("format", mcp.Description("Output format: json, table (default: json)")),
	)
}

func (t *Toolset) getHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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
	name, err := req.RequireString("name")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	namespace := req.GetString("namespace", "")
	format := req.GetString("format", "json")

	resourceType := rancher.SteveType(apiVersion, kind)
	res, err := t.client.Get(ctx, cluster, resourceType, namespace, name)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("%s %q not found: %v", kind, name, err)), nil
	}
	data := map[string]interface{}{
		"metadata": res.ObjectMeta,
		"spec":     res.Spec,
		"status":   res.Status,
	}
	out, err := t.formatter.Format(data, format)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("format: %v", err)), nil
	}
	return mcp.NewToolResultText(out), nil
}
