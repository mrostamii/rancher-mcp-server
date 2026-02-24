# rancher-mcp-server

Model Context Protocol (MCP) server for the **Rancher ecosystem**: multi-cluster Kubernetes, Harvester HCI (VMs, storage, networks), and Fleet GitOps.

## Features

- **Harvester toolset**: List/get VMs, images, volumes, networks, hosts; VM actions; addon list/switch (enable/disable)
- **Rancher toolset**: List clusters and projects, cluster get, overview (management API)
- **Kubernetes toolset**: List/get/create/patch/delete resources by apiVersion/kind; describe (resource + events), events, capacity
- **Rancher Steve API**: Single token, multi-cluster access; no CLI wrappers
- **Security**: Read-only default, disable-destructive, sensitive data masking
- **Config**: Flags, env (`RANCHER_MCP_*`), or file (YAML/TOML)

## Quick start

### Install

```bash
npm install -g rancher-mcp-server
```

### Cursor

Add to `.cursor/mcp.json` (project-level) or `~/.cursor/mcp.json` (global):

```json
{
  "mcpServers": {
    "rancher": {
      "command": "npx",
      "args": [
        "-y", "rancher-mcp-server",
        "--rancher-server-url", "https://rancher.example.com",
        "--rancher-token", "token-xxxxx:yyyy",
        "--toolsets", "harvester,rancher,kubernetes"
      ]
    }
  }
}
```

Restart Cursor after saving. Check **Settings → Tools & MCP** that **rancher** is listed and enabled.

### Claude Desktop

Add to your Claude Desktop config (`claude_desktop_config.json`):

```json
{
  "mcpServers": {
    "rancher": {
      "command": "npx",
      "args": [
        "-y", "rancher-mcp-server",
        "--rancher-server-url", "https://rancher.example.com",
        "--rancher-token", "token-xxxxx:yyyy",
        "--toolsets", "harvester,rancher,kubernetes"
      ]
    }
  }
}
```

### With env vars instead of args

If you prefer to keep the token out of the JSON config:

```json
{
  "mcpServers": {
    "rancher": {
      "command": "npx",
      "args": ["-y", "rancher-mcp-server"],
      "env": {
        "RANCHER_MCP_RANCHER_SERVER_URL": "https://rancher.example.com",
        "RANCHER_MCP_RANCHER_TOKEN": "token-xxxxx:yyyy",
        "RANCHER_MCP_TOOLSETS": "harvester,rancher,kubernetes"
      }
    }
  }
}
```

### Enable write operations

For VM create, snapshots, backups, image/volume create, addon switch, and Kubernetes create/patch/delete, add `--read-only=false`:

```json
{
  "mcpServers": {
    "rancher": {
      "command": "npx",
      "args": [
        "-y", "rancher-mcp-server",
        "--rancher-server-url", "https://rancher.example.com",
        "--rancher-token", "token-xxxxx:yyyy",
        "--toolsets", "harvester,rancher,kubernetes",
        "--read-only=false"
      ]
    }
  }
}
```

### HTTP/SSE transport

For web clients or remote access, add `--transport` and `--port`:

```json
{
  "mcpServers": {
    "rancher": {
      "command": "npx",
      "args": [
        "-y", "rancher-mcp-server",
        "--rancher-server-url", "https://rancher.example.com",
        "--rancher-token", "token-xxxxx:yyyy",
        "--transport", "http",
        "--port", "8080"
      ]
    }
  }
}
```

The server exposes the MCP endpoint over HTTP/SSE (e.g. `http://localhost:8080/sse`).

### Build from source

If you prefer to build the Go binary yourself:

```bash
go build -o rancher-mcp-server ./cmd/rancher-mcp-server
```

Then reference the binary directly in your MCP config:

```json
{
  "mcpServers": {
    "rancher": {
      "command": "/absolute/path/to/rancher-mcp-server",
      "args": [
        "--rancher-server-url", "https://rancher.example.com",
        "--rancher-token", "token-xxxxx:yyyy",
        "--toolsets", "harvester,rancher,kubernetes"
      ]
    }
  }
}
```

---

## Configuration


| Option                        | Env                                     | Default   | Description                                                               |
| ----------------------------- | --------------------------------------- | --------- | ------------------------------------------------------------------------- |
| `--rancher-server-url`        | `RANCHER_MCP_RANCHER_SERVER_URL`        | —         | Rancher server URL (required)                                             |
| `--rancher-token`             | `RANCHER_MCP_RANCHER_TOKEN`             | —         | Bearer token (required)                                                   |
| `--tls-insecure`              | `RANCHER_MCP_TLS_INSECURE`              | false     | Skip TLS verification                                                     |
| `--read-only`                 | `RANCHER_MCP_READ_ONLY`                 | true      | Disable write operations                                                  |
| `--disable-destructive`       | `RANCHER_MCP_DISABLE_DESTRUCTIVE`       | false     | Disable delete operations                                                 |
| `--toolsets`                  | `RANCHER_MCP_TOOLSETS`                  | harvester | Toolsets to enable: harvester, rancher, kubernetes                        |
| `--transport`                 | `RANCHER_MCP_TRANSPORT`                | stdio     | Transport: stdio or http (HTTP/SSE)                                      |
| `--port`                      | `RANCHER_MCP_PORT`                     | 0         | Port for HTTP/SSE (0 = stdio only)                                        |

