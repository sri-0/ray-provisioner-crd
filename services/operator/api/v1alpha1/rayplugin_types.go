package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type RayPluginPhase string

const (
	RayPluginPhasePending  RayPluginPhase = "Pending"
	RayPluginPhaseReady    RayPluginPhase = "Ready"
	RayPluginPhaseDisabled RayPluginPhase = "Disabled"
	RayPluginPhaseError    RayPluginPhase = "Error"
)

type RayClusterReference struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace,omitempty"`
}

type PluginResourceSpec struct {
	CPU    string `json:"cpu,omitempty"`
	Memory string `json:"memory,omitempty"`
	GPU    string `json:"gpu,omitempty"`
}

type PluginWorkerGroupSpec struct {
	Replicas       *int32             `json:"replicas,omitempty"`
	MinReplicas    *int32             `json:"minReplicas,omitempty"`
	MaxReplicas    *int32             `json:"maxReplicas,omitempty"`
	RayStartParams map[string]string  `json:"rayStartParams,omitempty"`
	Resources      PluginResourceSpec `json:"resources,omitempty"`
	Env            []corev1.EnvVar    `json:"env,omitempty"`
}

type PluginMetadataSpec struct {
	Entrypoint          string   `json:"entrypoint"`
	DataTypes           []string `json:"dataTypes,omitempty"`
	DependsOn           []string `json:"dependsOn,omitempty"`
	SupportsDistributed bool     `json:"supportsDistributed,omitempty"`
}

type RayPluginSpec struct {
	Enabled       bool                  `json:"enabled"`
	Image         string                `json:"image"`
	RayClusterRef RayClusterReference   `json:"rayClusterRef"`
	WorkerGroup   PluginWorkerGroupSpec `json:"workerGroup,omitempty"`
	Plugin        PluginMetadataSpec    `json:"plugin"`
}

type RayPluginStatus struct {
	Phase              RayPluginPhase     `json:"phase,omitempty"`
	WorkerGroupName    string             `json:"workerGroupName,omitempty"`
	ReadyReplicas      int32              `json:"readyReplicas,omitempty"`
	ObservedGeneration int64              `json:"observedGeneration,omitempty"`
	Message            string             `json:"message,omitempty"`
	Conditions         []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
type RayPlugin struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RayPluginSpec   `json:"spec,omitempty"`
	Status RayPluginStatus `json:"status,omitempty"`
}

func (in *RayPlugin) DeepCopyObject() runtime.Object {
	if in == nil {
		return nil
	}
	out := new(RayPlugin)
	in.DeepCopyInto(out)
	return out
}

func (in *RayPlugin) DeepCopyInto(out *RayPlugin) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	out.ObjectMeta = *in.ObjectMeta.DeepCopy()
	out.Spec = *in.Spec.DeepCopy()
	out.Status = *in.Status.DeepCopy()
}

func (in *RayPluginSpec) DeepCopy() *RayPluginSpec {
	if in == nil {
		return nil
	}
	out := new(RayPluginSpec)
	*out = *in
	if in.WorkerGroup.Replicas != nil {
		v := *in.WorkerGroup.Replicas
		out.WorkerGroup.Replicas = &v
	}
	if in.WorkerGroup.MinReplicas != nil {
		v := *in.WorkerGroup.MinReplicas
		out.WorkerGroup.MinReplicas = &v
	}
	if in.WorkerGroup.MaxReplicas != nil {
		v := *in.WorkerGroup.MaxReplicas
		out.WorkerGroup.MaxReplicas = &v
	}
	if in.WorkerGroup.RayStartParams != nil {
		out.WorkerGroup.RayStartParams = map[string]string{}
		for k, v := range in.WorkerGroup.RayStartParams {
			out.WorkerGroup.RayStartParams[k] = v
		}
	}
	out.WorkerGroup.Env = append([]corev1.EnvVar(nil), in.WorkerGroup.Env...)
	out.Plugin.DataTypes = append([]string(nil), in.Plugin.DataTypes...)
	out.Plugin.DependsOn = append([]string(nil), in.Plugin.DependsOn...)
	return out
}

func (in *RayPluginStatus) DeepCopy() *RayPluginStatus {
	if in == nil {
		return nil
	}
	out := new(RayPluginStatus)
	*out = *in
	out.Conditions = append([]metav1.Condition(nil), in.Conditions...)
	return out
}

// +kubebuilder:object:root=true
type RayPluginList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []RayPlugin `json:"items"`
}

func (in *RayPluginList) DeepCopyObject() runtime.Object {
	if in == nil {
		return nil
	}
	out := new(RayPluginList)
	*out = *in
	out.TypeMeta = in.TypeMeta
	out.ListMeta = in.ListMeta
	if in.Items != nil {
		out.Items = make([]RayPlugin, len(in.Items))
		for i := range in.Items {
			in.Items[i].DeepCopyInto(&out.Items[i])
		}
	}
	return out
}
