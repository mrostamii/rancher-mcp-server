package kubernetes

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mrostamii/rancher-mcp-server/pkg/client/rancher"
)

func (t *Toolset) createTool() mcp.Tool {
	return mcp.NewTool(
		"kubernetes_create",
		mcp.WithDescription("Create a Kubernetes resource from JSON or YAML (must include apiVersion, kind, metadata)"),
		mcp.WithString("cluster", mcp.Required(), mcp.Description("Cluster ID")),
		mcp.WithString("resource", mcp.Required(), mcp.Description("JSON or YAML body of the resource")),
		mcp.WithString("format", mcp.Description("Output format: json, table (default: json)")),
	)
}

func (t *Toolset) createHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	cluster, err := req.RequireString("cluster")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	bodyStr, err := req.RequireString("resource")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	format := req.GetString("format", "json")

	var body map[string]interface{}
	if err := json.Unmarshal([]byte(bodyStr), &body); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("invalid JSON resource: %v", err)), nil
	}
	apiVersion, _ := body["apiVersion"].(string)
	kind, _ := body["kind"].(string)
	if apiVersion == "" || kind == "" {
		return mcp.NewToolResultError("resource must include apiVersion and kind"), nil
	}
	meta, _ := body["metadata"].(map[string]interface{})
	namespace := ""
	if meta != nil {
		if ns, ok := meta["namespace"].(string); ok {
			namespace = ns
		}
	}
	resourceType := rancher.SteveType(apiVersion, kind)
	res, err := t.client.Create(ctx, cluster, resourceType, namespace, body)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("kubernetes_create: %v", err)), nil
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
