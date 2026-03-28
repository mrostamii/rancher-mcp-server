package rancher

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/mark3labs/mcp-go/mcp"
)

func (t *Toolset) normanAuthConfigListTool() mcp.Tool {
	return mcp.NewTool(
		"rancher_auth_config_list",
		mcp.WithDescription("List auth configs (/v3/authconfigs)"),
		mcp.WithNumber("limit", mcp.Description("Page size")),
		mcp.WithString("marker", mcp.Description("Pagination marker")),
	)
}

func (t *Toolset) normanAuthConfigListHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	q := url.Values{}
	if lim := req.GetInt("limit", 0); lim > 0 {
		q.Set("limit", fmt.Sprintf("%d", lim))
	}
	if m := req.GetString("marker", ""); m != "" {
		q.Set("marker", m)
	}
	raw, status, err := t.normanDo(ctx, http.MethodGet, "authconfigs", q, nil)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return t.normanTextResult(raw, status)
}

func (t *Toolset) normanAuthConfigGetTool() mcp.Tool {
	return mcp.NewTool(
		"rancher_auth_config_get",
		mcp.WithDescription("Get an auth config by id (/v3/authconfigs/{id})"),
		mcp.WithString("id", mcp.Required(), mcp.Description("Auth config id (e.g. github, local)")),
	)
}

func (t *Toolset) normanAuthConfigGetHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id, err := req.RequireString("id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	raw, status, err := t.normanDo(ctx, http.MethodGet, "authconfigs/"+url.PathEscape(id), nil, nil)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return t.normanTextResult(raw, status)
}

func (t *Toolset) normanAuthConfigUpdateTool() mcp.Tool {
	return mcp.NewTool(
		"rancher_auth_config_update",
		mcp.WithDescription("Replace an auth config (PUT /v3/authconfigs/{id}). Body must be full JSON resource."),
		mcp.WithString("id", mcp.Required(), mcp.Description("Auth config id")),
		mcp.WithString("body", mcp.Required(), mcp.Description("JSON body")),
	)
}

func (t *Toolset) normanAuthConfigUpdateHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if err := t.policy.CheckWrite(); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	id, err := req.RequireString("id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	body, err := req.RequireString("body")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	raw, status, err := t.normanDo(ctx, http.MethodPut, "authconfigs/"+url.PathEscape(id), nil, []byte(body))
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return t.normanTextResult(raw, status)
}
