package harvester

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mrostamii/rancher-mcp-server/pkg/client/rancher"
)

func (t *Toolset) vpcDeleteTool() mcp.Tool {
	return mcp.NewTool(
		"harvester_vpc_delete",
		mcp.WithDescription("Delete a KubeOVN VPC (requires kubeovn-operator addon enabled). Cannot delete ovn-cluster (default VPC)."),
		mcp.WithString("cluster", mcp.Required(), mcp.Description("Harvester cluster ID")),
		mcp.WithString("name", mcp.Required(), mcp.Description("VPC name to delete")),
	)
}

func (t *Toolset) vpcDeleteHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if err := t.policy.CheckDestructive(); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	cluster := req.GetString("cluster", "")
	name, err := req.RequireString("name")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	if name == "ovn-cluster" {
		return mcp.NewToolResultError("cannot delete ovn-cluster (default VPC)"), nil
	}

	// VPC is cluster-scoped, so namespace is empty
	err = t.client.Delete(ctx, cluster, rancher.TypeVpcs, "", name)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("harvester_vpc_delete: %v", err)), nil
	}
	return mcp.NewToolResultText(fmt.Sprintf("VPC %q deleted", name)), nil
}
