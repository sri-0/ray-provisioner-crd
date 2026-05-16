# Ray CRD Provisioner

Monorepo for dynamic Ray plugin provisioning on Kubernetes.

## What This Contains

- `services/operator`: Go Kubernetes operator and management API for the `RayPlugin` CRD.
- `apps/management-ui`: Vite React UI for listing, filtering, and enabling/disabling plugins.
- `packages/ray-framework`: Python framework skeleton for Ray orchestration and plugin DAG planning.
- `examples/plugins/csv-analyzer`: Example per-plugin repo layout with Dockerfile, CI, and `RayPlugin` manifest.
- `plans`: Architecture and implementation notes.

## Install Model

The install is split into two charts so day-to-day plugin operation does not require cluster-admin.

`RayPlugin` is a Kubernetes CRD, and CRDs are cluster-scoped resources. That means a cluster admin must install the CRD once. After that, each team or environment can run its own namespace-scoped operator using only namespaced RBAC.

| Chart | Scope | Installed By | Purpose |
| --- | --- | --- | --- |
| `services/operator/helm/ray-plugin-crds` | Cluster | Platform / cluster admin | Installs the `RayPlugin` CRD. |
| `services/operator/helm/ray-plugin-operator` | Namespace | Namespace owner | Runs the operator, API server, `Role`, and `RoleBinding`. |

Cluster admin installs CRDs once:

```bash
helm install ray-plugin-crds ./services/operator/helm/ray-plugin-crds
```

Namespace owners install the operator in their own namespace without cluster-admin, assuming KubeRay and `RayPlugin` CRDs already exist:

```bash
helm install ray-plugin-operator ./services/operator/helm/ray-plugin-operator \
  --namespace ray-system \
  --create-namespace
```

The namespace-scoped operator uses `Role` and `RoleBinding`, watches one namespace, and patches only `RayCluster` resources it can access in that namespace. If `RayPlugin.spec.rayClusterRef.namespace` points outside the watched namespace, the namespace-scoped install will not have permission to reconcile it.

For chart-specific details, see:

- `services/operator/helm/ray-plugin-crds/README.md`
- `services/operator/helm/ray-plugin-operator/README.md`
- `services/operator/README.md`

## Namespace Boundary

Recommended per-team layout:

| Namespace | Contains |
| --- | --- |
| `team-a-ray` | `RayCluster`, `RayPlugin` objects, plugin registry ConfigMap, operator, API server |
| `team-b-ray` | Independent `RayCluster`, `RayPlugin` objects, plugin registry ConfigMap, operator, API server |

This avoids giving one operator broad cluster permissions. A namespace owner can enable, disable, and update plugins in their namespace without being able to mutate another team's Ray cluster.

The tradeoff is intentional: a namespace-scoped operator cannot reconcile cross-namespace `RayCluster` references. Keep the `RayPlugin` and target `RayCluster` in the same namespace unless you deliberately deploy a cluster-scoped operator with broader RBAC.

## Operator Flow

1. A developer builds and pushes a plugin image from an isolated plugin repo.
2. CI applies or patches a `RayPlugin` object.
3. The operator reconciles that object into a KubeRay `RayCluster.spec.workerGroupSpecs` entry.
4. The worker group gets a Ray custom resource named after the plugin, for example `plugin-csv-analyzer`.
5. The operator writes the plugin registry ConfigMap.
6. The orchestrator reads the registry at job start, builds a dependency DAG, and routes tasks to plugin worker groups using Ray resource constraints.

## Common Commands

```bash
cd services/operator && go test ./...
```

```bash
npm install
npm --workspace apps/management-ui run test
npm --workspace apps/management-ui run build
```

```bash
cd packages/ray-framework && python3 -m unittest discover -s tests
```

```bash
helm template ray-plugin-crds ./services/operator/helm/ray-plugin-crds --include-crds
helm template ray-plugin-operator ./services/operator/helm/ray-plugin-operator --namespace ray-system
```

## Scale Position

Thousands of plugin definitions are fine. Thousands of active worker groups in one RayCluster are not the target operating model.

Use dedicated worker groups for hot or isolation-sensitive plugins. Keep cold plugins as disabled CRDs or scale-to-zero worker groups. Shard by RayCluster, data domain, tenant, or SLA before a single cluster accumulates hundreds of active plugin groups.

See `plans/2026-05-16-plugin-scale-notes.md` for details.