---

## Harvester tools


| Tool                      | Description                                                       |
| ------------------------- | ----------------------------------------------------------------- |
| `harvester_vm_list`       | List VMs with status, namespace, spec/status                      |
| `harvester_vm_get`        | Get one VM (full spec and status)                                 |
| `harvester_vm_action`     | start, stop, restart, pause, unpause, migrate                     |
| `harvester_vm_create`     | Create VM (when not read-only)                                    |
| `harvester_vm_snapshot`   | Create/list/restore/delete VM snapshots                            |
| `harvester_vm_backup`     | Create/list/restore VM backups (Backup Target)                    |
| `harvester_image_list`    | List VM images (VirtualMachineImage)                              |
| `harvester_image_create`  | Create VM image from URL (when not read-only)                      |
| `harvester_volume_list`   | List PVCs (Longhorn-backed volumes)                               |
| `harvester_volume_create` | Create volume/PVC (optionally from image)                          |
| `harvester_network_list`  | List NetworkAttachmentDefinition (VLANs)                          |
| `harvester_host_list`     | List nodes (Harvester hosts)                                      |
| `harvester_addon_list`    | List Harvester addons (enabled/disabled state)                     |
| `harvester_addon_switch`  | Enable or disable an addon (when not read-only)                   |

List tools accept `cluster` (required), `namespace`, `format` (json|table), `limit` (default 100). Write tools require `read_only: false`.

## Rancher tools


| Tool                   | Description                                   |
| ---------------------- | --------------------------------------------- |
| `rancher_cluster_list` | List Rancher clusters (management)            |
| `rancher_cluster_get`  | Get one cluster (health, version, node count) |
| `rancher_project_list` | List Rancher projects                         |
| `rancher_overview`     | Overview: cluster count and project count     |


Uses Rancher management API (cluster ID `local`). No `cluster` param.

## Kubernetes tools


| Tool                  | Description                                                         |
| --------------------- | ------------------------------------------------------------------- |
| `kubernetes_list`     | List resources by apiVersion/kind (e.g. v1 Pod, apps/v1 Deployment) |
| `kubernetes_get`      | Get one resource by apiVersion, kind, namespace, name               |
| `kubernetes_describe` | Get resource + recent events                                        |
| `kubernetes_events`   | List events in a namespace (optional involvedObject filter)         |
| `kubernetes_capacity` | Node capacity/allocatable summary per node                          |
| `kubernetes_create`   | Create resource from JSON (when not read-only)                      |
| `kubernetes_patch`    | Patch resource with JSON (when not read-only)                       |
| `kubernetes_delete`   | Delete resource (when destructive allowed)                          |


All tools take `cluster` (Rancher cluster ID). List/get support `namespace`, `format` (json|table), `limit`. Create/patch/delete are gated by `read_only` and `disable_destructive`.

---

## Setup: Rancher token & Harvester cluster ID

### Get a Rancher API token

1. Log in to your Rancher UI.
2. Click your **profile/avatar** (top right) → **Account & API Keys** (or **API & Keys**).
3. Click **Create API Key**, name it (e.g. `mcp-server`), then **Create**.
4. Copy the token once (format like `token-abc12:xyz...`). Use it as `--rancher-token` or `RANCHER_MCP_RANCHER_TOKEN`.

### Find your Harvester cluster ID

Harvester tools require the **cluster ID** (e.g. `c-tx8rn`) on each call.

- **From Rancher UI:** Go to Cluster Management → open your Harvester cluster. The URL contains the cluster ID: `.../c/<cluster-id>/...`.
- **From API:** `curl -s -H "Authorization: Bearer YOUR_TOKEN" "https://YOUR_RANCHER_URL/v1/management.cattle.io.clusters" | jq '.data[] | {name: .metadata.name}'`

---

## Supported platforms

- macOS (Apple Silicon & Intel)
- Linux (x64 & ARM64)
- Windows (x64)

---

## Troubleshooting


| Issue                                               | What to check                                                                                        |
| --------------------------------------------------- | ---------------------------------------------------------------------------------------------------- |
| "rancher-server-url and rancher-token are required" | Check `--rancher-server-url` and `--rancher-token` in args, or env vars `RANCHER_MCP_RANCHER_SERVER_URL` and `RANCHER_MCP_RANCHER_TOKEN`. |
| 401 Unauthorized                                    | Token expired or invalid. Create a new API key in Rancher.                                           |
| TLS / certificate errors                            | For self-signed Rancher, pass `--tls-insecure` (dev only).                                           |
| "cluster not found" or empty lists                  | Wrong cluster ID. Get it from Rancher UI URL or API; pass it as `cluster` to Harvester/Kubernetes tools. |
| Cursor doesn't show tools                           | Restart Cursor after editing `mcp.json`; check **Tools & MCP** that the server is enabled.           |
| Binary not found                                    | Use **absolute** paths in `mcp.json` for `command` when building from source.                        |


---

## License

Apache-2.0
