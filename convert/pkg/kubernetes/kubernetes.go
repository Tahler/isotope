// Package kubernetes converts service graphs into Kubernetes manifests.
package kubernetes

import (
	"fmt"
	"strings"
	"time"

	"github.com/Tahler/isotope/convert/pkg/consts"
	"github.com/Tahler/isotope/convert/pkg/graph"
	"github.com/Tahler/isotope/convert/pkg/graph/svc"
	"github.com/ghodss/yaml"
	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// ServiceGraphNamespace is the namespace all service graph related resources
	// (i.e. ConfigMap, Services, and Deployments) will reside in.
	ServiceGraphNamespace = "service-graph"

	numConfigMaps          = 1
	numManifestsPerService = 2

	configVolume           = "config-volume"
	serviceGraphConfigName = "service-graph-config"
)

var (
	serviceGraphAppLabels       = map[string]string{"app": "service-graph"}
	serviceGraphNodeLabels      = map[string]string{"role": "service"}
	prometheusScrapeAnnotations = map[string]string{
		"prometheus.io/scrape": "true"}
)

// ServiceGraphToKubernetesManifests converts a ServiceGraph to Kubernetes
// manifests.
func ServiceGraphToKubernetesManifests(
	serviceGraph graph.ServiceGraph,
	serviceNodeSelector map[string]string,
	serviceImage string,
	serviceMaxIdleConnectionsPerHost int,
	clientNodeSelector map[string]string,
	clientImage string) (yamlDoc []byte, err error) {
	numServices := len(serviceGraph.Services)
	numManifests := numManifestsPerService*numServices + numConfigMaps
	manifests := make([]string, 0, numManifests)

	appendManifest := func(manifest interface{}) error {
		yamlDoc, innerErr := yaml.Marshal(manifest)
		if innerErr != nil {
			return innerErr
		}
		manifests = append(manifests, string(yamlDoc))
		return nil
	}

	namespace := makeServiceGraphNamespace()
	err = appendManifest(namespace)
	if err != nil {
		return
	}

	configMap, err := makeConfigMap(serviceGraph)
	if err != nil {
		return
	}
	err = appendManifest(configMap)
	if err != nil {
		return
	}

	for _, service := range serviceGraph.Services {
		k8sDeployment, innerErr := makeDeployment(
			service, serviceNodeSelector, serviceImage,
			serviceMaxIdleConnectionsPerHost)
		if innerErr != nil {
			return nil, innerErr
		}
		innerErr = appendManifest(k8sDeployment)
		if innerErr != nil {
			return nil, innerErr
		}

		k8sService, innerErr := makeService(service)
		if innerErr != nil {
			return nil, innerErr
		}
		innerErr = appendManifest(k8sService)
		if innerErr != nil {
			return nil, innerErr
		}
	}

	fortioDeployment := makeFortioDeployment(
		clientNodeSelector, clientImage)
	err = appendManifest(fortioDeployment)
	if err != nil {
		return
	}

	fortioService := makeFortioService()
	err = appendManifest(fortioService)
	if err != nil {
		return
	}

	yamlDocString := strings.Join(manifests, "---\n")
	return []byte(yamlDocString), nil
}

func combineLabels(a, b map[string]string) map[string]string {
	c := make(map[string]string, len(a)+len(b))
	for k, v := range a {
		c[k] = v
	}
	for k, v := range b {
		c[k] = v
	}
	return c
}

func makeServiceGraphNamespace() (namespace apiv1.Namespace) {
	namespace.APIVersion = "v1"
	namespace.Kind = "Namespace"
	namespace.ObjectMeta.Name = consts.ServiceGraphNamespace
	namespace.ObjectMeta.Labels = map[string]string{"istio-injection": "enabled"}
	timestamp(&namespace.ObjectMeta)
	return
}

