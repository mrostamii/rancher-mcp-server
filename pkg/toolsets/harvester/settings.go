package harvester

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mrostamii/rancher-mcp-server/pkg/client/rancher"
	"github.com/mrostamii/rancher-mcp-server/pkg/formatter"
)

const clusterPodSecuritySetting = "cluster-pod-security-standard"

func (t *Toolset) settingsTool() mcp.Tool {
	return mcp.NewTool(
		"harvester_settings",
		mcp.WithDescription("List, get, or update Harvester cluster settings. Update requires read-only=false. Pod Security Standards use pss_* on cluster-pod-security-standard (comma-separated namespace lists)."),
		mcp.WithString("cluster", mcp.Required(), mcp.Description("Harvester cluster ID")),
		mcp.WithString("name", mcp.Description("Setting name (empty = list all). Required when setting value. Defaults to cluster-pod-security-standard when pss_* is set.")),
		mcp.WithString("value", mcp.Description("New setting value (string). Writes the named setting. Do not combine with pss_*.")),
		mcp.WithString("pss_enabled", mcp.Description("Enable cluster-pod-security-standard: true or false")),
		mcp.WithString("pss_whitelisted", mcp.Description("Comma-separated namespaces exempt from Pod Security labels")),
		mcp.WithString("pss_privileged", mcp.Description("Comma-separated namespaces enforced as privileged")),
		mcp.WithString("pss_restricted", mcp.Description("Comma-separated namespaces enforced as restricted")),
		mcp.WithString("format", mcp.Description("Output format: json, table (default: json)")),
		mcp.WithNumber("limit", mcp.Description("Max items when listing (default: 100)")),
		mcp.WithString("continue", mcp.Description("Pagination token from previous response (for next page; list mode only)")),
	)
}

func (t *Toolset) settingsHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	cluster, err := req.RequireString("cluster")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	name := req.GetString("name", "")
	format := req.GetString("format", "json")
	limit := req.GetInt("limit", 100)
	if limit <= 0 {
		limit = 100
	}
	continueToken := req.GetString("continue", "")

	updated, result, err := t.maybeUpdateSetting(ctx, req, cluster, name)
	if updated || err != nil {
		return result, err
	}

	if name != "" {
		// Get single setting (cluster-scoped, empty namespace)
		setting, err := t.client.Get(ctx, cluster, rancher.TypeSettings, "", name)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("harvester_settings get: %v", err)), nil
		}
		value := settingValue(*setting)
		data := map[string]interface{}{
			"name":  setting.ObjectMeta.Name,
			"value": value,
			"spec":  setting.Spec,
		}
		out, err := t.formatter.Format(data, format)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("format: %v", err)), nil
		}
		return mcp.NewToolResultText(out), nil
	}

	// List all settings
	opts := rancher.ListOpts{Limit: limit, Continue: continueToken}
	col, err := t.client.List(ctx, cluster, rancher.TypeSettings, opts)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("harvester_settings list: %v", err)), nil
	}
	items := make([]map[string]interface{}, 0, len(col.Data))
	for _, r := range col.Data {
		value := settingValue(r)
		items = append(items, map[string]interface{}{
			"name":  r.ObjectMeta.Name,
			"value": value,
		})
	}
	out, err := formatter.FormatListWithContinue(t.formatter, items, col.Continue, format)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("format: %v", err)), nil
	}
	return mcp.NewToolResultText(out), nil
}

func settingValue(setting rancher.SteveResource) string {
	if setting.Value != "" {
		return setting.Value
	}
	if spec, ok := setting.Spec.(map[string]interface{}); ok {
		if v, ok := spec["value"].(string); ok {
			return v
		}
	}
	return ""
}

type podSecuritySetting struct {
	Enabled                   bool   `json:"enabled"`
	WhitelistedNamespacesList string `json:"whitelistedNamespacesList"`
	PrivilegedNamespacesList  string `json:"privilegedNamespacesList"`
	RestrictedNamespacesList  string `json:"restrictedNamespacesList"`
}

