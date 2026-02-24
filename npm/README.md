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

Add to your `.cursor/mcp.json` (or Claude Desktop config):

```json
{
  "mcpServers": {
    "rancher": {
      "command": "npx",
      "args": ["-y", "rancher-mcp-server"],
      "env": {
        "RANCHER_MCP_RANCHER_SERVER_URL": "https://rancher.example.com",
        "RANCHER_MCP_RANCHER_TOKEN": "token-xxxxx:yyyy"
      }
    }
  }
}
```

## Features

- **Harvester toolset**: List/get VMs, images, volumes, networks, hosts; VM actions (start/stop/restart/pause/unpause/migrate); VM create, snapshots, backups
- **Rancher toolset**: List clusters and projects, cluster get, overview
- **Kubernetes toolset**: List/get/create/patch/delete resources by apiVersion/kind; describe, events, capacity
- **Security**: Read-only default, disable-destructive, sensitive data masking

## Configuration

| Option | Env | Default | Description |
|---|---|---|---|
| `--rancher-server-url` | `RANCHER_MCP_RANCHER_SERVER_URL` | — | Rancher server URL (required) |
| `--rancher-token` | `RANCHER_MCP_RANCHER_TOKEN` | — | Bearer token (required) |
| `--tls-insecure` | `RANCHER_MCP_TLS_INSECURE` | false | Skip TLS verification |
| `--read-only` | `RANCHER_MCP_READ_ONLY` | true | Disable write operations |
| `--disable-destructive` | `RANCHER_MCP_DISABLE_DESTRUCTIVE` | false | Disable delete operations |
| `--toolsets` | `RANCHER_MCP_TOOLSETS` | harvester | Toolsets: harvester, rancher, kubernetes |
| `--transport` | `RANCHER_MCP_TRANSPORT` | stdio | Transport: stdio or http |
| `--port` | `RANCHER_MCP_PORT` | 0 | Port for HTTP/SSE |

## Supported Platforms

- macOS (Apple Silicon & Intel)
- Linux (x64 & ARM64)
- Windows (x64)

## License

Apache-2.0

## Links

- [GitHub](https://github.com/mrostamii/rancher-mcp-server)
- [Issues](https://github.com/mrostamii/rancher-mcp-server/issues)
