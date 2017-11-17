import os
from os.path import islink, samefile

import common


def run_valid(d, conf, punkt):
    conf_file = common.create_conf_file(d, conf)

    res = punkt.run(conf_file, ['ensure'])
    assert res.returncode == 0

    for obj in conf['symlinks']:
        to = str(d.join(obj.get('to')))
        fromlink = str(d.join(obj.get('from')))

        # assert islink(to)
        assert samefile(fromlink, to)


def test_simple(tmpdir, punkt):
    conf = {
        'symlinks': [{
            'from': 'a.txt',
            'to': 'b.txt',
        }, {
            'from': 'a.txt',
            'to': 'c.txt'
        }],
        'backends': {},
        'tasks': [],
        'pkgdbs': 'packages'
    }

    d = tmpdir.mkdir('simple')
    f = d.join('a.txt')
    f.write('foo')
    run_valid(d, conf, punkt)


def test_creates_necessary_directories(tmpdir, punkt):
    conf = {
        'symlinks': [{
            'to': 'dir/a.txt',
            'from': 'a.txt',
        }, {
            'to': 'a/n/o/ther/dir/e/cto/ry/b.txt',
            'from': 'a.txt',
        }, {
            'to': 'a/n/o/ther/dir/ec/to/c.txt',
            'from': 'a.txt',
        }],
        'backends': {},
        'tasks': [],
        'pkgdbs': 'packages'
    }

    d = tmpdir.mkdir('directories')
    f = d.join('a.txt')
    f.write('foo')

    run_valid(d, conf, punkt)


def test_does_not_fail_if_symlink_exists(tmpdir, punkt):
    d = tmpdir.mkdir('foo')
    src = d.join('a.txt')
    src.write('foo')
    target = d.mkdir('dir').join('a.txt')

    conf = {
        'symlinks': [{
            'to': 'dir/b.txt',
            'from': 'a.txt',
        }],
        'backends': {},
        'tasks': [],
        'pkgdbs': 'packages'
    }

    os.symlink(str(src), str(target))
    run_valid(d, conf, punkt)


def test_fails_if_file_already_exists(tmpdir, punkt):
    conf = {
        'symlinks': {
            'b.txt': 'a.txt'
        },
        'backends': {},
        'tasks': [],
        'pkgdbs': 'packages'
    }

    d = tmpdir.mkdir('non_existant')
    f = d.join('b.txt')
    f.write('foo')

    conf_file = common.create_conf_file(d, conf)
    res = punkt.run(conf_file, ['ensure'])
    assert res.returncode != 0


def test_does_nothing_when_dry_running(tmpdir, punkt):
    conf = {
        'symlinks': [{
            'to': 'b.txt',
            'from': 'a.txt',
        }, {
            'to': 'dir/a.txt',
            'from': 'a.txt',
        }],
        'backends': {},
        'tasks': [],
        'pkgdbs': 'packages'
    }

    d = tmpdir.mkdir('directories')
    f = d.join('a.txt')
    f.write('foo')
    conf_file = common.create_conf_file(d, conf)

    snapshot = os.listdir(str(d))

    res = punkt.run(conf_file, ['-n'])
    assert res.returncode == 0

    assert snapshot == os.listdir(str(d))
