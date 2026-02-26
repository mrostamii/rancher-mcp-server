package harvester

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mrostamii/rancher-mcp-server/pkg/client/rancher"
)

func (t *Toolset) vmCreateTool() mcp.Tool {
	return mcp.NewTool(
		"harvester_vm_create",
		mcp.WithDescription("Create a Harvester VM. For KubeOVN VPC with external internet: use network (NAD name), interface_type=managedtap, and ensure the subnet has nat_outgoing=true."),
		mcp.WithString("cluster", mcp.Required(), mcp.Description("Harvester cluster ID")),
		mcp.WithString("namespace", mcp.Required(), mcp.Description("Namespace for the VM")),
		mcp.WithString("name", mcp.Required(), mcp.Description("VM name")),
		mcp.WithNumber("cpu", mcp.Description("CPU cores (default: 2)")),
		mcp.WithString("memory", mcp.Description("Memory size e.g. 2Gi, 4096Mi (default: 4Gi)")),
		mcp.WithString("image", mcp.Required(), mcp.Description("VirtualMachineImage name (use harvester_image_list to list)")),
		mcp.WithString("network", mcp.Description("Network name - NAD for KubeOVN overlay (use harvester_network_list; default: default pod network)")),
		mcp.WithString("interface_type", mcp.Description("Interface: managedtap (KubeOVN, recommended), bridge, or masquerade (default: masquerade for pod, managedtap for custom network)")),
		mcp.WithString("subnet", mcp.Description("KubeOVN logical_switch (subnet name) for IP assignment; optional, provider from network is used if unset")),
		mcp.WithNumber("disk_size_gib", mcp.Description("Root disk size in GiB (default: 20)")),
		mcp.WithString("run_strategy", mcp.Description("RunStrategy: Always, RerunOnFailure, Halted (default: RerunOnFailure)")),
	)
}

