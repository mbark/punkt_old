[![Build Status](https://travis-ci.org/mbark/punkt.svg?branch=master)](https://travis-ci.org/mbark/punkt) [![Go Report Card](https://goreportcard.com/badge/gojp/goreportcard)](https://goreportcard.com/report/gojp/goreportcard)

# punkt
A tool to manage your dotfiles and your environment.

Project goals:
- Easy to setup a simple initial dotfile repo;
- Easily scale the dotfiles as the config gets more complex;
- Back up what packages to install from which package manager;
- Easily bootstrap a new environment on a clean system;
- Reliable and robust, when something fails it should do so with clear error messages;
- Good looking CLI-app that you want to use.

## Current status and roadmap
A list of features that I want included and their status. 

- [ ] Create a base structure for the ansible config
- [ ] Config
  - [ ] Read a configuration file and use options as overrides
- [ ] Package managers
  - [ ] Add general support to add a package manager
    - [ ] `dump` environment to be installed
    - [ ] `ensure` environment is up to date with the given dump
    - [ ] `update` to latest version of packages for each manager
  - [ ] Support `homebrew` via `bundle` and `geerlingguy.homebrew`
    - [ ] Generate a brewfile when doing a `dump`
    - [ ] Run `geerlingguy.homebrew` (and manage sudo)
  - [ ] Support `apt-get`
- [ ] Symlinks
  - [ ] Create structure for symlinks
  - [ ] Search `~` and `~/.config` with a given depth for symlinks and allow the user to add these
  - [ ] Make it possible to directories and depth when searching for symlinks
  - [ ] Support `add`ing symlinks that are outside of the searched directories
- [ ] Tasks
  - [ ] Allow the user to configure their own tasks to be run
- [ ] devops
  - [ ] Set up a test suite via `pytest`
  - [ ] Run tests automatically via `travis`
  - [ ] Use `git hooks` to ensure that all tests pass
  - [ ] Write unit tests for the code where applicable
  - [ ] Provide a way of testing all supported package managers 
- [ ] Config
  - [ ] Store configuration for the app in a file
- [ ] UX/Beauty
  - [ ] Use [survey](https://github.com/AlecAivazis/survey) to manage interaction (such as selecting backends to update)
  - [ ] Print ansible output more beautifully via callback plugin
  - [ ] Print the `--help` section with a little touch of colors (and maybe an emoji) 
- [ ] Documentation
  - [ ] Provide documentation for all features inside the app (via the cli-library)
  - [ ] Provide documentation in the readme and/or guide
  - [ ] Provide an example gif and/or asciinema of the program in action
      - [ ] Provide guidelines for how to contribute 
