package harvester

import (
	"context"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mrostamii/rancher-mcp-server/pkg/client/rancher"
)

func (t *Toolset) vpcCreateTool() mcp.Tool {
	return mcp.NewTool(
		"harvester_vpc_create",
		mcp.WithDescription("Create a KubeOVN VPC (requires kubeovn-operator addon enabled)"),
		mcp.WithString("cluster", mcp.Required(), mcp.Description("Harvester cluster ID")),
		mcp.WithString("name", mcp.Required(), mcp.Description("VPC name")),
		mcp.WithString("namespaces", mcp.Description("Comma-separated list of namespaces that can use this VPC (empty = all namespaces)")),
	)
}

func (t *Toolset) vpcCreateHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if err := t.policy.CheckWrite(); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	cluster := req.GetString("cluster", "")
	name, err := req.RequireString("name")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	namespacesStr := req.GetString("namespaces", "")

	spec := map[string]interface{}{}
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

	body := map[string]interface{}{
		"apiVersion": "kubeovn.io/v1",
		"kind":       "Vpc",
		"metadata": map[string]interface{}{
			"name": name,
		},
		"spec": spec,
	}

	// VPC is cluster-scoped, so namespace is empty
	_, err = t.client.Create(ctx, cluster, rancher.TypeVpcs, "", body)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("harvester_vpc_create: %v", err)), nil
	}
	return mcp.NewToolResultText(fmt.Sprintf("VPC %q created", name)), nil
}