func (t *Toolset) vmCreateHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if err := t.policy.CheckWrite(); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	cluster := req.GetString("cluster", "")
	namespace, err := req.RequireString("namespace")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	name, err := req.RequireString("name")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	image, err := req.RequireString("image")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	cpu := req.GetInt("cpu", 2)
	if cpu < 1 {
		cpu = 1
	}
	memory := req.GetString("memory", "4Gi")
	if memory == "" {
		memory = "4Gi"
	}
	network := req.GetString("network", "default")
	interfaceType := req.GetString("interface_type", "")
	subnet := req.GetString("subnet", "")
	diskSizeGiB := req.GetInt("disk_size_gib", 20)
	if diskSizeGiB < 1 {
		diskSizeGiB = 20
	}
	runStrategy := req.GetString("run_strategy", "RerunOnFailure")

	// Look up the VirtualMachineImage to get the storageClassName assigned by Harvester.
	// Harvester creates a per-image StorageClass (e.g. "longhorn-image-<id>") that the PVC
	// must reference so Longhorn can clone the backing image into the new volume.
	imgRes, err := t.client.Get(ctx, cluster, rancher.TypeVirtualMachineImages, namespace, image)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("image %q not found in namespace %q: %v", image, namespace, err)), nil
	}
	storageClassName := "longhorn" // safe fallback
	if imgRes.Status != nil {
		if statusMap, ok := imgRes.Status.(map[string]interface{}); ok {
			if sc, ok := statusMap["storageClassName"].(string); ok && sc != "" {
				storageClassName = sc
			}
		}
	}

	// Create root disk PVC from the Harvester image.
	// Harvester provisions volumes via a per-image Longhorn StorageClass (longhorn-image-<id>).
	// The annotation harvesterhci.io/imageId tells the Harvester controller which image to clone;
	// no dataSourceRef/VolumePopulator is needed or supported.
	pvcName := name + "-disk-0"
	pvc := map[string]interface{}{
		"apiVersion": "v1",
		"kind":       "PersistentVolumeClaim",
		"metadata": map[string]interface{}{
			"name":      pvcName,
			"namespace": namespace,
			"annotations": map[string]string{
				"harvesterhci.io/imageId": fmt.Sprintf("%s/%s", namespace, image),
			},
		},
		"spec": map[string]interface{}{
			"accessModes": []string{"ReadWriteMany"},
			"resources": map[string]interface{}{
				"requests": map[string]interface{}{
					"storage": fmt.Sprintf("%dGi", diskSizeGiB),
				},
			},
			"storageClassName": storageClassName,
			"volumeMode":       "Block",
		},
	}
	_, err = t.client.Create(ctx, cluster, rancher.TypePersistentVolumeClaims, namespace, pvc)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to create root disk PVC from image %q: %v", image, err)), nil
	}

	// Build interface and network based on network type.
	// For KubeOVN VPC: use multus + managedtap; subnet needs nat_outgoing=true for external internet.
	useKubeOVN := network != "" && network != "default"
	if interfaceType == "" {
		if useKubeOVN {
			interfaceType = "managedtap"
		} else {
			interfaceType = "masquerade"
		}
	}

	templateMetadata := map[string]interface{}{
		"labels": map[string]string{"harvesterhci.io/vmName": name},
	}
	if useKubeOVN && (interfaceType == "managedtap" || interfaceType == "bridge") {
		templateMetadata["annotations"] = map[string]string{
			"kubevirt.io/allow-pod-bridge-network-live-migration": "true",
		}
	}
	if subnet != "" && useKubeOVN {
		if ann, ok := templateMetadata["annotations"].(map[string]string); ok {
			ann["ovn.kubernetes.io/logical_switch"] = subnet
		} else {
			templateMetadata["annotations"] = map[string]string{
				"kubevirt.io/allow-pod-bridge-network-live-migration": "true",
				"ovn.kubernetes.io/logical_switch":                    subnet,
			}
		}
	}

	iface := buildVMInterface("default", interfaceType)
	net := buildVMNetwork("default", network, namespace)

	spec := map[string]interface{}{
		"runStrategy": runStrategy,
		"template": map[string]interface{}{
			"metadata": templateMetadata,
			"spec": map[string]interface{}{
				"domain": map[string]interface{}{
					"cpu": map[string]interface{}{
						"cores": uint32(cpu),
					},
					"memory": map[string]interface{}{
						"guest": memory,
					},
					"devices": map[string]interface{}{
						"disks": []map[string]interface{}{
							{
								"name":      "disk-0",
								"disk":      map[string]interface{}{"bus": "virtio"},
								"bootOrder": 1,
							},
						},
						"interfaces": []map[string]interface{}{iface},
					},
					"machine": map[string]interface{}{"type": "q35"},
					"resources": map[string]interface{}{
						"requests": map[string]interface{}{
							"memory": memory,
							"cpu":    fmt.Sprintf("%d", cpu),
						},
						"limits": map[string]interface{}{
							"memory": memory,
							"cpu":    fmt.Sprintf("%d", cpu),
						},
					},
				},
				"volumes": []map[string]interface{}{
					{
						"name": "disk-0",
						"persistentVolumeClaim": map[string]interface{}{
							"claimName": pvcName,
						},
					},
				},
				"networks": []map[string]interface{}{net},
			},
		},
	}

	vm := map[string]interface{}{
		"apiVersion": "kubevirt.io/v1",
		"kind":       "VirtualMachine",
		"metadata": map[string]interface{}{
			"name":      name,
			"namespace": namespace,
		},
		"spec": spec,
	}

	_, err = t.client.Create(ctx, cluster, rancher.TypeVirtualMachines, namespace, vm)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("harvester_vm_create: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("VM %q created in namespace %q with root disk %q from image %q (%dGi)", name, namespace, pvcName, image, diskSizeGiB)), nil
}

// buildVMInterface returns the interface spec for the given type.
func buildVMInterface(name, ifaceType string) map[string]interface{} {
	out := map[string]interface{}{
		"name":  name,
		"model": "virtio",
	}
	switch ifaceType {
	case "managedtap":
		out["binding"] = map[string]interface{}{"name": "managedtap"}
	case "bridge":
		out["bridge"] = map[string]interface{}{}
	default:
		out["masquerade"] = map[string]interface{}{}
	}
	return out
}

// buildVMNetwork returns the network spec. For custom networks, uses multus with default=true.
func buildVMNetwork(name, network, namespace string) map[string]interface{} {
	if network == "" || network == "default" {
		return map[string]interface{}{
			"name": name,
			"pod":  map[string]interface{}{},
		}
	}
	networkName := fmt.Sprintf("%s/%s", namespace, network)
	return map[string]interface{}{
		"name": name,
		"multus": map[string]interface{}{
			"default":     true,
			"networkName": networkName,
		},
	}
}
