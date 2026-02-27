package helm

import (
	"context"
	"fmt"
	"os"

	"github.com/mark3labs/mcp-go/mcp"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/repo"
)

func (t *Toolset) repoListTool() mcp.Tool {
	return mcp.NewTool(
		"helm_repo_list",
		mcp.WithDescription("List configured Helm chart repositories (from local config file)"),
		mcp.WithString("config_path", mcp.Description("Path to repositories.yaml (default: ~/.config/helm/repositories.yaml)")),
		mcp.WithString("format", mcp.Description("Output format: json, table (default: json)")),
	)
}

func (t *Toolset) repoListHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	configPath := req.GetString("config_path", "")
	format := req.GetString("format", "json")
	if configPath == "" {
		settings := cli.New()
		configPath = settings.RepositoryConfig
	}

	f, err := repo.LoadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			// No repo config - return empty list
			out, _ := t.formatter.Format([]map[string]interface{}{}, format)
			return mcp.NewToolResultText(out), nil
		}
		return mcp.NewToolResultError(fmt.Sprintf("helm repo list: %v", err)), nil
	}
	items := make([]map[string]interface{}, 0, len(f.Repositories))
	for _, r := range f.Repositories {
		items = append(items, map[string]interface{}{
			"name": r.Name,
			"url":  r.URL,
		})
	}
	out, err := t.formatter.Format(items, format)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("format: %v", err)), nil
	}
	return mcp.NewToolResultText(out), nil
}
