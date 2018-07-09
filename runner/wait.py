import collections
import datetime
import logging
import subprocess
import time
from typing import Callable

from . import consts, sh

PROMETHEUS_SCRAPE_INTERVAL = datetime.timedelta(seconds=30)
RETRY_INTERVAL = datetime.timedelta(seconds=5)


def until(predicate: Callable[[], bool]) -> None:
    while not predicate():
        time.sleep(RETRY_INTERVAL.seconds)


def _until_rollouts_complete(resource_type: str, namespace: str) -> None:
    proc = sh.run_kubectl(
        [
            '--namespace', namespace, 'get', resource_type, '-o',
            'jsonpath={.items[*].metadata.name}'
        ],
        check=True)
    resources = collections.deque(proc.stdout.split(' '))
    logging.info('waiting for %ss in %s (%s) to rollout', resource_type,
                 namespace, ', '.join(resources))
    while len(resources) > 0:
        resource = resources.popleft()
        try:
            # kubectl blocks until ready.
            sh.run_kubectl(
                [
                    '--namespace', namespace, 'rollout', 'status',
                    resource_type, resource
                ],
                check=True)
        except subprocess.CalledProcessError as e:
            msg = 'failed to check rollout status of {}'.format(resource)
            if 'watch closed' in e.stderr:
                logging.debug('%s; retrying later', msg)
                resources.append(resource)
            else:
                logging.error(msg)


def until_deployments_are_ready(
        namespace: str = consts.DEFAULT_NAMESPACE) -> None:
    _until_rollouts_complete('deployment', namespace)


def until_stateful_sets_are_ready(
        namespace: str = consts.DEFAULT_NAMESPACE) -> None:
    _until_rollouts_complete('statefulsets', namespace)


def until_prometheus_has_scraped() -> None:
    logging.info('allowing Prometheus time to scrape final metrics')
    time.sleep(PROMETHEUS_SCRAPE_INTERVAL.seconds)


def until_namespace_is_deleted(
        namespace: str = consts.DEFAULT_NAMESPACE) -> None:
    until(lambda: _namespace_is_deleted(namespace))


def _namespace_is_deleted(namespace: str = consts.DEFAULT_NAMESPACE) -> bool:
    proc = sh.run_kubectl(['get', 'namespace', namespace])
    return proc.returncode != 0


def until_service_graph_is_ready() -> None:
    until(_service_graph_is_ready)


def _service_graph_is_ready() -> bool:
    proc = sh.run_kubectl(
        [
            '--namespace', consts.SERVICE_GRAPH_NAMESPACE, 'get', 'pods',
            '--selector', consts.SERVICE_GRAPH_SERVICE_SELECTOR, '-o',
            'jsonpath={.items[*].status.conditions[?(@.type=="Ready")].status}'
        ],
        check=True)
    out = proc.stdout
    all_services_ready = out != '' and 'False' not in out
    return all_services_ready
