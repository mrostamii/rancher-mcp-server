package kubernetes

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mrostamii/rancher-mcp-server/pkg/client/rancher"
)

func (t *Toolset) describeTool() mcp.Tool {
	return mcp.NewTool(
		"kubernetes_describe",
		mcp.WithDescription("Get a Kubernetes resource and its recent events (describe-style output)"),
		mcp.WithString("cluster", mcp.Required(), mcp.Description("Cluster ID")),
		mcp.WithString("api_version", mcp.Required(), mcp.Description("apiVersion (e.g. v1, apps/v1)")),
		mcp.WithString("kind", mcp.Required(), mcp.Description("Kind (e.g. Pod, Deployment)")),
		mcp.WithString("namespace", mcp.Description("Namespace (for namespaced resources)")),
		mcp.WithString("name", mcp.Required(), mcp.Description("Resource name")),
		mcp.WithString("format", mcp.Description("Output format: json, table (default: json)")),
		mcp.WithNumber("events_limit", mcp.Description("Max events to include (default: 20)")),
	)
}

func (t *Toolset) describeHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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
	eventsLimit := req.GetInt("events_limit", 20)
	if eventsLimit <= 0 {
		eventsLimit = 20
	}

	resourceType := rancher.SteveType(apiVersion, kind)
	res, err := t.client.Get(ctx, cluster, resourceType, namespace, name)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("%s %q not found: %v", kind, name, err)), nil
	}

	// Fetch events for this resource (events are namespaced; for cluster-scoped resources use default namespace or skip)
	eventsList := []map[string]interface{}{}
	if namespace != "" {
		fieldSel := fmt.Sprintf("involvedObject.name=%s,involvedObject.namespace=%s", name, namespace)
		opts := rancher.ListOpts{Namespace: namespace, FieldSelector: fieldSel, Limit: eventsLimit}
		col, err := t.client.List(ctx, cluster, rancher.TypeEvents, opts)
		if err == nil {
			for _, e := range col.Data {
				eventsList = append(eventsList, map[string]interface{}{
					"metadata":   e.ObjectMeta,
					"reason":     getEventReason(e),
					"message":    getEventMessage(e),
					"count":      getEventCount(e),
					"lastTimestamp": getEventLastTimestamp(e),
				})
			}
		}
	}

	data := map[string]interface{}{
		"resource": map[string]interface{}{
			"metadata": res.ObjectMeta,
			"spec":     res.Spec,
			"status":   res.Status,
		},
		"events": eventsList,
	}
	out, err := t.formatter.Format(data, format)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("format: %v", err)), nil
	}
	return mcp.NewToolResultText(out), nil
}

func getEventReason(e rancher.SteveResource) string {
	if m, ok := e.Status.(map[string]interface{}); ok {
		if r, ok := m["reason"].(string); ok {
			return r
		}
	}
	return ""
}

func getEventMessage(e rancher.SteveResource) string {
	if m, ok := e.Status.(map[string]interface{}); ok {
		if msg, ok := m["message"].(string); ok {
			return msg
		}
	}
	return ""
}

func getEventCount(e rancher.SteveResource) int64 {
	if m, ok := e.Status.(map[string]interface{}); ok {
		if n, ok := m["count"].(float64); ok {
			return int64(n)
		}
	}
	return 0
}

func getEventLastTimestamp(e rancher.SteveResource) string {
	if m, ok := e.Status.(map[string]interface{}); ok {
		if t, ok := m["lastTimestamp"].(string); ok {
			return t
		}
	}
	return ""
}
