package main

import (
	"flag"
	"os"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"

	v1alpha1 "github.com/ray-provisioner/ray-crd-provisioner/services/operator/api/v1alpha1"
	"github.com/ray-provisioner/ray-crd-provisioner/services/operator/controllers"
)

var scheme = runtime.NewScheme()

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(corev1.AddToScheme(scheme))
	utilruntime.Must(v1alpha1.AddToScheme(scheme))
}

func main() {
	var metricsAddr string
	var probeAddr string
	var leaderElect bool
	var rayClusterName string
	var rayClusterNamespace string
	var registryConfigMap string

	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.BoolVar(&leaderElect, "leader-elect", false, "Enable leader election for controller manager.")
	flag.StringVar(&rayClusterName, "ray-cluster-name", "raycluster", "Default KubeRay RayCluster name.")
	flag.StringVar(&rayClusterNamespace, "ray-cluster-namespace", "", "Default KubeRay RayCluster namespace. Empty means RayPlugin namespace.")
	flag.StringVar(&registryConfigMap, "registry-configmap", "ray-plugin-registry", "Plugin registry ConfigMap name.")
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseDevMode(true)))

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                 scheme,
		Metrics:                metricsserver.Options{BindAddress: metricsAddr},
		HealthProbeBindAddress: probeAddr,
		LeaderElection:         leaderElect,
		LeaderElectionID:       "ray-plugin-operator.ray.provisioner.io",
	})
	if err != nil {
		ctrl.Log.Error(err, "unable to start manager")
		os.Exit(1)
	}

	if err := (&controllers.RayPluginReconciler{
		Client:                mgr.GetClient(),
		Scheme:                mgr.GetScheme(),
		DefaultRayClusterName: rayClusterName,
		DefaultRayClusterNS:   rayClusterNamespace,
		RegistryConfigMapName: registryConfigMap,
	}).SetupWithManager(mgr); err != nil {
		ctrl.Log.Error(err, "unable to create RayPlugin controller")
		os.Exit(1)
	}

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		ctrl.Log.Error(err, "unable to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		ctrl.Log.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		ctrl.Log.Error(err, "manager exited")
		os.Exit(1)
	}
}
