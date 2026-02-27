package helm

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
)

func (t *Toolset) upgradeTool() mcp.Tool {
	return mcp.NewTool(
		"helm_upgrade",
		mcp.WithDescription("Upgrade a Helm release (requires read_only=false)"),
		mcp.WithString("cluster", mcp.Required(), mcp.Description("Rancher cluster ID")),
		mcp.WithString("release", mcp.Required(), mcp.Description("Release name")),
		mcp.WithString("chart", mcp.Required(), mcp.Description("Chart name or repo/name or OCI/URL")),
		mcp.WithString("namespace", mcp.Description("Namespace (default: default)")),
		mcp.WithString("repo_url", mcp.Description("Repository URL")),
		mcp.WithString("version", mcp.Description("Chart version constraint")),
		mcp.WithString("values", mcp.Description("JSON object of values to override")),
		mcp.WithBoolean("wait", mcp.Description("Wait for resources (default: false)")),
		mcp.WithBoolean("install", mcp.Description("Install if release not found (default: true)")),
	)
}

func (t *Toolset) upgradeHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if err := t.policy.CheckWrite(); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	cluster, err := req.RequireString("cluster")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	releaseName, err := req.RequireString("release")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	chartRef, err := req.RequireString("chart")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	namespace := req.GetString("namespace", "default")
	repoURL := req.GetString("repo_url", "")
	version := req.GetString("version", "")
	valuesStr := req.GetString("values", "{}")
	wait := req.GetBool("wait", false)
	install := req.GetBool("install", true)

	var vals map[string]interface{}
	if err := json.Unmarshal([]byte(valuesStr), &vals); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("invalid values JSON: %v", err)), nil
	}
	if vals == nil {
		vals = map[string]interface{}{}
	}

	cfg, err := t.actionConfigFor(cluster, namespace)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("helm action config: %v", err)), nil
	}
	upgradeAction := action.NewUpgrade(cfg)
	upgradeAction.Namespace = namespace
	upgradeAction.Wait = wait
	upgradeAction.Install = install
	upgradeAction.ChartPathOptions.RepoURL = repoURL
	upgradeAction.ChartPathOptions.Version = version

	settings := cli.New()
	if repoURL != "" {
		if err := prepareHelmSettingsForRepoURL(settings.RepositoryConfig, settings.RepositoryCache); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("prepare helm repo settings: %v", err)), nil
		}
	}
	chartPath, err := locateChartNoPanic(func() (string, error) {
		return upgradeAction.ChartPathOptions.LocateChart(chartRef, settings)
	})
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("locate chart: %v", err)), nil
	}
	chartRequested, err := loader.Load(chartPath)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("load chart: %v", err)), nil
	}

	rel, err := upgradeAction.RunWithContext(ctx, releaseName, chartRequested, vals)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("helm upgrade: %v", err)), nil
	}
	out := fmt.Sprintf("Upgraded %s in namespace %s (revision %d)", rel.Name, rel.Namespace, rel.Version)
	return mcp.NewToolResultText(out), nil
}
