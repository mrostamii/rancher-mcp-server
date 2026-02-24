package harvester

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mrostamii/rancher-mcp-server/pkg/client/rancher"
)

func (t *Toolset) vmGetTool() mcp.Tool {
	return mcp.NewTool(
		"harvester_vm_get",
		mcp.WithDescription("Get a single virtual machine with full spec and status (metrics, disks, networks)"),
		mcp.WithString("cluster", mcp.Required(), mcp.Description("Harvester cluster ID")),
		mcp.WithString("namespace", mcp.Required(), mcp.Description("Namespace")),
		mcp.WithString("name", mcp.Required(), mcp.Description("VM name")),
		mcp.WithString("format", mcp.Description("Output format: json, table (default: json)")),
	)
}

func (t *Toolset) vmGetHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	cluster := req.GetString("cluster", "")
	namespace, err := req.RequireString("namespace")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	name, err := req.RequireString("name")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	format := req.GetString("format", "json")

	res, err := t.client.Get(ctx, cluster, rancher.TypeVirtualMachines, namespace, name)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("VM %q not found in namespace %q: %v", name, namespace, err)), nil
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
