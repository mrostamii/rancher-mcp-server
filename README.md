# rancher-mcp-server

**Source:** [github.com/mrostamii/rancher-mcp-server](https://github.com/mrostamii/rancher-mcp-server)

Model Context Protocol (MCP) server for the **Rancher ecosystem**: multi-cluster Kubernetes, Harvester HCI (VMs, storage, networks), and Fleet GitOps.

## Features

- **Harvester toolset**: List/get VMs, images, volumes, networks, hosts; VM actions (start/stop/restart/pause/unpause/migrate)
- **Rancher toolset**: List clusters and projects, cluster get, overview (management API)
- **Kubernetes toolset**: List/get/create/patch/delete resources by apiVersion/kind; describe (resource + events), events, capacity
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


| Option                        | Env                                     | Default   | Description                                                               |
| ----------------------------- | --------------------------------------- | --------- | ------------------------------------------------------------------------- |
| `--rancher-server-url`        | `RANCHER_MCP_RANCHER_SERVER_URL`        | —         | Rancher server URL (required)                                             |
| `--rancher-token`             | `RANCHER_MCP_RANCHER_TOKEN`             | —         | Bearer token (required)                                                   |
| `--tls-insecure`              | `RANCHER_MCP_TLS_INSECURE`              | false     | Skip TLS verification                                                     |
| `--default-harvester-cluster` | `RANCHER_MCP_DEFAULT_HARVESTER_CLUSTER` | —         | Default Harvester cluster ID (e.g. c-tx8rn); used when tools omit cluster |
| `--read-only`                 | `RANCHER_MCP_READ_ONLY`                 | true      | Disable write operations                                                  |
| `--disable-destructive`       | `RANCHER_MCP_DISABLE_DESTRUCTIVE`       | false     | Disable delete operations                                                 |
| `--toolsets`                  | `RANCHER_MCP_TOOLSETS`                  | harvester | Toolsets to enable: harvester, rancher, kubernetes                        |


## Harvester tools


| Tool                     | Description                                   |
| ------------------------ | --------------------------------------------- |
| `harvester_vm_list`      | List VMs with status, namespace, spec/status  |
| `harvester_vm_get`       | Get one VM (full spec and status)             |
| `harvester_vm_action`    | start, stop, restart, pause, unpause, migrate |
| `harvester_image_list`   | List VM images (VirtualMachineImage)          |
| `harvester_volume_list`  | List PVCs (Longhorn-backed volumes)           |
| `harvester_network_list` | List NetworkAttachmentDefinition (VLANs)      |
| `harvester_host_list`    | List nodes (Harvester hosts)                  |


List tools accept `cluster` (optional if `default_harvester_cluster` is set in config), `namespace`, `format` (json|table), `limit` (default 100).

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
4. Copy the token once (format like `token-abc12:xyz...`). Use it as `RANCHER_MCP_RANCHER_TOKEN` or in `config.yaml` as `rancher_token`.

### Find your Harvester cluster ID

MCP tools need the **cluster ID** of the Harvester cluster (e.g. `c-tx8rn`).

- **From Rancher UI:** Go to Cluster Management → open your Harvester cluster. The URL contains the cluster ID: `.../c/<cluster-id>/...`.
- **From API:** `curl -s -H "Authorization: Bearer YOUR_TOKEN" "https://YOUR_RANCHER_URL/v1/management.cattle.io.clusters" | jq '.data[] | {name: .metadata.name}'`

Set it in config as `default_harvester_cluster` so you don’t have to pass it on every request:

```yaml
default_harvester_cluster: "c-tx8rn"
```

---

## Config file (recommended for Cursor)

```bash
cp config.example.yaml config.yaml
```

Edit `config.yaml`:

```yaml
rancher_server_url: https://YOUR_RANCHER_URL    # no trailing slash
rancher_token: token-xxxxx:yyyy                 # your API token
tls_insecure: false                             # set true only for self-signed certs

# Optional: default Harvester cluster (e.g. c-tx8rn). Tools use this when cluster is not passed.
default_harvester_cluster: "c-tx8rn"

transport: stdio
port: 0
log_level: 2
read_only: true
disable_destructive: false
show_sensitive_data: false
toolsets:
  - harvester
  - rancher
  - kubernetes
```

`config.yaml` is gitignored so your token is not committed.

---

## Cursor: add as MCP server

Cursor reads `**.cursor/mcp.json**` in the project (or `~/.cursor/mcp.json` globally).

1. **Create** `.cursor/mcp.json` (or copy from `.cursor/mcp.json.example`):
  ```bash
   cp .cursor/mcp.json.example .cursor/mcp.json
  ```
2. **Edit** `.cursor/mcp.json` and set **absolute** paths for the binary and config:
  ```json
   {
     "mcpServers": {
       "rancher": {
         "command": "/ABSOLUTE/PATH/TO/rancher-mcp-server/rancher-mcp-server",
         "args": ["--config", "/ABSOLUTE/PATH/TO/rancher-mcp-server/config.yaml"]
       }
     }
   }
  ```
3. **Restart Cursor** so it reloads MCP servers. Check **Settings** → **Tools & MCP** that **rancher** is enabled.

Then in a new chat you can ask e.g. “List Harvester VMs” or “List VM images”; the server uses your default cluster from config if you don’t pass one.

### Alternative: env in MCP config

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

---

## Troubleshooting


| Issue                                               | What to check                                                                                        |
| --------------------------------------------------- | ---------------------------------------------------------------------------------------------------- |
| “rancher-server-url and rancher-token are required” | Config path in `args`, or env vars `RANCHER_MCP_RANCHER_SERVER_URL` and `RANCHER_MCP_RANCHER_TOKEN`. |
| 401 Unauthorized                                    | Token expired or invalid. Create a new API key in Rancher.                                           |
| TLS / certificate errors                            | For self-signed Rancher, set `tls_insecure: true` in config (dev only).                              |
| “cluster not found” or empty lists                  | Wrong cluster ID. Get it from Rancher UI URL or API; set `default_harvester_cluster` in config.      |
| Cursor doesn’t show tools                           | Restart Cursor after editing `mcp.json`; check **Tools & MCP** that the server is enabled.           |
| Binary not found                                    | Use **absolute** paths in `mcp.json` for `command` and in `args` for `--config`.                     |


---

## License

Apache-2.0