func makeConfigMap(
	graph graph.ServiceGraph) (configMap apiv1.ConfigMap, err error) {
	graphYAMLBytes, err := yaml.Marshal(graph)
	if err != nil {
		return
	}
	configMap.APIVersion = "v1"
	configMap.Kind = "ConfigMap"
	configMap.ObjectMeta.Name = serviceGraphConfigName
	configMap.ObjectMeta.Namespace = ServiceGraphNamespace
	configMap.ObjectMeta.Labels = serviceGraphAppLabels
	timestamp(&configMap.ObjectMeta)
	configMap.Data = map[string]string{
		consts.ServiceGraphConfigMapKey: string(graphYAMLBytes),
	}
	return
}

func makeService(service svc.Service) (k8sService apiv1.Service, err error) {
	k8sService.APIVersion = "v1"
	k8sService.Kind = "Service"
	k8sService.ObjectMeta.Name = service.Name
	k8sService.ObjectMeta.Namespace = ServiceGraphNamespace
	k8sService.ObjectMeta.Labels = serviceGraphAppLabels
	k8sService.ObjectMeta.Annotations = prometheusScrapeAnnotations
	timestamp(&k8sService.ObjectMeta)
	k8sService.Spec.Ports = []apiv1.ServicePort{{Port: consts.ServicePort}, {Port: consts.ServiceGrpcPort}}
	k8sService.Spec.Selector = map[string]string{"name": service.Name}
	return
}

func makeDeployment(
	service svc.Service, nodeSelector map[string]string,
	serviceImage string, serviceMaxIdleConnectionsPerHost int) (
	k8sDeployment appsv1.Deployment, err error) {
	k8sDeployment.APIVersion = "apps/v1"
	k8sDeployment.Kind = "Deployment"
	k8sDeployment.ObjectMeta.Name = service.Name
	k8sDeployment.ObjectMeta.Namespace = ServiceGraphNamespace
	k8sDeployment.ObjectMeta.Labels = serviceGraphAppLabels
	k8sDeployment.ObjectMeta.Annotations = prometheusScrapeAnnotations
	timestamp(&k8sDeployment.ObjectMeta)
	k8sDeployment.Spec = appsv1.DeploymentSpec{
		Replicas: &service.NumReplicas,
		Selector: &metav1.LabelSelector{
			MatchLabels: map[string]string{
				"name": service.Name,
			},
		},
		Template: apiv1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Labels: combineLabels(
					serviceGraphNodeLabels,
					map[string]string{
						"name": service.Name,
					}),
			},
			Spec: apiv1.PodSpec{
				NodeSelector: nodeSelector,
				Containers: []apiv1.Container{
					apiv1.Container{
						Name:  consts.ServiceContainerName,
						Image: serviceImage,
						Args: []string{
							fmt.Sprintf(
								"--max-idle-connections-per-host=%v",
								serviceMaxIdleConnectionsPerHost),
						},
						Env: []apiv1.EnvVar{
							{Name: consts.ServiceNameEnvKey, Value: service.Name},
						},
						VolumeMounts: []apiv1.VolumeMount{
							{
								Name:      configVolume,
								MountPath: consts.ConfigPath,
							},
						},
						Ports: []apiv1.ContainerPort{
							{
								ContainerPort: consts.ServicePort,
							},
							{
								ContainerPort: consts.ServiceGrpcPort,
							},
						},
					},
				},
				Volumes: []apiv1.Volume{
					{
						Name: configVolume,
						VolumeSource: apiv1.VolumeSource{
							ConfigMap: &apiv1.ConfigMapVolumeSource{
								LocalObjectReference: apiv1.LocalObjectReference{
									Name: serviceGraphConfigName,
								},
								Items: []apiv1.KeyToPath{
									{
										Key:  consts.ServiceGraphConfigMapKey,
										Path: consts.ServiceGraphYAMLFileName,
									},
								},
							},
						},
					},
				},
			},
		},
	}
	timestamp(&k8sDeployment.Spec.Template.ObjectMeta)
	return
}

func timestamp(objectMeta *metav1.ObjectMeta) {
	objectMeta.CreationTimestamp = metav1.Time{Time: time.Now()}
}
