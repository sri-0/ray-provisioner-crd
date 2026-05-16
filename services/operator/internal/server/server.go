package server

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	v1alpha1 "github.com/ray-provisioner/ray-crd-provisioner/services/operator/api/v1alpha1"
	"github.com/ray-provisioner/ray-crd-provisioner/services/operator/internal/registry"
)

type Handler struct {
	client    client.Client
	namespace string
}

type patchPluginRequest struct {
	Enabled     *bool  `json:"enabled,omitempty"`
	Image       string `json:"image,omitempty"`
	Replicas    *int32 `json:"replicas,omitempty"`
	MinReplicas *int32 `json:"minReplicas,omitempty"`
	MaxReplicas *int32 `json:"maxReplicas,omitempty"`
}

func NewHandler(kubeClient client.Client, namespace string) http.Handler {
	return &Handler{client: kubeClient, namespace: namespace}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch {
	case r.Method == http.MethodGet && r.URL.Path == "/healthz":
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	case r.Method == http.MethodGet && r.URL.Path == "/api/v1/plugins":
		h.listPlugins(w, r)
	case r.Method == http.MethodGet && strings.HasPrefix(r.URL.Path, "/api/v1/plugins/"):
		h.getPlugin(w, r)
	case r.Method == http.MethodPatch && strings.HasPrefix(r.URL.Path, "/api/v1/plugins/"):
		h.patchPlugin(w, r)
	default:
		writeError(w, http.StatusNotFound, "not found")
	}
}

func (h *Handler) listPlugins(w http.ResponseWriter, r *http.Request) {
	list := &v1alpha1.RayPluginList{}
	if err := h.client.List(r.Context(), list, client.InNamespace(h.namespace)); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	data, err := registry.BuildRegistryData(list.Items)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(data[registry.RegistryJSONKey]))
}

func (h *Handler) getPlugin(w http.ResponseWriter, r *http.Request) {
	name := pluginNameFromPath(r.URL.Path)
	plugin, err := h.loadPlugin(r.Context(), name)
	if err != nil {
		writeKubernetesError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, plugin)
}

func (h *Handler) patchPlugin(w http.ResponseWriter, r *http.Request) {
	name := pluginNameFromPath(r.URL.Path)
	plugin, err := h.loadPlugin(r.Context(), name)
	if err != nil {
		writeKubernetesError(w, err)
		return
	}

	var req patchPluginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}

	base := plugin.DeepCopyObject().(*v1alpha1.RayPlugin)
	if req.Enabled != nil {
		plugin.Spec.Enabled = *req.Enabled
	}
	if req.Image != "" {
		plugin.Spec.Image = req.Image
	}
	if req.Replicas != nil {
		plugin.Spec.WorkerGroup.Replicas = req.Replicas
	}
	if req.MinReplicas != nil {
		plugin.Spec.WorkerGroup.MinReplicas = req.MinReplicas
	}
	if req.MaxReplicas != nil {
		plugin.Spec.WorkerGroup.MaxReplicas = req.MaxReplicas
	}

	if err := h.client.Patch(r.Context(), plugin, client.MergeFrom(base)); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, plugin)
}

func (h *Handler) loadPlugin(ctx context.Context, name string) (*v1alpha1.RayPlugin, error) {
	if name == "" {
		return nil, errors.New("plugin name is required")
	}
	plugin := &v1alpha1.RayPlugin{}
	if err := h.client.Get(ctx, client.ObjectKey{Name: name, Namespace: h.namespace}, plugin); err != nil {
		return nil, err
	}
	return plugin, nil
}

func pluginNameFromPath(path string) string {
	name := strings.TrimPrefix(path, "/api/v1/plugins/")
	name = strings.Trim(name, "/")
	return name
}

func writeKubernetesError(w http.ResponseWriter, err error) {
	if apierrors.IsNotFound(err) {
		writeError(w, http.StatusNotFound, "plugin not found")
		return
	}
	writeError(w, http.StatusInternalServerError, err.Error())
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func writeJSON(w http.ResponseWriter, status int, value interface{}) {
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
