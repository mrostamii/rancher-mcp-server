package harvester

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mrostamii/rancher-mcp-server/pkg/client/rancher"
)

var vmActions = map[string]bool{"start": true, "stop": true, "restart": true, "pause": true, "unpause": true, "migrate": true}

func (t *Toolset) vmActionTool() mcp.Tool {
	return mcp.NewTool(
		"harvester_vm_action",
		mcp.WithDescription("Run lifecycle action on a VM: start, stop, restart, pause, unpause, migrate"),
		mcp.WithString("cluster", mcp.Description("Harvester cluster ID (optional if default_harvester_cluster is set in config)")),
		mcp.WithString("namespace", mcp.Required(), mcp.Description("Namespace")),
		mcp.WithString("name", mcp.Required(), mcp.Description("VM name")),
		mcp.WithString("action", mcp.Required(), mcp.Description("Action: start, stop, restart, pause, unpause, migrate")),
	)
}

func (t *Toolset) vmActionHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	cluster := t.cluster(req)
	if cluster == "" {
		return mcp.NewToolResultError("cluster is required (or set default_harvester_cluster in config)"), nil
	}
	namespace, err := req.RequireString("namespace")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	name, err := req.RequireString("name")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	action, err := req.RequireString("action")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	if !vmActions[action] {
		return mcp.NewToolResultError(fmt.Sprintf("invalid action %q; allowed: start, stop, restart, pause, unpause, migrate", action)), nil
	}

	err = t.client.Action(ctx, cluster, rancher.TypeVirtualMachines, namespace, name, action, nil)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("harvester_vm_action: %v", err)), nil
	}
	return mcp.NewToolResultText(fmt.Sprintf("VM %q action %q completed", name, action)), nil
}
