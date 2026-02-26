package harvester

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mrostamii/rancher-mcp-server/pkg/client/rancher"
)

func (t *Toolset) networkDeleteTool() mcp.Tool {
	return mcp.NewTool(
		"harvester_network_delete",
		mcp.WithDescription("Delete a VM network (NetworkAttachmentDefinition). Delete any Subnets using this network as provider first."),
		mcp.WithString("cluster", mcp.Required(), mcp.Description("Harvester cluster ID")),
		mcp.WithString("namespace", mcp.Required(), mcp.Description("Namespace")),
		mcp.WithString("name", mcp.Required(), mcp.Description("Network name to delete")),
	)
}

func (t *Toolset) networkDeleteHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if err := t.policy.CheckDestructive(); err != nil {
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

	err = t.client.Delete(ctx, cluster, rancher.TypeNetworkAttachmentDefinition, namespace, name)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("harvester_network_delete: %v", err)), nil
	}
	return mcp.NewToolResultText(fmt.Sprintf("Network %q deleted", name)), nil
}
