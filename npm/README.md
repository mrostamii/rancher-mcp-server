# rancher-mcp-server

Model Context Protocol (MCP) server for the **Rancher ecosystem**: multi-cluster Kubernetes, Harvester HCI (VMs, storage, networks), and Fleet GitOps.

## Install

```bash
npm install -g rancher-mcp-server
```

Or run directly with npx:

```bash
npx rancher-mcp-server --rancher-server-url https://rancher.example.com --rancher-token 'token-xxxxx:yyyy'
```

## Usage with Cursor / Claude Desktop

Add to `.cursor/mcp.json` (or Claude Desktop config):

```json
{
  "mcpServers": {
    "rancher": {
      "command": "npx",
      "args": [
        "-y", "rancher-mcp-server",
        "--rancher-server-url", "https://rancher.example.com",
        "--rancher-token", "token-xxxxx:yyyy",
        "--toolsets", "harvester,rancher,kubernetes,fleet"
      ]
    }
  }
}
```

If you prefer env vars:

```json
{
  "mcpServers": {
    "rancher": {
      "command": "npx",
      "args": ["-y", "rancher-mcp-server"],
      "env": {
        "RANCHER_MCP_RANCHER_SERVER_URL": "https://rancher.example.com",
        "RANCHER_MCP_RANCHER_TOKEN": "token-xxxxx:yyyy",
        "RANCHER_MCP_TOOLSETS": "harvester,rancher,kubernetes,fleet"
      }
    }
  }
}
```

Enable write operations when needed:

```json
{
  "mcpServers": {
    "rancher": {
      "command": "npx",
      "args": [
        "-y", "rancher-mcp-server",
        "--rancher-server-url", "https://rancher.example.com",
        "--rancher-token", "token-xxxxx:yyyy",
        "--toolsets", "harvester,rancher,kubernetes,fleet",
        "--read-only=false"
      ]
    }
  }
}
```

## Features

- **Harvester toolset**: VMs, snapshots, backups, images, volumes, networks, subnets, VPCs, hosts, addons
- **Rancher toolset**: Clusters/projects/overview (Steve); Norman `/v3` management API (schemas, users, tokens, auth, global role bindings, registration tokens, node drivers, cloud credentials, catalogs, cluster repos, feature flags, settings, audit when exposed); optional writes/deletes when flags allow
- **Kubernetes toolset**: List/get/create/patch/delete by `apiVersion`/`kind`, plus describe/events/capacity
- **Helm toolset**: List/get/history of releases; install, upgrade, rollback, uninstall; repo list
- **Fleet toolset**: GitRepo list/get/create/delete/action/clone; Bundle list; Fleet cluster list; drift detection
- **Security**: Read-only default, optional destructive-op guardrails, sensitive data masking (`--show-sensitive-data` to show Norman token/credential fields)

The [main README](https://github.com/mrostamii/rancher-mcp-server/blob/main/README.md) lists every Rancher tool (Steve + Norman), write/delete gating, and notes on catalog cluster repos and `unavailable` responses.

## Configuration

| Option | Env | Default | Description |
|---|---|---|---|
| `--rancher-server-url` | `RANCHER_MCP_RANCHER_SERVER_URL` | — | Rancher server URL (required) |
| `--rancher-token` | `RANCHER_MCP_RANCHER_TOKEN` | — | Bearer token (required) |
| `--tls-insecure` | `RANCHER_MCP_TLS_INSECURE` | false | Skip TLS verification |
| `--read-only` | `RANCHER_MCP_READ_ONLY` | true | Disable write operations |
| `--disable-destructive` | `RANCHER_MCP_DISABLE_DESTRUCTIVE` | false | Disable delete operations |
| `--show-sensitive-data` | `RANCHER_MCP_SHOW_SENSITIVE_DATA` | false | Show Norman token/credential fields without redaction |
| `--toolsets` | `RANCHER_MCP_TOOLSETS` | harvester | Toolsets: harvester, rancher, kubernetes, helm, fleet |
| `--transport` | `RANCHER_MCP_TRANSPORT` | stdio | Transport: stdio or http (Streamable HTTP; default path `/mcp`) |
| `--port` | `RANCHER_MCP_PORT` | 0 | Port for HTTP |

## Supported platforms

- macOS (Apple Silicon & Intel)
- Linux (x64 & ARM64)
- Windows (x64)

## License

Apache-2.0

## Links

- [GitHub](https://github.com/mrostamii/rancher-mcp-server)
- [Documentation](https://github.com/mrostamii/rancher-mcp-server/blob/main/README.md)
- [Issues](https://github.com/mrostamii/rancher-mcp-server/issues)
