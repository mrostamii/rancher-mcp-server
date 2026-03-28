package rancher

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	rancherapi "github.com/mrostamii/rancher-mcp-server/pkg/client/rancher"
)

// Primary namespace for global ClusterRepo CRs when Norman /v3/clusterrepos is absent.
const clusterRepoSteveNamespace = "cattle-global-data"

// Other namespaces where Fleet / Rancher sometimes store catalog ClusterRepos.
var clusterRepoTryNamespaces = []string{
	clusterRepoSteveNamespace,
	"fleet-default",
	"fleet-local",
	"cattle-fleet-system",
}

func (t *Toolset) normanCatalogListTool() mcp.Tool {
	return mcp.NewTool(
		"rancher_catalog_list",
		mcp.WithDescription("List legacy catalogs (Norman /v3/catalogs). On any non-success (including transport errors), returns JSON with _source unavailable instead of a hard tool error."),
		mcp.WithNumber("limit", mcp.Description("Page size")),
		mcp.WithString("marker", mcp.Description("Pagination marker")),
	)
}

func (t *Toolset) normanCatalogListHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	q := url.Values{}
	if lim := req.GetInt("limit", 0); lim > 0 {
		q.Set("limit", fmt.Sprintf("%d", lim))
	}
	if m := req.GetString("marker", ""); m != "" {
		q.Set("marker", m)
	}
	raw, status, err := t.normanDo(ctx, http.MethodGet, "catalogs", q, nil)
	if err != nil {
		return t.catalogListUnavailableResult(0, nil, err)
	}
	if status == http.StatusOK {
		return t.normanTextResult(raw, status)
	}
	return t.catalogListUnavailableResult(status, raw, nil)
}

// catalogListUnavailableResult returns JSON with _source unavailable (never a hard tool error) for legacy /v3/catalogs list.
func (t *Toolset) catalogListUnavailableResult(httpStatus int, rancherBody []byte, reqErr error) (*mcp.CallToolResult, error) {
	out := map[string]interface{}{
		"_source":    "unavailable",
		"suggestion": "Use rancher_cluster_repo_list (ClusterRepo) or kubernetes_list for catalog.cattle.io/v1 ClusterRepo on cluster local.",
		"data":       []interface{}{},
	}
	if reqErr != nil {
		out["_note"] = fmt.Sprintf("Norman /v3/catalogs request failed: %v", reqErr)
		out["attempt_errors"] = []string{reqErr.Error()}
	} else {
		if httpStatus > 0 {
			out["_http_status"] = httpStatus
		}
		switch httpStatus {
		case http.StatusNotFound:
			out["_note"] = "Norman /v3/catalogs is not registered on this Rancher (legacy catalog API)."
		default:
			out["_note"] = fmt.Sprintf("Norman /v3/catalogs returned HTTP %d; legacy catalog list is not usable for this request.", httpStatus)
		}
		if len(rancherBody) > 0 {
			out["rancher_body"] = string(rancherBody)
		}
	}
	b, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return mcp.NewToolResultText(string(b)), nil
}

func (t *Toolset) catalogLegacyUnavailableResult() (*mcp.CallToolResult, error) {
	return t.catalogListUnavailableResult(http.StatusNotFound, nil, nil)
}

func (t *Toolset) normanCatalogGetTool() mcp.Tool {
	return mcp.NewTool(
		"rancher_catalog_get",
		mcp.WithDescription("Get a catalog by id (/v3/catalogs/{id})"),
		mcp.WithString("id", mcp.Required(), mcp.Description("Catalog id")),
	)
}

func (t *Toolset) normanCatalogGetHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id, err := req.RequireString("id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	raw, status, err := t.normanDo(ctx, http.MethodGet, "catalogs/"+url.PathEscape(id), nil, nil)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	if status == http.StatusOK {
		return t.normanTextResult(raw, status)
	}
	if status == http.StatusNotFound {
		return t.catalogLegacyUnavailableResult()
	}
	return t.normanTextResult(raw, status)
}

func (t *Toolset) normanCatalogRefreshTool() mcp.Tool {
	return mcp.NewTool(
		"rancher_catalog_refresh",
		mcp.WithDescription("Refresh a catalog (POST /v3/catalogs/{id}?action=refresh). Action name may vary by Rancher version."),
		mcp.WithString("id", mcp.Required(), mcp.Description("Catalog id")),
	)
}

func (t *Toolset) normanCatalogRefreshHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if err := t.policy.CheckWrite(); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	id, err := req.RequireString("id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	q := url.Values{}
	q.Set("action", "refresh")
	raw, status, err := t.normanDo(ctx, http.MethodPost, "catalogs/"+url.PathEscape(id), q, []byte("{}"))
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return t.normanTextResult(raw, status)
}

