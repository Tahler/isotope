package srv

const (
	// ServiceContainerName is the name to assign the container when it is run.
	ServiceContainerName = "mock-service"

	// ServicePort is the port for HTTP requests.
	ServiceHTTPPort = 8080

	// ServicePort is the port for GRPC requests
	ServiceGRPCPort = 8081

	// ServiceGraphNamespace is the name of the namespace that all service graph
	// related components will live in.
	ServiceGraphNamespace = "service-graph"

	// ConfigPath is the parent directory of all service configuration files.
	ConfigPath = "/etc/config"
	// ServiceGraphYAMLFileName is the name of the file which contains the
	// YAML-unmarshallable ServiceGraph.
	ServiceGraphYAMLFileName = "service-graph.yaml"
	// ServiceGraphConfigMapKey is the key of the Kubernetes config map entry
	// holding the ServiceGraph's YAML to be mounted in
	// "${ConfigPath}/${ServiceGraphYAMLFileName}".
	ServiceGraphConfigMapKey = "service-graph"

	// ServiceNameEnvKey is the key of the environment variable whose value is
	// the name of the service.
	ServiceNameEnvKey = "SERVICE_NAME"

	// FortioMetricsPort is the port on which /metrics is available.
	FortioMetricsPort = 42422
)
