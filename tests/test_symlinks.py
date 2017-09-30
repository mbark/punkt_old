import os
import os.path
import yaml


def create_conf_file(d, conf):
    conf_file = d.join("config.yaml")
    conf_file.write(yaml.dump(conf))
    return conf_file


def run_valid(d, conf, goot):
    conf_file = create_conf_file(d, conf)

    res = goot.run(conf_file)
    assert res.returncode == 0

    for key in conf['symlinks']:
        assert os.path.islink(str(d.join(key)))


def test_simple(tmpdir, goot):
    conf = {
        'symlinks': {
            'b.txt': 'a.txt',
            'c.txt': 'a.txt'
        },
        'backends': {},
        'tasks': []
    }

    d = tmpdir.mkdir("simple")
    f = d.join("a.txt")
    f.write("foo")
    run_valid(d, conf, goot)


def test_creates_necessary_directories(tmpdir, goot):
    conf = {
        'symlinks': {
            'dir/a.txt': 'a.txt',
            'a/n/o/ther/dir/e/cto/ry/b.txt': 'a.txt',
            'a/n/o/ther/dir/e/cto/c.txt': 'a.txt',
        },
        'backends': {},
        'tasks': []
    }

    d = tmpdir.mkdir("directories")
    f = d.join("a.txt")
    f.write("foo")

    run_valid(d, conf, goot)


def test_fails_if_file_already_exists(tmpdir, goot):
    conf = {
        'symlinks': {
            'b.txt': 'a.txt'
        },
        'backends': {},
        'tasks': []
    }

    d = tmpdir.mkdir("non_existant")
    f = d.join("b.txt")
    f.write("foo")

    conf_file = create_conf_file(d, conf)
    res = goot.run(conf_file)
    assert res.returncode != 0


def test_does_nothing_when_dry_running(tmpdir, goot):
    conf = {
        'symlinks': {
            'b.txt': 'a.txt',
            'dir/a.txt': 'a.txt',
        },
        'backends': {},
        'tasks': []
    }

    d = tmpdir.mkdir("directories")
    f = d.join("a.txt")
    f.write("foo")
    conf_file = create_conf_file(d, conf)

    snapshot = os.listdir(str(d))

    res = goot.run(conf_file, ['-n'])
    assert res.returncode == 0

    assert snapshot == os.listdir(str(d))
