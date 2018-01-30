[![Build Status](https://travis-ci.org/mbark/punkt.svg?branch=master)](https://travis-ci.org/mbark/punkt) [![Go Report Card](https://goreportcard.com/badge/mbark/punkt)](https://goreportcard.com/report/mbark/punkt) [![Coverage Status](https://coveralls.io/repos/github/mbark/punkt/badge.svg?branch=master)](https://coveralls.io/github/mbark/punkt?branch=master)

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

- [x] symlinks
  - [x] dump
  - [x] ensure
  - [x] use `survey` to select which directories to add
  - [x] configure directories to ignore
  - [x] add a file as a symlink
- [x] brew
  - [x] dump
  - [x] ensure
  - [x] update
- [x] yarn
  - [x] dump
  - [x] ensure
  - [x] update
- [x] git
  - [x] dump
  - [x] ensure
  - [x] update
- [ ] tests
  - [ ] cli
    - [x] sanity checks (e.g. all commands have --help flag)
    - [ ] add
    - [ ] dump
    - [ ] ensure
    - [ ] update
  - [ ] unit tests
