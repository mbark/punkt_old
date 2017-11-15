def test_has_help(punkt):
    res = punkt.run('--help')
    assert res.returncode == 0

    res = punkt.run('help')
    assert res.returncode == 0


def test_prints_help_without_command(punkt):
    res_with_help = punkt.run('--help')
    res_with_none = punkt.run()

    assert res_with_help.stdout == res_with_none.stdout


def fails_with_non_existant_configuratino_file(punkt):
    res = punkt.run(config='nonexistant.file')
    assert res.returncode != 0
