package kubernetes

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mrostamii/rancher-mcp-server/pkg/client/rancher"
)

func (t *Toolset) deleteTool() mcp.Tool {
	return mcp.NewTool(
		"kubernetes_delete",
		mcp.WithDescription("Delete a Kubernetes resource by apiVersion, kind, namespace, name"),
		mcp.WithString("cluster", mcp.Required(), mcp.Description("Cluster ID")),
		mcp.WithString("api_version", mcp.Required(), mcp.Description("apiVersion (e.g. v1, apps/v1)")),
		mcp.WithString("kind", mcp.Required(), mcp.Description("Kind (e.g. Pod, Deployment)")),
		mcp.WithString("namespace", mcp.Description("Namespace (for namespaced resources)")),
		mcp.WithString("name", mcp.Required(), mcp.Description("Resource name")),
	)
}

func (t *Toolset) deleteHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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

	resourceType := rancher.SteveType(apiVersion, kind)
	if err := t.client.Delete(ctx, cluster, resourceType, namespace, name); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("kubernetes_delete: %v", err)), nil
	}
	return mcp.NewToolResultText(fmt.Sprintf("Deleted %s %q", kind, name)), nil
}
