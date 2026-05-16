# Ray Plugin Operator Chart

This chart installs the namespace-scoped Ray plugin operator and management API server.

## Scope

This chart is designed for namespace owners. It installs namespaced resources only:

| Resource | Scope |
| --- | --- |
| `Deployment` for the operator | Namespace |
| `Deployment` for the API server | Namespace |
| `Service` for the API server | Namespace |
| `ServiceAccount` | Namespace |
| `Role` | Namespace |
| `RoleBinding` | Namespace |

It does not install CRDs. Install `ray-plugin-crds` first as a cluster admin.

## Install

```bash
helm install ray-plugin-operator ./services/operator/helm/ray-plugin-operator \
  --namespace team-a-ray \
  --create-namespace
```

By default, the chart uses the Helm release namespace for:

- The operator Deployment namespace
- `--watch-namespace`
- `--ray-cluster-namespace`
- The API server namespace
- The `Role` and `RoleBinding` namespace

## Values

| Value | Default | Purpose |
| --- | --- | --- |
| `namespace` | `""` | Optional override. Empty means `.Release.Namespace`. |
| `image.repository` | `registry.example.com/ray-plugin-operator` | Operator/API container image repository. |
| `image.tag` | `latest` | Operator/API container image tag. |
| `image.pullPolicy` | `IfNotPresent` | Kubernetes image pull policy. |
| `operator.rayClusterName` | `raycluster` | Default target KubeRay `RayCluster` name. |
| `operator.rayClusterNamespace` | `""` | Default target `RayCluster` namespace. Empty means watch namespace. |
| `operator.watchNamespace` | `""` | Namespace watched by the operator. Empty means release namespace. |
| `operator.registryConfigMap` | `ray-plugin-registry` | ConfigMap written with plugin registry JSON. |
| `operator.replicas` | `1` | Operator replica count. Use leader election before increasing. |
| `apiServer.enabled` | `true` | Whether to install the management API server. |
| `apiServer.replicas` | `1` | API server replica count. |
| `apiServer.port` | `8090` | API server container port. |
| `serviceAccount.name` | `ray-plugin-operator` | ServiceAccount used by operator and API server. |

## Namespace Boundary

The default install expects `RayPlugin` and `RayCluster` to live in the same namespace:

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

Cross-namespace `RayCluster` references are intentionally not supported by the default RBAC. If you set `operator.rayClusterNamespace` to another namespace, you must also provide RBAC that allows patching `RayCluster` in that namespace.

## Verify Permissions

After install, check that the service account can patch `RayCluster` only in its namespace:

```bash
kubectl auth can-i patch rayclusters.ray.io \
  --as system:serviceaccount:team-a-ray:ray-plugin-operator \
  --namespace team-a-ray
```

```bash
kubectl auth can-i patch rayclusters.ray.io \
  --as system:serviceaccount:team-a-ray:ray-plugin-operator \
  --namespace other-namespace
```

The first command should return `yes`. The second should return `no` unless you intentionally granted broader access.

## Render Locally

```bash
helm template ray-plugin-operator ./services/operator/helm/ray-plugin-operator \
  --namespace team-a-ray
```
