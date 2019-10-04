package consts

const (
	// ServiceContainerName is the name to assign the container when it is run.
	ServiceContainerName = "perf-test-service"
	// ServiceImageName is the name / path of the image to pull.
	ServiceImageName = "tahler/perf-test-service"

	// ServicePort is the port the service will run on.
	ServicePort = 8080

	// ServiceGraphNamespace is the name of the namespace that all service graph
	// related components will live in.
	ServiceGraphNamespace = "service-graph"

	// ServiceGraphConfigMapKey is the key of the Kubernetes config map entry
	// holding the ServiceGraph's YAML to be mounted in
	// "${ConfigPath}/${ServiceGraphYAMLFileName}".
	ServiceGraphConfigMapKey = "service-graph"
	// LabelsYAMLFileName is the name of the file which contains the YAML
	// representing the key-value pairs for all Prometheus metrics.
	LabelsYAMLFileName = "labels.yaml"
	// LabelsConfigMapKey is the key of the Kubernetes config map entry
	// holding the labels's YAML to be mounted in
	// "${ConfigPath}/${LabelsYAMLFileName}".
	LabelsConfigMapKey = "labels"

	// ServiceNameEnvKey is the key of the environment variable whose value is
	// the name of the service.
	ServiceNameEnvKey = "SERVICE_NAME"
)
