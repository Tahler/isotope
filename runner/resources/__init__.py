import os

_RESOURCES_DIR = os.path.realpath(
    os.path.join(os.getcwd(), os.path.dirname(__file__)))

HELM_SERVICE_ACCOUNT_YAML_PATH = os.path.join(_RESOURCES_DIR,
                                              'helm-service-account.yaml')
PROMETHEUS_STORAGE_VALUES_YAML_PATH = os.path.join(
    _RESOURCES_DIR, 'values-prometheus-storage.yaml')
PERSISTENT_VOLUME_YAML_PATH = os.path.join(_RESOURCES_DIR,
                                           'persistent-volume.yaml')

PROMETHEUS_VALUES_GEN_YAML_PATH = os.path.join(_RESOURCES_DIR,
                                               'values-prometheus.gen.yaml')
SERVICE_GRAPH_GEN_YAML_PATH = os.path.join(_RESOURCES_DIR,
                                           'service-graph.gen.yaml')
ISTIO_GEN_YAML_PATH = os.path.join(_RESOURCES_DIR, 'istio.gen.yaml')
ISTIO_INGRESS_YAML_PATH = os.path.join(_RESOURCES_DIR,
                                       'istio-ingress.gen.yaml')
