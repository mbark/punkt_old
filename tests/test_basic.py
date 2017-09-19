def test_has_help(goot):
    res = goot.run("--help")
    assert res.returncode == 0


def fails_with_non_existant_configuratino_file(goot):
    res = goot.run("nonexistant.file")
    assert res.returncode != 0
