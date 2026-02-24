package harvester

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mrostamii/rancher-mcp-server/pkg/client/rancher"
)

func (t *Toolset) vmSnapshotTool() mcp.Tool {
	return mcp.NewTool(
		"harvester_vm_snapshot",
		mcp.WithDescription("Create, list, restore, or delete a Harvester VM snapshot (KubeVirt VirtualMachineSnapshot, snapshot.kubevirt.io)."),
		mcp.WithString("cluster", mcp.Required(), mcp.Description("Harvester cluster ID")),
		mcp.WithString("action", mcp.Required(), mcp.Description("Action: create, list, restore, delete")),
		mcp.WithString("namespace", mcp.Description("Namespace (required for create/restore/delete)")),
		mcp.WithString("vm_name", mcp.Description("VM name (required for create)")),
		mcp.WithString("snapshot_name", mcp.Description("Snapshot name (required for restore/delete, optional for create)")),
		mcp.WithString("format", mcp.Description("Output format for list: json, table (default: json)")),
		mcp.WithNumber("limit", mcp.Description("Max items for list (default: 100)")),
	)
}

func (t *Toolset) vmSnapshotHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	cluster := req.GetString("cluster", "")
	action := req.GetString("action", "")
	namespace := req.GetString("namespace", "")
	vmName := req.GetString("vm_name", "")
	snapshotName := req.GetString("snapshot_name", "")
	format := req.GetString("format", "json")
	limit := req.GetInt("limit", 100)
	if limit <= 0 {
		limit = 100
	}

	switch action {
	case "create":
		if err := t.policy.CheckWrite(); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		if namespace == "" || vmName == "" {
			return mcp.NewToolResultError("harvester_vm_snapshot create requires namespace and vm_name"), nil
		}
		name := snapshotName
		if name == "" {
			name = "snapshot-" + vmName
		}
		spec := map[string]interface{}{
			"source": map[string]interface{}{
				"apiGroup": "kubevirt.io",
				"kind":     "VirtualMachine",
				"name":     vmName,
			},
		}
		body := map[string]interface{}{
			"apiVersion": "snapshot.kubevirt.io/v1beta1",
			"kind":       "VirtualMachineSnapshot",
			"metadata": map[string]interface{}{
				"name":      name,
				"namespace": namespace,
			},
			"spec": spec,
		}
		_, err := t.client.Create(ctx, cluster, rancher.TypeVirtualMachineSnapshots, namespace, body)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("harvester_vm_snapshot create: %v", err)), nil
		}
		return mcp.NewToolResultText(fmt.Sprintf("Snapshot %q created for VM %q in namespace %q", name, vmName, namespace)), nil

	case "list":
		opts := rancher.ListOpts{Limit: limit}
		if namespace != "" {
			opts.Namespace = namespace
		}
		col, err := t.client.List(ctx, cluster, rancher.TypeVirtualMachineSnapshots, opts)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("harvester_vm_snapshot list: %v", err)), nil
		}
		items := make([]map[string]interface{}, 0, len(col.Data))
		for _, r := range col.Data {
			items = append(items, map[string]interface{}{
				"name": r.ObjectMeta.Name, "namespace": r.ObjectMeta.Namespace,
				"spec": r.Spec, "status": r.Status,
			})
		}
		out, err := t.formatter.Format(items, format)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("format: %v", err)), nil
		}
		return mcp.NewToolResultText(out), nil

	case "restore":
		if err := t.policy.CheckWrite(); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		if namespace == "" || snapshotName == "" {
			return mcp.NewToolResultError("harvester_vm_snapshot restore requires namespace and snapshot_name"), nil
		}
		// Create VirtualMachineRestore targeting the snapshot.
		restoreName := "restore-" + snapshotName
		spec := map[string]interface{}{
			"virtualMachineSnapshotName": snapshotName,
			"target": map[string]interface{}{
				"apiGroup": "kubevirt.io",
				"kind":     "VirtualMachine",
				"name":     vmName,
			},
		}
		if vmName == "" {
			return mcp.NewToolResultError("harvester_vm_snapshot restore requires vm_name (target VM to restore into)"), nil
		}
		body := map[string]interface{}{
			"apiVersion": "harvesterhci.io/v1beta1",
			"kind":       "VirtualMachineRestore",
			"metadata": map[string]interface{}{
				"name":      restoreName,
				"namespace": namespace,
			},
			"spec": spec,
		}
		_, err := t.client.Create(ctx, cluster, rancher.TypeVirtualMachineRestores, namespace, body)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("harvester_vm_snapshot restore: %v", err)), nil
		}
		return mcp.NewToolResultText(fmt.Sprintf("Restore %q created from snapshot %q for VM %q", restoreName, snapshotName, vmName)), nil

	case "delete":
		if err := t.policy.CheckWrite(); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		if err := t.policy.CheckDestructive(); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		if namespace == "" || snapshotName == "" {
			return mcp.NewToolResultError("harvester_vm_snapshot delete requires namespace and snapshot_name"), nil
		}
		err := t.client.Delete(ctx, cluster, rancher.TypeVirtualMachineSnapshots, namespace, snapshotName)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("harvester_vm_snapshot delete: %v", err)), nil
		}
		return mcp.NewToolResultText(fmt.Sprintf("Snapshot %q deleted", snapshotName)), nil

	default:
		return mcp.NewToolResultError(fmt.Sprintf("invalid action %q; use create, list, restore, or delete", action)), nil
	}
}