func (t *Toolset) normanClusterRepoListTool() mcp.Tool {
	return mcp.NewTool(
		"rancher_cluster_repo_list",
		mcp.WithDescription("List app catalog cluster repos. Tries Norman /v3/clusterrepos; on failure or non-success, lists catalog.cattle.io/v1 ClusterRepo on local via Steve/Kubernetes, trying namespaces cattle-global-data, fleet-default, fleet-local, cattle-fleet-system. Returns either _source kubernetes_api_fallback with data or _source unavailable with attempt_errors (no hard tool error). Pagination marker applies only to the first successful namespace (usually cattle-global-data)."),
		mcp.WithNumber("limit", mcp.Description("Page size")),
		mcp.WithString("marker", mcp.Description("Pagination: Norman marker, or Kubernetes continue token when using the Steve/K8s fallback")),
	)
}

func (t *Toolset) normanClusterRepoListHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	q := url.Values{}
	limit := req.GetInt("limit", 0)
	marker := req.GetString("marker", "")
	if limit > 0 {
		q.Set("limit", fmt.Sprintf("%d", limit))
	}
	if marker != "" {
		q.Set("marker", marker)
	}
	raw, status, err := t.normanDo(ctx, http.MethodGet, "clusterrepos", q, nil)
	var priorErrs []string
	if err != nil {
		priorErrs = append(priorErrs, fmt.Sprintf("norman: %v", err))
		return t.clusterRepoListSteveFallback(ctx, limit, marker, priorErrs)
	}
	if status == http.StatusOK {
		return t.normanTextResult(raw, status)
	}
	if status != http.StatusNotFound {
		msg := fmt.Sprintf("norman: HTTP %d", status)
		if len(raw) > 0 {
			msg += ": " + truncateString(string(raw), 240)
		}
		priorErrs = append(priorErrs, msg)
	}
	return t.clusterRepoListSteveFallback(ctx, limit, marker, priorErrs)
}

func truncateString(s string, max int) string {
	if max <= 0 || len(s) <= max {
		return s
	}
	return s[:max] + "…"
}

func (t *Toolset) clusterRepoListSteveFallback(ctx context.Context, limit int, continueToken string, priorErrs []string) (*mcp.CallToolResult, error) {
	if limit <= 0 {
		limit = 100
	}

	// Pagination must stay on the same store/namespace as the first page.
	if continueToken != "" {
		col, err := t.client.List(ctx, localCluster, rancherapi.TypeCatalogClusterRepos, rancherapi.ListOpts{
			Limit:     limit,
			Continue:  continueToken,
			Namespace: clusterRepoSteveNamespace,
		})
		if err != nil {
			attempt := append(append([]string{}, priorErrs...), fmt.Sprintf("%s: %v", clusterRepoSteveNamespace, err))
			return t.clusterRepoUnavailableResult(
				"Pagination with the given continue token failed for catalog.cattle.io/v1 ClusterRepo on the local cluster.",
				attempt,
			)
		}
		return t.clusterRepoSuccessResult(clusterRepoSteveNamespace, col)
	}

	var errs []string
	for _, ns := range clusterRepoTryNamespaces {
		col, err := t.client.List(ctx, localCluster, rancherapi.TypeCatalogClusterRepos, rancherapi.ListOpts{
			Limit:     limit,
			Namespace: ns,
		})
		if err != nil {
			errs = append(errs, fmt.Sprintf("%s: %v", ns, err))
			continue
		}
		return t.clusterRepoSuccessResult(ns, col)
	}

	allErrs := append(append([]string{}, priorErrs...), errs...)
	return t.clusterRepoUnavailableResult(
		"Norman /v3/clusterrepos is not registered and catalog.cattle.io/v1 ClusterRepo could not be listed on the local cluster for the tried namespaces (CRD may be absent or API not proxied for this token).",
		allErrs,
	)
}

func (t *Toolset) clusterRepoSuccessResult(namespace string, col *rancherapi.SteveCollection) (*mcp.CallToolResult, error) {
	items := make([]map[string]interface{}, 0, len(col.Data))
	for _, r := range col.Data {
		items = append(items, map[string]interface{}{
			"name":      r.ObjectMeta.Name,
			"namespace": r.ObjectMeta.Namespace,
			"metadata":  r.ObjectMeta,
			"spec":      r.Spec,
			"status":    r.Status,
		})
	}
	out := map[string]interface{}{
		"_source":    "kubernetes_api_fallback",
		"_note":      fmt.Sprintf("Norman /v3/clusterrepos not available; listed catalog.cattle.io/v1 ClusterRepo in namespace %q.", namespace),
		"_namespace": namespace,
		"data":       items,
		"continue":   col.Continue,
	}
	b, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	outBytes, err := rancherapi.RedactNormanSecrets(t.policy.ShowSensitiveData, b)
	if err != nil {
		outBytes = b
	}
	return mcp.NewToolResultText(string(outBytes)), nil
}

