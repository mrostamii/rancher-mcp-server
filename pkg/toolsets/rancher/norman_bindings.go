package rancher

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/mark3labs/mcp-go/mcp"
)

func (t *Toolset) normanGlobalRoleBindingListTool() mcp.Tool {
	return mcp.NewTool(
		"rancher_global_role_binding_list",
		mcp.WithDescription("List global role bindings (/v3/globalrolebindings)"),
		mcp.WithNumber("limit", mcp.Description("Page size")),
		mcp.WithString("marker", mcp.Description("Pagination marker")),
	)
}

func (t *Toolset) normanGlobalRoleBindingListHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	q := url.Values{}
	if lim := req.GetInt("limit", 0); lim > 0 {
		q.Set("limit", fmt.Sprintf("%d", lim))
	}
	if m := req.GetString("marker", ""); m != "" {
		q.Set("marker", m)
	}
	raw, status, err := t.normanDo(ctx, http.MethodGet, "globalrolebindings", q, nil)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return t.normanTextResult(raw, status)
}

func (t *Toolset) normanGlobalRoleBindingGetTool() mcp.Tool {
	return mcp.NewTool(
		"rancher_global_role_binding_get",
		mcp.WithDescription("Get a global role binding by id (/v3/globalrolebindings/{id})"),
		mcp.WithString("id", mcp.Required(), mcp.Description("Binding id")),
	)
}

func (t *Toolset) normanGlobalRoleBindingGetHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id, err := req.RequireString("id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	raw, status, err := t.normanDo(ctx, http.MethodGet, "globalrolebindings/"+url.PathEscape(id), nil, nil)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return t.normanTextResult(raw, status)
}

func (t *Toolset) normanGlobalRoleBindingCreateTool() mcp.Tool {
	return mcp.NewTool(
		"rancher_global_role_binding_create",
		mcp.WithDescription("Create a global role binding (POST /v3/globalrolebindings)"),
		mcp.WithString("body", mcp.Required(), mcp.Description("JSON body")),
	)
}

func (t *Toolset) normanGlobalRoleBindingCreateHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if err := t.policy.CheckWrite(); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	body, err := req.RequireString("body")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	raw, status, err := t.normanDo(ctx, http.MethodPost, "globalrolebindings", nil, []byte(body))
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return t.normanTextResult(raw, status)
}

func (t *Toolset) normanGlobalRoleBindingDeleteTool() mcp.Tool {
	return mcp.NewTool(
		"rancher_global_role_binding_delete",
		mcp.WithDescription("Delete a global role binding (DELETE /v3/globalrolebindings/{id})"),
		mcp.WithString("id", mcp.Required(), mcp.Description("Binding id")),
	)
}

func (t *Toolset) normanGlobalRoleBindingDeleteHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if err := t.policy.CheckDestructive(); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	id, err := req.RequireString("id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	raw, status, err := t.normanDo(ctx, http.MethodDelete, "globalrolebindings/"+url.PathEscape(id), nil, nil)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return t.normanTextResult(raw, status)
}
