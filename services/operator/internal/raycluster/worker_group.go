package raycluster

import (
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"strings"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	v1alpha1 "github.com/ray-provisioner/ray-crd-provisioner/services/operator/api/v1alpha1"
)

var invalidResourceNameChars = regexp.MustCompile(`[^a-z0-9-]+`)

func WorkerGroupName(pluginName string) string {
	name := strings.ToLower(pluginName)
	name = invalidResourceNameChars.ReplaceAllString(name, "-")
	name = strings.Trim(name, "-")
	if name == "" {
		name = "unnamed"
	}
	return "plugin-" + name
}

func BuildWorkerGroupSpec(plugin v1alpha1.RayPlugin) (map[string]interface{}, error) {
	if plugin.Spec.Image == "" {
		return nil, fmt.Errorf("plugin %q image is required", plugin.Name)
	}

	workerGroupName := WorkerGroupName(plugin.Name)
	replicas := valueOrDefault(plugin.Spec.WorkerGroup.Replicas, 1)
	minReplicas := valueOrDefault(plugin.Spec.WorkerGroup.MinReplicas, 0)
	maxReplicas := valueOrDefault(plugin.Spec.WorkerGroup.MaxReplicas, replicas)
	if !plugin.Spec.Enabled {
		replicas = 0
		minReplicas = 0
		maxReplicas = 0
	}

	resourcePayload, err := json.Marshal(map[string]int{workerGroupName: 100})
	if err != nil {
		return nil, err
	}
	rayStartParams := map[string]interface{}{}
	for key, value := range plugin.Spec.WorkerGroup.RayStartParams {
		rayStartParams[key] = value
	}
	rayStartParams["resources"] = string(resourcePayload)

	container := map[string]interface{}{
		"name":  "ray-worker",
		"image": plugin.Spec.Image,
		"env":   buildEnv(plugin),
	}
	if resources := buildContainerResources(plugin.Spec.WorkerGroup.Resources); len(resources) > 0 {
		container["resources"] = resources
	}

	return map[string]interface{}{
		"groupName":      workerGroupName,
		"replicas":       int64(replicas),
		"minReplicas":    int64(minReplicas),
		"maxReplicas":    int64(maxReplicas),
		"rayStartParams": rayStartParams,
		"template": map[string]interface{}{
			"spec": map[string]interface{}{
				"containers": []interface{}{container},
			},
		},
	}, nil
}

func UpsertWorkerGroupSpec(rayCluster *unstructured.Unstructured, workerGroup map[string]interface{}) (bool, error) {
	groupName, ok := workerGroup["groupName"].(string)
	if !ok || groupName == "" {
		return false, fmt.Errorf("worker group must include groupName")
	}

	workerGroups, _, err := unstructured.NestedSlice(rayCluster.Object, "spec", "workerGroupSpecs")
	if err != nil {
		return false, err
	}

	for i, existing := range workerGroups {
		existingMap, ok := existing.(map[string]interface{})
		if !ok {
			continue
		}
		if existingMap["groupName"] != groupName {
			continue
		}
		if reflect.DeepEqual(existingMap, workerGroup) {
			return false, nil
		}
		workerGroups[i] = workerGroup
		return true, unstructured.SetNestedSlice(rayCluster.Object, workerGroups, "spec", "workerGroupSpecs")
	}

	workerGroups = append(workerGroups, workerGroup)
	return true, unstructured.SetNestedSlice(rayCluster.Object, workerGroups, "spec", "workerGroupSpecs")
}

func valueOrDefault(value *int32, fallback int32) int32 {
	if value == nil {
		return fallback
	}
	return *value
}

func buildEnv(plugin v1alpha1.RayPlugin) []interface{} {
	env := []interface{}{
		map[string]interface{}{"name": "RAY_PLUGIN_NAME", "value": plugin.Name},
		map[string]interface{}{"name": "RAY_PLUGIN_ENTRYPOINT", "value": plugin.Spec.Plugin.Entrypoint},
	}
	for _, item := range plugin.Spec.WorkerGroup.Env {
		env = append(env, envVarToMap(item))
	}
	return env
}

func envVarToMap(env corev1.EnvVar) map[string]interface{} {
	result := map[string]interface{}{"name": env.Name}
	if env.Value != "" {
		result["value"] = env.Value
	}
	if env.ValueFrom != nil {
		result["valueFrom"] = env.ValueFrom
	}
	return result
}

func buildContainerResources(spec v1alpha1.PluginResourceSpec) map[string]interface{} {
	requests := map[string]interface{}{}
	limits := map[string]interface{}{}
	if spec.CPU != "" {
		requests["cpu"] = spec.CPU
		limits["cpu"] = spec.CPU
	}
	if spec.Memory != "" {
		requests["memory"] = spec.Memory
		limits["memory"] = spec.Memory
	}
	if spec.GPU != "" {
		requests["nvidia.com/gpu"] = spec.GPU
		limits["nvidia.com/gpu"] = spec.GPU
	}
	if len(requests) == 0 && len(limits) == 0 {
		return nil
	}
	return map[string]interface{}{
		"requests": requests,
		"limits":   limits,
	}
}
