package operatorconfig

import (
	"strings"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
)

type ManagerConfig struct {
	Scheme           *runtime.Scheme
	MetricsAddr      string
	ProbeAddr        string
	LeaderElect      bool
	WatchNamespace   string
	LeaderElectionID string
}

func BuildManagerOptions(config ManagerConfig) ctrl.Options {
	watchNamespace := strings.TrimSpace(config.WatchNamespace)

	options := ctrl.Options{
		Scheme:                 config.Scheme,
		Metrics:                metricsserver.Options{BindAddress: config.MetricsAddr},
		HealthProbeBindAddress: config.ProbeAddr,
		LeaderElection:         config.LeaderElect,
		LeaderElectionID:       config.LeaderElectionID,
	}

	if watchNamespace != "" {
		options.Cache = cache.Options{
			DefaultNamespaces: map[string]cache.Config{
				watchNamespace: {},
			},
		}
		options.LeaderElectionNamespace = watchNamespace
	}

	return options
}
