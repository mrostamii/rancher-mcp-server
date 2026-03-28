package rancher

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/mark3labs/mcp-go/mcp"
)

func (t *Toolset) normanClusterRegistrationTokenListTool() mcp.Tool {
	return mcp.NewTool(
		"rancher_cluster_registration_token_list",
		mcp.WithDescription("List cluster registration tokens (/v3/clusterregistrationtokens). Optional cluster_id filters by cluster."),
		mcp.WithString("cluster_id", mcp.Description("Filter by cluster id (query param clusterId)")),
		mcp.WithNumber("limit", mcp.Description("Page size")),
		mcp.WithString("marker", mcp.Description("Pagination marker")),
	)
}

func (t *Toolset) normanClusterRegistrationTokenListHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	q := url.Values{}
	if cid := req.GetString("cluster_id", ""); cid != "" {
		q.Set("clusterId", cid)
	}
	if lim := req.GetInt("limit", 0); lim > 0 {
		q.Set("limit", fmt.Sprintf("%d", lim))
	}
	if m := req.GetString("marker", ""); m != "" {
		q.Set("marker", m)
	}
	raw, status, err := t.normanDo(ctx, http.MethodGet, "clusterregistrationtokens", q, nil)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return t.normanTextResult(raw, status)
}

func (t *Toolset) normanClusterRegistrationTokenGetTool() mcp.Tool {
	return mcp.NewTool(
		"rancher_cluster_registration_token_get",
		mcp.WithDescription("Get a cluster registration token by id"),
		mcp.WithString("id", mcp.Required(), mcp.Description("Token resource id")),
	)
}

func (t *Toolset) normanClusterRegistrationTokenGetHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id, err := req.RequireString("id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	raw, status, err := t.normanDo(ctx, http.MethodGet, "clusterregistrationtokens/"+url.PathEscape(id), nil, nil)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return t.normanTextResult(raw, status)
}

func (t *Toolset) normanClusterRegistrationTokenCreateTool() mcp.Tool {
	return mcp.NewTool(
		"rancher_cluster_registration_token_create",
		mcp.WithDescription("Create a cluster registration token (POST /v3/clusterregistrationtokens). Body must include clusterId. Token value redacted unless show-sensitive-data."),
		mcp.WithString("body", mcp.Required(), mcp.Description("JSON body")),
	)
}

func (t *Toolset) normanClusterRegistrationTokenCreateHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if err := t.policy.CheckWrite(); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	body, err := req.RequireString("body")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	raw, status, err := t.normanDo(ctx, http.MethodPost, "clusterregistrationtokens", nil, []byte(body))
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return t.normanTextResult(raw, status)
}

func (t *Toolset) normanClusterRegistrationTokenDeleteTool() mcp.Tool {
	return mcp.NewTool(
		"rancher_cluster_registration_token_delete",
		mcp.WithDescription("Delete a cluster registration token"),
		mcp.WithString("id", mcp.Required(), mcp.Description("Token resource id")),
	)
}

func (t *Toolset) normanClusterRegistrationTokenDeleteHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if err := t.policy.CheckDestructive(); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	id, err := req.RequireString("id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	raw, status, err := t.normanDo(ctx, http.MethodDelete, "clusterregistrationtokens/"+url.PathEscape(id), nil, nil)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return t.normanTextResult(raw, status)
}
