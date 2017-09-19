from subprocess import run

import pytest
import yaml


@pytest.mark.xfail
@pytest.mark.docker
def test_install_from_package_manager(config_file):
    conf = {'symlinks': {}, 'backends': {'rustup': 'rustup.yaml'}, 'tasks': {}}

    rustup_conf = {
        'bootstrap': 'curl https://sh.rustup.rs -sSf | sh',
        'list': "rustup show | tail -n 2 | awk '{print $1}'",
        'update': 'rustup update',
        'install': 'rustup install'
    }
    (d, conf_file) = config_file(conf)

    rustup_file = d.join("rustup.yaml")
    rustup_file.write(yaml.dump(rustup_conf))

    res = run(["cargo", "run", "--", str(conf_file)])
    assert res.returncode == 0
