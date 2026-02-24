package harvester

import (
	"context"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mrostamii/rancher-mcp-server/pkg/client/rancher"
)

func (t *Toolset) vmBackupTool() mcp.Tool {
	return mcp.NewTool(
		"harvester_vm_backup",
		mcp.WithDescription("Create, list, or restore a Harvester VM backup (VirtualMachineBackup). Create requires the cluster backup target to be configured first (Settings > backup-target in Harvester UI; NFS or S3). That configuration is manual or done in a separate setup phase—this tool cannot configure it. List and restore work regardless."),
		mcp.WithString("cluster", mcp.Required(), mcp.Description("Harvester cluster ID")),
		mcp.WithString("action", mcp.Required(), mcp.Description("Action: create, list, restore")),
		mcp.WithString("namespace", mcp.Description("Namespace (required for create/restore)")),
		mcp.WithString("vm_name", mcp.Description("VM name (required for create)")),
		mcp.WithString("backup_name", mcp.Description("Backup name (required for restore, optional for create)")),
		mcp.WithString("format", mcp.Description("Output format for list: json, table (default: json)")),
		mcp.WithNumber("limit", mcp.Description("Max items for list (default: 100)")),
	)
}

func (t *Toolset) vmBackupHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	cluster := req.GetString("cluster", "")
	action := req.GetString("action", "")
	namespace := req.GetString("namespace", "")
	vmName := req.GetString("vm_name", "")
	backupName := req.GetString("backup_name", "")
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
			return mcp.NewToolResultError("harvester_vm_backup create requires namespace and vm_name"), nil
		}
		name := backupName
		if name == "" {
			name = "backup-" + vmName
		}
		spec := map[string]interface{}{
			"source": map[string]interface{}{
				"apiGroup": "kubevirt.io",
				"kind":     "VirtualMachine",
				"name":     vmName,
			},
		}
		body := map[string]interface{}{
			"apiVersion": "harvesterhci.io/v1beta1",
			"kind":       "VirtualMachineBackup",
			"metadata": map[string]interface{}{
				"name":      name,
				"namespace": namespace,
			},
			"spec": spec,
		}
		_, err := t.client.Create(ctx, cluster, rancher.TypeVirtualMachineBackups, namespace, body)
		if err != nil {
			errStr := err.Error()
			if strings.Contains(errStr, "422") && strings.Contains(errStr, "backup target") {
				return mcp.NewToolResultError("harvester_vm_backup create failed: backup target is not set. This must be done manually or in a separate setup phase (e.g. Harvester UI Settings > backup-target, or another MCP/workflow). This tool cannot configure the backup target; list and restore remain available."), nil
			}
			return mcp.NewToolResultError(fmt.Sprintf("harvester_vm_backup create: %v", err)), nil
		}
		return mcp.NewToolResultText(fmt.Sprintf("Backup %q created for VM %q in namespace %q", name, vmName, namespace)), nil

	case "list":
		opts := rancher.ListOpts{Limit: limit}
		if namespace != "" {
			opts.Namespace = namespace
		}
		col, err := t.client.List(ctx, cluster, rancher.TypeVirtualMachineBackups, opts)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("harvester_vm_backup list: %v", err)), nil
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
		if namespace == "" || backupName == "" || vmName == "" {
			return mcp.NewToolResultError("harvester_vm_backup restore requires namespace, backup_name, and vm_name (target VM name)"), nil
		}
		restoreName := "restore-" + backupName
		spec := map[string]interface{}{
			"virtualMachineBackupName": backupName,
			"target": map[string]interface{}{
				"apiGroup": "kubevirt.io",
				"kind":     "VirtualMachine",
				"name":     vmName,
			},
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
			return mcp.NewToolResultError(fmt.Sprintf("harvester_vm_backup restore: %v", err)), nil
		}
		return mcp.NewToolResultText(fmt.Sprintf("Restore %q created from backup %q for VM %q", restoreName, backupName, vmName)), nil

	default:
		return mcp.NewToolResultError(fmt.Sprintf("invalid action %q; use create, list, or restore", action)), nil
	}
}
