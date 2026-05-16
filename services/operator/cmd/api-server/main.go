package main

import (
	"flag"
	"net/http"
	"os"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	v1alpha1 "github.com/ray-provisioner/ray-crd-provisioner/services/operator/api/v1alpha1"
	"github.com/ray-provisioner/ray-crd-provisioner/services/operator/internal/server"
)

func main() {
	var addr string
	var namespace string
	flag.StringVar(&addr, "addr", ":8090", "HTTP listen address.")
	flag.StringVar(&namespace, "namespace", "ray-system", "Namespace containing RayPlugin resources.")
	flag.Parse()

	scheme := runtime.NewScheme()
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(corev1.AddToScheme(scheme))
	utilruntime.Must(v1alpha1.AddToScheme(scheme))

	kubeClient, err := client.New(ctrl.GetConfigOrDie(), client.Options{Scheme: scheme})
	if err != nil {
		ctrl.Log.Error(err, "create kubernetes client")
		os.Exit(1)
	}

	if err := http.ListenAndServe(addr, server.NewHandler(kubeClient, namespace)); err != nil {
		ctrl.Log.Error(err, "api server exited")
		os.Exit(1)
	}
}
