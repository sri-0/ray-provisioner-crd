package server

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	v1alpha1 "github.com/ray-provisioner/ray-crd-provisioner/services/operator/api/v1alpha1"
)

func TestPatchPluginUpdatesEnabledFlag(t *testing.T) {
	scheme := runtime.NewScheme()
	if err := v1alpha1.AddToScheme(scheme); err != nil {
		t.Fatalf("AddToScheme failed: %v", err)
	}

	plugin := &v1alpha1.RayPlugin{
		ObjectMeta: metav1.ObjectMeta{Name: "csv-analyzer", Namespace: "ray-system"},
		Spec: v1alpha1.RayPluginSpec{
			Enabled: false,
			Image:   "registry.example.com/csv:v1",
			Plugin:  v1alpha1.PluginMetadataSpec{Entrypoint: "plugins.csv.Plugin"},
		},
	}
	kubeClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(plugin).Build()
	handler := NewHandler(kubeClient, "ray-system")

	req := httptest.NewRequest(http.MethodPatch, "/api/v1/plugins/csv-analyzer", bytes.NewBufferString(`{"enabled":true}`))
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", recorder.Code, recorder.Body.String())
	}

	updated := &v1alpha1.RayPlugin{}
	if err := kubeClient.Get(context.Background(), client.ObjectKey{Name: "csv-analyzer", Namespace: "ray-system"}, updated); err != nil {
		t.Fatalf("get updated plugin: %v", err)
	}
	if !updated.Spec.Enabled {
		t.Fatal("updated plugin enabled = false, want true")
	}
}
