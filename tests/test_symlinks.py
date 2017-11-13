import os
from os.path import islink

import common


def run_valid(d, conf, punkt):
    conf_file = common.create_conf_file(d, conf)

    res = punkt.run(conf_file, ['ensure'])
    assert res.returncode == 0

    for key in conf['symlinks']:
        assert islink(str(d.join(key)))


def test_simple(tmpdir, punkt):
    conf = {
        'symlinks': {
            'b.txt': 'a.txt',
            'c.txt': 'a.txt'
        },
        'backends': {},
        'tasks': [],
        'package_files': 'packages'
    }

    d = tmpdir.mkdir("simple")
    f = d.join("a.txt")
    f.write("foo")
    run_valid(d, conf, punkt)


def test_creates_necessary_directories(tmpdir, punkt):
    conf = {
        'symlinks': {
            'dir/a.txt': 'a.txt',
            'a/n/o/ther/dir/e/cto/ry/b.txt': 'a.txt',
            'a/n/o/ther/dir/e/cto/c.txt': 'a.txt',
        },
        'backends': {},
        'tasks': [],
        'package_files': 'packages'
    }

    d = tmpdir.mkdir("directories")
    f = d.join("a.txt")
    f.write("foo")

    run_valid(d, conf, punkt)


def test_fails_if_file_already_exists(tmpdir, punkt):
    conf = {
        'symlinks': {
            'b.txt': 'a.txt'
        },
        'backends': {},
        'tasks': [],
        'package_files': 'packages'
    }

    d = tmpdir.mkdir("non_existant")
    f = d.join("b.txt")
    f.write("foo")

    conf_file = common.create_conf_file(d, conf)
    res = punkt.run(conf_file, ['ensure'])
    assert res.returncode != 0


def test_does_nothing_when_dry_running(tmpdir, punkt):
    conf = {
        'symlinks': {
            'b.txt': 'a.txt',
            'dir/a.txt': 'a.txt',
        },
        'backends': {},
        'tasks': [],
        'package_files': 'packages'
    }

    d = tmpdir.mkdir("directories")
    f = d.join("a.txt")
    f.write("foo")
    conf_file = common.create_conf_file(d, conf)

    snapshot = os.listdir(str(d))

    res = punkt.run(conf_file, ['-n'])
    assert res.returncode == 0

    assert snapshot == os.listdir(str(d))
