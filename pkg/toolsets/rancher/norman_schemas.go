package rancher

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/mark3labs/mcp-go/mcp"
)

func (t *Toolset) normanSchemaListTool() mcp.Tool {
	return mcp.NewTool(
		"rancher_norman_schema_list",
		mcp.WithDescription("List Norman /v3 API schemas (discover types, collection links, and actions for this Rancher)"),
		mcp.WithNumber("limit", mcp.Description("Page size (optional; depends on Rancher version)")),
		mcp.WithString("marker", mcp.Description("Pagination marker from a previous response")),
	)
}

func (t *Toolset) normanSchemaListHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	q := url.Values{}
	if lim := req.GetInt("limit", 0); lim > 0 {
		q.Set("limit", fmt.Sprintf("%d", lim))
	}
	if m := req.GetString("marker", ""); m != "" {
		q.Set("marker", m)
	}
	raw, status, err := t.normanDo(ctx, http.MethodGet, "schemas", q, nil)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return t.normanTextResult(raw, status)
}

func (t *Toolset) normanSchemaGetTool() mcp.Tool {
	return mcp.NewTool(
		"rancher_norman_schema_get",
		mcp.WithDescription("Get a single Norman schema by id (e.g. user, token, cluster) from /v3/schemas/{id}"),
		mcp.WithString("schema_id", mcp.Required(), mcp.Description("Schema id from rancher_norman_schema_list")),
	)
}

func (t *Toolset) normanSchemaGetHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id, err := req.RequireString("schema_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	path := "schemas/" + url.PathEscape(id)
	raw, status, err := t.normanDo(ctx, http.MethodGet, path, nil, nil)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return t.normanTextResult(raw, status)
}
