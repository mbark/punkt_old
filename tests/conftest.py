import pytest
from subprocess import run


def pytest_addoption(parser):
    parser.addoption(
        "--docker",
        action="store_true",
        default=False,
        help="Run tests that are made to run from inside Docker")


def pytest_collection_modifyitems(config, items):
    if config.getoption("--docker"):
        return
    skip_docker = pytest.mark.skip(reason="need --docker option to run")
    for item in items:
        if "docker" in item.keywords:
            item.add_marker(skip_docker)


@pytest.fixture
def goot():
    g = Goot()
    g.build()
    return g


class Goot:
    def __init__(self):
        pass

    def build(self):
        res = run(["go", "build"], cwd="..")
        print(res)
        assert res.returncode == 0

    def run(self, conf_file):
        return run(["./goot", str(conf_file)], cwd="..")
