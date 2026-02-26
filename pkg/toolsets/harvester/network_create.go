package harvester

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mrostamii/rancher-mcp-server/pkg/client/rancher"
)

func (t *Toolset) networkCreateTool() mcp.Tool {
	return mcp.NewTool(
		"harvester_network_create",
		mcp.WithDescription("Create a VM network (NetworkAttachmentDefinition). For KubeOVN overlay, creates NAD with type kube-ovn. Use harvester_subnet_create to add a Subnet with this network as provider."),
		mcp.WithString("cluster", mcp.Required(), mcp.Description("Harvester cluster ID")),
		mcp.WithString("namespace", mcp.Required(), mcp.Description("Namespace for the network")),
		mcp.WithString("name", mcp.Required(), mcp.Description("Network name")),
		mcp.WithString("type", mcp.Description("Network type: kubeovn (default, KubeOVN overlay) or vlan")),
		mcp.WithString("vlan_id", mcp.Description("VLAN ID (required when type=vlan)")),
		mcp.WithString("config", mcp.Description("Raw CNI config JSON (overrides type/vlan when set; advanced use)")),
	)
}

func (t *Toolset) networkCreateHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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
	netType := req.GetString("type", "kubeovn")
	vlanID := req.GetString("vlan_id", "")
	configStr := req.GetString("config", "")

	var config string
	if configStr != "" {
		config = configStr
	} else if netType == "vlan" {
		if vlanID == "" {
			return mcp.NewToolResultError("vlan_id is required when type=vlan"), nil
		}
		// Harvester VLAN: bridge CNI with vlan (bridge name may vary; use mgmt-br as common default)
		vlanNum, _ := strconv.Atoi(vlanID)
		if vlanNum == 0 && vlanID != "0" {
			return mcp.NewToolResultError("vlan_id must be a valid number"), nil
		}
		cfg := map[string]interface{}{
			"cniVersion":  "0.3.1",
			"name":        name,
			"type":        "bridge",
			"bridge":      "mgmt-br",
			"promiscMode": true,
			"vlan":        vlanNum,
		}
		b, _ := json.Marshal(cfg)
		config = string(b)
	} else {
		// KubeOVN overlay: provider = name.namespace.ovn
		provider := fmt.Sprintf("%s.%s.ovn", name, namespace)
		cfg := map[string]interface{}{
			"cniVersion":     "0.3.0",
			"type":           "kube-ovn",
			"server_socket":  "/run/openvswitch/kube-ovn-daemon.sock",
			"provider":       provider,
		}
		b, _ := json.Marshal(cfg)
		config = string(b)
	}

	body := map[string]interface{}{
		"apiVersion": "k8s.cni.cncf.io/v1",
		"kind":       "NetworkAttachmentDefinition",
		"metadata": map[string]interface{}{
			"name":      name,
			"namespace": namespace,
		},
		"spec": map[string]interface{}{
			"config": config,
		},
	}

	_, err = t.client.Create(ctx, cluster, rancher.TypeNetworkAttachmentDefinition, namespace, body)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("harvester_network_create: %v", err)), nil
	}
	return mcp.NewToolResultText(fmt.Sprintf("Network %q created in namespace %s", name, namespace)), nil
}
