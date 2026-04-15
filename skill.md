---
name: rancher-mcp-server
description: Use when managing Rancher, Kubernetes, Harvester, Helm, or Fleet resources through this MCP server and you need safe, tool-oriented operational workflows.
---

# Rancher MCP Operator Skill

## What this server is for
- Operate Rancher ecosystems from AI tools using MCP.
- Manage clusters, projects, VMs, workloads, Helm releases, and Fleet GitOps resources.
- Use one Rancher API token for Steve (`/k8s/...`) and Norman (`/v3/...`) APIs.

## When to use each toolset
- `rancher_*`: Rancher management data (clusters, projects, users, tokens, settings, feature flags).
- `kubernetes_*`: downstream cluster resources by `apiVersion` and `kind`.
- `harvester_*`: VM, image, volume, network, subnet, host, addon, and VPC operations.
- `helm_*`: release lifecycle (`list`, `get`, `history`, `install`, `upgrade`, `rollback`, `uninstall`).
- `fleet_*`: GitRepo and bundle operations for GitOps/fleet state.

## Safety defaults and write gates
- Default mode is read-only.
- Write operations require server startup with `read-only=false`.
- Delete/destructive operations require both:
  - `read-only=false`
  - `disable-destructive=false`
- Sensitive Norman data is redacted unless `show-sensitive-data=true`.

## Cluster scoping rules
- `rancher_*` tools target management scope and typically do not need a `cluster` argument.
- `kubernetes_*`, `harvester_*`, and `helm_*` require `cluster` (Rancher cluster ID).
- `fleet_*` works from management scope (`local`) with optional `namespace`.

## API behavior that prevents confusion
- Steve endpoints can return 404 for some resources depending on Rancher setup.
- Native Kubernetes API proxy paths are generally reliable for downstream resources.
- Some Norman collections are missing on certain Rancher versions; tools may return `_source: "unavailable"` instead of hard failure.
- If one catalog/cluster repo path is unavailable, use Kubernetes/Helm/Fleet alternatives.

## Practical workflow (token-efficient)
1. Start with read/list tools to discover names, namespaces, and IDs.
2. Narrow with get/describe tools before any write action.
3. For paginated data, use `limit` and `continue`.
4. Switch to write tools only when server gates allow it.
5. Prefer smallest safe action first (patch/update before delete).

## Minimal troubleshooting checklist
- `401` or auth errors: verify Rancher token validity.
- Empty results / not found: verify correct cluster ID and namespace.
- TLS issues: enable `tls-insecure` only for trusted self-signed setups.
- Logs/stream proxy errors (e.g., intermittent `503`): reduce log scope and retry.
