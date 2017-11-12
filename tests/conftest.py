import pytest
from subprocess import run


def pytest_addoption(parser):
    parser.addoption(
        '--docker',
        action='store_true',
        default=False,
        help='Run tests that are made to run from inside Docker')


def pytest_collection_modifyitems(config, items):
    if config.getoption('--docker'):
        return
    skip_docker = pytest.mark.skip(reason='need --docker option to run')
    for item in items:
        if 'docker' in item.keywords:
            item.add_marker(skip_docker)


@pytest.fixture(scope='session')
def goot():
    g = Goot()
    return g


class Goot:
    def __init__(self):
        self.build()

    def build(self):
        return run(['go', 'build'], cwd='..')

    def run(self, conf_file, flags=None):
        if not flags:
            flags = []

        return run(['./goot', str(conf_file)] + flags, cwd='..')
