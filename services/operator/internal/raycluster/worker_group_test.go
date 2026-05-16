package raycluster

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	v1alpha1 "github.com/ray-provisioner/ray-crd-provisioner/services/operator/api/v1alpha1"
)

func int32ptr(v int32) *int32 { return &v }

func TestBuildWorkerGroupSpecCreatesRoutableEnabledWorkerGroup(t *testing.T) {
	plugin := v1alpha1.RayPlugin{
		ObjectMeta: metav1.ObjectMeta{Name: "csv-analyzer"},
		Spec: v1alpha1.RayPluginSpec{
			Enabled: true,
			Image:   "registry.example.com/plugins/csv-analyzer:v1",
			WorkerGroup: v1alpha1.PluginWorkerGroupSpec{
				Replicas:    int32ptr(2),
				MinReplicas: int32ptr(1),
				MaxReplicas: int32ptr(20),
				Resources: v1alpha1.PluginResourceSpec{
					CPU:    "4",
					Memory: "8Gi",
				},
				Env: []corev1.EnvVar{{Name: "PLUGIN_MODE", Value: "active"}},
			},
			Plugin: v1alpha1.PluginMetadataSpec{Entrypoint: "plugins.csv.Plugin"},
		},
	}

	workerGroup, err := BuildWorkerGroupSpec(plugin)
	if err != nil {
		t.Fatalf("BuildWorkerGroupSpec returned error: %v", err)
	}

	if got := workerGroup["groupName"]; got != "plugin-csv-analyzer" {
		t.Fatalf("groupName = %v, want plugin-csv-analyzer", got)
	}
	if got := workerGroup["replicas"]; got != int64(2) {
		t.Fatalf("replicas = %v, want 2", got)
	}
	rayStartParams := workerGroup["rayStartParams"].(map[string]interface{})
	if got := rayStartParams["resources"]; got != "{\"plugin-csv-analyzer\":100}" {
		t.Fatalf("resources start param = %v", got)
	}
	template := workerGroup["template"].(map[string]interface{})
	spec := template["spec"].(map[string]interface{})
	containers := spec["containers"].([]interface{})
	container := containers[0].(map[string]interface{})
	if got := container["image"]; got != "registry.example.com/plugins/csv-analyzer:v1" {
		t.Fatalf("container image = %v", got)
	}
}

func TestBuildWorkerGroupSpecScalesDisabledPluginToZero(t *testing.T) {
	plugin := v1alpha1.RayPlugin{
		ObjectMeta: metav1.ObjectMeta{Name: "pdf-parser"},
		Spec: v1alpha1.RayPluginSpec{
			Enabled: false,
			Image:   "registry.example.com/plugins/pdf-parser:v2",
			WorkerGroup: v1alpha1.PluginWorkerGroupSpec{
				Replicas:    int32ptr(5),
				MinReplicas: int32ptr(1),
				MaxReplicas: int32ptr(50),
			},
		},
	}

	workerGroup, err := BuildWorkerGroupSpec(plugin)
	if err != nil {
		t.Fatalf("BuildWorkerGroupSpec returned error: %v", err)
	}

	for _, key := range []string{"replicas", "minReplicas", "maxReplicas"} {
		if got := workerGroup[key]; got != int64(0) {
			t.Fatalf("%s = %v, want 0 for disabled plugin", key, got)
		}
	}
}

func TestUpsertWorkerGroupSpecReplacesExistingGroup(t *testing.T) {
	rayCluster := &unstructured.Unstructured{Object: map[string]interface{}{
		"spec": map[string]interface{}{
			"workerGroupSpecs": []interface{}{
				map[string]interface{}{"groupName": "plugin-csv-analyzer", "replicas": int64(1)},
				map[string]interface{}{"groupName": "plugin-pdf-parser", "replicas": int64(3)},
			},
		},
	}}

	changed, err := UpsertWorkerGroupSpec(rayCluster, map[string]interface{}{
		"groupName": "plugin-csv-analyzer",
		"replicas":  int64(4),
	})
	if err != nil {
		t.Fatalf("UpsertWorkerGroupSpec returned error: %v", err)
	}
	if !changed {
		t.Fatal("changed = false, want true")
	}

	groups, _, _ := unstructured.NestedSlice(rayCluster.Object, "spec", "workerGroupSpecs")
	if len(groups) != 2 {
		t.Fatalf("worker group count = %d, want 2", len(groups))
	}
	first := groups[0].(map[string]interface{})
	if got := first["replicas"]; got != int64(4) {
		t.Fatalf("updated replicas = %v, want 4", got)
	}
}
