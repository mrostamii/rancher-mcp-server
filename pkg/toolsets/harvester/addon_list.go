package harvester

import (
	"context"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mrostamii/rancher-mcp-server/pkg/client/rancher"
	"github.com/mrostamii/rancher-mcp-server/pkg/formatter"
)

// Common namespaces where Harvester addons are deployed.
var addonNamespaces = []string{"harvester-system", "cattle-logging-system", "cattle-monitoring-system", "kube-system"}

func (t *Toolset) addonListTool() mcp.Tool {
	return mcp.NewTool(
		"harvester_addon_list",
		mcp.WithDescription("List Harvester addons with their enabled/disabled state"),
		mcp.WithString("cluster", mcp.Required(), mcp.Description("Harvester cluster ID")),
		mcp.WithString("namespace", mcp.Description("Namespace (empty = all common addon namespaces: harvester-system, cattle-logging-system, cattle-monitoring-system, kube-system)")),
		mcp.WithString("format", mcp.Description("Output format: json, table (default: json)")),
		mcp.WithNumber("limit", mcp.Description("Max items per namespace (default: 100)")),
		mcp.WithString("continue", mcp.Description("Pagination token from previous response (for next page; only when namespace is specified)")),
	)
}

func (t *Toolset) addonListHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	cluster := req.GetString("cluster", "")
	namespace := req.GetString("namespace", "")
	format := req.GetString("format", "json")
	limit := req.GetInt("limit", 100)
	if limit <= 0 {
		limit = 100
	}

	namespaces := addonNamespaces
	if namespace != "" {
		namespaces = []string{namespace}
	}
	continueToken := req.GetString("continue", "")

	items := make([]map[string]interface{}, 0)
	var paginationContinue string
	for _, ns := range namespaces {
		opts := rancher.ListOpts{Namespace: ns, Limit: limit}
		if namespace != "" && continueToken != "" {
			opts.Continue = continueToken
		}
		col, err := t.client.List(ctx, cluster, rancher.TypeAddons, opts)
		if err != nil {
			// Skip namespaces that don't exist or have no addons (404/403)
			if strings.Contains(err.Error(), "404") || strings.Contains(err.Error(), "403") || strings.Contains(err.Error(), "not found") {
				continue
			}
			return mcp.NewToolResultError(fmt.Sprintf("failed to list addons in %s: %v", ns, err)), nil
		}
		if namespace != "" {
			paginationContinue = col.Continue
		}
		for _, r := range col.Data {
			enabled := false
			if spec, ok := r.Spec.(map[string]interface{}); ok {
				if e, ok := spec["enabled"].(bool); ok {
					enabled = e
				}
			}
			state := "Disabled"
			if enabled {
				state = "DeploySuccessful"
				if status, ok := r.Status.(map[string]interface{}); ok {
					if s, ok := status["status"].(string); ok && s != "" {
						state = s
					}
				}
			}
			items = append(items, map[string]interface{}{
				"name":        r.ObjectMeta.Name,
				"namespace":   r.ObjectMeta.Namespace,
				"enabled":     enabled,
				"state":       state,
				"labels":      r.ObjectMeta.Labels,
				"description": getAddonDescription(r.ObjectMeta.Name),
			})
		}
	}

	out, err := formatter.FormatListWithContinue(t.formatter, items, paginationContinue, format)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("format: %v", err)), nil
	}
	return mcp.NewToolResultText(out), nil
}

// getAddonDescription returns a short description for known addons.
func getAddonDescription(name string) string {
	descriptions := map[string]string{
		"harvester-csi-driver-lvm":           "Create PVC through LVM with local devices",
		"harvester-seeder":                   "IPMI/Redfish hardware discovery and out-of-band operations",
		"kubeovn-operator":                   "KubeOVN networking addon",
		"nvidia-driver-toolkit":              "vGPU devices for Harvester VMs",
		"pcidevices-controller":              "Discover PCI devices for PCI passthrough",
		"rancher-logging":                    "Collect logs, events, and audits",
		"rancher-monitoring":                 "Collect cluster and VM metrics, dashboards, alerts",
		"virtual-machine-auto-balance":       "Evict suboptimal VM placements (descheduler)",
		"descheduler":                        "Evict suboptimal VM placements (virtual-machine-auto-balance)",
		"vm-import-controller":               "Migrate VMs from other clusters to Harvester",
		"harvester-vm-dhcp-controller":       "Managed DHCP for VMs",
		"rancher-vcluster":                   "Rancher vCluster integration",
	}
	if d, ok := descriptions[name]; ok {
		return d
	}
	return ""
}