// maybeUpdateSetting writes a setting when value or pss_* arguments are present.
// updated is false when the call is a list or get.
func (t *Toolset) maybeUpdateSetting(ctx context.Context, req mcp.CallToolRequest, cluster, name string) (bool, *mcp.CallToolResult, error) {
	rawValue, valueSet := explicitStringArg(req, "value")
	pssEnabledSet := argPresent(req, "pss_enabled")
	whitelistSet := argPresent(req, "pss_whitelisted")
	privilegedSet := argPresent(req, "pss_privileged")
	restrictedSet := argPresent(req, "pss_restricted")
	pssSet := pssEnabledSet || whitelistSet || privilegedSet || restrictedSet
	if !valueSet && !pssSet {
		return false, nil, nil
	}
	if err := t.policy.CheckWrite(); err != nil {
		return true, mcp.NewToolResultError(err.Error()), nil
	}
	if valueSet && pssSet {
		return true, mcp.NewToolResultError("set either value or pss_* fields, not both"), nil
	}

	var value string
	if pssSet {
		if name != "" && name != clusterPodSecuritySetting {
			return true, mcp.NewToolResultError("pss_* fields only apply to " + clusterPodSecuritySetting), nil
		}
		if !pssEnabledSet {
			return true, mcp.NewToolResultError("pss_enabled is required when setting Pod Security namespaces"), nil
		}
		enabled, _, err := optionalBoolArg(req, "pss_enabled")
		if err != nil {
			return true, mcp.NewToolResultError(err.Error()), nil
		}
		whitelist := normalizeCommaList(req.GetString("pss_whitelisted", ""))
		privileged := normalizeCommaList(req.GetString("pss_privileged", ""))
		restricted := normalizeCommaList(req.GetString("pss_restricted", ""))
		if err := rejectOverlappingNamespaces(whitelist, privileged, restricted); err != nil {
			return true, mcp.NewToolResultError(err.Error()), nil
		}
		raw, err := json.Marshal(podSecuritySetting{
			Enabled:                   enabled,
			WhitelistedNamespacesList: whitelist,
			PrivilegedNamespacesList:  privileged,
			RestrictedNamespacesList:  restricted,
		})
		if err != nil {
			return true, mcp.NewToolResultError(fmt.Sprintf("encode pod security setting: %v", err)), nil
		}
		name = clusterPodSecuritySetting
		value = string(raw)
	} else {
		if name == "" {
			return true, mcp.NewToolResultError("harvester_settings update requires name"), nil
		}
		value = rawValue
	}

	existing, err := t.client.Get(ctx, cluster, rancher.TypeSettings, "", name)
	if err != nil {
		return true, mcp.NewToolResultError(fmt.Sprintf("harvester_settings get before update: %v", err)), nil
	}
	body := map[string]interface{}{
		"apiVersion": "harvesterhci.io/v1beta1",
		"kind":       "Setting",
		"metadata": map[string]interface{}{
			"name":            name,
			"resourceVersion": existing.ObjectMeta.ResourceVersion,
		},
		"value": value,
	}
	if existing.Default != "" {
		body["default"] = existing.Default
	}
	if _, err := t.client.Update(ctx, cluster, rancher.TypeSettings, "", name, body); err != nil {
		return true, mcp.NewToolResultError(fmt.Sprintf("harvester_settings update: %v", err)), nil
	}
	return true, mcp.NewToolResultText(fmt.Sprintf("Setting %q updated", name)), nil
}

func normalizeCommaList(raw string) string {
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return strings.Join(out, ",")
}

func rejectOverlappingNamespaces(lists ...string) error {
	seen := map[string]struct{}{}
	for _, list := range lists {
		for _, name := range strings.Split(list, ",") {
			if name == "" {
				continue
			}
			if _, ok := seen[name]; ok {
				return fmt.Errorf("namespace %q is listed in more than one Pod Security list", name)
			}
			seen[name] = struct{}{}
		}
	}
	return nil
}
