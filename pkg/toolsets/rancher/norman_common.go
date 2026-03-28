package rancher

import (
	"context"
	"fmt"
	"net/url"

	"github.com/mark3labs/mcp-go/mcp"
	rancherapi "github.com/mrostamii/rancher-mcp-server/pkg/client/rancher"
)

func (t *Toolset) normanDo(ctx context.Context, method, path string, query url.Values, body []byte) ([]byte, int, error) {
	if t.norman == nil {
		return nil, 0, fmt.Errorf("Rancher Norman API client is not configured")
	}
	return t.norman.Do(ctx, method, path, query, body)
}

func (t *Toolset) normanTextResult(raw []byte, status int) (*mcp.CallToolResult, error) {
	if status < 200 || status >= 300 {
		return mcp.NewToolResultError(fmt.Sprintf("Rancher API returned HTTP %d: %s", status, string(raw))), nil
	}
	out, err := rancherapi.RedactNormanSecrets(t.policy.ShowSensitiveData, raw)
	if err != nil {
		out = raw
	}
	return mcp.NewToolResultText(string(out)), nil
}
