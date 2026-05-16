# Ray Plugin Operator Monorepo Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use test-driven-development for behavior code. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a monorepo that provisions Ray plugin worker groups dynamically through a Kubernetes operator, exposes plugin management APIs/UI, and includes the Ray orchestration framework plus sample plugin layout.

**Architecture:** The Go operator owns a `RayPlugin` CRD and reconciles enabled plugins into KubeRay `RayCluster.spec.workerGroupSpecs` entries. The operator writes a compact plugin registry ConfigMap that the Python Ray orchestrator reads at job start to build the plugin DAG. The management UI talks to the operator API for plugin listing, filtering, and enable/disable operations.

**Tech Stack:** Go 1.22, controller-runtime, Kubernetes CRDs, KubeRay RayCluster CRs, Vite React TypeScript UI, Python Ray framework skeleton.

---

## File Structure

- `go.work`: root Go workspace for operator modules.
- `package.json`: root npm workspace wrapper for UI commands.
- `services/operator`: Go Kubernetes operator and management API.
- `services/operator/api/v1alpha1/rayplugin_types.go`: `RayPlugin` CRD schema.
- `services/operator/internal/raycluster`: pure helpers for RayCluster worker group patches.
- `services/operator/internal/registry`: pure helpers for plugin registry ConfigMap generation.
- `services/operator/controllers`: controller-runtime reconciler.
- `services/operator/cmd/operator`: operator process entry point.
- `services/operator/cmd/api-server`: REST API process entry point.
- `services/operator/config`: CRD, RBAC, and sample manifests.
- `services/operator/helm/ray-plugin-operator`: Helm chart.
- `apps/management-ui`: Vite React plugin management UI.
- `packages/ray-framework`: Python Ray orchestration framework skeleton.
- `examples/plugins/csv-analyzer`: sample plugin repository layout.
- `plans`: implementation and architecture plans only.

## Task 1: Operator API Types

**Files:**
- Create: `services/operator/api/v1alpha1/rayplugin_types.go`
- Create: `services/operator/api/v1alpha1/groupversion_info.go`
- Create: `services/operator/go.mod`

- [ ] Write tests for CRD helper behavior before implementation.
- [ ] Define plugin spec fields: image, enabled, worker group sizing/resources, plugin metadata, dependencies.
- [ ] Define status fields: phase, worker group name, ready replicas, observed generation, conditions.
- [ ] Run `go test ./...` from `services/operator`.

## Task 2: RayCluster Patch Helpers

**Files:**
- Create: `services/operator/internal/raycluster/worker_group.go`
- Create: `services/operator/internal/raycluster/worker_group_test.go`

- [ ] Write failing tests for enabled plugin worker group creation.
- [ ] Write failing tests for disabled plugin scale-to-zero behavior.
- [ ] Implement pure worker group spec builder using KubeRay-compatible unstructured maps.
- [ ] Run `go test ./internal/raycluster`.

## Task 3: Registry ConfigMap Helpers

**Files:**
- Create: `services/operator/internal/registry/registry.go`
- Create: `services/operator/internal/registry/registry_test.go`

- [ ] Write failing tests for deterministic plugin registry JSON.
- [ ] Implement registry JSON serialization sorted by plugin name.
- [ ] Run `go test ./internal/registry`.

## Task 4: Controller and API Server

**Files:**
- Create: `services/operator/controllers/rayplugin_controller.go`
- Create: `services/operator/cmd/operator/main.go`
- Create: `services/operator/cmd/api-server/main.go`
- Create: `services/operator/internal/server/server.go`

- [ ] Add reconciler that watches `RayPlugin` and patches target `RayCluster` worker groups.
- [ ] Add reconciler registry ConfigMap writer.
- [ ] Add HTTP routes: list plugins, get plugin, patch enable/config.
- [ ] Run `go test ./...`.

## Task 5: Kubernetes Packaging

**Files:**
- Create: `services/operator/config/crd/bases/ray.provisioner.io_rayplugins.yaml`
- Create: `services/operator/config/rbac/*.yaml`
- Create: `services/operator/config/samples/rayplugin.yaml`
- Create: `services/operator/helm/ray-plugin-operator/*`

- [ ] Add CRD manifest matching the Go type.
- [ ] Add least-privilege RBAC for RayPlugin, RayCluster, ConfigMap, Events.
- [ ] Add sample plugin manifest.
- [ ] Add Helm chart values for target namespace/ray cluster/name/images.

## Task 6: Management UI

**Files:**
- Create: `apps/management-ui/package.json`
- Create: `apps/management-ui/src/*`

- [ ] Build plugin list view with status, image, data types, dependencies, and ready replicas.
- [ ] Build filtering by enabled state and data type.
- [ ] Build enable/disable mutation using the operator API.
- [ ] Run `npm install` and `npm run build` from `apps/management-ui` when dependencies are available.

## Task 7: Framework and Sample Plugin

**Files:**
- Create: `packages/ray-framework/ray_framework/*`
- Create: `examples/plugins/csv-analyzer/*`

- [ ] Add base plugin class.
- [ ] Add registry ConfigMap reader.
- [ ] Add DAG builder with dependency cycle detection.
- [ ] Add sample plugin Dockerfile and `RayPlugin` manifest.

## Task 8: Scalability Notes

**Files:**
- Create: `plans/2026-05-16-plugin-scale-notes.md`

- [ ] Document where 1,000 plugins scales well.
- [ ] Document where 1,000 worker groups becomes operationally expensive.
- [ ] Recommend practical grouping strategy: cold plugins disabled, autoscaled groups, shared dependency profiles, and optional plugin bundles for rarely used plugins.
