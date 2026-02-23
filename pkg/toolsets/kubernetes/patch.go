package kubernetes

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mrostamii/rancher-mcp-server/pkg/client/rancher"
)

// deepMerge merges patch into base recursively. Map values are merged; other values overwrite.
func deepMerge(base, patch map[string]interface{}) {
	for k, pv := range patch {
		if pv == nil {
			base[k] = pv
			continue
		}
		pm, ok := pv.(map[string]interface{})
		if !ok {
			base[k] = pv
			continue
		}
		bm, ok := base[k].(map[string]interface{})
		if !ok {
			base[k] = pv
			continue
		}
		deepMerge(bm, pm)
	}
}

func (t *Toolset) patchTool() mcp.Tool {
	return mcp.NewTool(
		"kubernetes_patch",
		mcp.WithDescription("Patch (update) a Kubernetes resource with JSON merge patch"),
		mcp.WithString("cluster", mcp.Required(), mcp.Description("Cluster ID")),
		mcp.WithString("api_version", mcp.Required(), mcp.Description("apiVersion (e.g. v1, apps/v1)")),
		mcp.WithString("kind", mcp.Required(), mcp.Description("Kind (e.g. Pod, Deployment)")),
		mcp.WithString("namespace", mcp.Description("Namespace (for namespaced resources)")),
		mcp.WithString("name", mcp.Required(), mcp.Description("Resource name")),
		mcp.WithString("patch", mcp.Required(), mcp.Description("JSON merge patch body")),
		mcp.WithString("format", mcp.Description("Output format: json, table (default: json)")),
	)
}

func (t *Toolset) patchHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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
	patchStr, err := req.RequireString("patch")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	namespace := req.GetString("namespace", "")
	format := req.GetString("format", "json")

	var patch map[string]interface{}
	if err := json.Unmarshal([]byte(patchStr), &patch); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("invalid JSON patch: %v", err)), nil
	}
	resourceType := rancher.SteveType(apiVersion, kind)
	// Steve API expects full resource body for PUT; we do a get then merge patch then put
	existing, err := t.client.Get(ctx, cluster, resourceType, namespace, name)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("resource not found: %v", err)), nil
	}
	merged := map[string]interface{}{
		"metadata": existing.ObjectMeta,
		"spec":     existing.Spec,
		"status":   existing.Status,
	}
	if existing.TypeMeta.Kind != "" {
		merged["kind"] = existing.TypeMeta.Kind
	}
	if existing.TypeMeta.APIVersion != "" {
		merged["apiVersion"] = existing.TypeMeta.APIVersion
	}
	deepMerge(merged, patch)
	res, err := t.client.Update(ctx, cluster, resourceType, namespace, name, merged)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("kubernetes_patch: %v", err)), nil
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