func (t *Toolset) clusterRepoUnavailableResult(summary string, attemptErrs []string) (*mcp.CallToolResult, error) {
	out := map[string]interface{}{
		"_source":    "unavailable",
		"_note":      summary,
		"suggestion": "This Rancher may not install catalog.cattle.io ClusterRepos, or the management cluster API is not reachable as catalog CRDs for your token. Try kubernetes_list on cluster local with api_version catalog.cattle.io/v1 and kind ClusterRepo (various namespaces) or use Helm/Fleet tools.",
		"data":       []interface{}{},
	}
	if len(attemptErrs) > 0 {
		out["attempt_errors"] = attemptErrs
	}
	b, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return mcp.NewToolResultText(string(b)), nil
}

func (t *Toolset) normanClusterRepoGetTool() mcp.Tool {
	return mcp.NewTool(
		"rancher_cluster_repo_get",
		mcp.WithDescription("Get a cluster repo by id. Tries Norman /v3/clusterrepos/{id} first; on 404, reads catalog.cattle.io/v1 ClusterRepo from the local cluster. Use id \"namespace/name\" or \"name\" (defaults to namespace cattle-global-data)."),
		mcp.WithString("id", mcp.Required(), mcp.Description("Norman id or Kubernetes name; use namespace/name for fallback if not default")),
	)
}

func (t *Toolset) normanClusterRepoGetHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id, err := req.RequireString("id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	raw, status, err := t.normanDo(ctx, http.MethodGet, "clusterrepos/"+url.PathEscape(id), nil, nil)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	if status == http.StatusOK {
		return t.normanTextResult(raw, status)
	}
	if status == http.StatusNotFound {
		return t.clusterRepoGetSteveFallback(ctx, id)
	}
	return t.normanTextResult(raw, status)
}

func clusterRepoNamespaceAndName(id string) (namespace, name string) {
	id = strings.TrimSpace(id)
	if i := strings.Index(id, "/"); i >= 0 {
		return id[:i], id[i+1:]
	}
	if i := strings.Index(id, ":"); i >= 0 {
		return id[:i], id[i+1:]
	}
	return clusterRepoSteveNamespace, id
}

func (t *Toolset) clusterRepoGetSteveFallback(ctx context.Context, id string) (*mcp.CallToolResult, error) {
	explicitNS := strings.Contains(id, "/") || strings.Contains(id, ":")
	var name string
	var tryNSList []string
	if explicitNS {
		ns, n := clusterRepoNamespaceAndName(id)
		name = n
		tryNSList = []string{ns}
	} else {
		name = strings.TrimSpace(id)
		tryNSList = append([]string{}, clusterRepoTryNamespaces...)
	}
	var res *rancherapi.SteveResource
	var errs []string
	for _, tryNS := range tryNSList {
		r, err := t.client.Get(ctx, localCluster, rancherapi.TypeCatalogClusterRepos, tryNS, name)
		if err != nil {
			errs = append(errs, fmt.Sprintf("%s/%s: %v", tryNS, name, err))
			continue
		}
		res = r
		break
	}
	if res == nil {
		return t.clusterRepoUnavailableResult(
			fmt.Sprintf("Norman cluster repo %q returned 404 and catalog.cattle.io ClusterRepo get failed.", id),
			errs,
		)
	}
	out := map[string]interface{}{
		"_source":    "kubernetes_api_fallback",
		"_note":      "Norman /v3/clusterrepos not available; object is catalog.cattle.io/v1 ClusterRepo from the local cluster.",
		"name":       res.ObjectMeta.Name,
		"namespace":  res.ObjectMeta.Namespace,
		"metadata":   res.ObjectMeta,
		"spec":       res.Spec,
		"status":     res.Status,
	}
	b, encErr := json.MarshalIndent(out, "", "  ")
	if encErr != nil {
		return mcp.NewToolResultError(encErr.Error()), nil
	}
	outBytes, redErr := rancherapi.RedactNormanSecrets(t.policy.ShowSensitiveData, b)
	if redErr != nil {
		outBytes = b
	}
	return mcp.NewToolResultText(string(outBytes)), nil
}
