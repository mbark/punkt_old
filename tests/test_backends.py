import common
import yaml
import os

script_path = os.path.dirname(os.path.realpath(__file__))


def test_parses_valid_backend_file(tmpdir, punkt):
    d = tmpdir.mkdir("parses")

    conf = {
        'symlinks': {},
        'backends': {
            'apt': 'backend/apt.yaml'
        },
        'tasks': [],
        'package_files': 'packages'
    }
    conf_file = common.create_conf_file(d, conf)

    apt_conf = {
        'list': "apt list --installed | cut -d/ -f1",
        'update': 'apt upgrade',
        'install': 'apt install'
    }
    common.create_conf_file(d.mkdir('backend'), apt_conf, 'apt')
    res = punkt.run(conf_file, ['ensure'])
    assert res.returncode == 0


def test_creates_database_file(tmpdir, punkt):
    d = tmpdir.mkdir("bootstrap")

    conf = {
        'symlinks': {},
        'backends': {
            'fake': 'backend/fake.yaml'
        },
        'tasks': [],
        'package_files': 'packages'
    }
    conf_file = common.create_conf_file(d, conf)

    cmd = '%s/fake_backend.sh' % script_path

    backend_conf = {
        'list': '%s list' % cmd,
        'update': '%s upgrade' % cmd,
        'install': '%s install' % cmd
    }
    common.create_conf_file(d.mkdir('backend'), backend_conf, 'fake')

    res = punkt.run(conf_file, ['ensure'])
    assert res.returncode is 0

    assert d.join('packages').check()
    assert d.join('packages').join('fake.yaml').check()

    contents = yaml.load(d.join('packages').join('fake.yaml').read())

    packages = ['package1', 'package2', 'package3', 'package4']
    assert packages.sort() == contents.sort()
