package fleet

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mrostamii/rancher-mcp-server/pkg/client/rancher"
)

func (t *Toolset) gitrepoCloneTool() mcp.Tool {
	return mcp.NewTool(
		"fleet_gitrepo_clone",
		mcp.WithDescription("Clone a Fleet GitRepo to a new name (copy spec from existing; requires read_only=false)"),
		mcp.WithString("name", mcp.Required(), mcp.Description("Source GitRepo name")),
		mcp.WithString("clone_name", mcp.Required(), mcp.Description("Name for the new clone")),
		mcp.WithString("namespace", mcp.Description("Namespace (default: fleet-default)")),
		mcp.WithString("format", mcp.Description("Output format: json, table (default: json)")),
	)
}

func (t *Toolset) gitrepoCloneHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if err := t.policy.CheckWrite(); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	name, err := req.RequireString("name")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	cloneName, err := req.RequireString("clone_name")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	namespace := req.GetString("namespace", "fleet-default")
	format := req.GetString("format", "json")

	source, err := t.client.Get(ctx, localCluster, rancher.TypeFleetGitRepos, namespace, name)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("source GitRepo %q not found: %v", name, err)), nil
	}

	spec, ok := source.Spec.(map[string]interface{})
	if !ok {
		return mcp.NewToolResultError("invalid source GitRepo spec"), nil
	}

	body := map[string]interface{}{
		"apiVersion": "fleet.cattle.io/v1alpha1",
		"kind":       "GitRepo",
		"metadata": map[string]interface{}{
			"name":      cloneName,
			"namespace": namespace,
		},
		"spec": spec,
	}

	res, err := t.client.Create(ctx, localCluster, rancher.TypeFleetGitRepos, namespace, body)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("fleet_gitrepo_clone: %v", err)), nil
	}
	data := map[string]interface{}{
		"metadata": res.ObjectMeta,
		"spec":     res.Spec,
		"status":   res.Status,
	}
	out, err := t.formatter.Format(data, format)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("format: %v", err)), nil
	}
	return mcp.NewToolResultText(out), nil
}
