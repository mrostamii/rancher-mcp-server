package harvester

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mrostamii/rancher-mcp-server/pkg/client/rancher"
)

var vmActions = map[string]bool{"start": true, "stop": true, "restart": true, "pause": true, "unpause": true, "migrate": true}

func (t *Toolset) vmActionTool() mcp.Tool {
	return mcp.NewTool(
		"harvester_vm_action",
		mcp.WithDescription("Run lifecycle action on a VM: start, stop, restart, pause, unpause, migrate"),
		mcp.WithString("cluster", mcp.Required(), mcp.Description("Harvester cluster ID")),
		mcp.WithString("namespace", mcp.Required(), mcp.Description("Namespace")),
		mcp.WithString("name", mcp.Required(), mcp.Description("VM name")),
		mcp.WithString("action", mcp.Required(), mcp.Description("Action: start, stop, restart, pause, unpause, migrate")),
	)
}

func (t *Toolset) vmActionHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	cluster := req.GetString("cluster", "")
	namespace, err := req.RequireString("namespace")
	if err != nil {
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
	if !vmActions[action] {
		return mcp.NewToolResultError(fmt.Sprintf("invalid action %q; allowed: start, stop, restart, pause, unpause, migrate", action)), nil
	}

	switch action {
	case "stop":
		// Patch runStrategy to Halted. Using the Steve ?action=stop on a VM with
		// runStrategy=RerunOnFailure causes KubeVirt to treat the shutdown as a failure
		// and immediately restart the VM. Setting Halted is the correct way to stop
		// and keep the VM off (this is what the Harvester UI does).
		patch := map[string]interface{}{
			"spec": map[string]interface{}{
				"runStrategy": "Halted",
			},
		}
		_, err = t.client.Patch(ctx, cluster, rancher.TypeVirtualMachines, namespace, name, patch)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("harvester_vm_action stop: %v", err)), nil
		}

	case "start":
		// Restore the run strategy stored in the Harvester annotation, which records
		// what the strategy was before the VM was halted. Default to RerunOnFailure.
		vm, err := t.client.Get(ctx, cluster, rancher.TypeVirtualMachines, namespace, name)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("harvester_vm_action start: %v", err)), nil
		}
		runStrategy := "RerunOnFailure"
		if prev, ok := vm.ObjectMeta.Annotations["harvesterhci.io/vmRunStrategy"]; ok && prev != "" && prev != "Halted" {
			runStrategy = prev
		}
		patch := map[string]interface{}{
			"spec": map[string]interface{}{
				"runStrategy": runStrategy,
			},
		}
		_, err = t.client.Patch(ctx, cluster, rancher.TypeVirtualMachines, namespace, name, patch)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("harvester_vm_action start: %v", err)), nil
		}

	default:
		// restart, pause, unpause, migrate: use the Steve action endpoint.
		err = t.client.Action(ctx, cluster, rancher.TypeVirtualMachines, namespace, name, action, nil)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("harvester_vm_action: %v", err)), nil
		}
	}

	return mcp.NewToolResultText(fmt.Sprintf("VM %q action %q completed", name, action)), nil
}
