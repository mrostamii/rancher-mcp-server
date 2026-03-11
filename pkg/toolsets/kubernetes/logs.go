package kubernetes

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
)

func (t *Toolset) logsTool() mcp.Tool {
	return mcp.NewTool(
		"kubernetes_logs",
		mcp.WithDescription("Get recent logs from a pod's container (tail only; no follow). Use for debugging. Cluster is Rancher cluster ID."),
		mcp.WithString("cluster", mcp.Required(), mcp.Description("Cluster ID")),
		mcp.WithString("namespace", mcp.Required(), mcp.Description("Namespace of the pod")),
		mcp.WithString("pod", mcp.Required(), mcp.Description("Pod name")),
		mcp.WithString("container", mcp.Description("Container name (optional; default is first/only container)")),
		mcp.WithNumber("tail_lines", mcp.Description("Number of lines from the end (default: 100, max practical ~1000)")),
		mcp.WithNumber("since_seconds", mcp.Description("Only logs from last N seconds (optional)")),
	)
}

func (t *Toolset) logsHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	cluster, err := req.RequireString("cluster")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	namespace, err := req.RequireString("namespace")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	pod, err := req.RequireString("pod")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	if err := t.policy.CheckNamespace(namespace); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	container := req.GetString("container", "")
	tailLines := req.GetInt("tail_lines", 100)
	sinceSeconds := req.GetInt("since_seconds", 0)
	if tailLines <= 0 {
		tailLines = 100
	}
	if tailLines > 5000 {
		tailLines = 5000
	}

	logs, err := t.client.GetPodLogs(ctx, cluster, namespace, pod, container, tailLines, sinceSeconds)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("pod logs: %v", err)), nil
	}
	return mcp.NewToolResultText(logs), nil
}
