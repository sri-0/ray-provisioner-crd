# Plugin Scale Notes

## Short Answer

This architecture can scale to thousands of plugin definitions, but it should not naively run thousands of active Ray worker groups in one Ray cluster.

The correct model is:

- Thousands of `RayPlugin` CRDs: acceptable.
- Hundreds or thousands of disabled/cold plugins: acceptable.
- Dozens to low hundreds of active worker groups per RayCluster: realistic depending on worker count and autoscaler behavior.
- Thousands of active worker groups in one RayCluster: likely to get out of hand.

## What Scales Well

`RayPlugin` CRDs scale well because they are small Kubernetes objects. The operator watches them, reconciles them, and writes registry state. Kubernetes can handle thousands of small custom resources if the controller is written carefully.

Namespace-scoped operator instances also scale operational ownership better than one cluster-wide operator. Each team can run an operator in its own namespace with namespaced RBAC, while platform admins install the CRDs once.

Per-plugin CI also scales well. A plugin repo builds one image and applies one `RayPlugin` manifest. That keeps plugin releases isolated and avoids full-cluster image rebuilds.

Disabled plugins scale well. A disabled plugin can remain present as a CRD and registry entry without running pods.

## What Can Get Expensive

One worker group per plugin becomes expensive if too many are enabled at once.

The pressure points are:

- The `RayCluster` spec grows because every worker group is embedded in `spec.workerGroupSpecs`.
- KubeRay reconciliation gets slower as the worker group list grows.
- Kubernetes scheduling gets noisier if every enabled plugin creates unique pods and autoscaling rules.
- Ray scheduling has more custom resources to track.
- Cluster autoscaler behavior can become fragmented because every plugin has its own pod shape.
- The plugin registry ConfigMap has a 1 MiB Kubernetes object limit. A compact 1,000-plugin registry may fit, but metadata-heavy plugin records can exceed it.

## Recommended Scale Strategy

Keep the `RayPlugin` CRD as the source of truth for every plugin, but only provision worker groups for plugins that are enabled and operationally hot.

Use these tiers:

- Hot plugins: dedicated worker group, autoscaled, custom image, custom Ray resource.
- Warm plugins: worker group exists with `minReplicas: 0`, scales on demand.
- Cold plugins: CRD exists, but worker group is not present or is scaled to zero until enabled.
- Long-tail plugins: consider shared execution profiles or bundled images if strict per-plugin image isolation is not worth the control-plane cost.

## Practical Cluster Limits

Start with soft limits rather than hard-coded assumptions:

- Alert above 100 active worker groups in one RayCluster.
- Require review above 250 active worker groups in one RayCluster.
- Shard by data domain, tenant, or SLA before reaching 500 active worker groups.
- Split the plugin registry ConfigMap by namespace, data type, or hash bucket before it approaches 750 KiB.

These are operational guardrails, not Kubernetes hard limits.

## Operator Changes Needed Before Serious Scale

The first implementation writes one registry ConfigMap and patches one RayCluster worker group list. That is fine for the first slice.

Before operating thousands of plugin definitions, add:

- Registry sharding: `ray-plugin-registry-0..N`, or one registry per data type family.
- Reconcile rate limiting and workqueue backoff.
- Server-side apply or strategic ownership markers for generated worker groups.
- Admission validation to prevent impossible DAGs, duplicate data type ownership if required, and invalid dependency cycles.
- Status aggregation from RayCluster worker pod readiness.
- UI pagination or virtualization for large plugin inventories.

## Recommendation

Proceed with this architecture, but treat "one worker group per plugin" as a deployment policy for hot or isolation-sensitive plugins, not as a universal rule for all 1,000 plugins.

The source-of-truth CRD scales to thousands. The runtime execution topology should be adaptive.

Use the split install model for multi-team clusters: `ray-plugin-crds` is cluster-scoped and installed once by an admin; `ray-plugin-operator` is namespace-scoped and can be installed per team or environment without cluster-admin.
