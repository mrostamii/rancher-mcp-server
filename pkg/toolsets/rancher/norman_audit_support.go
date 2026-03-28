package rancher

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/mark3labs/mcp-go/mcp"
)

func (t *Toolset) normanAuditLogListTool() mcp.Tool {
	return mcp.NewTool(
		"rancher_audit_log_list",
		mcp.WithDescription("Attempts GET /v3/auditlogs (Norman). Many Rancher builds return 404 (not exposed) or 405 (GET not supported); in those cases the tool returns JSON with _source unavailable, _http_status, and an explanatory _note (e.g. GET not supported for 405) instead of failing."),
		mcp.WithNumber("limit", mcp.Description("Page size")),
		mcp.WithString("marker", mcp.Description("Pagination marker")),
	)
}

func (t *Toolset) normanAuditLogListHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	q := url.Values{}
	if lim := req.GetInt("limit", 0); lim > 0 {
		q.Set("limit", fmt.Sprintf("%d", lim))
	}
	if m := req.GetString("marker", ""); m != "" {
		q.Set("marker", m)
	}
	raw, status, err := t.normanDo(ctx, http.MethodGet, "auditlogs", q, nil)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	if status == http.StatusOK {
		return t.normanTextResult(raw, status)
	}
	if status == http.StatusNotFound || status == http.StatusMethodNotAllowed {
		note := "Norman /v3/auditlogs is not registered (404). Audit data may be disabled or exported elsewhere (e.g. log aggregation)."
		if status == http.StatusMethodNotAllowed {
			note = "GET is not supported for listing Norman /v3/auditlogs on this Rancher (HTTP 405). The collection may not expose a list via GET; audit data may be exported elsewhere (e.g. log aggregation)."
		}
		out := map[string]interface{}{
			"_source":      "unavailable",
			"_http_status": status,
			"_note":        note,
			"rancher_body": string(raw),
			"data":         []interface{}{},
		}
		b, encErr := json.MarshalIndent(out, "", "  ")
		if encErr != nil {
			return mcp.NewToolResultError(encErr.Error()), nil
		}
		return mcp.NewToolResultText(string(b)), nil
	}
	return t.normanTextResult(raw, status)
}

func (t *Toolset) normanSupportBundleGenerateTool() mcp.Tool {
	return mcp.NewTool(
		"rancher_supportbundle_generate",
		mcp.WithDescription("Request a support bundle for a downstream cluster (POST /v3/clusters/{cluster_id}?action=generateSupportBundle). Exact action name may vary by Rancher version; use rancher_norman_action if this fails."),
		mcp.WithString("cluster_id", mcp.Required(), mcp.Description("Downstream cluster id (c-xxxxx)")),
		mcp.WithString("body", mcp.Description("Optional JSON body for the action")),
	)
}

func (t *Toolset) normanSupportBundleGenerateHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if err := t.policy.CheckWrite(); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	cid, err := req.RequireString("cluster_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	bodyStr := req.GetString("body", "{}")
	if bodyStr == "" {
		bodyStr = "{}"
	}
	q := url.Values{}
	q.Set("action", "generateSupportBundle")
	path := "clusters/" + url.PathEscape(cid)
	raw, status, err := t.normanDo(ctx, http.MethodPost, path, q, []byte(bodyStr))
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return t.normanTextResult(raw, status)
}
