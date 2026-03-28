package rancher

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/mark3labs/mcp-go/mcp"
)

func (t *Toolset) normanFeatureListTool() mcp.Tool {
	return mcp.NewTool(
		"rancher_feature_flag_list",
		mcp.WithDescription("List feature flags (/v3/features)"),
		mcp.WithNumber("limit", mcp.Description("Page size")),
		mcp.WithString("marker", mcp.Description("Pagination marker")),
	)
}

func (t *Toolset) normanFeatureListHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	q := url.Values{}
	if lim := req.GetInt("limit", 0); lim > 0 {
		q.Set("limit", fmt.Sprintf("%d", lim))
	}
	if m := req.GetString("marker", ""); m != "" {
		q.Set("marker", m)
	}
	raw, status, err := t.normanDo(ctx, http.MethodGet, "features", q, nil)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return t.normanTextResult(raw, status)
}

func (t *Toolset) normanFeatureGetTool() mcp.Tool {
	return mcp.NewTool(
		"rancher_feature_flag_get",
		mcp.WithDescription("Get a feature flag by id (/v3/features/{id})"),
		mcp.WithString("id", mcp.Required(), mcp.Description("Feature id")),
	)
}

func (t *Toolset) normanFeatureGetHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id, err := req.RequireString("id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	raw, status, err := t.normanDo(ctx, http.MethodGet, "features/"+url.PathEscape(id), nil, nil)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return t.normanTextResult(raw, status)
}

func (t *Toolset) normanFeatureSetTool() mcp.Tool {
	return mcp.NewTool(
		"rancher_feature_flag_set",
		mcp.WithDescription("Update a feature flag (PUT /v3/features/{id}). Body is full JSON resource."),
		mcp.WithString("id", mcp.Required(), mcp.Description("Feature id")),
		mcp.WithString("body", mcp.Required(), mcp.Description("JSON body")),
	)
}

func (t *Toolset) normanFeatureSetHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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
	raw, status, err := t.normanDo(ctx, http.MethodPut, "features/"+url.PathEscape(id), nil, []byte(body))
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return t.normanTextResult(raw, status)
}

func (t *Toolset) normanSettingListTool() mcp.Tool {
	return mcp.NewTool(
		"rancher_setting_list",
		mcp.WithDescription("List global Rancher settings (/v3/settings)"),
		mcp.WithNumber("limit", mcp.Description("Page size")),
		mcp.WithString("marker", mcp.Description("Pagination marker")),
	)
}

func (t *Toolset) normanSettingListHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	q := url.Values{}
	if lim := req.GetInt("limit", 0); lim > 0 {
		q.Set("limit", fmt.Sprintf("%d", lim))
	}
	if m := req.GetString("marker", ""); m != "" {
		q.Set("marker", m)
	}
	raw, status, err := t.normanDo(ctx, http.MethodGet, "settings", q, nil)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return t.normanTextResult(raw, status)
}

func (t *Toolset) normanSettingGetTool() mcp.Tool {
	return mcp.NewTool(
		"rancher_setting_get",
		mcp.WithDescription("Get a setting by id (/v3/settings/{id})"),
		mcp.WithString("id", mcp.Required(), mcp.Description("Setting id (e.g. server-url)")),
	)
}

func (t *Toolset) normanSettingGetHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id, err := req.RequireString("id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	raw, status, err := t.normanDo(ctx, http.MethodGet, "settings/"+url.PathEscape(id), nil, nil)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return t.normanTextResult(raw, status)
}

func (t *Toolset) normanSettingUpdateTool() mcp.Tool {
	return mcp.NewTool(
		"rancher_setting_update",
		mcp.WithDescription("Update a setting (PUT /v3/settings/{id}). Body is full JSON resource."),
		mcp.WithString("id", mcp.Required(), mcp.Description("Setting id")),
		mcp.WithString("body", mcp.Required(), mcp.Description("JSON body")),
	)
}

func (t *Toolset) normanSettingUpdateHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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
	raw, status, err := t.normanDo(ctx, http.MethodPut, "settings/"+url.PathEscape(id), nil, []byte(body))
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return t.normanTextResult(raw, status)
}
