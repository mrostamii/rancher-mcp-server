package fleet

import (
	"context"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mrostamii/rancher-mcp-server/pkg/client/rancher"
)

func (t *Toolset) gitrepoCreateTool() mcp.Tool {
	return mcp.NewTool(
		"fleet_gitrepo_create",
		mcp.WithDescription("Create a Fleet GitRepo (requires read_only=false)"),
		mcp.WithString("name", mcp.Required(), mcp.Description("GitRepo name")),
		mcp.WithString("repo", mcp.Required(), mcp.Description("Git repository URL (HTTPS or git)")),
		mcp.WithString("namespace", mcp.Description("Namespace (default: fleet-default)")),
		mcp.WithString("branch", mcp.Description("Branch to track (default: main)")),
		mcp.WithString("paths", mcp.Description("Comma-separated paths in repo (e.g. path1,path2)")),
		mcp.WithString("format", mcp.Description("Output format: json, table (default: json)")),
	)
}

func (t *Toolset) gitrepoCreateHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	name, err := req.RequireString("name")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	repo, err := req.RequireString("repo")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	namespace := req.GetString("namespace", "fleet-default")

	if err := t.policy.CheckNamespace(namespace); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	branch := req.GetString("branch", "main")
	pathsStr := req.GetString("paths", "")
	format := req.GetString("format", "json")

	spec := map[string]interface{}{
		"repo": repo,
	}
	if branch != "" {
		spec["branch"] = branch
	}
	if pathsStr != "" {
		var paths []string
		for _, p := range strings.Split(pathsStr, ",") {
			p = strings.TrimSpace(p)
			if p != "" {
				paths = append(paths, p)
			}
		}
		if len(paths) > 0 {
			spec["paths"] = paths
		}
	}

	body := map[string]interface{}{
		"apiVersion": "fleet.cattle.io/v1alpha1",
		"kind":       "GitRepo",
		"metadata": map[string]interface{}{
			"name":      name,
			"namespace": namespace,
		},
		"spec": spec,
	}

	res, err := t.client.Create(ctx, localCluster, rancher.TypeFleetGitRepos, namespace, body)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("fleet_gitrepo_create: %v", err)), nil
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
