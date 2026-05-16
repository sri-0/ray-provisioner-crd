# Ray CRD Provisioner

Monorepo for dynamic Ray plugin provisioning on Kubernetes.

## What This Contains

- `services/operator`: Go Kubernetes operator and management API for the `RayPlugin` CRD.
- `apps/management-ui`: Vite React UI for listing, filtering, and enabling/disabling plugins.
- `packages/ray-framework`: Python framework skeleton for Ray orchestration and plugin DAG planning.
- `examples/plugins/csv-analyzer`: Example per-plugin repo layout with Dockerfile, CI, and `RayPlugin` manifest.
- `plans`: Architecture and implementation notes.

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
helm template ray-plugin-operator ./services/operator/helm/ray-plugin-operator
```

## Scale Position

Thousands of plugin definitions are fine. Thousands of active worker groups in one RayCluster are not the target operating model.

Use dedicated worker groups for hot or isolation-sensitive plugins. Keep cold plugins as disabled CRDs or scale-to-zero worker groups. Shard by RayCluster, data domain, tenant, or SLA before a single cluster accumulates hundreds of active plugin groups.

See `plans/2026-05-16-plugin-scale-notes.md` for details.
