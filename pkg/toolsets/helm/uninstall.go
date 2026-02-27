package helm

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"helm.sh/helm/v3/pkg/action"
)

func (t *Toolset) uninstallTool() mcp.Tool {
	return mcp.NewTool(
		"helm_uninstall",
		mcp.WithDescription("Uninstall a Helm release (requires read_only=false, destructive allowed)"),
		mcp.WithString("cluster", mcp.Required(), mcp.Description("Rancher cluster ID")),
		mcp.WithString("release", mcp.Required(), mcp.Description("Release name")),
		mcp.WithString("namespace", mcp.Description("Namespace (default: default)")),
		mcp.WithBoolean("keep_history", mcp.Description("Keep release history (default: false)")),
		mcp.WithBoolean("wait", mcp.Description("Wait for resources to be deleted (default: false)")),
	)
}

func (t *Toolset) uninstallHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if err := t.policy.CheckDestructive(); err != nil {
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
	namespace := req.GetString("namespace", "default")
	keepHistory := req.GetBool("keep_history", false)
	wait := req.GetBool("wait", false)

	cfg, err := t.actionConfigFor(cluster, namespace)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("helm action config: %v", err)), nil
	}
	uninstallAction := action.NewUninstall(cfg)
	uninstallAction.KeepHistory = keepHistory
	uninstallAction.Wait = wait

	res, err := uninstallAction.Run(releaseName)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("helm uninstall: %v", err)), nil
	}
	out := fmt.Sprintf("Uninstalled %s from namespace %s", res.Release.Name, res.Release.Namespace)
	return mcp.NewToolResultText(out), nil
}
