package controllers

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	v1alpha1 "github.com/ray-provisioner/ray-crd-provisioner/services/operator/api/v1alpha1"
	"github.com/ray-provisioner/ray-crd-provisioner/services/operator/internal/raycluster"
	"github.com/ray-provisioner/ray-crd-provisioner/services/operator/internal/registry"
)

var rayClusterGVK = schema.GroupVersionKind{Group: "ray.io", Version: "v1", Kind: "RayCluster"}

type RayPluginReconciler struct {
	client.Client
	Scheme                *runtime.Scheme
	DefaultRayClusterName string
	DefaultRayClusterNS   string
	RegistryConfigMapName string
}

func (r *RayPluginReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	plugin := &v1alpha1.RayPlugin{}
	if err := r.Get(ctx, req.NamespacedName, plugin); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	workerGroup, err := raycluster.BuildWorkerGroupSpec(*plugin)
	if err != nil {
		r.setStatus(ctx, plugin, v1alpha1.RayPluginPhaseError, "", err.Error())
		return ctrl.Result{}, err
	}

	rayCluster := &unstructured.Unstructured{}
	rayCluster.SetGroupVersionKind(rayClusterGVK)
	rayClusterKey := r.rayClusterKey(plugin)
	if err := r.Get(ctx, rayClusterKey, rayCluster); err != nil {
		message := fmt.Sprintf("load RayCluster %s/%s: %v", rayClusterKey.Namespace, rayClusterKey.Name, err)
		r.setStatus(ctx, plugin, v1alpha1.RayPluginPhaseError, "", message)
		return ctrl.Result{}, err
	}

	changed, err := raycluster.UpsertWorkerGroupSpec(rayCluster, workerGroup)
	if err != nil {
		r.setStatus(ctx, plugin, v1alpha1.RayPluginPhaseError, "", err.Error())
		return ctrl.Result{}, err
	}
	if changed {
		if err := r.Update(ctx, rayCluster); err != nil {
			message := fmt.Sprintf("update RayCluster %s/%s: %v", rayClusterKey.Namespace, rayClusterKey.Name, err)
			r.setStatus(ctx, plugin, v1alpha1.RayPluginPhaseError, "", message)
			return ctrl.Result{}, err
		}
	}

	phase := v1alpha1.RayPluginPhaseReady
	if !plugin.Spec.Enabled {
		phase = v1alpha1.RayPluginPhaseDisabled
	}
	r.setStatus(ctx, plugin, phase, raycluster.WorkerGroupName(plugin.Name), "")

	if err := r.syncRegistryConfigMap(ctx, plugin.Namespace); err != nil {
		logger.Error(err, "sync plugin registry ConfigMap")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *RayPluginReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.RayPlugin{}).
		Complete(r)
}

func (r *RayPluginReconciler) rayClusterKey(plugin *v1alpha1.RayPlugin) client.ObjectKey {
	name := plugin.Spec.RayClusterRef.Name
	if name == "" {
		name = r.DefaultRayClusterName
	}
	namespace := plugin.Spec.RayClusterRef.Namespace
	if namespace == "" {
		namespace = r.DefaultRayClusterNS
	}
	if namespace == "" {
		namespace = plugin.Namespace
	}
	return client.ObjectKey{Name: name, Namespace: namespace}
}

func (r *RayPluginReconciler) syncRegistryConfigMap(ctx context.Context, namespace string) error {
	list := &v1alpha1.RayPluginList{}
	if err := r.List(ctx, list, client.InNamespace(namespace)); err != nil {
		return err
	}
	data, err := registry.BuildRegistryData(list.Items)
	if err != nil {
		return err
	}

	name := r.RegistryConfigMapName
	if name == "" {
		name = "ray-plugin-registry"
	}

	configMap := &corev1.ConfigMap{}
	key := client.ObjectKey{Name: name, Namespace: namespace}
	if err := r.Get(ctx, key, configMap); err != nil {
		if !apierrors.IsNotFound(err) {
			return err
		}
		configMap = &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace},
			Data:       data,
		}
		return r.Create(ctx, configMap)
	}

	base := configMap.DeepCopy()
	configMap.Data = data
	return r.Patch(ctx, configMap, client.MergeFrom(base))
}

func (r *RayPluginReconciler) setStatus(ctx context.Context, plugin *v1alpha1.RayPlugin, phase v1alpha1.RayPluginPhase, workerGroupName string, message string) {
	base := plugin.DeepCopyObject().(*v1alpha1.RayPlugin)
	plugin.Status.Phase = phase
	plugin.Status.WorkerGroupName = workerGroupName
	plugin.Status.ObservedGeneration = plugin.Generation
	plugin.Status.Message = message
	if err := r.Status().Patch(ctx, plugin, client.MergeFrom(base)); err != nil {
		log.FromContext(ctx).Error(err, "patch RayPlugin status")
	}
}
