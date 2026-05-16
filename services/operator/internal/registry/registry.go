package registry

import (
	"encoding/json"
	"sort"

	v1alpha1 "github.com/ray-provisioner/ray-crd-provisioner/services/operator/api/v1alpha1"
	"github.com/ray-provisioner/ray-crd-provisioner/services/operator/internal/raycluster"
)

const RegistryJSONKey = "plugins.json"

type PluginRecord struct {
	Name                string   `json:"name"`
	Namespace           string   `json:"namespace"`
	Enabled             bool     `json:"enabled"`
	Image               string   `json:"image"`
	Entrypoint          string   `json:"entrypoint"`
	DataTypes           []string `json:"dataTypes"`
	DependsOn           []string `json:"dependsOn"`
	SupportsDistributed bool     `json:"supportsDistributed"`
	WorkerGroupName     string   `json:"workerGroupName"`
	RayResourceName     string   `json:"rayResourceName"`
	ReadyReplicas       int32    `json:"readyReplicas"`
}

type registryDocument struct {
	Plugins []PluginRecord `json:"plugins"`
}

func BuildRegistryData(plugins []v1alpha1.RayPlugin) (map[string]string, error) {
	records := make([]PluginRecord, 0, len(plugins))
	for _, plugin := range plugins {
		workerGroupName := plugin.Status.WorkerGroupName
		if workerGroupName == "" {
			workerGroupName = raycluster.WorkerGroupName(plugin.Name)
		}
		records = append(records, PluginRecord{
			Name:                plugin.Name,
			Namespace:           plugin.Namespace,
			Enabled:             plugin.Spec.Enabled,
			Image:               plugin.Spec.Image,
			Entrypoint:          plugin.Spec.Plugin.Entrypoint,
			DataTypes:           append([]string(nil), plugin.Spec.Plugin.DataTypes...),
			DependsOn:           append([]string(nil), plugin.Spec.Plugin.DependsOn...),
			SupportsDistributed: plugin.Spec.Plugin.SupportsDistributed,
			WorkerGroupName:     workerGroupName,
			RayResourceName:     workerGroupName,
			ReadyReplicas:       plugin.Status.ReadyReplicas,
		})
	}

	sort.Slice(records, func(i, j int) bool {
		return records[i].Name < records[j].Name
	})

	payload, err := json.MarshalIndent(registryDocument{Plugins: records}, "", "  ")
	if err != nil {
		return nil, err
	}

	return map[string]string{RegistryJSONKey: string(payload)}, nil
}
