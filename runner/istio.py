import contextlib
import logging
import os
import tempfile
from typing import Dict, Generator

import yaml

from . import consts, sh

_HELM_ISTIO_NAME = 'istio'


@contextlib.contextmanager
def latest(hub: str, tag: str) -> Generator[None, None, None]:
    _install_latest(hub, tag)
    yield
    sh.run_helm(['delete', '--purge', _HELM_ISTIO_NAME])


def _install_latest(hub: str, tag: str) -> None:
    """Installs Istio from master, using hub:tag for the images.

    Requires Helm to be present.

    This clones the repo in a temporary directory, builds and pushes the
    images, then runs `helm install`.
    """
    with tempfile.TemporaryDirectory() as tmp_go_path:
        repo_path = os.path.join(tmp_go_path, 'src', 'istio.io', 'istio')
        _clone(repo_path)
        _push_images(tmp_go_path, repo_path, hub, tag)

        chart_path = os.path.join(repo_path, 'install', 'kubernetes', 'helm',
                                  'istio')
        values_path = os.path.join(chart_path, 'values-isotope.yaml')
        _gen_helm_values(values_path, hub, tag)

        logging.info('installing Helm chart for Istio')
        _install_helm_chart(chart_path, values_path, _HELM_ISTIO_NAME,
                            consts.ISTIO_NAMESPACE)


def _clone(path: str) -> None:
    """Clones github.com/istio.io/istio to path."""
    logging.info('cloning istio.io/istio to %s', path)
    sh.run(
        ['git', 'clone', 'https://github.com/istio/istio.git', path],
        check=True)


def _push_images(go_path: str, repo_path: str, hub: str, tag: str) -> None:
    logging.info('pushing images to %s with tag %s', hub, tag)
    with _work_dir(repo_path):
        env = _set_env({
            'GOPATH': go_path,
            'HUB': hub,
            'TAG': tag,
        })
        sh.run(['make', 'docker.push'], env=env, check=True)


def _gen_helm_values(path: str, hub: str, tag: str) -> str:
    parent_dir = os.path.dirname(path)
    if not os.path.exists(parent_dir):
        os.makedirs(parent_dir)

    with open(path, 'w') as f:
        return yaml.dump({'global': {'hub': hub, 'tag': tag}}, f)


def _install_helm_chart(chart_path: str,
                        values_path: str,
                        name: str,
                        namespace: str = consts.DEFAULT_NAMESPACE) -> None:
    sh.run_helm(
        [
            'install', chart_path, '--values', values_path, '--name', name,
            '--namespace', namespace
        ],
        check=True)


def _set_env(overrides: Dict[str, str]) -> Dict[str, str]:
    copy = os.environ.copy()
    copy.update(overrides)
    return copy


@contextlib.contextmanager
def _work_dir(path: str) -> Generator[None, None, None]:
    prev_path = os.getcwd()
    if not os.path.exists(path):
        os.makedirs(path)
    os.chdir(path)
    yield
    os.chdir(prev_path)