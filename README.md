# rancher-mcp-server

Model Context Protocol (MCP) server for the **Rancher ecosystem**: multi-cluster Kubernetes, Harvester HCI (VMs, storage, networks), and Fleet GitOps — through a single configuration.

## Features

- **Harvester toolset**: List/get VMs, images, volumes, networks, hosts; VM actions (start/stop/restart/pause/unpause/migrate)
- **Rancher Steve API**: Single token, multi-cluster access; no CLI wrappers
- **Security**: Read-only default, disable-destructive, sensitive data masking
- **Config**: Flags, env (`RANCHER_MCP_*`), or file (YAML/TOML)

## Requirements

- Go 1.23+
- A Rancher server URL and API token (or access key)
- Harvester cluster(s) managed by Rancher (for Harvester tools)

## Quick start

1. **Build**
   ```bash
   go build -o rancher-mcp-server ./cmd/rancher-mcp-server
   ```

2. **Run (stdio — for Claude Desktop, Cursor, etc.)**
   ```bash
   ./rancher-mcp-server \
     --rancher-server-url https://rancher.example.com \
     --rancher-token 'token-xxxxx:yyyy' \
     --toolsets harvester
   ```

3. **Or use env vars**
   ```bash
   export RANCHER_MCP_RANCHER_SERVER_URL=https://rancher.example.com
   export RANCHER_MCP_RANCHER_TOKEN=token-xxxxx:yyyy
   ./rancher-mcp-server
   ```

4. **Or a config file**
   ```bash
   cp config.example.yaml config.yaml
   # Edit config.yaml with your Rancher URL and token
   ./rancher-mcp-server --config config.yaml
   ```

## Configuration

| Option | Env | Default | Description |
|--------|-----|---------|-------------|
| `--rancher-server-url` | `RANCHER_MCP_RANCHER_SERVER_URL` | — | Rancher server URL (required) |
| `--rancher-token` | `RANCHER_MCP_RANCHER_TOKEN` | — | Bearer token (required) |
| `--tls-insecure` | `RANCHER_MCP_TLS_INSECURE` | false | Skip TLS verification |
| `--read-only` | `RANCHER_MCP_READ_ONLY` | true | Disable write operations |
| `--disable-destructive` | `RANCHER_MCP_DISABLE_DESTRUCTIVE` | false | Disable delete operations |
| `--toolsets` | `RANCHER_MCP_TOOLSETS` | harvester | Toolsets to enable (currently only harvester) |

## Harvester tools

| Tool | Description |
|------|-------------|
| `harvester_vm_list` | List VMs with status, namespace, spec/status |
| `harvester_vm_get` | Get one VM (full spec and status) |
| `harvester_vm_action` | start, stop, restart, pause, unpause, migrate |
| `harvester_image_list` | List VM images (VirtualMachineImage) |
| `harvester_volume_list` | List PVCs (Longhorn-backed volumes) |
| `harvester_network_list` | List NetworkAttachmentDefinition (VLANs) |
| `harvester_host_list` | List nodes (Harvester hosts) |

All list tools accept `cluster` (required), `namespace` (optional), `format` (json|table), `limit` (default 100).

## Cursor / Claude Desktop

Add to your MCP config (e.g. Cursor `mcp.json` or Claude `claude_desktop_config.json`):

```json
{
  "mcpServers": {
    "rancher": {
      "command": "/path/to/rancher-mcp-server",
      "args": ["--config", "/path/to/config.yaml"]
    }
  }
}
```

Or with env:

```json
{
  "mcpServers": {
    "rancher": {
      "command": "/path/to/rancher-mcp-server",
      "env": {
        "RANCHER_MCP_RANCHER_SERVER_URL": "https://rancher.example.com",
        "RANCHER_MCP_RANCHER_TOKEN": "token-xxxxx:yyyy"
      }
    }
  }
}
```

## License

Apache-2.0
