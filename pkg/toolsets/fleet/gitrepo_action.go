package fleet

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mrostamii/rancher-mcp-server/pkg/client/rancher"
)

var gitrepoActions = map[string]bool{
	"pause": true, "unpause": true,
	"disablePolling": true, "enablePolling": true,
	"forceUpdate": true,
}

func (t *Toolset) gitrepoActionTool() mcp.Tool {
	return mcp.NewTool(
		"fleet_gitrepo_action",
		mcp.WithDescription("Run action on a Fleet GitRepo: pause, unpause, disablePolling, enablePolling, forceUpdate (requires read_only=false)"),
		mcp.WithString("name", mcp.Required(), mcp.Description("GitRepo name")),
		mcp.WithString("namespace", mcp.Description("Namespace (default: fleet-default)")),
		mcp.WithString("action", mcp.Required(), mcp.Description("Action: pause, unpause, disablePolling, enablePolling, forceUpdate")),
	)
}

func (t *Toolset) gitrepoActionHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if err := t.policy.CheckWrite(); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	name, err := req.RequireString("name")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	action, err := req.RequireString("action")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	namespace := req.GetString("namespace", "fleet-default")

	if err := t.policy.CheckNamespace(namespace); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	if !gitrepoActions[action] {
		return mcp.NewToolResultError(fmt.Sprintf("invalid action %q; allowed: pause, unpause, disablePolling, enablePolling, forceUpdate", action)), nil
	}

	var patch map[string]interface{}
	switch action {
	case "pause":
		patch = map[string]interface{}{"spec": map[string]interface{}{"paused": true}}
	case "unpause":
		patch = map[string]interface{}{"spec": map[string]interface{}{"paused": false}}
	case "disablePolling":
		patch = map[string]interface{}{"spec": map[string]interface{}{"disablePolling": true}}
	case "enablePolling":
		patch = map[string]interface{}{"spec": map[string]interface{}{"disablePolling": false}}
	case "forceUpdate":
		existing, err := t.client.Get(ctx, localCluster, rancher.TypeFleetGitRepos, namespace, name)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("fleet_gitrepo_action forceUpdate: %v", err)), nil
		}
		gen := int64(0)
		if spec, ok := existing.Spec.(map[string]interface{}); ok {
			if g, ok := spec["forceSyncGeneration"].(float64); ok {
				gen = int64(g)
			}
		}
		gen++
		patch = map[string]interface{}{"spec": map[string]interface{}{"forceSyncGeneration": gen}}
	}

	_, err = t.client.Patch(ctx, localCluster, rancher.TypeFleetGitRepos, namespace, name, patch)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("fleet_gitrepo_action %s: %v", action, err)), nil
	}
	return mcp.NewToolResultText(fmt.Sprintf("GitRepo %q action %q completed", name, action)), nil
}
