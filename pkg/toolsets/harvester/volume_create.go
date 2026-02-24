package harvester

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mrostamii/rancher-mcp-server/pkg/client/rancher"
)

func (t *Toolset) volumeCreateTool() mcp.Tool {
	return mcp.NewTool(
		"harvester_volume_create",
		mcp.WithDescription("Create a Harvester volume (PersistentVolumeClaim backed by Longhorn). Optionally clone from an existing VM image."),
		mcp.WithString("cluster", mcp.Required(), mcp.Description("Harvester cluster ID")),
		mcp.WithString("namespace", mcp.Required(), mcp.Description("Namespace for the volume")),
		mcp.WithString("name", mcp.Required(), mcp.Description("Volume (PVC) name")),
		mcp.WithString("size", mcp.Required(), mcp.Description("Size e.g. 10Gi, 100Gi")),
		mcp.WithString("storage_class", mcp.Description("Storage class (default: longhorn)")),
		mcp.WithString("image_name", mcp.Description("Optional: VirtualMachineImage name to clone from (creates volume from image)")),
		mcp.WithString("image_namespace", mcp.Description("Namespace of the image when image_name is set (default: same as namespace)")),
	)
}

func (t *Toolset) volumeCreateHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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
	size := req.GetString("size", "")
	if size == "" {
		return mcp.NewToolResultError("harvester_volume_create requires size (e.g. 10Gi)"), nil
	}
	storageClass := req.GetString("storage_class", "longhorn")
	imageName := req.GetString("image_name", "")
	imageNamespace := req.GetString("image_namespace", namespace)

	if imageName != "" {
		// Clone from image: use image's storageClassName and harvesterhci.io/imageId (same as vm_create).
		imgRes, err := t.client.Get(ctx, cluster, rancher.TypeVirtualMachineImages, imageNamespace, imageName)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("image %q not found in namespace %q: %v", imageName, imageNamespace, err)), nil
		}
		if imgRes.Status != nil {
			if statusMap, ok := imgRes.Status.(map[string]interface{}); ok {
				if sc, ok := statusMap["storageClassName"].(string); ok && sc != "" {
					storageClass = sc
				}
			}
		}
	}

	spec := map[string]interface{}{
		"accessModes": []string{"ReadWriteOnce"},
		"resources": map[string]interface{}{
			"requests": map[string]interface{}{
				"storage": size,
			},
		},
		"storageClassName": storageClass,
	}
	metadata := map[string]interface{}{
		"name":      name,
		"namespace": namespace,
	}
	if imageName != "" {
		metadata["annotations"] = map[string]string{
			"harvesterhci.io/imageId": fmt.Sprintf("%s/%s", imageNamespace, imageName),
		}
		spec["accessModes"] = []string{"ReadWriteMany"}
		spec["volumeMode"] = "Block"
	}

	body := map[string]interface{}{
		"apiVersion": "v1",
		"kind":       "PersistentVolumeClaim",
		"metadata":   metadata,
		"spec":       spec,
	}

	_, err = t.client.Create(ctx, cluster, rancher.TypePersistentVolumeClaims, namespace, body)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("harvester_volume_create: %v", err)), nil
	}
	if imageName != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Volume %q created in namespace %q (%s) from image %q", name, namespace, size, imageName)), nil
	}
	return mcp.NewToolResultText(fmt.Sprintf("Volume %q created in namespace %q (%s)", name, namespace, size)), nil
}
