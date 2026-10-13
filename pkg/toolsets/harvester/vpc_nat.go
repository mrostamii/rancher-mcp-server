package harvester

import (
	"context"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mrostamii/rancher-mcp-server/pkg/client/rancher"
	"github.com/mrostamii/rancher-mcp-server/pkg/formatter"
)

func (t *Toolset) vpcNatGatewayTool() mcp.Tool {
	return mcp.NewTool(
		"harvester_vpc_nat_gateway",
		mcp.WithDescription("Manage Kube-OVN VPC NAT gateways and their EIP, DNAT, and SNAT rules (cluster-scoped, kubeovn.io/v1). kind is gateway, eip, dnat, or snat. Create requires read-only=false. Delete requires destructive operations."),
		mcp.WithString("cluster", mcp.Required(), mcp.Description("Harvester cluster ID")),
		mcp.WithString("action", mcp.Required(), mcp.Description("Action: create, list, delete")),
		mcp.WithString("kind", mcp.Description("Object kind: gateway (default), eip, dnat, snat")),
		mcp.WithString("name", mcp.Description("Object name (required for create and delete)")),
		mcp.WithString("vpc", mcp.Description("VPC name (gateway create)")),
		mcp.WithString("subnet", mcp.Description("Subnet name for the gateway LAN IP (gateway create)")),
		mcp.WithString("lan_ip", mcp.Description("Gateway LAN IP (gateway create, optional)")),
		mcp.WithString("external_subnets", mcp.Description("Comma-separated external subnet names (gateway create)")),
		mcp.WithString("nat_gw", mcp.Description("NAT gateway name (spec.natGwDp) for EIP create")),
		mcp.WithString("eip", mcp.Description("EIP name (DNAT and SNAT create)")),
		mcp.WithString("external_port", mcp.Description("External port (DNAT create)")),
		mcp.WithString("internal_ip", mcp.Description("Internal IP (DNAT create)")),
		mcp.WithString("internal_port", mcp.Description("Internal port (DNAT create)")),
		mcp.WithString("protocol", mcp.Description("DNAT protocol: tcp (default) or udp")),
		mcp.WithString("internal_cidr", mcp.Description("Internal CIDR (SNAT create)")),
		mcp.WithString("format", mcp.Description("Output format for list: json, table (default: json)")),
		mcp.WithNumber("limit", mcp.Description("Max items for list (default: 100)")),
	)
}

func (t *Toolset) vpcNatGatewayHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	cluster := req.GetString("cluster", "")
	action := req.GetString("action", "")
	kind := req.GetString("kind", "gateway")
	if kind == "" {
		kind = "gateway"
	}
	name := strings.TrimSpace(req.GetString("name", ""))
	resourceType, apiKind, err := natKind(kind)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	switch action {
	case "list":
		format := req.GetString("format", "json")
		limit := req.GetInt("limit", 100)
		if limit <= 0 {
			limit = 100
		}
		col, err := t.client.List(ctx, cluster, resourceType, rancher.ListOpts{Limit: limit})
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("harvester_vpc_nat_gateway list: %v", err)), nil
		}
		items := make([]map[string]interface{}, 0, len(col.Data))
		for _, r := range col.Data {
			items = append(items, map[string]interface{}{
				"name": r.ObjectMeta.Name, "kind": kind, "spec": r.Spec, "status": r.Status,
			})
		}
		out, err := formatter.FormatListWithContinue(t.formatter, items, col.Continue, format)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("format: %v", err)), nil
		}
		return mcp.NewToolResultText(out), nil

	case "create":
		if err := t.policy.CheckWrite(); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		if err := requireResourceName(kind, name); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		body, err := natCreateBody(req, kind, apiKind, name)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		if _, err := t.client.Create(ctx, cluster, resourceType, "", body); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("harvester_vpc_nat_gateway create: %v", err)), nil
		}
		return mcp.NewToolResultText(fmt.Sprintf("%s %q created", apiKind, name)), nil

	case "delete":
		if err := t.policy.CheckDestructive(); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		if err := requireResourceName(kind, name); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		if err := t.client.Delete(ctx, cluster, resourceType, "", name); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("harvester_vpc_nat_gateway delete: %v", err)), nil
		}
		return mcp.NewToolResultText(fmt.Sprintf("%s %q deleted", apiKind, name)), nil

	default:
		return mcp.NewToolResultError(fmt.Sprintf("invalid action %q; use create, list, or delete", action)), nil
	}
}

func natKind(kind string) (resourceType, apiKind string, err error) {
	switch kind {
	case "gateway":
		return rancher.TypeVpcNatGateways, "VpcNatGateway", nil
	case "eip":
		return rancher.TypeIptablesEIPs, "IptablesEIP", nil
	case "dnat":
		return rancher.TypeIptablesDnatRules, "IptablesDnatRule", nil
	case "snat":
		return rancher.TypeIptablesSnatRules, "IptablesSnatRule", nil
	default:
		return "", "", fmt.Errorf("invalid kind %q; use gateway, eip, dnat, or snat", kind)
	}
}

func natCreateBody(req mcp.CallToolRequest, kind, apiKind, name string) (map[string]interface{}, error) {
	spec := map[string]interface{}{}
	switch kind {
	case "gateway":
		vpc := strings.TrimSpace(req.GetString("vpc", ""))
		subnet := strings.TrimSpace(req.GetString("subnet", ""))
		if vpc == "" || subnet == "" {
			return nil, fmt.Errorf("gateway create requires vpc and subnet")
		}
		spec["vpc"] = vpc
		spec["subnet"] = subnet
		if lan := strings.TrimSpace(req.GetString("lan_ip", "")); lan != "" {
			spec["lanIp"] = lan
		}
		if ext := normalizeCommaList(req.GetString("external_subnets", "")); ext != "" {
			spec["externalSubnets"] = strings.Split(ext, ",")
		}
	case "eip":
		gw := strings.TrimSpace(req.GetString("nat_gw", ""))
		if gw == "" {
			return nil, fmt.Errorf("eip create requires nat_gw")
		}
		spec["natGwDp"] = gw
	case "dnat":
		eip := strings.TrimSpace(req.GetString("eip", ""))
		externalPort := strings.TrimSpace(req.GetString("external_port", ""))
		internalIP := strings.TrimSpace(req.GetString("internal_ip", ""))
		internalPort := strings.TrimSpace(req.GetString("internal_port", ""))
		protocol := strings.ToLower(strings.TrimSpace(req.GetString("protocol", "tcp")))
		if protocol == "" {
			protocol = "tcp"
		}
		if eip == "" || externalPort == "" || internalIP == "" || internalPort == "" {
			return nil, fmt.Errorf("dnat create requires eip, external_port, internal_ip, and internal_port")
		}
		if protocol != "tcp" && protocol != "udp" {
			return nil, fmt.Errorf("protocol must be tcp or udp")
		}
		spec["eip"] = eip
		spec["externalPort"] = externalPort
		spec["internalIp"] = internalIP
		spec["internalPort"] = internalPort
		spec["protocol"] = protocol
	case "snat":
		eip := strings.TrimSpace(req.GetString("eip", ""))
		cidr := strings.TrimSpace(req.GetString("internal_cidr", ""))
		if eip == "" || cidr == "" {
			return nil, fmt.Errorf("snat create requires eip and internal_cidr")
		}
		spec["eip"] = eip
		spec["internalCIDR"] = cidr
	}
	return map[string]interface{}{
		"apiVersion": "kubeovn.io/v1",
		"kind":       apiKind,
		"metadata":   map[string]interface{}{"name": name},
		"spec":       spec,
	}, nil
}
