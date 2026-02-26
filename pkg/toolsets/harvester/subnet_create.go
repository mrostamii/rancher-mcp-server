package harvester

import (
	"context"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mrostamii/rancher-mcp-server/pkg/client/rancher"
)

func (t *Toolset) subnetCreateTool() mcp.Tool {
	return mcp.NewTool(
		"harvester_subnet_create",
		mcp.WithDescription("Create a KubeOVN Subnet for VM networks (requires kubeovn-operator). DHCP, DNS, and NAT are enabled by default. Create the Network (NAD) first with harvester_network_create, then create Subnet with provider = {network}.{namespace}.ovn"),
		mcp.WithString("cluster", mcp.Required(), mcp.Description("Harvester cluster ID")),
		mcp.WithString("name", mcp.Required(), mcp.Description("Subnet name")),
		mcp.WithString("cidr_block", mcp.Required(), mcp.Description("CIDR e.g. 10.20.0.0/24")),
		mcp.WithString("gateway", mcp.Description("Gateway IP (default: first IP in CIDR)")),
		mcp.WithString("vpc", mcp.Description("VPC name (default: ovn-cluster)")),
		mcp.WithString("provider", mcp.Required(), mcp.Description("Provider = {network-name}.{namespace}.ovn (links to NetworkAttachmentDefinition)")),
		mcp.WithString("gateway_type", mcp.Description("distributed or centralized (default: distributed)")),
		mcp.WithString("nat_outgoing", mcp.Description("Outbound NAT (default: true; VM subnets need this for external access)")),
		mcp.WithString("enable_dhcp", mcp.Description("Enable DHCP for VMs (default: true; VM subnets need this)")),
		mcp.WithString("dns_servers", mcp.Description("Comma-separated DNS servers for DHCP (default: 8.8.8.8)")),
		mcp.WithString("namespaces", mcp.Description("Comma-separated namespaces that can use this subnet (empty = all)")),
	)
}

func (t *Toolset) subnetCreateHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if err := t.policy.CheckWrite(); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	cluster := req.GetString("cluster", "")
	name, err := req.RequireString("name")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	cidrBlock, err := req.RequireString("cidr_block")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	provider, err := req.RequireString("provider")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	gateway := req.GetString("gateway", "")
	vpc := req.GetString("vpc", "ovn-cluster")
	gatewayType := req.GetString("gateway_type", "distributed")
	natOutgoingStr := req.GetString("nat_outgoing", "true")
	enableDHCPStr := req.GetString("enable_dhcp", "true")
	dnsServers := req.GetString("dns_servers", "8.8.8.8")
	namespacesStr := req.GetString("namespaces", "")

	natOutgoing := natOutgoingStr == "true" || natOutgoingStr == "1"
	enableDHCP := enableDHCPStr == "true" || enableDHCPStr == "1"

	spec := map[string]interface{}{
		"protocol":    "IPv4",
		"cidrBlock":   cidrBlock,
		"provider":    provider,
		"vpc":         vpc,
		"gatewayType": gatewayType,
		"natOutgoing": natOutgoing,
		"enableDHCP":  enableDHCP,
		"default":     false,
	}
	if gateway != "" {
		spec["gateway"] = gateway
	}
	if enableDHCP && dnsServers != "" {
		// OVN dhcpV4Options format: dns_server="{8.8.8.8}" or dns_server="{8.8.8.8,8.8.4.4}"
		parts := strings.Split(dnsServers, ",")
		trimmed := make([]string, 0, len(parts))
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if p != "" {
				trimmed = append(trimmed, p)
			}
		}
		if len(trimmed) > 0 {
			spec["dhcpV4Options"] = "dns_server={" + strings.Join(trimmed, ",") + "}"
		}
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

	body := map[string]interface{}{
		"apiVersion": "kubeovn.io/v1",
		"kind":       "Subnet",
		"metadata": map[string]interface{}{
			"name": name,
		},
		"spec": spec,
	}

	_, err = t.client.Create(ctx, cluster, rancher.TypeSubnets, "", body)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("harvester_subnet_create: %v", err)), nil
	}
	return mcp.NewToolResultText(fmt.Sprintf("Subnet %q created", name)), nil
}
