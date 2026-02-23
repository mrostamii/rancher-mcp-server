package kubernetes

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mrostamii/rancher-mcp-server/pkg/client/rancher"
)

func (t *Toolset) capacityTool() mcp.Tool {
	return mcp.NewTool(
		"kubernetes_capacity",
		mcp.WithDescription("Summarize cluster capacity and allocatable resources from Nodes"),
		mcp.WithString("cluster", mcp.Required(), mcp.Description("Cluster ID")),
		mcp.WithString("format", mcp.Description("Output format: json, table (default: json)")),
	)
}

func (t *Toolset) capacityHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	cluster, err := req.RequireString("cluster")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	format := req.GetString("format", "json")

	col, err := t.client.List(ctx, cluster, rancher.TypeNodes, rancher.ListOpts{Limit: 500})
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("list nodes: %v", err)), nil
	}

	nodes := make([]map[string]interface{}, 0, len(col.Data))
	for _, n := range col.Data {
		nodes = append(nodes, map[string]interface{}{
			"name":        n.ObjectMeta.Name,
			"capacity":    getNodeCapacity(n),
			"allocatable": getNodeAllocatable(n),
		})
	}

	data := map[string]interface{}{
		"node_count": len(col.Data),
		"nodes":      nodes,
	}
	out, err := t.formatter.Format(data, format)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("format: %v", err)), nil
	}
	return mcp.NewToolResultText(out), nil
}

func getNodeCapacity(n rancher.SteveResource) map[string]interface{} {
	if m, ok := n.Status.(map[string]interface{}); ok {
		if cap, ok := m["capacity"].(map[string]interface{}); ok {
			return cap
		}
	}
	return nil
}

func getNodeAllocatable(n rancher.SteveResource) map[string]interface{} {
	if m, ok := n.Status.(map[string]interface{}); ok {
		if a, ok := m["allocatable"].(map[string]interface{}); ok {
			return a
		}
	}
	return nil
}

