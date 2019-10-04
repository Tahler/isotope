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

	ServiceGraphConfigMapKey = "service-graph"

	// ServiceNameEnvKey is the key of the environment variable whose value is
	// the name of the service.
	ServiceNameEnvKey = "SERVICE_NAME"

	// FortioMetricsPort is the port on which /metrics is available.
	FortioMetricsPort = 42422
)
