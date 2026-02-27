package kubernetes

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mrostamii/rancher-mcp-server/pkg/client/rancher"
	"github.com/mrostamii/rancher-mcp-server/pkg/formatter"
)

func (t *Toolset) eventsTool() mcp.Tool {
	return mcp.NewTool(
		"kubernetes_events",
		mcp.WithDescription("List Kubernetes events in a namespace (optionally filtered by involvedObject)"),
		mcp.WithString("cluster", mcp.Required(), mcp.Description("Cluster ID")),
		mcp.WithString("namespace", mcp.Required(), mcp.Description("Namespace")),
		mcp.WithString("involved_object_name", mcp.Description("Filter by involvedObject.name")),
		mcp.WithString("format", mcp.Description("Output format: json, table (default: json)")),
		mcp.WithNumber("limit", mcp.Description("Max events (default: 50)")),
		mcp.WithString("continue", mcp.Description("Pagination token from previous response (for next page)")),
	)
}

func (t *Toolset) eventsHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	cluster, err := req.RequireString("cluster")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	namespace, err := req.RequireString("namespace")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	involvedName := req.GetString("involved_object_name", "")
	format := req.GetString("format", "json")
	limit := req.GetInt("limit", 50)
	if limit <= 0 {
		limit = 50
	}
	continueToken := req.GetString("continue", "")

	opts := rancher.ListOpts{Namespace: namespace, Limit: limit, Continue: continueToken}
	if involvedName != "" {
		opts.FieldSelector = fmt.Sprintf("involvedObject.name=%s", involvedName)
	}
	col, err := t.client.List(ctx, cluster, rancher.TypeEvents, opts)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("list events: %v", err)), nil
	}
	items := make([]map[string]interface{}, 0, len(col.Data))
	for _, e := range col.Data {
		items = append(items, map[string]interface{}{
			"name":           e.ObjectMeta.Name,
			"namespace":      e.ObjectMeta.Namespace,
			"reason":         getEventReason(e),
			"message":        getEventMessage(e),
			"count":          getEventCount(e),
			"lastTimestamp":  getEventLastTimestamp(e),
			"involvedObject": getEventInvolvedObject(e),
		})
	}
	out, err := formatter.FormatListWithContinue(t.formatter, items, col.Continue, format)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("format: %v", err)), nil
	}
	return mcp.NewToolResultText(out), nil
}

func getEventInvolvedObject(e rancher.SteveResource) interface{} {
	if m, ok := e.Status.(map[string]interface{}); ok {
		return m["involvedObject"]
	}
	return nil
}
