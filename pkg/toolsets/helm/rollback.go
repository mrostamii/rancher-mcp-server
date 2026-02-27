package helm

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"helm.sh/helm/v3/pkg/action"
)

func (t *Toolset) rollbackTool() mcp.Tool {
	return mcp.NewTool(
		"helm_rollback",
		mcp.WithDescription("Rollback a Helm release to a previous revision (requires read_only=false)"),
		mcp.WithString("cluster", mcp.Required(), mcp.Description("Rancher cluster ID")),
		mcp.WithString("release", mcp.Required(), mcp.Description("Release name")),
		mcp.WithString("namespace", mcp.Description("Namespace (default: default)")),
		mcp.WithNumber("revision", mcp.Description("Revision to rollback to (0 = previous)")),
		mcp.WithBoolean("wait", mcp.Description("Wait for resources (default: false)")),
		mcp.WithBoolean("force", mcp.Description("Force resource update (default: false)")),
	)
}

func (t *Toolset) rollbackHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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
	namespace := req.GetString("namespace", "default")
	revision := req.GetInt("revision", 0)
	wait := req.GetBool("wait", false)
	force := req.GetBool("force", false)

	cfg, err := t.actionConfigFor(cluster, namespace)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("helm action config: %v", err)), nil
	}
	rollbackAction := action.NewRollback(cfg)
	rollbackAction.Version = revision
	rollbackAction.Wait = wait
	rollbackAction.Force = force

	if err := rollbackAction.Run(releaseName); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("helm rollback: %v", err)), nil
	}
	out := fmt.Sprintf("Rolled back %s in namespace %s to revision %d", releaseName, namespace, revision)
	if revision == 0 {
		out = fmt.Sprintf("Rolled back %s in namespace %s to previous revision", releaseName, namespace)
	}
	return mcp.NewToolResultText(out), nil
}
