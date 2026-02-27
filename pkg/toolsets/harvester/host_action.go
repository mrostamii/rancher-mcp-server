package harvester

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mrostamii/rancher-mcp-server/pkg/client/rancher"
)

func (t *Toolset) hostActionTool() mcp.Tool {
	return mcp.NewTool(
		"harvester_host_action",
		mcp.WithDescription("Enable or disable maintenance mode on a Harvester host (cordon/uncordon). Requires read_only=false."),
		mcp.WithString("cluster", mcp.Required(), mcp.Description("Harvester cluster ID")),
		mcp.WithString("host", mcp.Required(), mcp.Description("Host (node) name")),
		mcp.WithString("action", mcp.Required(), mcp.Description("Action: enable_maintenance (cordon) or disable_maintenance (uncordon)")),
	)
}

func (t *Toolset) hostActionHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if err := t.policy.CheckWrite(); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	cluster, err := req.RequireString("cluster")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	hostName, err := req.RequireString("host")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	action, err := req.RequireString("action")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	var unschedulable bool
	switch action {
	case "enable_maintenance", "cordon":
		unschedulable = true
	case "disable_maintenance", "uncordon":
		unschedulable = false
	default:
		return mcp.NewToolResultError(`invalid action; allowed: enable_maintenance, disable_maintenance (or cordon, uncordon)`), nil
	}

	// Node is cluster-scoped (empty namespace)
	patch := map[string]interface{}{
		"spec": map[string]interface{}{
			"unschedulable": unschedulable,
		},
	}
	_, err = t.client.Patch(ctx, cluster, rancher.TypeNodes, "", hostName, patch)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("harvester_host_action: %v", err)), nil
	}
	op := "enabled"
	if !unschedulable {
		op = "disabled"
	}
	return mcp.NewToolResultText(fmt.Sprintf("Maintenance mode %s on host %s", op, hostName)), nil
}
