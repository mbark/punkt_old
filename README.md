# goot
A tool to manage your dotfiles and your environment.

Project goals:
- Manage dotfiles through simple yet powerful-if-necessary configuration files;
- Provide functionality to easily keep what packages you have installed with package managers up to date and checked in;
- Easily bootstrap a clean system through a standalone binary;
- Reliable and robust, when something fails it should do so with clear error messages;
- Good looking CLI-app that you want to use.

## Current status and roadmap
A list of features that I want included and their status.

- [x] Parse a config file written in yaml
- [ ] UX/Beauty
  - [ ] Use [survey](https://github.com/AlecAivazis/survey) to manage interaction (such as selecting backends to update)
  - [ ] Display current progress in a nice way
  - [ ] Print errors that occur nicely
  - [ ] Print the `--help` section with a little touch of colors (and maybe an emoji)
- [x] Symlinks
  - [x] Allow symlinks to be specified and created
  - [x] Create directories necessary to allow the symlinks to be created
- [ ] Backends
  - [x] Use the `list` command to create a list of installed packages for a backend
  - [ ] Allow all packages for a given backend to be updated
  - [ ] Allow an individual package for a given backend to be updated
  - [ ] Allow installing all non-installed packages for a given backend
- [ ] Tasks
  - [ ] Run through all shell-tasks
  - [ ] Handle backend-tasks by installing all packages specified in the package file
- [ ] Config
  - [ ] Store configuration for the app in a file
  - [ ] Allow settings a package as ignored and undoing the operation
- [ ] devops
  - [x] Set up a test suite via `pytest`
  - [ ] Run tests automatically via travis
  - [ ] Provide a way of running the test suite under different conditions via `Docker`
  - [ ] Use git hooks to ensure that all tests pass
  - [ ] Write unit tests for the code where applicable
- [ ] Documentation
  - [ ] Provide documentation for all features inside the app (via the cli-library)
  - [ ] Provide documentation in the readme and/or guide
  - [ ] Provide an example gif and/or asciinema of the program in action
  - [ ] Provide guidelines for how to contribute
