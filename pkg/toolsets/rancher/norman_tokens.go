package rancher

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/mark3labs/mcp-go/mcp"
)

func (t *Toolset) normanTokenListTool() mcp.Tool {
	return mcp.NewTool(
		"rancher_token_list",
		mcp.WithDescription("List Norman API tokens (/v3/tokens). Sensitive token values are redacted unless show-sensitive-data is enabled."),
		mcp.WithNumber("limit", mcp.Description("Page size")),
		mcp.WithString("marker", mcp.Description("Pagination marker")),
	)
}

func (t *Toolset) normanTokenListHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	q := url.Values{}
	if lim := req.GetInt("limit", 0); lim > 0 {
		q.Set("limit", fmt.Sprintf("%d", lim))
	}
	if m := req.GetString("marker", ""); m != "" {
		q.Set("marker", m)
	}
	raw, status, err := t.normanDo(ctx, http.MethodGet, "tokens", q, nil)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return t.normanTextResult(raw, status)
}

func (t *Toolset) normanTokenGetTool() mcp.Tool {
	return mcp.NewTool(
		"rancher_token_get",
		mcp.WithDescription("Get a Norman token by id (/v3/tokens/{id}). Token value redacted unless show-sensitive-data is enabled."),
		mcp.WithString("id", mcp.Required(), mcp.Description("Token id")),
	)
}

func (t *Toolset) normanTokenGetHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id, err := req.RequireString("id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	raw, status, err := t.normanDo(ctx, http.MethodGet, "tokens/"+url.PathEscape(id), nil, nil)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return t.normanTextResult(raw, status)
}

func (t *Toolset) normanTokenCreateTool() mcp.Tool {
	return mcp.NewTool(
		"rancher_token_create",
		mcp.WithDescription("Create a Norman API token (POST /v3/tokens). Body is JSON per Rancher schema for token."),
		mcp.WithString("body", mcp.Required(), mcp.Description("JSON body")),
	)
}

func (t *Toolset) normanTokenCreateHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if err := t.policy.CheckWrite(); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	body, err := req.RequireString("body")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	raw, status, err := t.normanDo(ctx, http.MethodPost, "tokens", nil, []byte(body))
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return t.normanTextResult(raw, status)
}

func (t *Toolset) normanTokenDeleteTool() mcp.Tool {
	return mcp.NewTool(
		"rancher_token_delete",
		mcp.WithDescription("Delete a Norman API token (DELETE /v3/tokens/{id})"),
		mcp.WithString("id", mcp.Required(), mcp.Description("Token id")),
	)
}

func (t *Toolset) normanTokenDeleteHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if err := t.policy.CheckDestructive(); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	id, err := req.RequireString("id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	raw, status, err := t.normanDo(ctx, http.MethodDelete, "tokens/"+url.PathEscape(id), nil, nil)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return t.normanTextResult(raw, status)
}
