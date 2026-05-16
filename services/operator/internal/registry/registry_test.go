package registry

import (
	"encoding/json"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	v1alpha1 "github.com/ray-provisioner/ray-crd-provisioner/services/operator/api/v1alpha1"
)

func TestBuildRegistryDataSerializesPluginsInDeterministicNameOrder(t *testing.T) {
	plugins := []v1alpha1.RayPlugin{
		{
			ObjectMeta: metav1.ObjectMeta{Name: "z-plugin"},
			Spec: v1alpha1.RayPluginSpec{
				Enabled: true,
				Image:   "z:v1",
				Plugin: v1alpha1.PluginMetadataSpec{
					Entrypoint: "plugins.z.Plugin",
					DataTypes:  []string{"application/json"},
				},
			},
			Status: v1alpha1.RayPluginStatus{WorkerGroupName: "plugin-z-plugin", ReadyReplicas: 2},
		},
		{
			ObjectMeta: metav1.ObjectMeta{Name: "a-plugin"},
			Spec: v1alpha1.RayPluginSpec{
				Enabled: false,
				Image:   "a:v1",
				Plugin: v1alpha1.PluginMetadataSpec{
					Entrypoint: "plugins.a.Plugin",
					DataTypes:  []string{"text/csv"},
					DependsOn:  []string{"base-normalizer"},
				},
			},
			Status: v1alpha1.RayPluginStatus{WorkerGroupName: "plugin-a-plugin"},
		},
	}

	data, err := BuildRegistryData(plugins)
	if err != nil {
		t.Fatalf("BuildRegistryData returned error: %v", err)
	}

	var decoded struct {
		Plugins []PluginRecord `json:"plugins"`
	}
	if err := json.Unmarshal([]byte(data[RegistryJSONKey]), &decoded); err != nil {
		t.Fatalf("registry json did not decode: %v", err)
	}

	if got := decoded.Plugins[0].Name; got != "a-plugin" {
		t.Fatalf("first plugin = %s, want a-plugin", got)
	}
	if got := decoded.Plugins[1].Name; got != "z-plugin" {
		t.Fatalf("second plugin = %s, want z-plugin", got)
	}
	if decoded.Plugins[0].Enabled {
		t.Fatal("a-plugin enabled = true, want false")
	}
	if got := decoded.Plugins[0].DependsOn[0]; got != "base-normalizer" {
		t.Fatalf("dependency = %s, want base-normalizer", got)
	}
}
