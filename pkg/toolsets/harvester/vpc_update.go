package harvester

import (
	"context"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mrostamii/rancher-mcp-server/pkg/client/rancher"
)

func (t *Toolset) vpcUpdateTool() mcp.Tool {
	return mcp.NewTool(
		"harvester_vpc_update",
		mcp.WithDescription("Update a KubeOVN VPC's namespaces (requires kubeovn-operator addon enabled)"),
		mcp.WithString("cluster", mcp.Required(), mcp.Description("Harvester cluster ID")),
		mcp.WithString("name", mcp.Required(), mcp.Description("VPC name")),
		mcp.WithString("namespaces", mcp.Description("Comma-separated list of namespaces that can use this VPC (empty = all namespaces)")),
	)
}

func (t *Toolset) vpcUpdateHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if err := t.policy.CheckWrite(); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	cluster := req.GetString("cluster", "")
	name, err := req.RequireString("name")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	namespacesStr := req.GetString("namespaces", "")

	var ns []string
	if namespacesStr != "" {
		parts := strings.Split(namespacesStr, ",")
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if p != "" {
				ns = append(ns, p)
			}
		}
	}

	patch := map[string]interface{}{
		"spec": map[string]interface{}{
			"namespaces": ns,
		},
	}

	// VPC is cluster-scoped, so namespace is empty
	_, err = t.client.Patch(ctx, cluster, rancher.TypeVpcs, "", name, patch)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("harvester_vpc_update: %v", err)), nil
	}
	return mcp.NewToolResultText(fmt.Sprintf("VPC %q updated", name)), nil
}
