package operatorconfig

import (
	"testing"

	"k8s.io/apimachinery/pkg/runtime"
)

func TestBuildManagerOptionsScopesCacheWhenWatchNamespaceIsSet(t *testing.T) {
	scheme := runtime.NewScheme()

	options := BuildManagerOptions(ManagerConfig{
		Scheme:           scheme,
		MetricsAddr:      ":8080",
		ProbeAddr:        ":8081",
		LeaderElect:      true,
		WatchNamespace:   "team-a-ray",
		LeaderElectionID: "ray-plugin-operator.ray.provisioner.io",
	})

	if options.Scheme != scheme {
		t.Fatal("manager options did not preserve the provided scheme")
	}
	if _, ok := options.Cache.DefaultNamespaces["team-a-ray"]; !ok {
		t.Fatalf("watch namespace was not added to cache default namespaces: %#v", options.Cache.DefaultNamespaces)
	}
	if options.LeaderElectionNamespace != "team-a-ray" {
		t.Fatalf("leader election namespace = %q, want team-a-ray", options.LeaderElectionNamespace)
	}
}

func TestBuildManagerOptionsLeavesCacheClusterScopedWhenWatchNamespaceIsEmpty(t *testing.T) {
	options := BuildManagerOptions(ManagerConfig{
		Scheme:           runtime.NewScheme(),
		MetricsAddr:      ":8080",
		ProbeAddr:        ":8081",
		LeaderElectionID: "ray-plugin-operator.ray.provisioner.io",
	})

	if options.Cache.DefaultNamespaces != nil {
		t.Fatalf("default namespaces = %#v, want nil", options.Cache.DefaultNamespaces)
	}
	if options.LeaderElectionNamespace != "" {
		t.Fatalf("leader election namespace = %q, want empty", options.LeaderElectionNamespace)
	}
}
