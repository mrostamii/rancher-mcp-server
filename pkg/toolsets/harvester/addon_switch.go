package harvester

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mrostamii/rancher-mcp-server/pkg/client/rancher"
)

func (t *Toolset) addonSwitchTool() mcp.Tool {
	return mcp.NewTool(
		"harvester_addon_switch",
		mcp.WithDescription("Enable or disable a Harvester addon by setting spec.enabled"),
		mcp.WithString("cluster", mcp.Required(), mcp.Description("Harvester cluster ID")),
		mcp.WithString("namespace", mcp.Required(), mcp.Description("Addon namespace (e.g. harvester-system, kube-system, cattle-logging-system)")),
		mcp.WithString("name", mcp.Required(), mcp.Description("Addon name (e.g. kubeovn-operator, rancher-logging)")),
		mcp.WithString("enabled", mcp.Required(), mcp.Description("Enable or disable: true or false")),
	)
}

func (t *Toolset) addonSwitchHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if err := t.policy.CheckWrite(); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	cluster := req.GetString("cluster", "")
	namespace, err := req.RequireString("namespace")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	name, err := req.RequireString("name")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	enabledStr, err := req.RequireString("enabled")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	var enabled bool
	switch enabledStr {
	case "true", "1":
		enabled = true
	case "false", "0":
		enabled = false
	default:
		return mcp.NewToolResultError("enabled must be true or false"), nil
	}

	patch := map[string]interface{}{
		"spec": map[string]interface{}{
			"enabled": enabled,
		},
	}
	_, err = t.client.Patch(ctx, cluster, rancher.TypeAddons, namespace, name, patch)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("harvester_addon_switch: %v", err)), nil
	}

	action := "disabled"
	if enabled {
		action = "enabled"
	}
	return mcp.NewToolResultText(fmt.Sprintf("Addon %q %s in namespace %s", name, action, namespace)), nil
}
