package rancher

import (
	"context"
	"net/http"
	"net/url"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
)

func (t *Toolset) normanActionTool() mcp.Tool {
	return mcp.NewTool(
		"rancher_norman_action",
		mcp.WithDescription("POST a Norman action on a resource, e.g. generateSupportBundle on a cluster (?action=...). resource_path is relative to /v3 (e.g. clusters/c-abc123)."),
		mcp.WithString("resource_path", mcp.Required(), mcp.Description("Path under /v3 without leading slash, e.g. clusters/c-xxxxx")),
		mcp.WithString("action", mcp.Required(), mcp.Description("Action name, e.g. generateSupportBundle")),
		mcp.WithString("body", mcp.Description("Optional JSON body for the action")),
	)
}

func (t *Toolset) normanActionHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if err := t.policy.CheckWrite(); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	rp, err := req.RequireString("resource_path")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	action, err := req.RequireString("action")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	rp = strings.Trim(rp, "/")
	bodyStr := req.GetString("body", "")
	var body []byte
	if strings.TrimSpace(bodyStr) != "" {
		body = []byte(bodyStr)
	}
	q := url.Values{}
	q.Set("action", action)
	raw, status, err := t.normanDo(ctx, http.MethodPost, rp, q, body)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return t.normanTextResult(raw, status)
}
