import common
import yaml
import os

script_path = os.path.dirname(os.path.realpath(__file__))


def test_parses_valid_backend_file(tmpdir, punkt):
    d = tmpdir.mkdir("parses")

    conf = {
        'symlinks': {},
        'backends': {
            'apt': {
                'list': "apt list --installed | cut -d/ -f1",
                'update': 'apt upgrade',
                'install': 'apt install'
            }
        },
        'tasks': [],
        'pkgdbs': 'packages'
    }
    conf_file = common.create_conf_file(d, conf)

    res = punkt.run(conf_file)
    assert res.returncode == 0


def test_creates_database_file(tmpdir, punkt):
    d = tmpdir.mkdir("bootstrap")
    cmd = '%s/fake_backend.sh' % script_path

    conf = {
        'symlinks': {},
        'backends': {
            'fake': {
                'list': '%s list' % cmd,
                'update': '%s upgrade' % cmd,
                'install': '%s install' % cmd
            },
        },
        'tasks': [],
        'pkgdbs': 'packages'
    }

    conf_file = common.create_conf_file(d, conf)

    res = punkt.run(conf_file, ['ensure'])
    assert res.returncode is 0

    assert d.join('packages').check()
    assert d.join('packages').join('fake.yaml').check()

    contents = yaml.load(d.join('packages').join('fake.yaml').read())

    packages = ['package1', 'package2', 'package3', 'package4']
    assert packages.sort() == contents.sort()
