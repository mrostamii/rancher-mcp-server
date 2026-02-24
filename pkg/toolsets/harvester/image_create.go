package harvester

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mrostamii/rancher-mcp-server/pkg/client/rancher"
)

func (t *Toolset) imageCreateTool() mcp.Tool {
	return mcp.NewTool(
		"harvester_image_create",
		mcp.WithDescription("Create a Harvester VM image (VirtualMachineImage) from a download URL (qcow2, raw, or ISO)."),
		mcp.WithString("cluster", mcp.Required(), mcp.Description("Harvester cluster ID")),
		mcp.WithString("namespace", mcp.Required(), mcp.Description("Namespace for the image")),
		mcp.WithString("name", mcp.Required(), mcp.Description("Image resource name")),
		mcp.WithString("url", mcp.Required(), mcp.Description("Download URL for the image")),
		mcp.WithString("display_name", mcp.Description("Human-readable display name (default: same as name)")),
		mcp.WithString("checksum", mcp.Description("Optional checksum for verification")),
	)
}

func (t *Toolset) imageCreateHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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
	url, err := req.RequireString("url")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	displayName := req.GetString("display_name", name)
	checksum := req.GetString("checksum", "")

	spec := map[string]interface{}{
		"displayName": displayName,
		"sourceType":  "download",
		"url":         url,
	}
	if checksum != "" {
		spec["checksum"] = checksum
	}

	body := map[string]interface{}{
		"apiVersion": "harvesterhci.io/v1beta1",
		"kind":       "VirtualMachineImage",
		"metadata": map[string]interface{}{
			"name":      name,
			"namespace": namespace,
		},
		"spec": spec,
	}

	_, err = t.client.Create(ctx, cluster, rancher.TypeVirtualMachineImages, namespace, body)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("harvester_image_create: %v", err)), nil
	}
	return mcp.NewToolResultText(fmt.Sprintf("Image %q created in namespace %q from %s", name, namespace, url)), nil
}
