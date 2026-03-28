package rancher

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/mark3labs/mcp-go/mcp"
)

func (t *Toolset) normanUserListTool() mcp.Tool {
	return mcp.NewTool(
		"rancher_user_list",
		mcp.WithDescription("List Rancher users (/v3/users)"),
		mcp.WithNumber("limit", mcp.Description("Page size")),
		mcp.WithString("marker", mcp.Description("Pagination marker")),
	)
}

func (t *Toolset) normanUserListHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	q := url.Values{}
	if lim := req.GetInt("limit", 0); lim > 0 {
		q.Set("limit", fmt.Sprintf("%d", lim))
	}
	if m := req.GetString("marker", ""); m != "" {
		q.Set("marker", m)
	}
	raw, status, err := t.normanDo(ctx, http.MethodGet, "users", q, nil)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return t.normanTextResult(raw, status)
}

func (t *Toolset) normanUserGetTool() mcp.Tool {
	return mcp.NewTool(
		"rancher_user_get",
		mcp.WithDescription("Get a user by id (/v3/users/{id})"),
		mcp.WithString("id", mcp.Required(), mcp.Description("User id")),
	)
}

func (t *Toolset) normanUserGetHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id, err := req.RequireString("id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	raw, status, err := t.normanDo(ctx, http.MethodGet, "users/"+url.PathEscape(id), nil, nil)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return t.normanTextResult(raw, status)
}

func (t *Toolset) normanUserCreateTool() mcp.Tool {
	return mcp.NewTool(
		"rancher_user_create",
		mcp.WithDescription("Create a user (POST /v3/users). Body is JSON per Rancher user schema."),
		mcp.WithString("body", mcp.Required(), mcp.Description("JSON body")),
	)
}

func (t *Toolset) normanUserCreateHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if err := t.policy.CheckWrite(); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	body, err := req.RequireString("body")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	raw, status, err := t.normanDo(ctx, http.MethodPost, "users", nil, []byte(body))
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return t.normanTextResult(raw, status)
}

func (t *Toolset) normanUserDisableTool() mcp.Tool {
	return mcp.NewTool(
		"rancher_user_disable",
		mcp.WithDescription("Disable a user (POST /v3/users/{id}?action=disable)"),
		mcp.WithString("id", mcp.Required(), mcp.Description("User id")),
	)
}

func (t *Toolset) normanUserDisableHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if err := t.policy.CheckWrite(); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	id, err := req.RequireString("id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	q := url.Values{}
	q.Set("action", "disable")
	raw, status, err := t.normanDo(ctx, http.MethodPost, "users/"+url.PathEscape(id), q, []byte("{}"))
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return t.normanTextResult(raw, status)
}

func (t *Toolset) normanUserEnableTool() mcp.Tool {
	return mcp.NewTool(
		"rancher_user_enable",
		mcp.WithDescription("Enable a user (POST /v3/users/{id}?action=enable)"),
		mcp.WithString("id", mcp.Required(), mcp.Description("User id")),
	)
}

func (t *Toolset) normanUserEnableHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if err := t.policy.CheckWrite(); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	id, err := req.RequireString("id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	q := url.Values{}
	q.Set("action", "enable")
	raw, status, err := t.normanDo(ctx, http.MethodPost, "users/"+url.PathEscape(id), q, []byte("{}"))
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return t.normanTextResult(raw, status)
}
