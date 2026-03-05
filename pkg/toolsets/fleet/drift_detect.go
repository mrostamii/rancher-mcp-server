package fleet

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mrostamii/rancher-mcp-server/pkg/client/rancher"
)

func (t *Toolset) driftDetectTool() mcp.Tool {
	return mcp.NewTool(
		"fleet_drift_detect",
		mcp.WithDescription("Report Fleet drift: BundleDeployments with Modified state (resources changed outside GitOps)"),
		mcp.WithString("namespace", mcp.Description("Namespace (default: fleet-default)")),
		mcp.WithString("gitrepo", mcp.Description("Filter by GitRepo name (optional)")),
		mcp.WithString("format", mcp.Description("Output format: json, table (default: json)")),
		mcp.WithNumber("limit", mcp.Description("Max items (default: 100)")),
	)
}

func (t *Toolset) driftDetectHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	namespace := req.GetString("namespace", "fleet-default")

	if err := t.policy.CheckNamespace(namespace); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	gitrepoFilter := req.GetString("gitrepo", "")
	format := req.GetString("format", "json")
	limit := req.GetInt("limit", 100)
	if limit <= 0 {
		limit = 100
	}

	opts := rancher.ListOpts{Namespace: namespace, Limit: limit}
	col, err := t.client.List(ctx, localCluster, rancher.TypeFleetBundleDeployments, opts)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to list BundleDeployments: %v", err)), nil
	}

	var drifted []map[string]interface{}
	var all []map[string]interface{}
	for _, r := range col.Data {
		if gitrepoFilter != "" {
			// BundleDeployment name is typically "bundle-name-cluster-name"; check labels or spec for gitrepo
			if labels := r.ObjectMeta.Labels; labels != nil {
				if gr, ok := labels["fleet.cattle.io/gitrepo-name"]; ok && gr != gitrepoFilter {
					continue
				}
			}
		}
		displayState := extractDisplayState(r.Status)
		modified := displayState == "Modified"
		item := map[string]interface{}{
			"name":          r.ObjectMeta.Name,
			"namespace":     r.ObjectMeta.Namespace,
			"displayState": displayState,
			"drifted":       modified,
			"metadata":      r.ObjectMeta,
			"spec":          r.Spec,
			"status":        r.Status,
		}
		all = append(all, item)
		if modified {
			drifted = append(drifted, item)
		}
	}

	result := map[string]interface{}{
		"driftedCount": len(drifted),
		"totalCount":   len(all),
		"drifted":      drifted,
		"all":          all,
	}
	out, err := t.formatter.Format(result, format)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("format: %v", err)), nil
	}
	return mcp.NewToolResultText(out), nil
}

// extractDisplayState reads the display state from Fleet BundleDeployment status.
// Fleet uses status.display.state or similar; we handle both map and nested structures.
func extractDisplayState(status interface{}) string {
	if status == nil {
		return ""
	}
	m, ok := toMap(status)
	if !ok {
		return ""
	}
	// Fleet v1alpha1: status.display.state
	if d, ok := m["display"].(map[string]interface{}); ok {
		if s, ok := d["state"].(string); ok {
			return s
		}
	}
	// Alternative: status.conditions with type Ready
	if conds, ok := m["conditions"].([]interface{}); ok {
		for _, c := range conds {
			cm, ok := c.(map[string]interface{})
			if !ok {
				continue
			}
			if t, _ := cm["type"].(string); t == "Ready" {
				if reason, ok := cm["reason"].(string); ok {
					return reason
				}
			}
		}
	}
	return ""
}

func toMap(v interface{}) (map[string]interface{}, bool) {
	if m, ok := v.(map[string]interface{}); ok {
		return m, true
	}
	b, err := json.Marshal(v)
	if err != nil {
		return nil, false
	}
	var m map[string]interface{}
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, false
	}
	return m, true
}
