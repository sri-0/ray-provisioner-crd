# Ray Plugin CRDs Chart

This chart installs cluster-scoped CRDs for the Ray plugin provisioner.

## Scope

This is a cluster-scoped chart. It must be installed by a user with permission to create `CustomResourceDefinition` objects.

It installs:

| Resource | Scope | Purpose |
| --- | --- | --- |
| `rayplugins.ray.provisioner.io` | Cluster | Defines the namespaced `RayPlugin` resource. |

## Install

```bash
helm install ray-plugin-crds ./services/operator/helm/ray-plugin-crds
```

## Upgrade

```bash
helm upgrade ray-plugin-crds ./services/operator/helm/ray-plugin-crds
```

## Render Locally

```bash
helm template ray-plugin-crds ./services/operator/helm/ray-plugin-crds --include-crds
```

## Relationship To Namespace Operators

Install this chart once per cluster. After it exists, teams can install `ray-plugin-operator` into their own namespaces without cluster-admin.

CRDs are cluster-scoped, but `RayPlugin` objects are namespaced. A namespace-scoped operator only watches and reconciles `RayPlugin` objects in its configured namespace.
