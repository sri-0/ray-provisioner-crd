# Ray Plugin Operator

The operator reconciles `RayPlugin` resources into KubeRay `RayCluster.spec.workerGroupSpecs` entries. It also writes the plugin registry ConfigMap consumed by the Ray orchestration framework.

## Install Scopes

There are two supported install scopes.

| Mode | RBAC | Watches | Use Case |
| --- | --- | --- | --- |
| Namespace-scoped | `Role` / `RoleBinding` | One namespace | Recommended default for teams and environments. |
| Cluster-scoped | `ClusterRole` / `ClusterRoleBinding` | All namespaces | Not packaged by default; only use when a platform team deliberately wants one operator for many namespaces. |

The current Helm chart is namespace-scoped. It does not install CRDs and does not require cluster-admin after the CRDs already exist.

## Required Cluster Prerequisites

A cluster admin must install these once:

- KubeRay CRDs, including `RayCluster`
- Ray plugin CRDs from `services/operator/helm/ray-plugin-crds`

Then a namespace owner can install the operator chart in their namespace.

## Namespace-Scoped Behavior

The operator binary accepts `--watch-namespace`. When set, controller-runtime only caches and watches objects in that namespace.

Example arguments:

```bash
ray-plugin-operator \
  --watch-namespace=team-a-ray \
  --ray-cluster-name=raycluster \
  --ray-cluster-namespace=team-a-ray \
  --registry-configmap=ray-plugin-registry
```

The namespace-scoped operator can reconcile this:

```yaml
apiVersion: ray.provisioner.io/v1alpha1
kind: RayPlugin
metadata:
  name: csv-analyzer
  namespace: team-a-ray
spec:
  rayClusterRef:
    name: raycluster
    namespace: team-a-ray
```

It should not be used for this unless broader RBAC is intentionally granted:

```yaml
apiVersion: ray.provisioner.io/v1alpha1
kind: RayPlugin
metadata:
  name: csv-analyzer
  namespace: team-a-ray
spec:
  rayClusterRef:
    name: raycluster
    namespace: shared-ray
```

## RBAC

The namespace chart grants access to these resources inside the release namespace:

| Resource | API Group | Why |
| --- | --- | --- |
| `rayplugins` | `ray.provisioner.io` | Watch plugin definitions. |
| `rayplugins/status` | `ray.provisioner.io` | Update reconciliation status. |
| `rayclusters` | `ray.io` | Patch KubeRay worker group specs. |
| `configmaps` | core | Write the plugin registry. |
| `events` | core | Emit Kubernetes events. |
| `leases` | `coordination.k8s.io` | Support leader election when enabled. |

## Local Verification

```bash
go test ./...
```

```bash
helm template ray-plugin-crds ./helm/ray-plugin-crds --include-crds
helm template ray-plugin-operator ./helm/ray-plugin-operator --namespace team-a-ray
```
