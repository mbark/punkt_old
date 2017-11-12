from subprocess import PIPE, run

import common
import pytest
import yaml


def test_parses_valid_backend_file(tmpdir, goot):
    d = tmpdir.mkdir("parses")

    conf = {
        'symlinks': {},
        'backends': {
            'apt': 'backend/apt.yaml'
        },
        'tasks': []
    }
    conf_file = common.create_conf_file(d, conf)

    apt_conf = {
        'list': "apt list --installed | cut -d/ -f1",
        'update': 'apt upgrade',
        'install': 'apt install'
    }
    common.create_conf_file(d.mkdir('backend'), apt_conf, 'apt')
    res = goot.run(conf_file)
    assert res.returncode == 0


@pytest.mark.docker
def test_creates_database_file(tmpdir, goot):
    d = tmpdir.mkdir("bootstrap")

    conf = {
        'symlinks': {},
        'backends': {
            'apt': 'backend/rustp.yaml'
        },
        'tasks': [],
        'package_files': 'packages'
    }
    conf_file = common.create_conf_file(d, conf)

    apt_cmd = 'apt list --installed | cut -d/ -f1'
    apt_conf = {
        'list': apt_cmd,
        'update': 'apt upgrade',
        'install': 'apt install'
    }
    common.create_conf_file(d.mkdir('backend'), apt_conf, 'apt')

    res = run(apt_cmd, stdout=PIPE, shell=True)
    assert res.returncode is 0

    res = goot.run(conf_file, ['--verify'])
    assert res.returncode is 0

    assert d.join('packages').check()
    assert d.join('packages').join('apt.yaml').check()

    contents = yaml.load(d.join('packages').join('apt.yaml').read())

    packages = res.stdout.decode('utf-8').splitlines()
    assert packages.sort() == contents.sort()
