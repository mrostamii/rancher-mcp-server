package cmd

import (
	"fmt"

	"github.com/mark3labs/mcp-go/server"
	"github.com/mrostamii/rancher-mcp-server/internal/config"
	"github.com/mrostamii/rancher-mcp-server/internal/security"
	"github.com/mrostamii/rancher-mcp-server/pkg/client/rancher"
	harvesterToolset "github.com/mrostamii/rancher-mcp-server/pkg/toolsets/harvester"
	kubernetesToolset "github.com/mrostamii/rancher-mcp-server/pkg/toolsets/kubernetes"
	rancherToolset "github.com/mrostamii/rancher-mcp-server/pkg/toolsets/rancher"
)

const version = "0.1.0"

func runServe(cfg *config.Config) error {
	if cfg.RancherServerURL == "" || cfg.RancherToken == "" {
		return fmt.Errorf("rancher-server-url and rancher-token are required (or set RANCHER_MCP_RANCHER_SERVER_URL and RANCHER_MCP_RANCHER_TOKEN)")
	}

	s := server.NewMCPServer(
		"rancher-mcp-server",
		version,
		server.WithToolCapabilities(true),
		server.WithRecovery(),
	)

	policy := &security.Policy{
		ReadOnly:           cfg.ReadOnly,
		DisableDestructive: cfg.DisableDestructive,
		ShowSensitiveData:  cfg.ShowSensitiveData,
	}

	steveClient := rancher.NewSteveClient(cfg.RancherServerURL, cfg.RancherToken, cfg.TLSInsecure)

	for _, name := range cfg.Toolsets {
		switch name {
		case "harvester":
			harvesterToolset.NewToolset(steveClient, policy).Register(s)
		case "rancher":
			rancherToolset.NewToolset(steveClient, policy).Register(s)
		case "kubernetes":
			kubernetesToolset.NewToolset(steveClient, policy).Register(s)
		}
	}

	return server.ServeStdio(s)
}
