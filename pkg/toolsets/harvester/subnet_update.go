package harvester

import (
	"context"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mrostamii/rancher-mcp-server/pkg/client/rancher"
)

func (t *Toolset) subnetUpdateTool() mcp.Tool {
	return mcp.NewTool(
		"harvester_subnet_update",
		mcp.WithDescription("Update a KubeOVN Subnet (namespaces, natOutgoing, gatewayType)"),
		mcp.WithString("cluster", mcp.Required(), mcp.Description("Harvester cluster ID")),
		mcp.WithString("name", mcp.Required(), mcp.Description("Subnet name")),
		mcp.WithString("nat_outgoing", mcp.Description("true or false for outbound NAT")),
		mcp.WithString("namespaces", mcp.Description("Comma-separated namespaces (replaces existing)")),
	)
}

func (t *Toolset) subnetUpdateHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if err := t.policy.CheckWrite(); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	cluster := req.GetString("cluster", "")
	name, err := req.RequireString("name")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	natOutgoingStr := req.GetString("nat_outgoing", "")
	namespacesStr := req.GetString("namespaces", "")

	spec := map[string]interface{}{}
	if natOutgoingStr != "" {
		spec["natOutgoing"] = natOutgoingStr == "true" || natOutgoingStr == "1"
	}
	if namespacesStr != "" {
		parts := strings.Split(namespacesStr, ",")
		ns := make([]string, 0, len(parts))
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if p != "" {
				ns = append(ns, p)
			}
		}
		spec["namespaces"] = ns
	}
	if len(spec) == 0 {
		return mcp.NewToolResultError("at least one of nat_outgoing or namespaces is required"), nil
	}

	patch := map[string]interface{}{"spec": spec}

	_, err = t.client.Patch(ctx, cluster, rancher.TypeSubnets, "", name, patch)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("harvester_subnet_update: %v", err)), nil
	}
	return mcp.NewToolResultText(fmt.Sprintf("Subnet %q updated", name)), nil
}
