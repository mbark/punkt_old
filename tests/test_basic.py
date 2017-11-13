def test_has_help(punkt):
    res = punkt.run("--help")
    assert res.returncode == 0

    res = punkt.run("help")
    assert res.returncode == 0


def fails_with_non_existant_configuratino_file(punkt):
    res = punkt.run("nonexistant.file")
    assert res.returncode != 0
