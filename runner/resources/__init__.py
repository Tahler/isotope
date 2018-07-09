import contextlib
import logging
import os
from typing import Generator, Optional, Type

from .. import consts, context, sh, wait

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


@contextlib.contextmanager
def manifest(path: str) -> Generator[None, None, None]:
    """Runs `kubectl create -f path` on entry and opposing delete on exit."""
    _create_from_manifest(path)
    with context.confirm_clean_up_on_exception():
        yield
    _delete_from_manifest(path)


def _create_from_manifest(path: str) -> None:
    logging.info('creating from %s', path)
    sh.run_kubectl(['create', '-f', path], check=True)


def _delete_from_manifest(path: str) -> None:
    logging.info('deleting from %s', path)
    sh.run_kubectl(['delete', '-f', path])
