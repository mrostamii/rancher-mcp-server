package harvester

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mrostamii/rancher-mcp-server/pkg/client/rancher"
)

func (t *Toolset) subnetDeleteTool() mcp.Tool {
	return mcp.NewTool(
		"harvester_subnet_delete",
		mcp.WithDescription("Delete a KubeOVN Subnet. Cannot delete ovn-default or join (system subnets)."),
		mcp.WithString("cluster", mcp.Required(), mcp.Description("Harvester cluster ID")),
		mcp.WithString("name", mcp.Required(), mcp.Description("Subnet name to delete")),
	)
}

func (t *Toolset) subnetDeleteHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if err := t.policy.CheckDestructive(); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	cluster := req.GetString("cluster", "")
	name, err := req.RequireString("name")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	if name == "ovn-default" || name == "join" {
		return mcp.NewToolResultError("cannot delete system subnet (ovn-default or join)"), nil
	}

	err = t.client.Delete(ctx, cluster, rancher.TypeSubnets, "", name)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("harvester_subnet_delete: %v", err)), nil
	}
	return mcp.NewToolResultText(fmt.Sprintf("Subnet %q deleted", name)), nil
}
