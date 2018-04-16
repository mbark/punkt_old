# manager

`punkt` uses the concept of "managers" to provide an extended functionality and
better dotfile-management than just symlinking. A manager is anything which
implements the following api:

* `dump(): toml`
* `ensure(config)`
* `update(config)`

`punkt` has two fundamental managers: `git` and `symlink`, which can be used by
other manager's to ensure git repositories or symlinks exist. More about them
below.

## Configuration format

Managers are configured with `toml` files. There are two special keys:
`symlinks` and `repositories` which correspond to the `symlink` and `git`
managers respectively, allowing them to be used to create f.e. a symlink.

### Why toml?

`toml` was chosen for the following reasons:

* well-specified;
* readable and easily understood;
* good support for parsing it across languages;
* easy to read and write even in simpler languages like bash.

## API

The api for a manager. All commands are ran with `std{err,in,out}` being piped
to the user allowing interactive use if so desired.

### dump

`dump` should **only** build a configuration file. Running `dump` should never
have any side effects on the system.

* _input_: nothing
* _output_: toml with the manager's configuration

Example:

```console
$ mgr.sh dump
symlinks = [ "Users/user/.Brewfile" ]
[foo]
this = "is other toml stuff"
```

### ensure

`ensure` can only be ran after building a configuration file given by `dump`. It
should **never** modify the given configuration file, only read from it and
modify the external system.

* _input_: path to the config-file
* _output_: nothing

Example:

```console
$ mgr.sh ensure ~/.dotfiles/mgr.toml
```

### update

`update` can only be ran after building a configuration file with `dump`. It is
allowed to both modify the system and update the configuration file if necessary.

* _input_: path to the config-file
* _output_: nothing

Example:

```console
$ mgr.sh update ~/.dotfiles/mgr.toml
```

## Implement your own manager

A manager is specified in punkt's configuration file:
`$XDG_CONFIG_HOME/punkt/punkt.toml` under the header `managers`. You can specify
a command to run with the `command` key, you can also specify each of the
individual commands with `dump`, `ensure`, `update`. `command` is obligatory
unless all commands are deliberately specified. If a command is specified it
will be preferred over just proxying to `command`.

In the example below three managers are provided and showing the three different
ways to specify a manager.

```toml
[managers]
    [managers.brew]
    command = "brew.sh"

    [managers.yarn]
    dump = ""
    ensure = "cd $XDG_CONFIG_HOME/.config/yarn/global && yarn install"
    update = "yarn global upgrade"

    [managers.foo]
    command = "foobar"
    ensure = "barfoo ensure"
```

Names for managers must be unique and each will be given a corresponding
configuration file at `$XDG_CONFIG_HOME/punkt/$NAME.toml`.

## git

The `git` manager handles `git repositories` and is used to ensure the existence
of the `dotfiles` repository as well as any additional repositories. The manager
can be used by others if they want to ensure the existence of some repository.

### `repositories.toml` format

The `repositories.toml` file contains a list of repositories where a repository consists
of a name and a repository configuration, (the repository configuration being
`go-git`~s marshalled repository object).

Example:

```toml
[[repository]]
name = "punkt"
path = "~/.dotfiles"
    [repository.config]
    # go-git config here
```

## symlink

The `symlink` manager handles symlinks and is primarily used for two cases:

1. Symlinking your non-managed config-files such as `.Xresources`;
2. Providing a way for other managers to specify what configuration files they
   have that they want symlinked to the dotfiles directory.

Due to the fact that it doesn~t make sense to `dump` or `update` symlinks it is
effectively a `noop` to call these for the manager. However, it has an extended
API and provides:

* `add` to move a file to the dotfiles directory, create a symlink from the
  original place to the new placement for the file and store it in the
  `symlinks.toml` file;
* `remove` to remove a file that was added by removing the symlink, moving the
  file back and removing it from the `symlinks.toml` file.

### `symlinks.toml` format

The `symlinks.toml` file maps a file in the dotfiles directory to a list of
placements. Symlinks are specified as a list of tables (or map if you want to)
where each table contains a `file` entry specifying the file and optionally
a `link`.

If `link` is not given the file will be placed with the same relative path to
home as it has to the dotfiles directory, i.e. `~/.dotfiles/path/.file` will be
placed at `~/path/.file`. Otherwise the link will be created at the path of
`link`.

Example:

```toml
# placement is inferred from where the file is located in the repository
symlinks = [
    { file = "a.txt" },
    { file = "a.txt", link = "~/specific/placement.txt" },
]
```
