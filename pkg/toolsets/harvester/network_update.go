package harvester

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mrostamii/rancher-mcp-server/pkg/client/rancher"
)

func (t *Toolset) networkUpdateTool() mcp.Tool {
	return mcp.NewTool(
		"harvester_network_update",
		mcp.WithDescription("Update a VM network (NetworkAttachmentDefinition). Can update spec.config or metadata."),
		mcp.WithString("cluster", mcp.Required(), mcp.Description("Harvester cluster ID")),
		mcp.WithString("namespace", mcp.Required(), mcp.Description("Namespace")),
		mcp.WithString("name", mcp.Required(), mcp.Description("Network name")),
		mcp.WithString("config", mcp.Description("New CNI config JSON (replaces existing)")),
	)
}

func (t *Toolset) networkUpdateHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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
	config := req.GetString("config", "")

	if config == "" {
		return mcp.NewToolResultError("config is required for update"), nil
	}

	patch := map[string]interface{}{
		"spec": map[string]interface{}{
			"config": config,
		},
	}

	_, err = t.client.Patch(ctx, cluster, rancher.TypeNetworkAttachmentDefinition, namespace, name, patch)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("harvester_network_update: %v", err)), nil
	}
	return mcp.NewToolResultText(fmt.Sprintf("Network %q updated", name)), nil
}
